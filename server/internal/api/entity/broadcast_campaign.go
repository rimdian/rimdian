package entity

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/rimdian/rimdian/internal/api/common"
	"github.com/rotisserie/eris"
)

var (
	BroadcastCampaignStatusDraft     string = "draft"
	BroadcastCampaignStatusScheduled string = "scheduled"
	BroadcastCampaignStatusLaunched  string = "launched"
	BroadcastCampaignStatusSent      string = "sent"
	BroadcastCampaignStatusFailed    string = "failed"

	TaskKindLaunchBroadcastCampaign = "launch_broadcast_campaign"
)

type BroadcastCampaign struct {
	ID                string                             `db:"id" json:"id"` // ID is the utm_campaign value
	Name              string                             `db:"name" json:"name"`
	Channel           string                             `db:"channel" json:"channel"`                     // email | sms | push
	MessageTemplates  BroadcastCampaignMessageTemplates  `db:"message_templates" json:"message_templates"` // templates for A/B testing
	Status            string                             `db:"status" json:"status"`                       // draft | sending | sent | failed
	SubscriptionLists BroadcastCampaignSubscriptionLists `db:"subscription_lists" json:"subscription_lists"`
	UTMSource         string                             `db:"utm_source" json:"utm_source"`
	UTMMedium         string                             `db:"utm_medium" json:"utm_medium"`
	ScheduledAt       *string                            `db:"scheduled_at" json:"scheduled_at,omitempty"`
	Timezone          *string                            `db:"timezone" json:"timezone,omitempty"`
	LaunchedAt        *time.Time                         `db:"launched_at" json:"launched_at,omitempty"`
	DBCreatedAt       time.Time                          `db:"db_created_at" json:"db_created_at"`
	DBUpdatedAt       time.Time                          `db:"db_updated_at" json:"db_updated_at"`
}

func (bc *BroadcastCampaign) LaunchNow() {
	now := time.Now()
	bc.ScheduledAt = nil
	bc.LaunchedAt = &now
	bc.Status = BroadcastCampaignStatusLaunched
}

func (bc *BroadcastCampaign) GetScheduledAt() (*time.Time, error) {
	if bc.ScheduledAt == nil {
		return nil, eris.New("ScheduledAt is nil")
	}
	// parse date in location
	location, err := time.LoadLocation(*bc.Timezone)
	if err != nil {
		return nil, eris.Wrap(err, "Failed to load timezone")
	}

	datetime, err := time.ParseInLocation("2006-01-02 15:04", *bc.ScheduledAt, location)

	if err != nil {
		return nil, eris.Wrap(err, "Failed to parse datetime")
	}

	return &datetime, nil
}

func (bc *BroadcastCampaign) Validate() error {
	if bc.ID == "" {
		return eris.New("ID is required")
	}

	if bc.Name == "" {
		return eris.New("Name is required")
	}

	// only email channel is supported for now
	if bc.Channel != "email" {
		return eris.New("Channel is invalid")
	}

	if bc.MessageTemplates == nil || len(bc.MessageTemplates) == 0 {
		return eris.New("Templates is required")
	}

	if bc.SubscriptionLists == nil || len(bc.SubscriptionLists) == 0 {
		return eris.New("SubscriptionLists is required")
	}

	if bc.UTMSource == "" {
		return eris.New("UTMSource is required")
	}

	if bc.UTMMedium == "" {
		return eris.New("UTMMedium is required")
	}

	if bc.ScheduledAt != nil {

		if bc.Timezone != nil && !govalidator.IsIn(*bc.Timezone, common.Timezones...) {
			return eris.New("timezone is invalid")
		}

		// parse date in location
		location, err := time.LoadLocation(*bc.Timezone)
		if err != nil {
			return eris.Wrap(err, "Failed to load timezone")
		}
		datetime, err := time.ParseInLocation("2006-01-02 15:04", *bc.ScheduledAt, location)

		if err != nil {
			return eris.Wrap(err, "Failed to parse datetime")
		}

		if datetime.Before(time.Now()) {
			return eris.New("scheduled_at should be in the future")
		}
	}

	if bc.Status == "" {
		bc.Status = BroadcastCampaignStatusDraft
	}

	if bc.Status == BroadcastCampaignStatusDraft && bc.ScheduledAt != nil {
		bc.Status = BroadcastCampaignStatusScheduled
	}

	// switch back to draft if scheduled_at is removed
	if bc.ScheduledAt == nil && bc.Status == BroadcastCampaignStatusScheduled {
		bc.Status = BroadcastCampaignStatusDraft
	}

	if !govalidator.IsIn(bc.Status, BroadcastCampaignStatusDraft, BroadcastCampaignStatusScheduled, BroadcastCampaignStatusLaunched, BroadcastCampaignStatusSent, BroadcastCampaignStatusFailed) {
		return eris.New("Status is invalid")
	}

	return nil
}

type BroadcastCampaignMessageTemplates []*BroadcastCampaignMessageTemplate

func (x *BroadcastCampaignMessageTemplates) Scan(val interface{}) error {

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

func (x BroadcastCampaignMessageTemplates) Value() (driver.Value, error) {
	return json.Marshal(x)
}

type BroadcastCampaignMessageTemplate struct {
	ID string `json:"id"`
	// percentage for A/B testing
	Percentage int `json:"percentage"`
}

func (bct *BroadcastCampaignMessageTemplate) Validate() error {
	if bct.ID == "" {
		return eris.New("TemplateID is required")
	}

	if bct.Percentage < 0 || bct.Percentage > 100 {
		return eris.New("Percentage is invalid")
	}

	return nil
}

type BroadcastCampaignSubscriptionLists []*BroadcastCampaignSubscriptionList

func (x *BroadcastCampaignSubscriptionLists) Scan(val interface{}) error {

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

func (x BroadcastCampaignSubscriptionLists) Value() (driver.Value, error) {
	return json.Marshal(x)
}

type BroadcastCampaignSubscriptionList struct {
	ID string `json:"id"`
}

var BroadcastCampaignSchema = `CREATE ROWSTORE TABLE IF NOT EXISTS broadcast_campaign (
	id VARCHAR(64) NOT NULL,
	name VARCHAR(255) NOT NULL,
	channel VARCHAR(255) NOT NULL,
	message_templates JSON NOT NULL,
	status VARCHAR(24) NOT NULL,
	subscription_lists JSON NOT NULL,
	utm_source VARCHAR(255) NOT NULL,
	utm_medium VARCHAR(255) NOT NULL,
	scheduled_at VARCHAR(24),
	timezone VARCHAR(255),
	launched_at DATETIME,
	db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

	PRIMARY KEY (id)
) COLLATE utf8mb4_general_ci;`

var BroadcastCampaignSchemaMYSQL = `CREATE TABLE IF NOT EXISTS broadcast_campaign (
	id VARCHAR(64) NOT NULL,
	name VARCHAR(255) NOT NULL,
	channel VARCHAR(255) NOT NULL,
	message_templates JSON NOT NULL,
	status VARCHAR(24) NOT NULL,
	subscription_lists JSON NOT NULL,
	utm_source VARCHAR(255) NOT NULL,
	utm_medium VARCHAR(255) NOT NULL,
	scheduled_at VARCHAR(24),
	timezone VARCHAR(255),
	launched_at DATETIME,
	db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

	PRIMARY KEY (id)
) COLLATE utf8mb4_general_ci;`
