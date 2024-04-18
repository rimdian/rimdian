package entity

import (
	"time"

	"github.com/rimdian/rimdian/internal/api/common"
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
	Signature   string    `json:"signature"` // hmac of the task exec id
}

func NewScheduledTask(workspaceID string, taskExec TaskExec, scheduledAt time.Time) ScheduledTask {
	return ScheduledTask{
		WorkspaceID: workspaceID,
		TaskExec:    taskExec,
		ScheduledAt: scheduledAt,
	}
}

func (st *ScheduledTask) Sign(secretKey string) string {
	return common.ComputeHMAC256([]byte(st.TaskExec.ID), secretKey)
}

func (st *ScheduledTask) VerifySignature(secretKey string) bool {
	return st.Signature == st.Sign(secretKey)
}

func (st *ScheduledTask) Validate(secretKey string) error {
	if st.WorkspaceID == "" {
		return eris.New("WorkspaceID is required")
	}

	if err := st.TaskExec.Validate(); err != nil {
		return eris.Wrap(err, "TaskExec")
	}

	if st.ScheduledAt.IsZero() {
		return eris.New("ScheduledAt is required")
	}

	// scheduled mat max in 30 days
	if st.ScheduledAt.After(time.Now().Add(30 * 24 * time.Hour)) {
		return eris.New("ScheduledAt must be before 30 days in the future")
	}

	if !st.VerifySignature(secretKey) {
		return eris.New("Signature is invalid")
	}

	return nil
}
