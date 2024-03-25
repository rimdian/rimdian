package entity

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/rotisserie/eris"
)

var (
	BroadcastCampaignStatusDraft   string = "draft"
	BroadcastCampaignStatusSending string = "sending"
	BroadcastCampaignStatusSent    string = "sent"
	BroadcastCampaignStatusFailed  string = "failed"
)

type BroadcastCampaign struct {
	ID                string                             `json:"id"` // ID is the utm_campaign value
	Name              string                             `json:"name"`
	Channel           string                             `json:"channel"`   // email | sms | push
	Templates         BroadcastCampaignTemplates         `json:"templates"` // templates for A/B testing
	Status            string                             `json:"status"`    // draft | sending | sent | failed
	SubscriptionLists BroadcastCampaignSubscriptionLists `json:"subscription_lists"`
	UTMSource         string                             `json:"utm_source"`
	UTMMedium         string                             `json:"utm_medium"`
	DBCreatedAt       time.Time                          `json:"db_created_at"`
	DBUpdatedAt       time.Time                          `json:"db_updated_at"`
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

	if bc.Templates == nil || len(bc.Templates) == 0 {
		return eris.New("Templates is required")
	}

	if !govalidator.IsIn(bc.Status, BroadcastCampaignStatusDraft, BroadcastCampaignStatusSending, BroadcastCampaignStatusSent, BroadcastCampaignStatusFailed) {
		return eris.New("Status is invalid")
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

	return nil
}

type BroadcastCampaignTemplates []*BroadcastCampaignTemplate

func (x *BroadcastCampaignTemplates) Scan(val interface{}) error {

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

func (x BroadcastCampaignTemplates) Value() (driver.Value, error) {
	return json.Marshal(x)
}

type BroadcastCampaignTemplate struct {
	TemplateID string `json:"template_id"`
	// percentage for A/B testing
	Percentage int `json:"percentage"`
}

func (bct *BroadcastCampaignTemplate) Validate() error {
	if bct.TemplateID == "" {
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

var BroadcastCampaignSchema = `CREATE ROWSTORE TABLE IF NOT EXISTS broadcast_campaigns (
	id VARCHAR(64) NOT NULL,
	name VARCHAR(255) NOT NULL,
	channel VARCHAR(255) NOT NULL,
	templates JSON NOT NULL,
	status VARCHAR(255) NOT NULL,
	subscription_list_id VARCHAR(64) NOT NULL,
	utm_source VARCHAR(255) NOT NULL,
	utm_medium VARCHAR(255) NOT NULL,
	db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

	PRIMARY KEY (id)
) COLLATE utf8mb4_general_ci;`

var BroadcastCampaignSchemaMYSQL = `CREATE TABLE IF NOT EXISTS broadcast_campaigns (
	id VARCHAR(64) NOT NULL,
	name VARCHAR(255) NOT NULL,
	channel VARCHAR(255) NOT NULL,
	templates JSON NOT NULL,
	status VARCHAR(255) NOT NULL,
	subscription_list_id VARCHAR(64) NOT NULL,
	utm_source VARCHAR(255) NOT NULL,
	utm_medium VARCHAR(255) NOT NULL,
	db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

	PRIMARY KEY (id)
) COLLATE utf8mb4_general_ci;`
