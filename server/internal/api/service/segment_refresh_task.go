package service

import (
	"context"
	"fmt"
	"time"

	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"go.opencensus.io/trace"
)

func TaskExecRefreshOutdatedSegments(ctx context.Context, pipe *TaskExecPipeline) (result *entity.TaskExecResult) {

	spanCtx, span := trace.StartSpan(ctx, "TaskExecRefreshOutdatedSegments")
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
		pipe.Logger.Printf("TaskRefreshOutdatedSegments: workspace %s, taskExec %s, worker %d, took %s", pipe.Workspace.ID, pipe.TaskExec.ID, pipe.TaskExecPayload.WorkerID, time.Since(startedAt))
	}()

	// by default, keep current state
	mainState := pipe.TaskExec.State.Workers[0]

	bgCtx := context.Background()

	// get the segments
	segments, err := pipe.Repo().ListSegments(bgCtx, pipe.Workspace.ID, false)

	if err != nil {
		result.SetError(fmt.Sprintf("ListSegments err %v", err), false)
		return
	}

	retryDelay := 15

	// find segments that have a relative time filter

	for _, segment := range segments {
		if segment.HasRelativeTimeFilter() {

			// check if the segment has been processed
			if _, ok := mainState[segment.ID]; !ok {

				// create a task to recompute the segment
				// retry if a similar task is running
				state := entity.NewTaskState()
				state.Workers[0] = entity.TaskWorkerState{
					"segment_id":      segment.ID,
					"segment_version": float64(segment.Version), // float64 because of JSON
					"current_step":    entity.RecomputeSegmentStepMatchUsers,
				}

				code, err := pipe.CreateTask(ctx, &entity.TaskExec{
					TaskID:          entity.TaskKindRecomputeSegment,
					Name:            fmt.Sprintf("Refresh outdated segment %v, version %v", segment.Name, segment.Version),
					MultipleExecKey: entity.StringPtr(segment.ID),       // deduplicate tasks by segment ID
					OnMultipleExec:  entity.OnMultipleExecAbortExisting, // aborting existing task if segment is updated
					State:           state,
				})

				if err != nil {
					if code == 500 {
						result.SetError(err.Error(), false)
						return
					}
					result.SetError(err.Error(), true)
					return
				}

				// add it to the state
				mainState[segment.ID] = "processing"
				result.Message = entity.StringPtr(fmt.Sprintf("Processing segment %s", segment.ID))
				result.DelayNextRequestInSecs = &retryDelay
				result.UpdatedWorkerState = mainState
				return result
			}

			// if it is, check if it is still processing
			if mainState[segment.ID].(string) == "processing" {

				// TODO: check if a similar task is running for this segment
				tasks, _, _, _, err := pipe.Repo().ListTaskExecs(bgCtx, pipe.Workspace.ID, &dto.TaskExecListParams{
					WorkspaceID:     pipe.Workspace.ID,
					Limit:           1,
					TaskID:          &entity.TaskKindRecomputeSegment,
					MultipleExecKey: entity.StringPtr(segment.ID),
					Status:          &entity.TaskExecStatusProcessing,
				})

				if err != nil {
					result.SetError(fmt.Sprintf("ListTaskExecs err %v", err), false)
					return
				}

				// if its processing, check again in few secs
				if len(tasks) > 0 {
					result.DelayNextRequestInSecs = &retryDelay
					return result
				}

				// if it is not processing, mark it as done
				mainState[segment.ID] = "done"
				return result
			}
		}
	}

	// no more segments to process
	result.Message = entity.StringPtr("All segments processed")
	result.IsDone = true

	return
}
