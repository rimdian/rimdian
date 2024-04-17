package entity

import (
	"time"

	"github.com/rotisserie/eris"
)

type SubscriptionList struct {
	ID                string    `db:"id" json:"id"`
	Name              string    `db:"name" json:"name"`
	Color             string    `db:"color" json:"color"`
	Channel           string    `db:"channel" json:"channel"` // email | sms | push
	DoubleOptIn       bool      `db:"double_opt_in" json:"double_opt_in"`
	MessageTemplateID *string   `db:"message_template_id" json:"message_template_id,omitempty"`
	DBCreatedAt       time.Time `db:"db_created_at" json:"db_created_at"`
	DBUpdatedAt       time.Time `db:"db_updated_at" json:"db_updated_at"`

	// joined server side
	ActiveUsers       int64 `db:"active_users" json:"active_users"`
	PausedUsers       int64 `db:"paused_users" json:"paused_users"`
	UnsubscribedUsers int64 `db:"unsubscribed_users" json:"unsubscribed_users"`
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
			if sl.MessageTemplateID == nil {
				return eris.New("message_template_id is required")
			}
		}
	}

	return nil
}

// schema
var SubscriptionListSchema = `CREATE ROWSTORE TABLE IF NOT EXISTS subscription_list (
	id VARCHAR(64) NOT NULL,
	name VARCHAR(255) NOT NULL,
	color VARCHAR(32) NOT NULL,
	channel VARCHAR(32) NOT NULL,
	double_opt_in BOOLEAN NOT NULL,
	message_template_id VARCHAR(64),
	db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (id),
	SHARD KEY (id)
) COLLATE utf8mb4_general_ci`

var SubscriptionListSchemaMYSQL = `CREATE TABLE IF NOT EXISTS subscription_list (
	id VARCHAR(64) NOT NULL,
	name VARCHAR(255) NOT NULL,
	color VARCHAR(32) NOT NULL,
	channel VARCHAR(32) NOT NULL,
	double_opt_in BOOLEAN NOT NULL,
	message_template_id VARCHAR(64),
	db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (id)
) COLLATE utf8mb4_general_ci`
