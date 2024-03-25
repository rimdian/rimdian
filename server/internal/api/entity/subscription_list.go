package entity

import (
	"time"

	"github.com/rotisserie/eris"
)

type SubscriptionList struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Color           string    `json:"color"`
	Channel         string    `json:"channel"` // email | sms | push
	DoubleOptIn     bool      `json:"double_opt_in"`
	EmailTemplateID *string   `json:"email_template_id,omitempty"`
	DBCreatedAt     time.Time `json:"db_created_at"`
	DBUpdatedAt     time.Time `json:"db_updated_at"`

	// joined server side
	UsersCount int64 `json:"users_count,omitempty"`
}

func (sl *SubscriptionList) Validate() error {
	if sl.ID == "" {
		return eris.New("id is required")
	}

	if sl.Name == "" {
		return eris.New("name is required")
	}

	if sl.Color == "" {
		return eris.New("color is required")
	}

	if sl.Channel == "" {
		return eris.New("channel is required")
	}

	// only email channel is supported for now
	if sl.Channel != "email" {
		return eris.New("channel is invalid")
	}

	if sl.DoubleOptIn {

		if sl.Channel == "email" {
			if sl.EmailTemplateID == nil {
				return eris.New("template_id is required")
			}
		}
	}

	return nil
}

type SubscriptionListUser struct {
	SubscriptionListID string    `json:"subscription_list_id"`
	UserID             string    `json:"user_id"`
	DBCreatedAt        time.Time `json:"db_created_at"`
}

// schema
var SubscriptionListSchema = `CREATE ROWSTORE TABLE IF NOT EXISTS subscription_list (
	id VARCHAR(64) NOT NULL,
	name VARCHAR(255) NOT NULL,
	color VARCHAR(32) NOT NULL,
	channel VARCHAR(32) NOT NULL,
	double_opt_in BOOLEAN NOT NULL,
	template_id VARCHAR(64),
	db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (id),
	SHARD KEY (id)
)`

var SubscriptionListSchemaMYSQL = `CREATE TABLE IF NOT EXISTS subscription_list (
	id VARCHAR(64) NOT NULL,
	name VARCHAR(255) NOT NULL,
	color VARCHAR(32) NOT NULL,
	channel VARCHAR(32) NOT NULL,
	double_opt_in BOOLEAN NOT NULL,
	template_id VARCHAR(64),
	db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (id)
)`

var SubscriptionListUserSchema = `CREATE ROWSTORE TABLE IF NOT EXISTS subscription_list_user (
	subscription_list_id VARCHAR(64) NOT NULL,
	user_id VARCHAR(64) NOT NULL,
	db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (subscription_list_id, user_id),
	SHARD KEY (subscription_list_id)
)`

var SubscriptionListUserSchemaMYSQL = `CREATE TABLE IF NOT EXISTS subscription_list_user (
	subscription_list_id VARCHAR(64) NOT NULL,
	user_id VARCHAR(64) NOT NULL,
	db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (subscription_list_id, user_id)
)`
