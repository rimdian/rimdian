package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (svc *ServiceImpl) TaskRun(ctx context.Context, accountID string, params *dto.TaskRunParams) (code int, err error) {

	// fetch workspace
	workspace, err := svc.Repo.GetWorkspace(ctx, params.WorkspaceID)

	if err != nil {
		if sqlscan.NotFound(err) {
			return 400, err
		}
		return 500, eris.Wrap(err, "TaskRun")
	}

	// verify that token is owner of its organization
	isAccount, code, err := svc.IsAccountOfOrganization(ctx, accountID, workspace.OrganizationID)

	if err != nil {
		return code, eris.Wrap(err, "TaskRun")
	}

	if !isAccount {
		return 400, eris.New("account is not part of the organization")
	}

	// find task

	// get all the tasks that need to be launched
	task, err := svc.Repo.GetTask(ctx, params.WorkspaceID, params.ID, nil)
	if err != nil {
		if sqlscan.NotFound(err) {
			return 400, eris.Wrap(err, "TaskRun")
		}
		return 500, eris.Wrap(err, "TaskRun")
	}

	if task == nil {
		return 400, eris.Wrap(err, "TaskRun")
	}

	mainState := entity.NewTaskState()

	if params.MainWorkerState != nil {
		mainState.Workers[0] = params.MainWorkerState
	}

	taskExec := &entity.TaskExec{
		TaskID:         task.TaskManifest.ID,
		Name:           task.TaskManifest.Name,
		OnMultipleExec: task.OnMultipleExec,
		State:          mainState,
	}

	code, err = svc.doTaskCreate(ctx, task.WorkspaceID, taskExec)

	return code, err
}

// check if some app cron tasks need to be launched
func (svc *ServiceImpl) TaskWakeUpCron(ctx context.Context) (code int, err error) {

	scheduledTasks, err := svc.Repo.ListTasksToWakeUp(ctx)
	if err != nil {
		return 500, eris.Wrap(err, "TaskWakeUpCron")
	}

	// launch them
	for _, scheduledTask := range scheduledTasks {
		code, err = svc.Repo.RunInTransactionForSystem(ctx, func(ctx context.Context, tx *sql.Tx) (int, error) {

			// get scheduled task to make sure its not already running
			scheduledTask, err := svc.Repo.GetTask(ctx, scheduledTask.WorkspaceID, scheduledTask.ID, tx)

			if err != nil {
				return 500, eris.Wrap(err, "TaskWakeUpCron")
			}

			// should not happen but just in case
			if !scheduledTask.ShouldRun() {
				return 200, nil
			}

			// svc.Logger.Printf("will launch scheduledTask: %+v/n", scheduledTask)

			taskExec := &entity.TaskExec{
				TaskID:         scheduledTask.ID,
				Name:           scheduledTask.Name,
				OnMultipleExec: scheduledTask.OnMultipleExec,
				State:          entity.NewTaskState(),
			}

			code, err := svc.doTaskCreate(ctx, scheduledTask.WorkspaceID, taskExec)

			if err != nil {
				// return error only for errors that are not "already running"
				if !eris.Is(err, entity.ErrTaskAlreadyRunningDiscardNew) && !eris.Is(err, entity.ErrTaskAlreadyRunningWillRetry) {
					return code, eris.Wrap(err, "TaskWakeUpCron")
				}
			}

			scheduledTask.ComputeNextRun()
			scheduledTask.LastRun = entity.TimePtr(time.Now())

			// update scheduled task
			if err = svc.Repo.UpdateTask(ctx, scheduledTask, tx); err != nil {
				return 500, eris.Wrap(err, "TaskWakeUpCron")
			}

			return 200, nil
		})

		if err != nil {
			return code, eris.Wrap(err, "TaskWakeUpCron")
		}
	}

	return 200, nil
}

func (svc *ServiceImpl) TaskList(ctx context.Context, accountID string, params *dto.TaskListParams) (result *dto.TaskListResult, code int, err error) {

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, params.WorkspaceID, accountID)

	if err != nil {
		return nil, code, eris.Wrap(err, "TaskList")
	}
	// fetch segments
	result = &dto.TaskListResult{}

	result.Tasks, err = svc.Repo.ListTasks(ctx, workspace.ID)

	if err != nil {
		return nil, 500, err
	}

	return result, 200, nil
}
