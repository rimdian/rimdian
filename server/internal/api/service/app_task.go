package service

import (
	"context"
	"time"

	"github.com/rimdian/rimdian/internal/api/entity"
	"go.opencensus.io/trace"
)

func TaskExecUpgradeApp(ctx context.Context, pipe *TaskExecPipeline) (result *entity.TaskExecResult) {

	spanCtx, span := trace.StartSpan(ctx, "TaskExecUpgradeApp")
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
		pipe.Logger.Printf("TaskExecUpgradeApp: workspace %s, taskExec %s, worker %d, took %s", pipe.Workspace.ID, pipe.TaskExec.ID, pipe.TaskExecPayload.WorkerID, time.Since(startedAt))
	}()

	// by default, keep current state
	// mainState := pipe.TaskExec.State.Workers[0]

	// bgCtx := context.Background()

	// TODO

	// // update state
	// offset += limit
	// mainState["offset"] = offset
	// result.UpdatedWorkerState = mainState

	// result.Message = entity.StringPtr(fmt.Sprintf("subscription list: %v, offset: %d", subscriptionListID, offset))
	// result.ItemsToImport = items

	return result
}
