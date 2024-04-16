package entity

import (
	"time"

	"github.com/rotisserie/eris"
)

// a scheduled task is a JSON payload waiting in a Google Task Queue
// that will be pushed to the API when the time comes

var (
	ScheduledTasksQueueName = "scheduled-tasks"
	ScheduledTaskEndpoint   = "/api/scheduledTask.do"
)

type ScheduledTask struct {
	WorkspaceID string    `json:"workspace_id"`
	TaskExec    TaskExec  `json:"task_exec"` // the task to execute
	ScheduledAt time.Time `json:"scheduled_at"`
}

func NewScheduledTask(workspaceID string, taskExec TaskExec) ScheduledTask {
	return ScheduledTask{
		WorkspaceID: workspaceID,
		TaskExec:    taskExec,
	}
}

func (st *ScheduledTask) Validate() error {
	if st.WorkspaceID == "" {
		return eris.New("WorkspaceID is required")
	}

	if err := st.TaskExec.Validate(); err != nil {
		return eris.Wrap(err, "TaskExec")
	}

	// scheduled mat max in 30 days
	if st.ScheduledAt.Before(time.Now().Add(30 * 24 * time.Hour)) {
		return eris.New("ScheduledAt must be at least 30 days in the future")
	}

	return nil
}
