package dto

import (
	"net/http"
	"strconv"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

type MessageTemplateGetParams struct {
	WorkspaceID string `json:"workspace_id"`
	ID          string `json:"id"`
	Version     *int64 `json:"version,omitempty"`
}

func (params *MessageTemplateGetParams) FromRequest(r *http.Request) (err error) {

	params.WorkspaceID = r.FormValue("workspace_id")
	if params.WorkspaceID == "" {
		return eris.New("workspace_id is required")
	}

	params.ID = r.FormValue("id")
	if params.ID == "" {
		return eris.New("id is required")
	}

	// version
	version := r.FormValue("version")
	if version != "" {
		// convert to int
		i, err := strconv.Atoi(version)
		if err != nil {
			return eris.New("version is invalid")
		}
		i64 := int64(i)
		params.Version = &i64
	}

	return nil
}

type MessageTemplateListParams struct {
	WorkspaceID string  `json:"workspace_id"`
	Channel     *string `json:"channel"` // email | sms | push
	Category    *string `json:"category"`
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

	// category
	category := r.FormValue("category")
	if category != "" {
		if category != entity.MessageTemplateCategoryTransactional && category != entity.MessageTemplateCategoryCampaign && category != entity.MessageTemplateCategoryAutomation && category != entity.MessageTemplateCategoryOther {
			return eris.New("category is invalid")
		}
		params.Category = &category
	}

	return nil
}

type MessageTemplate struct {
	WorkspaceID     string                      `json:"workspace_id"`
	ID              string                      `json:"id"`
	Name            string                      `json:"name"`
	Channel         string                      `json:"channel"`  // email | sms | push
	Category        string                      `json:"category"` // email | sms | push
	Engine          string                      `json:"engine"`   // visual | mjml | nunchucks
	Email           entity.MessageTemplateEmail `json:"email"`
	TemplateMacroID *string                     `json:"template_macro_id,omitempty"`
	UTMSource       *string                     `json:"utm_source,omitempty"`
	UTMMedium       *string                     `json:"utm_medium,omitempty"`
	UTMCampaign     *string                     `json:"utm_campaign,omitempty"`
	Settings        entity.MapOfInterfaces      `json:"settings"`  // Channels specific 3rd-party settings
	TestData        string                      `json:"test_data"` // Test data for the template
}
