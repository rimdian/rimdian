package entity

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/rotisserie/eris"
)

var (
	ErrTaskIDRequired                = eris.New("task id is required")
	ErrTaskExternalIDRequired        = eris.New("task external id is required")
	ErrAppIDRequired                 = eris.New("app id is required")
	ErrScheduledTaskIntervalTooSmall = eris.New("task interval should be >= 10 minutes")
	ErrOnMultipleExecInvalid         = eris.New("on_multiple_exec is not valid")

	TaskNameGenerateDemo                  string = "Generate demo"
	TaskKindTestingNotDone                string = "system_testing_task_not_done"
	TaskKindTestingDone                   string = "system_testing_task_done"
	TaskKindTestingTimeout                string = "system_testing_task_timeout"
	TaskKindTestingPanic                  string = "system_testing_task_panic"
	TaskKindGenerateDemo                  string = "system_generate_demo"
	TaskKindDataLogReprocessUntil         string = "system_data_log_reprocess_until"
	TaskKindReattributeConversions        string = "system_reattribute_conversions"
	TaskKindImportUsersToSubscriptionList string = "system_import_users_to_subscription_list"
	TaskKindUpgradeApp                    string = "system_upgrade_app"
	TaskKindRecomputeSegment                     = "system_recompute_segment"
	TaskKindRefreshOutdatedSegments              = "system_refresh_outdated_segments"

	SystemTasks = []TaskManifest{
		{

			ID:             TaskKindGenerateDemo,
			Name:           "Generate demo",
			IsCron:         false,
			OnMultipleExec: OnMultipleExecDiscardNew,
		},
		{
			ID:             TaskKindDataLogReprocessUntil,
			Name:           "Reprocess data log until",
			IsCron:         false,
			OnMultipleExec: OnMultipleExecDiscardNew,
		},
		{
			ID:             TaskKindReattributeConversions,
			Name:           "Reattribute conversions",
			IsCron:         false,
			OnMultipleExec: OnMultipleExecAbortExisting,
		},
		{
			ID:             TaskKindImportUsersToSubscriptionList,
			Name:           "Import users to subscription list",
			IsCron:         false,
			OnMultipleExec: OnMultipleExecDiscardNew,
		},
		{
			ID:             TaskKindUpgradeApp,
			Name:           "Upgrade app",
			IsCron:         false,
			OnMultipleExec: OnMultipleExecDiscardNew,
		},
		{
			ID:             TaskKindRecomputeSegment,
			Name:           "Recompute segment",
			IsCron:         false,
			OnMultipleExec: OnMultipleExecAbortExisting,
		},
		{
			ID:              TaskKindRefreshOutdatedSegments,
			Name:            "Refresh outdated segments",
			IsCron:          true,
			OnMultipleExec:  OnMultipleExecDiscardNew,
			MinutesInterval: 60 * 12, // every 12 hours
		},
	}

	OnMultipleExecAllow         = "allow"          // mutliple tasks of the same kind are allowed to run at the same time
	OnMultipleExecDiscardNew    = "discard_new"    // discard the new task if one is already running
	OnMultipleExecRetryLater    = "retry_later"    // retry the new task until the current one is done
	OnMultipleExecAbortExisting = "abort_existing" // end the current task and start the new one

)

type TaskManifest struct {
	ID              string `db:"id" json:"id"`
	Name            string `db:"name" json:"name"`
	IsCron          bool   `db:"is_cron" json:"is_cron"`
	OnMultipleExec  string `db:"on_multiple_exec" json:"on_multiple_exec"` // 'allow' | 'discard_new' | 'retry_later' | 'abort_existing'
	MinutesInterval int64  `db:"minutes_interval" json:"minutes_interval"`
}

type TasksManifest []*TaskManifest

func (x *TasksManifest) Scan(val interface{}) error {

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

func (x TasksManifest) Value() (driver.Value, error) {
	return json.Marshal(x)
}

type Task struct {
	TaskManifest
	WorkspaceID string     `db:"workspace_id" json:"workspace_id"`
	AppID       string     `db:"app_id" json:"app_id"`
	IsActive    bool       `db:"is_active" json:"is_active"`
	IsCron      bool       `db:"is_cron" json:"is_cron"`
	NextRun     *time.Time `db:"next_run" json:"next_run,omitempty"`
	LastRun     *time.Time `db:"last_run" json:"last_run,omitempty"`
	// LastError *string    `db:"last_error" json:"last_error,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

func (x *Task) IsPersisted() bool {
	if x.CreatedAt.IsZero() {
		return false
	}
	return true
}

func (x *Task) ShouldRun() bool {
	if !x.IsActive || x.NextRun == nil {
		return false
	}
	return time.Now().After(*x.NextRun)
}

func (x *Task) ComputeNextRun() {
	now := time.Now()
	if x.NextRun == nil {
		x.NextRun = &now
	}
	nextRun := now.Add(time.Minute * time.Duration(x.MinutesInterval))
	x.NextRun = &nextRun
}

func (x *Task) Validate() error {

	// should have an id
	if x.ID == "" {
		return ErrTaskExternalIDRequired
	}
	// should have a workspace id
	if x.WorkspaceID == "" {
		return ErrWorkspaceIDRequired
	}

	// should have a name
	if x.Name == "" {
		return ErrTaskNameRequired
	}

	// should have an app id
	if x.AppID == "" {
		return ErrAppIDRequired
	}

	if x.IsCron {
		// should have an interval greater than 10
		if x.MinutesInterval < 10 {
			return ErrScheduledTaskIntervalTooSmall
		}
	}

	if !govalidator.IsIn(x.OnMultipleExec, OnMultipleExecDiscardNew, OnMultipleExecRetryLater, OnMultipleExecAbortExisting) {
		return ErrOnMultipleExecInvalid
	}

	return nil
}

type Tasks []*Task

func (x *Tasks) Scan(val interface{}) error {

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

func (x Tasks) Value() (driver.Value, error) {
	return json.Marshal(x)
}

var TaskSchema string = `CREATE ROWSTORE TABLE IF NOT EXISTS task (
	id VARCHAR(128) NOT NULL,
	workspace_id VARCHAR(128) NOT NULL,
	name VARCHAR(64) NOT NULL,
	on_multiple_exec VARCHAR(20) NOT NULL DEFAULT 'discard',
	app_id VARCHAR(64) NOT NULL,
	is_active BOOLEAN NOT NULL,
	is_cron BOOLEAN NOT NULL,
	minutes_interval INT NOT NULL,
	next_run DATETIME,
	last_run DATETIME,
	created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	
	PRIMARY KEY (workspace_id, id),
	KEY (next_run ASC),
    SHARD KEY (workspace_id)
) COLLATE utf8mb4_general_ci;`

var TaskSchemaMYSQL string = `CREATE TABLE IF NOT EXISTS task (
	id VARCHAR(128) NOT NULL,
	workspace_id VARCHAR(128) NOT NULL,
	name VARCHAR(64) NOT NULL,
	on_multiple_exec VARCHAR(20) NOT NULL DEFAULT 'discard',
	app_id VARCHAR(64) NOT NULL,
	is_active BOOLEAN NOT NULL,
	is_cron BOOLEAN NOT NULL,
	minutes_interval INT NOT NULL,
	next_run DATETIME,
	last_run DATETIME,
	created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	
	PRIMARY KEY (workspace_id, id),
	KEY (next_run ASC)
    -- SHARD KEY (id)
) COLLATE utf8mb4_general_ci;`
