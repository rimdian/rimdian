package dto

import (
	"net/http"

	"github.com/rotisserie/eris"
)

type MessageTemplateListParams struct {
	WorkspaceID string `json:"workspace_id"`
	Channel     string `json:"channel"` // email | sms | push
}

func (params *MessageTemplateListParams) FromRequest(r *http.Request) (err error) {

	params.WorkspaceID = r.FormValue("workspace_id")
	if params.WorkspaceID == "" {
		return eris.New("workspace_id is required")
	}

	// channel
	params.Channel = r.FormValue("channel")
	if params.Channel == "" {
		return eris.New("channel is required")
	}
	if params.Channel != "email" && params.Channel != "sms" && params.Channel != "push" {
		return eris.New("channel is invalid")
	}

	return nil
}
