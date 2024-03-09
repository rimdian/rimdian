package entity

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

type TaskExecStatusType = int

var (
	// status from 0 to 100, to have room to add new steps in between
	TaskExecStatusAborted       TaskExecStatusType = -2 // aborted (non-retryable error...)
	TaskExecStatusRetryingError TaskExecStatusType = -1 // is retrying...
	TaskExecStatusProcessing    TaskExecStatusType = 0  // processing...
	TaskExecStatusDone          TaskExecStatusType = 1  // done (success)

	TaskMaxRetries           = 7
	TaskRetryDelayInSecs int = 20

	TaskExecIDDev    string = "todo"
	TaskExecJobIDDev string = "dev"

	ErrTaskExecAlreadyExists   = eris.New("task_exec already exists")
	ErrTaskExecNotFound        = eris.New("task_exec not found")
	ErrTaskExecIDRequired      = eris.New("task_exec id is required")
	ErrTaskInvalid             = eris.New("task is not valid")
	ErrTaskNameRequired        = eris.New("task name is required")
	ErrTaskKindNotImplemented  = eris.New("task kind is not implemented")
	ErrTaskPayloadRequired     = eris.New("task payload is required")
	ErrTaskIsRequired          = eris.New("task is required")
	ErrTaskWorkerIDNotAllowed  = eris.New("task_exec add worker id invalid")
	ErrTaskRetryCountExceeded  = eris.New("task_exec retry count exceeded")
	ErrTaskWorkerStateRequired = eris.New("task worker state is required")

	ErrTaskAlreadyRunningDiscardNew = eris.New("New task discarded, a similar task is already running.")
	ErrTaskAlreadyRunningWillRetry  = eris.New("A similar task is already running, will retry.")

	ErrTaskExecJobNotFound = eris.New("task_exec_job not found")
)

type TaskWorkerState map[string]interface{}

type TaskExecResult struct {
	IsDone                 bool                `json:"is_done"`
	UpdatedWorkerState     TaskWorkerState     `json:"updated_worker_state"`
	AppStateMutations      []*AppStateMutation `json:"app_state_mutations,omitempty"`
	ItemsToImport          []string            `json:"items_to_import,omitempty"`
	WorkerID               int                 `json:"worker_id"`
	IsError                bool                `json:"is_error"`
	Message                *string             `json:"message"`
	DelayNextRequestInSecs *int                `json:"delay_next_request_in_secs,omitempty"`
}

func (taskExecResult *TaskExecResult) SetError(message string, isDone bool) {
	taskExecResult.IsError = true
	taskExecResult.Message = &message
	taskExecResult.IsDone = isDone
}

// generate a new task result with default values, used by native apps workers
func NewTaskExecResult(currentWorkerState TaskWorkerState, workerID int) *TaskExecResult {
	return &TaskExecResult{
		UpdatedWorkerState: currentWorkerState, // by default, keep current state
		WorkerID:           workerID,
	}
}

type TaskExec struct {
	ID              string        `db:"id" json:"id"`
	TaskID          string        `db:"task_id" json:"task_id"`
	Name            string        `db:"name" json:"name"`
	MultipleExecKey *string       `db:"multiple_exec_key" json:"multiple_exec_key,omitempty"` // used to identify tasks that can run at the same time or not
	OnMultipleExec  string        `db:"on_multiple_exec" json:"on_multiple_exec"`             // allow, discard, retry, replace
	State           TaskExecState `db:"state" json:"state"`
	RetryCount      int           `db:"retry_count" json:"retry_count"` // global for all workers
	Message         *string       `db:"message" json:"message"`         // eventual error message
	// IsError         bool          `db:"is_error" json:"is_error"`       // true if the task is done with an error
	Status      int        `db:"status" json:"status"` // pending, running, done, error
	DBCreatedAt *time.Time `db:"db_created_at" json:"db_created_at"`
	DBUpdatedAt *time.Time `db:"db_updated_at" json:"db_updated_at"`
}

func (taskExec *TaskExec) IsPersisted() bool {
	return taskExec.DBCreatedAt != nil
}

func (taskExec *TaskExec) ResetStatus() {
	taskExec.Status = TaskExecStatusProcessing
	taskExec.Message = nil
}

func (taskExec *TaskExec) EnsureID() {
	if taskExec.ID == "" {
		taskExec.ID = uuid.New().String()
	}
}

func (task *TaskExec) Validate() error {
	if task.ID == "" {
		return ErrTaskExecIDRequired
	}

	if task.State.Workers == nil {
		task.State.Workers = map[int]TaskWorkerState{
			0: map[string]interface{}{},
		}
	}

	if !govalidator.IsIn(task.OnMultipleExec, OnMultipleExecDiscardNew, OnMultipleExecRetryLater, OnMultipleExecAbortExisting) {
		return ErrOnMultipleExecInvalid
	}

	return nil
}

type TaskExecState struct {
	// map[workerID]workerState
	Workers map[int]TaskWorkerState `json:"workers"`
}

func NewTaskState() TaskExecState {
	return TaskExecState{
		Workers: map[int]TaskWorkerState{
			0: map[string]interface{}{},
		},
	}
}

func (x *TaskExecState) Scan(val interface{}) error {

	var data []byte

	if b, ok := val.([]byte); ok {
		// VERY IMPORTANT: we need to clone the bytes here
		// The sql driver will reuse the same bytes RAM slots for future queries
		// Thank you St Antoine De Padoue for helping me find this bug
		data = bytes.Clone(b)
	} else if s, ok := val.(string); ok {
		data = []byte(s)
	} else if val == nil {
		return nil
	}

	return json.Unmarshal(data, x)
}

func (x TaskExecState) Value() (driver.Value, error) {
	v, err := json.Marshal(x)
	return v, err
}

// index on created_at for faster ORDER BY created_at
// created_at microsecs for pagination ordering
var TaskExecSchema string = `CREATE ROWSTORE TABLE IF NOT EXISTS task_exec (
	id VARCHAR(128) NOT NULL,
	task_id VARCHAR(50) NOT NULL,
	name VARCHAR(64) NOT NULL,
	multiple_exec_key VARCHAR(256),
	on_multiple_exec VARCHAR(20) NOT NULL DEFAULT 'discard',
	state JSON NOT NULL,
	retry_count SMALLINT NOT NULL DEFAULT 0,
	message VARCHAR(512) SPARSE,
	-- is_error BOOLEAN NOT NULL DEFAULT FALSE,
	status TINYINT NOT NULL,
	db_created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
	db_updated_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	
	PRIMARY KEY (id),
	KEY (db_created_at DESC),
    SHARD KEY (id)
  ) COLLATE utf8mb4_general_ci;
`
var TaskExecSchemaMYSQL string = `CREATE TABLE IF NOT EXISTS task_exec (
	id VARCHAR(128) NOT NULL,
	task_id VARCHAR(50) NOT NULL,
	name VARCHAR(64) NOT NULL,
	multiple_exec_key VARCHAR(256),
	on_multiple_exec VARCHAR(20) NOT NULL DEFAULT 'discard',
	state JSON NOT NULL,
	retry_count SMALLINT NOT NULL DEFAULT 0,
	message VARCHAR(512),
	-- is_error BOOLEAN NOT NULL DEFAULT FALSE,
	status TINYINT NOT NULL,
	db_created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
	db_updated_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
	
	PRIMARY KEY (id),
	KEY (db_created_at DESC)
    -- SHARD KEY (id)
  ) COLLATE utf8mb4_general_ci;
`
