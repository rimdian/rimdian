package entity

import (
	"time"
)

var (
	AppWebhookKindDataHook = "data_hook"
	AppWebhookKindTaskExec = "task_exec_worker"

	RejectItem = "reject"
	UpdateItem = "update"
)

// payload sent to apps
type AppWebhookPayload struct {
	APIEndpoint       string           `json:"api_endpoint"`
	CollectorEndpoint string           `json:"collector_endpoint"`
	WorkspaceID       string           `json:"workspace_id"`
	AppID             string           `json:"app_id"`
	AppState          MapOfInterfaces  `json:"app_state,omitempty"`
	Kind              string           `json:"kind"` // data_hook | task_exec_worker
	TaskExecWorker    *TaskExecWorker  `json:"task_exec_worker,omitempty"`
	DataHook          *DataHookPayload `json:"data_hook,omitempty"`
}

type TaskExecWorker struct {
	TaskID            string          `json:"task_id"`
	TaskName          string          `json:"task_name"`
	TaskExecID        string          `json:"task_exec_id"`
	TaskExecCreatedAt time.Time       `json:"task_exec_created_at"`
	WorkerID          int             `json:"worker_id"` // 0 = parent thread
	WorkerState       TaskWorkerState `json:"worker_state"`
	RetryCount        int             `json:"retry_count"`
}

type DataHookPayload struct {
	DataHookID   string `json:"data_hook_id"`
	DataHookName string `json:"data_hook_name"`
	DataHookOn   string `json:"data_hook_on"` // on_validation, on_success
	DataLogID    string `json:"data_log_id"`
	DataLogItem  string `json:"data_log_item"` // raw json string of the item
	DataLogKind  string `json:"data_log_kind"` // order, segment, user...
	// fields provided for "on_success" hooks:
	DataLogAction         *string       `json:"data_log_action,omitempty"`
	DataLogItemID         *string       `json:"data_log_item_id,omitempty"`
	DataLogItemExternalID *string       `json:"data_log_item_external_id,omitempty"`
	DataLogUpdatedFields  UpdatedFields `json:"data_log_updated_fields,omitempty"`
	User                  *User         `json:"user,omitempty"` // user object if user_id is not none
}

type DataHookOnvalidationResponse struct {
	Action      string `json:"action"`       // reject | update (empty = accept)
	Message     string `json:"message"`      // reason for reject or update
	UpdatedItem string `json:"updated_item"` // order, segment, user...
}
