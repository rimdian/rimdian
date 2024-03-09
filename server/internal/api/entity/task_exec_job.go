package entity

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// tracking of Google Cloud Task queue running jobs
type TaskExecJob struct {
	ID          string     `db:"id" json:"id"`
	TaskExecID  string     `db:"task_exec_id" json:"task_exec_id"`
	DoneAt      *time.Time `db:"done_at" json:"done_at,omitempty"`
	DBCreatedAt time.Time  `db:"db_created_at" json:"db_created_at"`
	DBUpdatedAt time.Time  `db:"db_updated_at" json:"db_updated_at"`
}

func NewTaskExecJob(taskID string) *TaskExecJob {
	return &TaskExecJob{
		ID: fmt.Sprintf("%v_%v", uuid.New().String(), taskID), // random id first, for Google Taske Queue performance sharding
	}
}

// index on created_at for faster ORDER BY created_at
// created_at microsecs for pagination ordering
var TaskExecJobSchema string = `CREATE TABLE IF NOT EXISTS task_exec_job (
	id VARCHAR(128) NOT NULL,
	task_exec_id VARCHAR(128) NOT NULL,
	done_at DATETIME,
	db_created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
	db_updated_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	
	SORT KEY (task_exec_id),
	PRIMARY KEY (id, task_exec_id),
	SHARD KEY (task_exec_id)
  ) COLLATE utf8mb4_general_ci;
`
var TaskExecJobSchemaMYSQL string = `CREATE TABLE IF NOT EXISTS task_exec_job (
	id VARCHAR(128) NOT NULL,
	task_exec_id VARCHAR(128) NOT NULL,
	done_at DATETIME,
	db_created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
	db_updated_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
	
	PRIMARY KEY (id),
	KEY (task_exec_id)
  ) COLLATE utf8mb4_general_ci;
`
