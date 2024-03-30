package entity

import (
	"time"

	"github.com/rotisserie/eris"
	"github.com/tidwall/gjson"
)

var (
	// statuses
	SubscriptionListUserStatusActive       = int64(1)
	SubscriptionListUserStatusPaused       = int64(2) // used for double opt-in and bounces
	SubscriptionListUserStatusUnsubscribed = int64(3)
)

// attaches a user to a subscription list

type SubscriptionListUser struct {
	SubscriptionListID string          `db:"subscription_list_id" json:"subscription_list_id"`
	UserID             string          `db:"user_id" json:"user_id"`
	Status             *int64          `db:"status" json:"status,omitempty"`
	Comment            *NullableString `db:"comment" json:"comment,omitempty"` // optional, reason for status change (email bounce...)
	CreatedAt          time.Time       `db:"created_at" json:"created_at"`
	DBCreatedAt        time.Time       `db:"db_created_at" json:"db_created_at"`
	DBUpdatedAt        time.Time       `db:"db_updated_at" json:"db_updated_at"`
	MergedFromUserID   *string         `db:"merged_from_user_id" json:"merged_from_user_id,omitempty"`
	FieldsTimestamp    FieldsTimestamp `db:"fields_timestamp" json:"fields_timestamp"`

	// used to merge fields and append item_timeline at the right time
	UpdatedAt *time.Time `db:"-" json:"-"`
}

func NewSubscriptionListUserFromDataLog(dataLog *DataLog, clockDifference time.Duration, lists []*SubscriptionList) (subscribeToList *SubscriptionListUser, err error) {

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
			if value.Type != gjson.Number {
				err = eris.New("subscription_list_user.status must be a number")
				return false
			}

			subscribeToList.Status = Int64Ptr(int64(value.Int()))

		case "comment":
			if value.Type == gjson.Null {
				subscribeToList.Comment = NewNullableString(nil)
			} else {
				subscribeToList.Comment = NewNullableString(StringPtr(value.String()))
			}

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

	if subscribeToList.Status != nil && (*subscribeToList.Status < 0 || *subscribeToList.Status > 3) {
		return nil, eris.New("subscription_list_user.status is invalid")
	}

	// TODO : validate list exists + double opt in status
	var listFound *SubscriptionList

	for _, list := range lists {
		if list.ID == subscribeToList.SubscriptionListID {
			listFound = list
			break
		}
	}

	if listFound == nil {
		return nil, eris.New("subscription_list_user.subscription_list_id is invalid")
	}

	if listFound.DoubleOptIn && (subscribeToList.Status == nil || *subscribeToList.Status == 0) {
		subscribeToList.Status = Int64Ptr(SubscriptionListUserStatusPaused)
		subscribeToList.Comment = NewNullableString(StringPtr("waiting for double opt-in"))
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

func (s *SubscriptionListUser) GetFieldDate(field string) time.Time {
	// use updated_at if it has been passed in the API data import
	if s.UpdatedAt != nil && s.UpdatedAt.After(s.CreatedAt) {
		return *s.UpdatedAt
	}
	// or use the existing field timestamp
	if date, exists := s.FieldsTimestamp[field]; exists {
		return date
	}
	// or use the object creation date as a fallback
	return s.CreatedAt
}

func (s *SubscriptionListUser) SetStatus(value *int64, timestamp time.Time) (update *UpdatedField) {
	key := "status"
	// value cant be null
	if value == nil {
		return nil
	}
	// abort if values are equal
	if Int64Equal(s.Status, value) {
		return nil
	}
	existingValueTimestamp := s.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		s.Status = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: Int64PointerToInterface(s.Status),
		NewValue:  Int64PointerToInterface(value),
	}
	s.Status = value
	s.FieldsTimestamp[key] = timestamp
	return
}

func (o *SubscriptionListUser) SetComment(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "comme t"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && o.Comment != nil && o.Comment.IsNull == value.IsNull && o.Comment.String == value.String {
		return nil
	}
	existingValueTimestamp := o.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		o.Comment = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(o.Comment),
		NewValue:  NullableStringToInterface(value),
	}
	o.Comment = value
	o.FieldsTimestamp[key] = timestamp
	return
}

// merges two subs and returns the list of updated fields
func (fromSubscribe *SubscriptionListUser) MergeInto(toSubscribe *SubscriptionListUser) (updatedFields []*UpdatedField) {

	updatedFields = []*UpdatedField{} // init

	if toSubscribe.FieldsTimestamp == nil {
		toSubscribe.FieldsTimestamp = FieldsTimestamp{}
	}

	if fieldUpdate := toSubscribe.SetStatus(fromSubscribe.Status, fromSubscribe.GetFieldDate("status")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	if fieldUpdate := toSubscribe.SetComment(fromSubscribe.Comment, fromSubscribe.GetFieldDate("comment")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

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
	created_at DATETIME NOT NULL,
    db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
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
	created_at DATETIME NOT NULL,
  	db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    merged_from_user_id VARCHAR(64),
    fields_timestamp JSON NOT NULL,

	PRIMARY KEY (subscription_list_id, user_id)
) COLLATE utf8mb4_general_ci`
