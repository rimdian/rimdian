package service

import (
	"context"
	"fmt"
	"time"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	common "github.com/rimdian/rimdian/internal/common/dto"
	"go.opencensus.io/trace"
)

func TaskExecRecomputeSegment(ctx context.Context, pipe *TaskExecPipeline) (result *entity.TaskExecResult) {

	spanCtx, span := trace.StartSpan(ctx, "TaskExecRecomputeSegment")
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
		pipe.Logger.Printf("TaskRecomputeSegment: workspace %s, taskExec %s, worker %d, took %s", pipe.Workspace.ID, pipe.TaskExec.ID, pipe.TaskExecPayload.WorkerID, time.Since(startedAt))
	}()

	// by default, keep current state
	mainState := pipe.TaskExec.State.Workers[0]

	bgCtx := context.Background()

	// get the segment
	currentStep := mainState["current_step"].(string)
	segmentID := mainState["segment_id"].(string)
	segmentVersion := int(mainState["segment_version"].(float64))

	segment, err := pipe.Repo().GetSegment(bgCtx, pipe.Workspace.ID, segmentID)

	if err != nil {
		if sqlscan.NotFound(err) {
			result.SetError(fmt.Sprintf("Segment %s not found", segmentID), true)
			return
		}
		result.SetError(fmt.Sprintf("Segment err %v", err), false)
		return
	}

	// check that the segment version is the same
	if segment.Version != segmentVersion {
		result.SetError(fmt.Sprintf("Segment %s, version %v could not be activated (newer version building...)", segment.ID, segment.Version), true)
		return
	}

	// check that the segment is not deleted
	if segment.Status == entity.SegmentStatusDeleted {
		result.SetError(fmt.Sprintf("Segment %s is deleted", segment.ID), true)
		return
	}

	switch currentStep {
	case entity.RecomputeSegmentStepMatchUsers:
		// wipe the user_segment_enter + user_segment_exit tables
		if err := pipe.Repo().ClearUserSegmentQueue(bgCtx, pipe.Workspace.ID, segmentID, segmentVersion); err != nil {
			result.SetError(fmt.Sprintf("ClearUserSegmentQueue err %v", err), false)
			return
		}
		entersCount, exitCount, err := pipe.Repo().EnqueueMatchingSegmentUsers(bgCtx, pipe.Workspace.ID, segment)

		if err != nil {
			result.SetError(fmt.Sprintf("EnqueueMatchingSegmentUsers err %v", err), false)
			return
		}

		mainState["users_to_enter"] = float64(entersCount)
		mainState["users_to_exit"] = float64(exitCount)
		mainState["timeline_automations_processed"] = float64(0.0)

		mainState["current_step"] = entity.RecomputeSegmentStepEnterUsers
		result.Message = entity.StringPtr(fmt.Sprintf("Matched %d users to segment %s, enters %v, exit: %v", entersCount+exitCount, segmentID, entersCount, exitCount))

	case entity.RecomputeSegmentStepEnterUsers:
		// Insert all the entering users in the user_segment table
		if err := pipe.Repo().EnterUserSegmentFromQueue(bgCtx, pipe.Workspace.ID, segmentID, segmentVersion); err != nil {
			result.SetError(fmt.Sprintf("EnterUserSegmentFromQueue err %v", err), false)
			return
		}

		// update the state of the main thread
		mainState["current_step"] = entity.RecomputeSegmentStepExitUsers

	case entity.RecomputeSegmentStepExitUsers:

		// Insert all the entering users in the user_segment table
		if err := pipe.Repo().ExitUserSegmentFromQueue(bgCtx, pipe.Workspace.ID, segmentID, segmentVersion); err != nil {
			result.SetError(fmt.Sprintf("ExitUserSegmentFromQueue err %v", err), false)
			return
		}

		// set the segment as active
		didActivate, err := pipe.Repo().ActivateSegment(bgCtx, pipe.Workspace.ID, segment.ID, segment.Version)

		if err != nil {
			result.SetError(fmt.Sprintf("ActivateSegment err %v", err), false)
			return
		}

		if !didActivate {
			result.SetError(fmt.Sprintf("Segment %s, version %v could not be activated (newer version building...)", segment.ID, segment.Version), true)
			return
		}

		// update the state of the main thread
		mainState["current_step"] = entity.RecomputeSegmentStepEnterDataLogs

	case entity.RecomputeSegmentStepEnterDataLogs:

		mainState["has_hooks"] = false
		mainState["has_workflows"] = false
		checkpoint := entity.DataLogCheckpointDone

		// TODO: find if workflows are matched here

		for _, hook := range pipe.GetWorkspace().DataHooks {
			if hook.MatchesDataLog("segment", "enter") {
				mainState["has_hooks"] = true
				checkpoint = entity.DataLogCheckpointWorkflowsTriggered
			}
		}

		// Insert all the entering users in the item_timelnine table
		if err := pipe.Repo().InsertSegmentDataLogs(bgCtx, pipe.Workspace.ID, segmentID, segmentVersion, pipe.TaskExec.ID, true, *pipe.TaskExec.DBCreatedAt, checkpoint); err != nil {
			result.SetError(fmt.Sprintf("InsertSegmentDataLogs err %v", err), false)
			return
		}

		mainState["current_step"] = entity.RecomputeSegmentStepExitDataLogs

	case entity.RecomputeSegmentStepExitDataLogs:

		checkpoint := entity.DataLogCheckpointDone
		if mainState["has_hooks"].(bool) {
			checkpoint = entity.DataLogCheckpointWorkflowsTriggered
		}

		// TODO in future: find if workflows are matched here

		for _, hook := range pipe.GetWorkspace().DataHooks {
			if hook.MatchesDataLog("segment", "exit") {
				mainState["has_hooks"] = true
				checkpoint = entity.DataLogCheckpointWorkflowsTriggered
			}
		}

		for _, hook := range pipe.GetWorkspace().DataHooks {
			if hook.MatchesDataLog("segment", "exit") {
				mainState["has_hooks"] = true
				checkpoint = entity.DataLogCheckpointWorkflowsTriggered
			}
		}
		// Insert all the exiting users in the item_timelnine table
		if err := pipe.Repo().InsertSegmentDataLogs(bgCtx, pipe.Workspace.ID, segmentID, segmentVersion, pipe.TaskExec.ID, false, *pipe.TaskExec.DBCreatedAt, checkpoint); err != nil {
			result.SetError(fmt.Sprintf("InsertSegmentDataLogs err %v", err), false)
			return
		}

		// clear queue
		if err := pipe.Repo().ClearUserSegmentQueue(bgCtx, pipe.Workspace.ID, segmentID, segmentVersion); err != nil {
			result.SetError(fmt.Sprintf("ClearUserSegmentQueue err %v", err), false)
			return
		}

		mainState["current_step"] = entity.RecomputeSegmentStepEnqueueJobs

	case entity.RecomputeSegmentStepEnqueueJobs:

		if mainState["has_hooks"].(bool) || mainState["has_workflows"].(bool) {
			// fetch enough rows for the current process to feed the semaphore for 20 secs
			// because the workers are not yet implemented, we give the wordkerID = 0

			var withNextToken *string

			if _, ok := mainState["next_token"]; ok {
				withNextToken = entity.StringPtr(mainState["next_token"].(string))
			}

			// get 10 rows and enqueue them, and do it again until we have no more rows or we have less than 5 secs remaining

			shouldContinue := true
			hasMoreRows := true
			limit := 50 //  50 rows = 5 secs with 100ms enqueing latency per row

			for shouldContinue {

				// check if the we have less than 5 secs remaining
				if deadline, _ := spanCtx.Deadline(); time.Until(deadline) < 5*time.Second {
					shouldContinue = false
					pipe.Logger.Printf("TaskRecomputeSegment: deadline ellapsed, should continue = false")
					continue
				}

				// fetch 11 rows but will only enqueue 10
				// the last row will be used to determine if we have more rows
				rows, err := pipe.Repo().ListDataLogsToRespawn(spanCtx, pipe.Workspace.ID,
					common.DataLogOriginInternalTaskExec,
					pipe.TaskExec.ID,
					entity.DataLogCheckpointWorkflowsTriggered,
					limit+1,
					withNextToken,
				)

				if err != nil {
					if sqlscan.NotFound(err) {
						result.IsDone = true
						mainState["current_step"] = entity.RecomputeSegmentStepFinalize
						result.UpdatedWorkerState = mainState
						return
					}

					result.SetError(fmt.Sprintf("ListDataLogsToRespawn err %v", err), false)
					return
				}

				for _, row := range rows {
					replayID := row.ID // copy the value
					if err := DataLogEnqueue(ctx, pipe.Config, pipe.NetClient, &replayID, common.DataLogOriginInternalTaskExec, pipe.TaskExec.ID, pipe.Workspace.ID, []string{""}, false); err != nil {
						result.SetError(fmt.Sprintf("DataLogEnqueue err %v", err), false)
						return
					}
				}

				// we have no more rows
				if len(rows) < limit+1 {
					shouldContinue = false
					hasMoreRows = false
					continue
				} else {
					withNextToken = entity.StringPtr(dto.EncodePaginationToken(rows[len(rows)-1].ID, rows[len(rows)-1].EventAt))
				}
			}

			// compute next token if we have more rows
			if hasMoreRows {
				mainState["next_token"] = withNextToken
			} else {
				// delete the next token
				delete(mainState, "next_token")
				mainState["current_step"] = entity.RecomputeSegmentStepDone

				result.UpdatedWorkerState = mainState
				result.IsDone = true
				return
			}

		} else {
			result.UpdatedWorkerState = mainState
			result.IsDone = true
			mainState["current_step"] = entity.RecomputeSegmentStepDone
			return
		}

	default:
		result.SetError(fmt.Sprintf("invalid current step: %s", currentStep), true)
		return
	}

	// update the state of the main thread
	result.UpdatedWorkerState = mainState

	return
}

// semaphore example
// process the user_segment_queue table in a semaphore of 20 for 20 secs max
// max tickets in parallel
// concurrency := 10
// oneTicketTimeout := 6 * time.Second

// sem := semaphore.New(concurrency, oneTicketTimeout)

// var wg sync.WaitGroup

// shouldContinue := true
// index := 0
// totalRows := len(rows)
// noMoreRows := false
// // atomic sync counter
// var indexesProcessed atomic.Uint64

// for shouldContinue {

// 	// check if the we have less than 5 secs remaining
// 	if deadline, _ := spanCtx.Deadline(); time.Until(deadline) < 5*time.Second {
// 		shouldContinue = false
// 		pipe.Logger.Printf("TaskRecomputeSegment: deadline ellapsed, should continue = false")
// 		continue
// 	}

// 	// blocks until a ticket is available
// 	if err := sem.Acquire(); err != nil {
// 		continue
// 	}

// 	// tickets can wait in the Acquire() before the shouldContinue changed to false
// 	// check if we can still continue
// 	if !shouldContinue {
// 		pipe.Logger.Println("abort, shouldContinue == false")
// 		// exit the loop
// 		break
// 	}

// 	if index >= totalRows {
// 		noMoreRows = true
// 		pipe.Logger.Printf("done, index >= totalRows: %d >= %d", index, totalRows)
// 		// exit the loop
// 		break
// 	}

// 	wg.Add(1)

// 	go func(currentIndex int) {
// 		defer wg.Done()
// 		defer sem.Release()

// 		row := rows[currentIndex]
// 		if err := DataLogEnqueue(); err != nil {

// 		}

// 		// increase the counter of successfully processed rows
// 		indexesProcessed.Add(1)

// 	}(index)

// 	index += 1
// }

// wg.Wait()

// // compute next token if we have more rows
// if !noMoreRows && withNextToken != nil {
// 	// nextToken := rows[len(rows)-1].EventAt.UTC().Format(time.RFC3339Nano)
// 	lastIndex := indexesProcessed.Load() - 1
// 	nextToken := dto.EncodePaginationToken(rows[lastIndex].ID, rows[lastIndex].EventAt)
// 	mainState["next_token"] = nextToken
// }

// mainState["timeline_automations_processed"] = mainState["timeline_automations_processed"].(float64) + float64(indexesProcessed.Load())

// // check if we are done
// if noMoreRows && withNextToken == nil {
// 	mainState["current_step"] = entity.RecomputeSegmentStepDone
// 	result.IsDone = true
// 	timeTaken := time.Since(task.CreatedAt.UTC())
// 	result.Message = entity.StringPtr(fmt.Sprintf("Processed %d rows in %s", totalRows, timeTaken))
// }
