package dto

import (
	"net/http"

	"github.com/rotisserie/eris"
)

type SubscriptionListListParams struct {
	WorkspaceID    string `json:"workspace_id"`
	WithUsersCount bool   `json:"with_users_count,omitempty"`
}

func (params *SubscriptionListListParams) FromRequest(r *http.Request) (err error) {

	params.WorkspaceID = r.FormValue("workspace_id")
	if params.WorkspaceID == "" {
		return eris.New("workspace_id is required")
	}

	if r.FormValue("with_users_count") == "true" {
		params.WithUsersCount = true
	}

	return nil
}
