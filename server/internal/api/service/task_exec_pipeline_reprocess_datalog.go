package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/rimdian/rimdian/internal/api/entity"
	common "github.com/rimdian/rimdian/internal/common/dto"
	"github.com/rotisserie/eris"
	"go.opencensus.io/trace"
)

// requeue a batch of data imports that have not been processed after a certain delay
func TaskExecDataLogReprocessUntil(ctx context.Context, pipe *TaskExecPipeline) (result *entity.TaskExecResult) {

	spanCtx, span := trace.StartSpan(ctx, "TaskExecDataLogReprocessUntil")
	defer span.End()

	result = &entity.TaskExecResult{
		// keep current state by default
		UpdatedWorkerState: pipe.TaskExec.State.Workers[pipe.TaskExecPayload.WorkerID],
	}
	select {
	case <-spanCtx.Done():
		result.SetError("task execution timeout", false)
		return
	default:
	}

	// log time taken
	startedAt := time.Now()
	defer func() {
		log.Printf("TaskDataLogReprocessUntil: workspace %s, task %s, worker %d, took %s", pipe.Workspace.ID, pipe.TaskExec.ID, pipe.TaskExecPayload.WorkerID, time.Since(startedAt))
	}()

	bgCtx := context.Background()

	// get the state of the main thread
	mainState := pipe.TaskExec.State.Workers[0]

	// get the until date
	untilDateString, ok := mainState["until_date"].(string)

	if !ok {
		result.SetError(fmt.Sprintf("until_date not found in state, got %+v", mainState), true)
		return
	}

	untilDate, err := time.Parse(time.RFC3339, untilDateString)

	if err != nil {
		result.SetError(fmt.Sprintf("parse until_date: %s", err.Error()), false)
		return
	}

	// get the number of data imports that have been reprocessed
	reprocessed, ok := mainState["reprocessed"].(float64)

	if !ok {
		result.SetError(fmt.Sprintf("reprocessed not found in state, got %+v", mainState), true)
		return
	}

	// id of the last data import that was requeued
	lastID, ok := mainState["last_id"].(string)

	if !ok {
		result.SetError(fmt.Sprintf("last_id not found in state, got %+v", mainState), true)
		return
	}

	lastIDEventAt, ok := mainState["last_id_event_at"].(string)

	if !ok {
		result.SetError(fmt.Sprintf("last_id_event_at not found in state, got %+v", mainState), true)
		return
	}

	// use the until date for pagination by default
	paginationDate := untilDate

	// use the last ID received at date for pagination if it is available
	if lastIDEventAt != "" {
		lastIDEventAtDate, err := time.Parse(time.RFC3339, lastIDEventAt)

		if err != nil {
			result.SetError(fmt.Sprintf("parse last_id_event_at: %s", err.Error()), true)
			return
		}

		paginationDate = lastIDEventAtDate
	}

	limit := 300 // number of data imports to requeue at a time

	// get the data imports that need to be reprocessed
	dataLogs, err := pipe.Repo().ListDataLogsToReprocess(bgCtx, pipe.Workspace.ID, lastID, paginationDate, limit)

	if err != nil {
		result.SetError(fmt.Sprintf("list data logs to reprocess: %s", err.Error()), false)
		return
	}

	// log.Printf("found %v data imports", len(dataLogs))

	// process data imports in parallel with a wait group
	var wg sync.WaitGroup

	// buffered channel to receive errors from goroutines
	// buffered channel are non-blocking, we can write into it and read values later
	// as long as we don't write more values than its length
	errChan := make(chan error, len(dataLogs))

	// process data imports in parallel, as replay can only be done one by one
	for i := range dataLogs {

		wg.Add(1)

		dataLogRef := dataLogs[i]

		go func(dataLog entity.DataLog) {
			defer wg.Done()

			jsonItems := []string{dataLog.Item}

			pipe.DataLogEnqueue(spanCtx, &dataLog.ID, common.DataLogOriginInternalTaskExec, pipe.TaskExec.ID, pipe.Workspace.ID, jsonItems, false)

			if pipe.HasError() {
				errChan <- eris.Errorf("error while enqueueing data logs: %v", pipe.QueueResult.Error)
				return
			}

			log.Printf("requeued data_log %s, event_at %v", dataLogRef.ID, dataLogRef.EventAt)

		}(*dataLogRef)

		// increment the number of data imports that have been reprocessed
		reprocessed++

		// update the last ID
		lastID = dataLogRef.ID
		lastIDEventAt = dataLogRef.EventAt.Format(time.RFC3339)
	}

	// wait for all goroutines to finish
	wg.Wait()

	// close the channel, data written into it are still available
	close(errChan)

	// check for errors
	for err := range errChan {
		// retry the whole batch on error
		if err != nil {
			result.SetError(fmt.Sprintf("requeue data log: %s", err.Error()), false)
			return
		}
	}

	if len(dataLogs) < limit {
		result.IsDone = true
	} else {
		result.IsDone = false
	}

	// update the state of the main thread
	mainState["reprocessed"] = reprocessed
	mainState["last_id"] = lastID
	mainState["last_id_event_at"] = lastIDEventAt

	result.UpdatedWorkerState = mainState

	// log.Printf("data import result: %+v", result)

	return
}
