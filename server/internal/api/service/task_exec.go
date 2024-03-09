package service

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	common "github.com/rimdian/rimdian/internal/common/dto"
	"github.com/rimdian/rimdian/internal/common/taskorchestrator"
	"github.com/rotisserie/eris"
)

var (
	TaskTimeoutInSecs int64  = 25
	TaskExecEndpoint  string = "/api/taskExec.do"
	TasksQueueName           = "tasks"
)

func (svc *ServiceImpl) TaskExecJobs(ctx context.Context, accountID string, params *dto.TaskExecJobsParams) (result *dto.TaskExecJobsResult, code int, err error) {

	_, code, err = svc.GetWorkspaceForAccount(ctx, params.WorkspaceID, accountID)

	if err != nil {
		return nil, code, eris.Wrap(err, "TaskExecJobs")
	}

	if params.Offset < 0 {
		params.Offset = 0
	}

	if params.Limit < 1 || params.Limit > 50 {
		params.Limit = 5
	}

	// get jobs
	jobs, total, err := svc.Repo.GetTaskExecJobs(ctx, params.WorkspaceID, params.TaskExecID, params.Offset, params.Limit)

	if err != nil {
		return nil, 500, eris.Wrap(err, "TaskExecJobInfo")
	}

	result = &dto.TaskExecJobsResult{
		TaskExecJobs: jobs,
		Total:        total,
		Offset:       params.Offset,
		Limit:        params.Limit,
	}

	return result, 200, nil
}

func (svc *ServiceImpl) TaskExecJobInfo(ctx context.Context, accountID string, params *dto.TaskExecJobInfoParams) (jobInfo *dto.TaskExecJobInfoInfo, code int, err error) {

	_, code, err = svc.GetWorkspaceForAccount(ctx, params.WorkspaceID, accountID)

	if err != nil {
		return nil, code, eris.Wrap(err, "TaskExecJobInfo")
	}

	// get job
	job, err := svc.Repo.GetTaskExecJob(ctx, params.WorkspaceID, params.JobID)

	if err != nil {
		if eris.Is(err, entity.ErrTaskExecJobNotFound) {
			return nil, 400, err
		}
		return nil, 500, eris.Wrap(err, "TaskExecJobInfo")
	}

	if job.DoneAt != nil {
		return nil, 400, eris.New("the job is already done, info not available")
	}

	jobInfo, err = svc.TaskOrchestrator.GetTaskRunningJob(ctx, svc.Config.TASK_QUEUE_LOCATION, TasksQueueName, params.JobID)

	if err != nil {
		// code 400 to propagate the error to the browser UI
		return nil, 400, eris.Wrap(err, "TaskExecJobInfo")
	}

	return jobInfo, 200, nil
}

func (svc *ServiceImpl) TaskExecList(ctx context.Context, accountID string, params *dto.TaskExecListParams) (result *dto.TaskExecListResult, code int, err error) {

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, params.WorkspaceID, accountID)

	if err != nil {
		return nil, code, eris.Wrap(err, "TaskExecList")
	}

	// fetch tasks
	result = &dto.TaskExecListResult{}

	result.TaskExecs, result.NextToken, result.PreviousToken, code, err = svc.Repo.ListTaskExecs(ctx, workspace.ID, params)

	if err != nil {
		return nil, code, err
	}

	return result, 200, nil
}

// abort runnig task
func (svc *ServiceImpl) TaskExecAbort(ctx context.Context, accountID string, params *dto.TaskExecAbortParams) (code int, err error) {

	// fetch workspace
	workspace, err := svc.Repo.GetWorkspace(ctx, params.WorkspaceID)

	if err != nil {
		if sqlscan.NotFound(err) {
			return 400, err
		}
		return 500, eris.Wrap(err, "TaskAbort")
	}

	// verify that token is account of its organization
	isAccount, code, err := svc.IsAccountOfOrganization(ctx, accountID, workspace.OrganizationID)

	if err != nil {
		return code, eris.Wrap(err, "TaskAbort")
	}

	if !isAccount {
		return 400, eris.New("account is not part of the organization")
	}

	// automaticaly set the message if not provided
	if params.Message == "" {
		// get account
		account, err := svc.Repo.GetAccountFromID(ctx, accountID)

		if err != nil {
			return 500, eris.Wrap(err, "TaskAbort")
		}

		params.Message = "Task aborted by " + account.Email
	}

	return svc.Repo.RunInTransactionForWorkspace(ctx, workspace.ID, func(ctx context.Context, tx *sql.Tx) (int, error) {

		if err := svc.Repo.AbortTaskExec(ctx, params.TaskExecID, params.Message, tx); err != nil {
			return 500, err
		}

		return 200, nil
	})
}

// create new task
func (svc *ServiceImpl) TaskExecCreate(ctx context.Context, accountID string, params *dto.TaskExecCreateParams) (code int, err error) {

	// fetch workspace
	workspace, err := svc.Repo.GetWorkspace(ctx, params.WorkspaceID)

	if err != nil {
		if sqlscan.NotFound(err) {
			return 400, err
		}
		return 500, eris.Wrap(err, "TaskCreate")
	}

	// verify that token is account of its organization
	isAccount, code, err := svc.IsAccountOfOrganization(ctx, accountID, workspace.OrganizationID)

	if err != nil {
		return code, eris.Wrap(err, "TaskCreate")
	}

	if !isAccount {
		return 400, eris.New("account is not part of the organization")
	}

	if err := params.Validate(); err != nil {
		return 400, eris.Wrap(err, "TaskCreate")
	}

	state := entity.NewTaskState()
	state.Workers[0] = params.State

	taskExec := &entity.TaskExec{
		TaskID:          params.TaskID,
		Name:            params.Name,
		MultipleExecKey: params.MultipleExecKey,
		OnMultipleExec:  params.OnMultipleExec,
		State:           state,
	}

	return svc.doTaskCreate(ctx, workspace.ID, taskExec)
}

// receives a job from the Cloud Task Orchestrator
// it should return before 25 seconds deadline
// code 400 will persist error message, end task and avoid retrying
// code 500 will persist error message, and retry
// code 200 will clean eventual error previously persisted, and end task

func (svc *ServiceImpl) TaskExecDo(ctx context.Context, workspaceID string, payload *dto.TaskExecRequestPayload) (result *common.DataLogInQueueResult) {

	if payload == nil {
		return &common.DataLogInQueueResult{
			HasError:         true,
			Error:            "task payload required",
			QueueShouldRetry: false,
		}
	}

	// fetch workspace from DB
	workspace, err := svc.Repo.GetWorkspace(ctx, workspaceID)

	if err != nil {
		// check if not found
		if sqlscan.NotFound(err) {
			return &common.DataLogInQueueResult{
				HasError:         true,
				Error:            fmt.Sprintf("TaskExecDo: workspace not found: %v", workspaceID),
				QueueShouldRetry: false,
			}
		}

		return &common.DataLogInQueueResult{
			HasError:         true,
			Error:            fmt.Sprintf("TaskExecDo: %v", err),
			QueueShouldRetry: true,
		}
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Duration(TaskTimeoutInSecs)*time.Second)
	defer cancel()

	props := &TaskExecPipelineProps{
		Config:           svc.Config,
		NetClient:        svc.NetClient,
		Repository:       svc.Repo,
		Workspace:        workspace,
		TaskOrchestrator: svc.TaskOrchestrator,
		TaskExecPayload:  payload,
	}

	pipeline := NewTaskExecPipeline(props)
	pipeline.Execute(ctxWithTimeout)

	result = pipeline.GetQueueResult()
	return
}

// creates a new task and inserts it in DB & enqueue a new job to exec the task
func (svc *ServiceImpl) doTaskCreate(ctx context.Context, workspaceID string, taskExec *entity.TaskExec) (code int, err error) {

	if workspaceID == "" {
		return 500, eris.New("TaskCreate workspace id is required")
	}

	// insert task in db + enqueue job
	code, err = svc.Repo.RunInTransactionForWorkspace(ctx, workspaceID, func(ctx context.Context, tx *sql.Tx) (int, error) {

		if taskExec.OnMultipleExec != entity.OnMultipleExecAllow {
			runningTask, err := svc.Repo.GetRunningTaskExecByTaskID(ctx, taskExec.TaskID, taskExec.MultipleExecKey, tx)
			if err != nil {
				return 500, err
			}

			switch taskExec.OnMultipleExec {
			case entity.OnMultipleExecAllow:
				// do nothing
			case entity.OnMultipleExecDiscardNew:
				// discard new task
				if runningTask != nil {
					return 400, eris.Wrapf(entity.ErrTaskAlreadyRunningDiscardNew, ": %v", runningTask.ID)
				}
			case entity.OnMultipleExecRetryLater:
				// return a 503, Could Task will retry later
				if runningTask != nil {
					return http.StatusServiceUnavailable, entity.ErrTaskAlreadyRunningWillRetry
				}
			case entity.OnMultipleExecAbortExisting:
				// end existing task
				if runningTask != nil {
					if err := svc.Repo.AbortTaskExec(ctx, runningTask.ID, "A similar task has been launched", tx); err != nil {
						return 500, err
					}
				}
			default:
				// do nothing
			}
		}

		// create new job
		job := entity.NewTaskExecJob(taskExec.TaskID)

		googleTaskQueueJob := &taskorchestrator.TaskRequest{
			JobID:             &job.ID,
			QueueLocation:     svc.Config.TASK_QUEUE_LOCATION,
			QueueName:         TasksQueueName,
			PostEndpoint:      svc.Config.API_ENDPOINT + TaskExecEndpoint + "?workspace_id=" + workspaceID,
			TaskTimeoutInSecs: &TaskTimeoutInSecs,
		}

		taskExec.EnsureID()

		// customize task for its kind
		switch taskExec.TaskID {

		case entity.TaskKindGenerateDemo:
			taskExec.ID = workspaceID
			job.TaskExecID = taskExec.ID

			googleTaskQueueJob.Payload = dto.TaskExecRequestPayload{
				TaskExecID: taskExec.ID,
				WorkerID:   0,
				JobID:      job.ID,
			}

		case entity.TaskKindDataLogReprocessUntil:
			job.TaskExecID = taskExec.ID

			googleTaskQueueJob.Payload = dto.TaskExecRequestPayload{
				TaskExecID: taskExec.ID,
				WorkerID:   0,
				JobID:      job.ID,
			}

		case entity.TaskKindReattributeConversions:
			job.TaskExecID = taskExec.ID

			googleTaskQueueJob.Payload = dto.TaskExecRequestPayload{
				TaskExecID: taskExec.ID,
				WorkerID:   0,
				JobID:      job.ID,
			}

		case entity.TaskKindRecomputeSegment:
			job.TaskExecID = taskExec.ID

			googleTaskQueueJob.Payload = dto.TaskExecRequestPayload{
				TaskExecID: taskExec.ID,
				WorkerID:   0,
				JobID:      job.ID,
			}

		default:
			// check if task starts with "app_" and find an app to call
			if !strings.HasPrefix(taskExec.TaskID, "app_") && !strings.HasPrefix(taskExec.TaskID, "appx_") {
				return 500, eris.Wrapf(entity.ErrTaskKindNotImplemented, ": %v", taskExec.TaskID)
			}
			bits := strings.Split(taskExec.TaskID, "_")

			if bits == nil || len(bits) < 3 {
				return 500, eris.Wrapf(entity.ErrTaskKindNotImplemented, ": %v", taskExec.TaskID)
			}

			// verify that app exists
			appID := bits[0] + "_" + bits[1]

			_, err := svc.Repo.GetApp(ctx, workspaceID, appID)
			if err != nil {
				if sqlscan.NotFound(err) {
					return 400, err
				}
				return 500, err
			}

			job.TaskExecID = taskExec.ID

			googleTaskQueueJob.Payload = dto.TaskExecRequestPayload{
				TaskExecID: taskExec.ID,
				WorkerID:   0,
				JobID:      job.ID,
			}
		}

		// persist in db
		if err := svc.Repo.InsertTaskExec(ctx, workspaceID, taskExec, job, tx); err != nil {
			if eris.Is(err, entity.ErrTaskExecAlreadyExists) {
				return 400, eris.Wrap(entity.ErrTaskExecAlreadyExists, "TaskCreate")
			}
			return 500, err
		}

		// enqueue job
		if err := svc.TaskOrchestrator.PostRequest(ctx, googleTaskQueueJob); err != nil {
			return 500, err
		}

		return 200, nil
	})

	if err != nil {
		return code, eris.Wrap(err, "TaskCreate")
	}

	// Post-task creation actions
	// update workspace last conversion attribution date
	if taskExec.TaskID == entity.TaskKindReattributeConversions {

		// update workspace
		code, err = svc.Repo.RunInTransactionForSystem(ctx, func(ctx context.Context, tx *sql.Tx) (code int, err error) {

			workspace, err := svc.Repo.GetWorkspace(ctx, workspaceID)
			if err != nil {
				if sqlscan.NotFound(err) {
					return 400, err
				}
				return 500, err
			}

			workspace.OutdatedConversionsAttribution = false

			if err := svc.Repo.UpdateWorkspace(ctx, workspace, tx); err != nil {
				return 500, eris.Wrap(err, "TaskCreate")
			}

			return 200, nil
		})

		if err != nil {
			return code, eris.Wrap(err, "TaskCreate")
		}
	}

	return 200, nil
}
