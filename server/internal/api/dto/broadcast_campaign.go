package dto

import (
	"net/http"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

type BroadcastCampaignLaunchParams struct {
	WorkspaceID string `json:"workspace_id"`
	ID          string `json:"id"`
	// ScheduledAt *string `db:"scheduled_at" json:"scheduled_at,omitempty"`
	// Timezone    *string `db:"timezone" json:"timezone,omitempty"`
}

type BroadcastCampaignListParams struct {
	WorkspaceID string `json:"workspace_id"`
}

func (params *BroadcastCampaignListParams) FromRequest(r *http.Request) (err error) {

	params.WorkspaceID = r.FormValue("workspace_id")
	if params.WorkspaceID == "" {
		return eris.New("workspace_id is required")
	}

	return nil
}

type BroadcastCampaign struct {
	WorkspaceID string `json:"workspace_id"`
	*entity.BroadcastCampaign
}
