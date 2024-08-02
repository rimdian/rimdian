package service

import (
	"context"
	"time"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

// create a Task that wil fetch all the data imports that have not been processed after a certain delay
// en requeues them for processing
// this is useful when the Task Queue is down and we need to reprocess the data imports
// or when we want to reprocess a batch of data imports
func (svc *ServiceImpl) DataLogReprocessUntil(ctx context.Context, untilDate time.Time) (code int, err error) {

	// loop over workspaces, count if there are data imports that need to be reprocessed and create a task to reprocess them

	// get all the workspaces
	workspaces, err := svc.Repo.ListWorkspaces(ctx, nil)

	if err != nil {
		return 500, eris.Wrap(err, "DataLogReprocessUntil")
	}

	// loop over workspaces
	for _, workspace := range workspaces {

		// count the number of data imports that need to be reprocessed
		foundOne, err := svc.Repo.HasDataLogsToReprocess(ctx, workspace.ID, untilDate)

		if err != nil {
			return 500, eris.Wrap(err, "DataLogReprocessUntil")
		}

		if !foundOne {
			continue
		}

		// if there are data imports to reprocess, create a task to reprocess them

		taskState := entity.TaskExecState{
			Workers: map[int]entity.TaskWorkerState{
				// 0: main thread state
				0: map[string]interface{}{
					"until_date":       untilDate.Format(time.RFC3339),
					"reprocessed":      0,  // number of data imports that have been reprocessed
					"last_id":          "", // ID of the last data import that was reprocessed, used for pagination
					"last_id_event_at": "", // time when the last ID was received, used for pagination
				},
			},
		}

		onMultipleExec := entity.OnMultipleExecDiscardNew

		taskExec := &entity.TaskExec{
			TaskID:         entity.TaskKindDataLogReprocessUntil,
			Name:           "Reprocess data imports",
			OnMultipleExec: onMultipleExec,
			State:          taskState,
		}

		code, err = DoTaskCreate(ctx, svc.Repo, svc.Config, svc.TaskOrchestrator, workspace.ID, taskExec)

		if err != nil {
			// ignore error if a similar task is already running
			if !eris.Is(err, entity.ErrTaskAlreadyRunningDiscardNew) && !eris.Is(err, entity.ErrTaskAlreadyRunningWillRetry) {
				return code, eris.Wrap(err, "DataLogReprocessUntil")
			}
		}
	}

	return 200, nil
}
