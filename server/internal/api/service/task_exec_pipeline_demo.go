package service

import (
	"context"
	"math"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/internal/common/dto"
	"go.opencensus.io/trace"
)

// The demo task will:
// 1. init: launch 10 parallel workers
// 2. loading: each worker will inject data_logs into the Collector
// 3. processing: wait for data_logs to be processed
// 4. end

// if worker ID == 0
// - check task status (init | loading)
// - init: launches 10 workers and changes status to loading
// - loading: polls every 15 secs to checks when workers are all done
//
// if worker ID > 0
// - check task status (loading | processing)
// - loading: injects a fraction of the 60 days of fixtures
// - processing: waits for data_logs to complete

func TaskExecGenerateDemo(ctx context.Context, pipe *TaskExecPipeline) (result *entity.TaskExecResult) {

	spanCtx, span := trace.StartSpan(ctx, "TaskExecGenerateDemo")
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

	mainWorkerPollingDelay := 15
	workersCount := 10

	totalDaysGenerated := 60 // generate fake data for 60 days of activity, multiple of 10

	// we are in main thread
	if pipe.TaskExecPayload.WorkerID == 0 {

		mainWorkerState := pipe.TaskExec.State.Workers[0]

		// check current status
		if _, ok := mainWorkerState["status"]; !ok {
			result.SetError("generate demo task main thread state has no status", true)
			return
		}

		status := mainWorkerState["status"].(string)

		// init does creates 10 parallel workers to inject fixtures data
		// then does polling to check when workers are done to end the task
		if status == entity.DemoTaskStatusInit {

			// divide 60 days of data fixtures for 10 workers = 6 days per worker
			fromDay := 0
			increment := totalDaysGenerated / 10

			for i := 1; i <= workersCount; i++ {

				toDay := fromDay + increment
				workerState := map[string]interface{}{
					"status":          entity.DemoTaskStatusInit,
					"current_day":     fromDay + 1,
					"from_day":        fromDay + 1,
					"to_day":          toDay,
					"total_data_logs": 0,
				}

				pipe.TaskExecAddWorker(spanCtx, i, workerState)

				if pipe.HasError() {
					result.SetError("error while adding worker", false)
					return
				}

				fromDay = fromDay + increment
			}

			mainWorkerState["status"] = entity.DemoTaskStatusLoading

			result.UpdatedWorkerState = mainWorkerState
			result.IsDone = false
			result.DelayNextRequestInSecs = &mainWorkerPollingDelay

			return
		}

		if status == entity.DemoTaskStatusLoading {
			// check if all workers are done loading data
			isDoneLoading := true

			// starts at worker 1 (0 is parent)
			for i := 1; i < workersCount; i++ {
				if _, ok := pipe.TaskExec.State.Workers[i]; !ok {
					isDoneLoading = false
				}

				if _, ok := pipe.TaskExec.State.Workers[i]["status"]; !ok {
					isDoneLoading = false
				}

				if pipe.TaskExec.State.Workers[i]["status"].(string) != "done" {
					isDoneLoading = false
				}
			}

			if isDoneLoading {
				mainWorkerState["status"] = entity.DemoTaskStatusProcessing
			}

			result.UpdatedWorkerState = mainWorkerState
			result.IsDone = false
			result.DelayNextRequestInSecs = &mainWorkerPollingDelay
			return
		}

		// we are actually processing data_logs, check progress
		processedDataImports, err := pipe.Repository.CountSuccessfulDataLogsForDemo(spanCtx, pipe.Workspace.ID)

		if err != nil {
			result.SetError(err.Error(), false)
			return
		}

		mainWorkerState["processed_data_logs"] = processedDataImports

		totalDataImports := int64(0)

		for _, workerState := range pipe.TaskExec.State.Workers {
			if _, ok := workerState["total_data_logs"]; ok {
				totalDataImports += int64(workerState["total_data_logs"].(float64))
			}
		}

		// import is not yet done
		if processedDataImports < totalDataImports {

			result.UpdatedWorkerState = mainWorkerState
			result.IsDone = false
			return
		}

		// else is done
		mainWorkerState["status"] = entity.DemoTaskStatusDone
		result.UpdatedWorkerState = mainWorkerState
		result.IsDone = true
		return
	}

	// we are in a worker job
	currentWorkerState := pipe.TaskExec.State.Workers[pipe.TaskExecPayload.WorkerID]

	status := currentWorkerState["status"].(string)

	if status == "done" {
		// worker finished its job
		return
	}

	// fromDay := currentWorkerState["fromDay"].(float64)
	toDay := currentWorkerState["to_day"].(float64)
	currentDay := currentWorkerState["current_day"].(float64)

	if status == entity.DemoTaskStatusInit {
		currentWorkerState["status"] = entity.DemoTaskStatusLoading
	}

	// worker generate fixtures here
	items := []string{}

	if *pipe.Workspace.DemoKind == entity.WorkspaceDemoOrder {
		var err error
		items, err = generateOrderDemoFixtures(spanCtx, pipe.Workspace, int(currentDay), totalDaysGenerated)

		if err != nil {
			result.SetError(err.Error(), false)
			return
		}
	}

	itemsCount := len(items)

	// log.Printf("Day %v, %v items", currentDay, itemsCount)

	// cut items into batches of 10 and import them internally
	const maxBatchSize int = 10
	skip := 0

	batchCount := int(math.Ceil(float64(itemsCount) / float64(maxBatchSize)))

	for i := 1; i <= batchCount; i++ {

		lowerBound := skip
		upperBound := skip + maxBatchSize

		if upperBound > itemsCount {
			upperBound = itemsCount
		}

		skip += maxBatchSize

		// log.Printf("batch %v, %v to %v", i, lowerBound, upperBound)
		pipe.DataLogEnqueue(spanCtx, nil, dto.DataLogOriginInternalTaskExec, pipe.TaskExec.ID, pipe.Workspace.ID, items[lowerBound:upperBound], false)
		if pipe.HasError() {
			result.SetError("error while enqueueing data logs", false)
			return
		}
	}

	totalDataImports := int(currentWorkerState["total_data_logs"].(float64))

	// set next state
	if currentDay < toDay {
		currentWorkerState["current_day"] = currentDay + 1
		currentWorkerState["total_data_logs"] = totalDataImports + batchCount
	} else {
		currentWorkerState["status"] = "done"
		currentWorkerState["total_data_logs"] = totalDataImports + batchCount
		result.IsDone = true // end this worker
		result.UpdatedWorkerState = currentWorkerState
		return
	}

	// update worker state and return
	result.UpdatedWorkerState = currentWorkerState
	return
}
