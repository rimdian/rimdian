package entity

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/rotisserie/eris"
)

var (
	MessageTemplateChannelEmail = "email"
	MessageTemplateChannelSMS   = "sms"
	MessageTemplateChannelPush  = "push"

	MessageTemplateEngineVisual    = "visual"
	MessageTemplateEngineMJML      = "mjml"
	MessageTemplateEngineNunchucks = "nunchucks"
)

type MessageTemplate struct {
	ID              string               `json:"id"`
	Version         int                  `json:"version"`
	Name            string               `json:"name"`
	Channel         string               `json:"channel"` // email | sms | push
	Engine          string               `json:"engine"`  // visual | mjml | nunchucks
	Email           MessageTemplateEmail `json:"email"`
	TemplateMacroID *string              `json:"template_macro_id,omitempty"`
	UTMSource       *string              `json:"utm_source,omitempty"`
	UTMMedium       *string              `json:"utm_medium,omitempty"`
	UTMCampaign     *string              `json:"utm_campaign,omitempty"`
	UTMContent      *string              `json:"utm_content,omitempty"`
	Settings        MapOfInterfaces      `json:"settings"`  // Channels specific 3rd-party settings
	TestData        string               `json:"test_data"` // Test data for the template
	DBCreatedAt     time.Time            `json:"db_created_at"`
	DBUpdatedAt     time.Time            `json:"db_updated_at"`
}

func (e *MessageTemplate) Validate() (err error) {
	// trim spaces
	e.ID = strings.TrimSpace(e.ID)
	e.Name = strings.TrimSpace(e.Name)

	// validate required fields
	if e.ID == "" {
		return eris.New("id is required")
	}

	if e.Name == "" {
		return eris.New("name is required")
	}

	if e.Engine != MessageTemplateEngineVisual && e.Engine != MessageTemplateEngineMJML && e.Engine != MessageTemplateEngineNunchucks {
		return eris.New("engine is invalid")
	}

	if e.Channel != MessageTemplateChannelEmail && e.Channel != MessageTemplateChannelSMS && e.Channel != MessageTemplateChannelPush {
		return eris.New("channel is invalid")
	}

	if e.Channel == MessageTemplateChannelEmail {
		if err := e.Email.Validate(); err != nil {
			return err
		}
	}

	if e.Settings == nil {
		e.Settings = MapOfInterfaces{}
	}

	return nil
}

type MessageTemplateEmail struct {
	FromAdrress      string          `json:"from_address"`
	FromName         string          `json:"from_name"`
	ReplyTo          *string         `json:"reply_to,omitempty"`
	Subject          string          `json:"subject"`
	Content          string          `json:"content"` // html | mjml code | nunjucks code
	VisualEditorTree MapOfInterfaces `json:"visual_editor_tree"`
	Text             *string         `json:"text,omitempty"`
}

func (e *MessageTemplateEmail) Validate() (err error) {
	// trim spaces
	e.FromAdrress = strings.TrimSpace(e.FromAdrress)
	e.FromName = strings.TrimSpace(e.FromName)
	e.Subject = strings.TrimSpace(e.Subject)

	// validate required fields
	if e.FromAdrress == "" {
		return eris.New("from_address is required")
	}

	if !govalidator.IsEmail(e.FromAdrress) {
		return eris.New("from_address is not an email address")
	}

	if e.ReplyTo != nil && !govalidator.IsEmail(*e.ReplyTo) {
		return eris.New("reply_to is not an email address")
	}

	if e.FromName == "" {
		return eris.New("from_name is required")
	}

	if e.Subject == "" {
		return eris.New("subject is required")
	}

	if e.Content == "" {
		return eris.New("content is required")
	}

	if e.VisualEditorTree == nil {
		e.VisualEditorTree = MapOfInterfaces{}
	}

	return nil
}

// scan
func (x *MessageTemplateEmail) Scan(val interface{}) error {

	var data []byte

	if b, ok := val.([]byte); ok {
		// VERY IMPORTANT: we need to clone the bytes here
		// The sql driver will reuse the same bytes RAM slots for future queries
		// Thank you St Antoine De Padoue for helping me find this bug
		data = bytes.Clone(b)
	} else if s, ok := val.(string); ok {
		data = []byte(s)
	} else if val == nil {
		return nil
	}

	return json.Unmarshal(data, x)
}

func (x MessageTemplateEmail) Value() (driver.Value, error) {
	return json.Marshal(x)
}

var MessageTemplateSchema string = `CREATE ROWSTORE TABLE IF NOT EXISTS message_template (
	id VARCHAR(64) NOT NULL,
	version INT NOT NULL,
	name VARCHAR(255) NOT NULL,
	channel VARCHAR(255) NOT NULL,
	engine VARCHAR(255) NOT NULL,
	email JSON NOT NULL,
	template_macro_id VARCHAR(64),
	utm_source VARCHAR(255),
	utm_medium VARCHAR(255),
	utm_campaign VARCHAR(255),
	utm_content VARCHAR(255),
	settings JSON NOT NULL,
	test_data TEXT,
	db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	  
	PRIMARY KEY (id, version),
	SHARD KEY (id)
) COLLATE utf8mb4_general_ci;`

var MessageTemplateSchemaMYSQL string = `CREATE TABLE IF NOT EXISTS message_template (
	id VARCHAR(64) NOT NULL,
	version INT NOT NULL,
	name VARCHAR(255) NOT NULL,
	channel VARCHAR(255) NOT NULL,
	engine VARCHAR(255) NOT NULL,
	email JSON NOT NULL,
	template_macro_id VARCHAR(64),
	utm_source VARCHAR(255),
	utm_medium VARCHAR(255),
	utm_campaign VARCHAR(255),
	utm_content VARCHAR(255),
	settings JSON NOT NULL,
	test_data TEXT,
	db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

	PRIMARY KEY (id, version)
) COLLATE utf8mb4_general_ci;`
