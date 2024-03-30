package dto

import (
	"net/http"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

type TaskListResult struct {
	Tasks []*entity.Task `json:"tasks"`
}

type TaskListParams struct {
	WorkspaceID string `json:"workspace_id"`
}

var (
	ErrTaskListWorkspaceIDRequired error = eris.New("task list: workspace_id is required")
)

func (params *TaskListParams) FromRequest(r *http.Request) (err error) {

	params.WorkspaceID = r.FormValue("workspace_id")
	if params.WorkspaceID == "" {
		return ErrTaskListWorkspaceIDRequired
	}

	return nil
}

type TaskRunParams struct {
	ID              string                 `json:"id"`
	WorkspaceID     string                 `json:"workspace_id"`
	MainWorkerState entity.TaskWorkerState `json:"main_worker_state,omitempty"`
	MultipleExecKey *string                `json:"multiple_exec_key,omitempty"`
}
