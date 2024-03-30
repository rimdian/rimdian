package dto

import (
	"net/http"
	"time"

	"github.com/rimdian/rimdian/internal/api/entity"
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

type SubscriptionListCreate struct {
	WorkspaceID string `json:"workspace_id"`
	*entity.SubscriptionList
}

type UserToImportToSubscriptionList struct {
	ExternalID      string    `db:"external_id" json:"external_id"`
	IsAuthenticated bool      `db:"is_authenticated" json:"is_authenticated"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
}
