package dto

import (
	"net/http"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

type MessageTemplateListParams struct {
	WorkspaceID string  `json:"workspace_id"`
	Channel     *string `json:"channel"` // email | sms | push
}

func (params *MessageTemplateListParams) FromRequest(r *http.Request) (err error) {

	params.WorkspaceID = r.FormValue("workspace_id")
	if params.WorkspaceID == "" {
		return eris.New("workspace_id is required")
	}

	// channel
	channel := r.FormValue("channel")
	if channel != "" {
		if channel != "email" && channel != "sms" && channel != "push" {
			return eris.New("channel is invalid")
		}
		params.Channel = &channel
	}

	return nil
}

type MessageTemplate struct {
	WorkspaceID     string                      `json:"workspace_id"`
	ID              string                      `json:"id"`
	Name            string                      `json:"name"`
	Channel         string                      `json:"channel"` // email | sms | push
	Engine          string                      `json:"engine"`  // visual | mjml | nunchucks
	Email           entity.MessageTemplateEmail `json:"email"`
	TemplateMacroID *string                     `json:"template_macro_id,omitempty"`
	UTMSource       *string                     `json:"utm_source,omitempty"`
	UTMMedium       *string                     `json:"utm_medium,omitempty"`
	UTMCampaign     *string                     `json:"utm_campaign,omitempty"`
	UTMContent      *string                     `json:"utm_content,omitempty"`
	Settings        entity.MapOfInterfaces      `json:"settings"`  // Channels specific 3rd-party settings
	TestData        string                      `json:"test_data"` // Test data for the template
}