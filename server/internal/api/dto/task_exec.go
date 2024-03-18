package dto

import (
	"net/http"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

var (
	ErrInvalidWebhookKind        = eris.New("invalid webhook kind")
	ErrExternalWebhookIDRequired = eris.New("external webhook id is required")
)

type TaskExecRequestPayload struct {
	TaskExecID string `json:"id"`
	WorkerID   int    `json:"worker_id"` // 0 = parent thread
	JobID      string `json:"job_id"`
}

type TaskExecAbortParams struct {
	TaskExecID  string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	Message     string `json:"message"`
}

type TaskExecCreateParams struct {
	WorkspaceID     string                 `json:"workspace_id"`
	TaskID          string                 `json:"task_id"`
	Name            string                 `json:"name"`
	MultipleExecKey *string                `json:"multiple_exec_key,omitempty"`
	OnMultipleExec  string                 `json:"on_multiple_exec"`
	State           entity.TaskWorkerState `json:"state"`
}

func (params *TaskExecCreateParams) Validate() error {
	if params.WorkspaceID == "" {
		return entity.ErrWorkspaceIDRequired
	}

	if params.TaskID == "" {
		return entity.ErrTaskIDRequired
	}

	taskFound := false

	for _, t := range entity.SystemTasks {
		if t.ID == params.TaskID {
			taskFound = true
			break
		}
	}

	if !taskFound {
		return entity.ErrTaskInvalid
	}

	if params.Name == "" {
		return entity.ErrTaskNameRequired
	}

	if !govalidator.IsIn(params.OnMultipleExec, entity.OnMultipleExecAllow, entity.OnMultipleExecDiscardNew, entity.OnMultipleExecRetryLater, entity.OnMultipleExecAbortExisting) {
		return entity.ErrOnMultipleExecInvalid
	}

	return nil
}

type TaskExecJobsParams struct {
	WorkspaceID string `json:"workspace_id"`
	TaskExecID  string `json:"task_exec_id"`
	Offset      int    `json:"offset"`
	Limit       int    `json:"limit"`
}

func (params *TaskExecJobsParams) FromRequest(r *http.Request) (err error) {

	params.WorkspaceID = r.FormValue("workspace_id")
	params.TaskExecID = r.FormValue("task_exec_id")
	// default limit
	limit := 5

	if r.FormValue("limit") != "" {
		limit, err = strconv.Atoi(r.FormValue("limit"))
		if err != nil {
			return eris.Errorf("limit is not valid: %v", err)
		}
	}

	if limit < 1 || limit > 100 {
		limit = 5
	}

	params.Limit = limit

	offset := 0

	if r.FormValue("offset") != "" {
		offset, err = strconv.Atoi(r.FormValue("offset"))
		if err != nil {
			return eris.Errorf("offset is not valid: %v", err)
		}
	}

	if offset < 0 {
		offset = 0
	}

	params.Offset = offset

	return nil
}

type TaskExecJobsResult struct {
	TaskExecJobs []*entity.TaskExecJob `json:"task_exec_jobs"`
	Total        int                   `json:"total"`
	Offset       int                   `json:"offset"`
	Limit        int                   `json:"limit"`
}

type TaskExecListResult struct {
	TaskExecs     []*entity.TaskExec `json:"task_execs"`
	NextToken     string             `json:"next_token,omitempty"`
	PreviousToken string             `json:"previous_token,omitempty"`
}

type TaskExecListParams struct {
	WorkspaceID   string  `json:"workspace_id"`
	Limit         int     `json:"limit"`
	NextToken     *string `json:"next_token,omitempty"`
	PreviousToken *string `json:"previous_token,omitempty"`
	// filters:
	TaskExecID      *string `json:"task_exec_id,omitempty"`
	AppID           *string `json:"app_id,omitempty"`
	TaskID          *string `json:"task_id,omitempty"`
	MultipleExecKey *string `json:"multiple_exec_key,omitempty"`
	Status          *int    `json:"status,omitempty"`
	// pagination computed server side:
	NextID       string
	NextDate     time.Time
	PreviousID   string
	PreviousDate time.Time
}

var (
	ErrTaskExecListNextInvalid     error = eris.New("task_exec list: next is not valid")
	ErrTaskExecListPreviousInvalid error = eris.New("task_exec list: previous is not valid")
	ErrTaskExecListLimitInvalid    error = eris.New("task_exec list: limit is not valid")
)

func (params *TaskExecListParams) FromRequest(r *http.Request) (err error) {

	params.WorkspaceID = r.FormValue("workspace_id")

	appID := r.FormValue("app_id")
	if appID != "" {
		params.AppID = &appID
	}

	taskExecID := r.FormValue("task_exec_id")
	if taskExecID != "" {
		params.TaskExecID = &taskExecID
	}

	// parse pagination token, either next_token or previous_token

	if r.FormValue("next_token") != "" {
		params.NextID, params.NextDate, err = DecodePaginationToken(r.FormValue("next_token"))
		if err != nil {
			return eris.Errorf("next_token err: %v", err)
		}
	} else if r.FormValue("previous_token") != "" {
		params.PreviousID, params.PreviousDate, err = DecodePaginationToken(r.FormValue("previous_token"))
		if err != nil {
			return eris.Errorf("previous_token err: %v", err)
		}
	}

	// default limit
	limit := 25

	if r.FormValue("limit") != "" {
		limit, err = strconv.Atoi(r.FormValue("limit"))
		if err != nil {
			return eris.Wrapf(ErrTaskExecListLimitInvalid, "err: %v", err)
		}
	}

	if limit < 1 || limit > 100 {
		return ErrTaskExecListLimitInvalid
	}

	params.Limit = limit

	// filters

	statusString := r.FormValue("status")

	if statusString != "" {
		// convert to int
		status, err := strconv.Atoi(statusString)
		if err != nil {
			return eris.Errorf("status is not valid: %v", err)
		}
		params.Status = &status
	}

	taskID := r.FormValue("task_id")

	if taskID != "" {
		taskFound := false

		for _, t := range entity.SystemTasks {
			if t.ID == taskID {
				taskFound = true
				break
			}
		}

		if !taskFound {
			return entity.ErrTaskInvalid
		}

		params.TaskID = &taskID
	}

	multipleExecKey := r.FormValue("multiple_exec_key")

	if multipleExecKey != "" {
		params.MultipleExecKey = &multipleExecKey
	}

	return nil
}

type TaskExecJobInfoParams struct {
	JobID       string `json:"id"`
	TaskExecID  string `json:"task_exec_id"`
	WorkspaceID string `json:"workspace_id"`
}

func (params *TaskExecJobInfoParams) FromRequest(r *http.Request) (err error) {

	params.WorkspaceID = r.FormValue("workspace_id")
	params.TaskExecID = r.FormValue("task_exec_id")
	params.JobID = r.FormValue("id")

	return nil
}

type TaskExecJobInfoInfo struct {
	ID            string                  `json:"id"`
	ScheduleTime  *time.Time              `json:"schedule_time,omitempty"`
	CreateTime    *time.Time              `json:"create_time,omitempty"`
	DispatchCount int32                   `json:"dispatch_count"`
	ResponseCount int32                   `json:"response_count"`
	FirstAttempt  *TaskExecJobInfoAttempt `json:"first_attempt,omitempty"`
	LastAttempt   *TaskExecJobInfoAttempt `json:"last_attempt,omitempty"`
}

type TaskExecJobInfoAttempt struct {
	ScheduleTime   *time.Time `json:"schedule_time,omitempty"`
	DispatchTime   *time.Time `json:"dispatch_time,omitempty"`
	ResponseTime   *time.Time `json:"response_time,omitempty"`
	ResponseCode   *int32     `json:"response_code,omitempty"`
	ReponseMessage *string    `json:"response_message,omitempty"`
}
