package entity

import (
	"time"

	"github.com/rotisserie/eris"
	"github.com/tidwall/gjson"
)

var (
	// statuses
	SubscriptionListUserStatusActive       = 1
	SubscriptionListUserStatusPaused       = 2
	SubscriptionListUserStatusUnsubscribed = 3
)

// attaches a user to a subscription list

type SubscriptionListUser struct {
	SubscriptionListID string          `db:"subscription_list_id" json:"subscription_list_id"`
	UserID             string          `db:"user_id" json:"user_id"`
	Status             int             `db:"status" json:"status"`
	Comment            *string         `db:"comment" json:"comment,omitempty"` // optional, reason for status change (email bounce...)
	CreatedAt          time.Time       `db:"created_at" json:"created_at"`
	DBCreatedAt        time.Time       `db:"db_created_at" json:"db_created_at"`
	DBUpdatedAt        time.Time       `db:"db_updated_at" json:"db_updated_at"`
	MergedFromUserID   *string         `db:"merged_from_user_id" json:"merged_from_user_id,omitempty"`
	FieldsTimestamp    FieldsTimestamp `db:"fields_timestamp" json:"fields_timestamp"`

	// used to merge fields and append item_timeline at the right time
	UpdatedAt *time.Time `db:"-" json:"-"`
}

func NewSubscriptionListUserFromDataLog(dataLog *DataLog, clockDifference time.Duration) (subscribeToList *SubscriptionListUser, err error) {

	result := gjson.Get(dataLog.Item, "subscription_list_user")
	if !result.Exists() {
		return nil, eris.New("item has no subscription_list_user object")
	}

	if result.Type != gjson.JSON {
		return nil, eris.New("item subscription_list_user is not an object")
	}

	// init
	subscribeToList = &SubscriptionListUser{
		UserID: dataLog.UserID,
		// FieldsTimestamp: FieldsTimestamp{},
	}

	// loop over fields
	result.ForEach(func(key, value gjson.Result) bool {

		keyString := key.String()
		switch keyString {

		case "subscription_list_id":
			subscribeToList.SubscriptionListID = value.String()

		case "created_at":
			if subscribeToList.CreatedAt, err = time.Parse(time.RFC3339, value.String()); err != nil {
				err = eris.Wrap(err, "subscription_list_user.created_at")
				return false
			}

			// apply clock difference
			if subscribeToList.CreatedAt.After(time.Now()) {

				subscribeToList.CreatedAt = subscribeToList.CreatedAt.Add(clockDifference)
				if subscribeToList.CreatedAt.After(time.Now()) {
					err = eris.New("subscription_list_user.created_at cannot be in the future")
					return false
				}
			}

		case "updated_at":
			var updatedAt time.Time
			if updatedAt, err = time.Parse(time.RFC3339, value.String()); err != nil {
				err = eris.Wrap(err, "subscription_list_user.updated_at")
				return false
			}

			if updatedAt.After(time.Now()) {
				// apply clock difference
				updatedAtModified := updatedAt.Add(clockDifference)
				updatedAt = updatedAtModified
				if updatedAt.After(time.Now()) {
					err = eris.New("subscription_list_user.updated_at cannot be in the future")
					return false
				}
			}

			subscribeToList.UpdatedAt = &updatedAt

		case "status":
			subscribeToList.Status = int(value.Int())

		case "comment":
			comment := value.String()
			subscribeToList.Comment = &comment

		default:
			// ignore other fields
		}

		return true
	})

	if err != nil {
		return nil, err
	}

	// use data import createdAt as updatedAt if not provided
	if subscribeToList.UpdatedAt == nil {
		subscribeToList.UpdatedAt = &subscribeToList.CreatedAt
	}

	// Validation
	if subscribeToList.SubscriptionListID == "" {
		return nil, eris.New("subscription_list_user.subscription_list_id is required")
	}

	if subscribeToList.CreatedAt.IsZero() {
		return nil, eris.New("subscription_list_user.created_at is required")
	}

	if subscribeToList.Status <= 0 {
		return nil, eris.New("subscription_list_user.status is required")
	}

	if subscribeToList.Status > 3 {
		return nil, eris.New("subscription_list_user.status is invalid")
	}

	return subscribeToList, nil
}

func (s *SubscriptionListUser) UpdateFieldTimestamp(field string, timestamp *time.Time) {
	if timestamp == nil {
		return
	}
	if previousTimestamp, exists := s.FieldsTimestamp[field]; exists {
		if previousTimestamp.Before(*timestamp) {
			s.FieldsTimestamp[field] = *timestamp
		}
	} else {
		s.FieldsTimestamp[field] = *timestamp
	}
}
func (s *SubscriptionListUser) SetCreatedAt(value time.Time) {
	// update if current value is older
	if value.Before(s.CreatedAt) {
		s.CreatedAt = value
	}
}

// merges two subs and returns the list of updated fields
func (fromSubscribe *SubscriptionListUser) MergeInto(toSubscribe *SubscriptionListUser) (updatedFields []*UpdatedField) {

	updatedFields = []*UpdatedField{} // init

	if toSubscribe.FieldsTimestamp == nil {
		toSubscribe.FieldsTimestamp = FieldsTimestamp{}
	}

	// TODO
	// if fieldUpdate := toSubscribe.SetUnsubscribed(fromSubscribe.Unsubscribed, fromSubscribe.GetFieldDate("status")); fieldUpdate != nil {
	// 	updatedFields = append(updatedFields, fieldUpdate)
	// }

	// UpdatedAt is the timeOfEvent for ITs
	toSubscribe.UpdatedAt = fromSubscribe.UpdatedAt
	// priority to oldest date
	toSubscribe.SetCreatedAt(fromSubscribe.CreatedAt)

	return
}

var SubscriptionListUserSchema = `CREATE ROWSTORE TABLE IF NOT EXISTS subscription_list_user (
	subscription_list_id VARCHAR(64) NOT NULL,
	user_id VARCHAR(64) NOT NULL,
	status INT NOT NULL DEFAULT 0,
	comment VARCHAR(255),
	created_at TIMESTAMP NOT NULL,
    db_created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    merged_from_user_id VARCHAR(64),
    fields_timestamp JSON NOT NULL,

	PRIMARY KEY (subscription_list_id, user_id),
	SHARD KEY (subscription_list_id)
) COLLATE utf8mb4_general_ci`

var SubscriptionListUserSchemaMYSQL = `CREATE TABLE IF NOT EXISTS subscription_list_user (
	subscription_list_id VARCHAR(64) NOT NULL,
	user_id VARCHAR(64) NOT NULL,
	status INT NOT NULL DEFAULT 0,
	comment VARCHAR(255),
	created_at TIMESTAMP NOT NULL,
    db_created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    merged_from_user_id VARCHAR(64),
    fields_timestamp JSON NOT NULL,

	PRIMARY KEY (subscription_list_id, user_id)
) COLLATE utf8mb4_general_ci`
