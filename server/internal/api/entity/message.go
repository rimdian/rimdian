package entity

import (
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/rotisserie/eris"
	"github.com/tidwall/gjson"
)

var (
	MessageStatusPending      = int64(0)
	MessageStatusScheduled    = int64(10)
	MessageStatusQueued       = int64(20)
	MessageStatusRetrying     = int64(30)
	MessageStatusSent         = int64(40)
	MessageStatusFailed       = int64(50) // dropped, bounced, complained, spam
	MessageStatusDelivered    = int64(60)
	MessageStatusOpened       = int64(70)
	MessageStatusClicked      = int64(80)
	MessageStatusUnsubscribed = int64(90)
	MessageStatusComplained   = int64(100)

	DoubleOptInKeyword       = "double_opt_in_link"
	UnsubscribeKeyword       = "unsubscribe_link"
	OpenTrackingPixelKeyword = "open_tracking_pixel_src"

	// computed fields should be excluded from SELECT/INSERT while cloning rows
	MessageComputedFields []string = []string{
		"created_at_trunc",
	}
)

type GenerateEmailLinkOptions struct {
	DataLogID         string
	MessageExternalID string
	APIEndpoint       string
	Path              string
	SecretKey         string
	WorkspaceID       string
	SubscriptionList  *SubscriptionList
	User              *User
}

func GenerateEmailLink(opts GenerateEmailLinkOptions) (token string, err error) {

	// create a token with custom claims
	pasetoToken := paseto.NewToken()
	pasetoToken.SetAudience(opts.APIEndpoint)

	// claims should follow the dto.EmailTokenClaims struct
	pasetoToken.SetIssuedAt(time.Now())
	pasetoToken.SetString("dlid", opts.DataLogID)
	pasetoToken.SetString("mxid", opts.MessageExternalID)
	pasetoToken.SetString("wid", opts.WorkspaceID)
	pasetoToken.SetString("email", opts.User.Email.String)
	if opts.SubscriptionList != nil {
		pasetoToken.SetString("lid", opts.SubscriptionList.ID)
		pasetoToken.SetString("lname", opts.SubscriptionList.Name)
	}

	if opts.User.IsAuthenticated {
		pasetoToken.SetString("auth_uxid", opts.User.ExternalID)
	} else {
		pasetoToken.SetString("anon_uxid", opts.User.ExternalID)
	}

	key, err := paseto.V4SymmetricKeyFromBytes([]byte(opts.SecretKey))
	if err != nil {
		return "", eris.Wrap(err, "GenerateEmailLink V4SymmetricKeyFromBytes")
	}

	return opts.APIEndpoint + opts.Path + "?token=" + pasetoToken.V4Encrypt(key, nil), nil
}

type Message struct {
	ID               string          `db:"id" json:"id"`
	ExternalID       string          `db:"external_id" json:"external_id"`
	UserID           string          `db:"user_id" json:"user_id"`
	DomainID         *string         `db:"domain_id" json:"domain_id,omitempty"`
	SessionID        *string         `db:"session_id" json:"session_id,omitempty"`
	CreatedAt        time.Time       `db:"created_at" json:"created_at"`
	CreatedAtTrunc   time.Time       `db:"created_at_trunc" json:"created_at_trunc"` // date truncated at hour, used to optimize columnstore storage and queries
	DBCreatedAt      time.Time       `db:"db_created_at" json:"db_created_at"`
	DBUpdatedAt      time.Time       `db:"db_updated_at" json:"db_updated_at"`
	IsDeleted        bool            `db:"is_deleted" json:"is_deleted,omitempty"` // deleting rows in transactions cause deadlocks in singlestore, we use an update
	MergedFromUserID *string         `db:"merged_from_user_id" json:"merged_from_user_id,omitempty"`
	FieldsTimestamp  FieldsTimestamp `db:"fields_timestamp" json:"fields_timestamp"`

	Channel   string `db:"channel" json:"channel"`
	IsInbound bool   `db:"is_inbound" json:"is_inbound"` // used for conversation threads
	// outbound message:
	IsTransactional        *bool           `db:"is_transactional" json:"is_transactional"` // used to bypass marketing pressure filters
	MessageTemplateID      *string         `db:"message_template_id" json:"message_template_id,omitempty"`
	MessageTemplateVersion *int64          `db:"message_template_version" json:"message_template_version,omitempty"`
	SubscriptionListID     *string         `db:"subscription_list_id" json:"subscription_list_id,omitempty"`
	Data                   MapOfInterfaces `db:"data" json:"data"`
	// states:
	Status       *int64          `db:"status" json:"status"`
	StatusAt     *time.Time      `db:"status_at" json:"status_at"`                     // used to track when the status was last updated
	Comment      *NullableString `db:"comment" json:"comment,omitempty"`               // optional, reason for status change (email bounce...)
	RetryCount   int64           `db:"retry_count" json:"retry_count"`                 // used to track the number of retries
	IsSent       bool            `db:"is_sent" json:"is_sent"`                         // used to track if the message was sent
	SentAt       *time.Time      `db:"sent_at" json:"sent_at,omitempty"`               // used to track when the message was sent
	ScheduledAt  *time.Time      `db:"scheduled_at" json:"scheduled_at,omitempty"`     // used to track when the message was scheduled
	DeliveredAt  *time.Time      `db:"delivered_at" json:"delivered_at,omitempty"`     // used to track when the message was delivered
	FirstOpenAt  *time.Time      `db:"first_open_at" json:"first_open_at,omitempty"`   // used to track when the message was seen by the recipient
	FirstClickAt *time.Time      `db:"first_click_at" json:"first_click_at,omitempty"` // used to track when the message was clicked

	// used to merge fields
	UpdatedAt *time.Time `db:"-" json:"-"`
	// attached in data pipeline for easy access:
	SubscriptionList *SubscriptionList `db:"-" json:"-"`
	MessageTemplate  *MessageTemplate  `db:"-" json:"-"`
}

func (message *Message) BeforeInsert() {
	// set status
	if message.Status == nil {
		message.Status = &MessageStatusQueued
	}

	if message.StatusAt == nil {
		message.StatusAt = &message.CreatedAt
	}

	if message.ScheduledAt != nil && message.ScheduledAt.After(time.Now()) {
		message.Status = &MessageStatusScheduled
	}

	if message.IsTransactional == nil {
		message.IsTransactional = BoolPtr(false)
	}
}

func NewMessageFromDataLog(dataLog *DataLog, clockDifference time.Duration) (message *Message, err error) {

	result := gjson.Get(dataLog.Item, "message")
	if !result.Exists() {
		return nil, eris.New("item has no message object")
	}

	if result.Type != gjson.JSON {
		return nil, eris.New("item message is not an object")
	}

	// init
	message = &Message{
		UserID: dataLog.UserID,
		Data:   MapOfInterfaces{},
		// FieldsTimestamp: FieldsTimestamp{},
	}

	// loop over fields
	result.ForEach(func(key, value gjson.Result) bool {

		keyString := key.String()
		switch keyString {

		case "domain_id":
			if value.Type == gjson.Null {
				message.DomainID = nil
			} else {
				message.DomainID = StringPtr(ComputeSessionID(value.String()))
			}

		case "session_external_id":
			if value.Type == gjson.Null {
				message.SessionID = nil
			} else {
				message.SessionID = StringPtr(ComputeSessionID(value.String()))
			}

		case "created_at":
			if message.CreatedAt, err = time.Parse(time.RFC3339, value.String()); err != nil {
				err = eris.Wrap(err, "message.created_at")
				return false
			}

			// apply clock difference
			if message.CreatedAt.After(time.Now()) {

				message.CreatedAt = message.CreatedAt.Add(clockDifference)
				if message.CreatedAt.After(time.Now()) {
					err = eris.New("message.created_at cannot be in the future")
					return false
				}
			}

		case "updated_at":
			var updatedAt time.Time
			if updatedAt, err = time.Parse(time.RFC3339, value.String()); err != nil {
				err = eris.Wrap(err, "message.updated_at")
				return false
			}

			if updatedAt.After(time.Now()) {
				// apply clock difference
				updatedAtModified := updatedAt.Add(clockDifference)
				updatedAt = updatedAtModified
				if updatedAt.After(time.Now()) {
					err = eris.New("message.updated_at cannot be in the future")
					return false
				}
			}

			message.UpdatedAt = &updatedAt

		case "channel":
			message.Channel = value.String()

		case "is_inbound":
			message.IsInbound = value.Bool()

		case "is_transactional":
			if value.Type == gjson.Null {
				message.IsTransactional = BoolPtr(value.Bool())
			}

		case "message_template_id":
			if value.Type != gjson.Null {
				message.MessageTemplateID = StringPtr(value.String())
			}

		case "message_template_version":
			if value.Type != gjson.Null {
				message.MessageTemplateVersion = Int64Ptr(int64(value.Int()))
			}

		case "subscription_list_id":
			if value.Type != gjson.Null {
				message.SubscriptionListID = StringPtr(value.String())
			}

		case "data":
			if value.Type != gjson.Null {
				if value.Type != gjson.JSON {
					err = eris.New("message.data must be an object")
					return false
				}
				value.ForEach(func(key, value gjson.Result) bool {
					message.Data[key.String()] = value.Value()
					return true
				})
			}

		case "status":
			if value.Type != gjson.Number {
				err = eris.New("message.status must be a number")
				return false
			}

			message.Status = Int64Ptr(int64(value.Int()))

		case "status_at":
			if value.Type != gjson.Null {
				var statusAt time.Time
				if statusAt, err = time.Parse(time.RFC3339, value.String()); err != nil {
					err = eris.Wrap(err, "message.status_at")
					return false
				}
				message.StatusAt = &statusAt
			}

		case "comment":
			if value.Type == gjson.Null {
				message.Comment = nil
			} else {
				message.Comment = NewNullableString(StringPtr(value.String()))
			}

		case "retry_count":
			message.RetryCount = int64(value.Int())

		case "is_sent":
			message.IsSent = value.Bool()

		case "sent_at":
			if value.Type != gjson.Null {
				var sentAt time.Time
				if sentAt, err = time.Parse(time.RFC3339, value.String()); err != nil {
					err = eris.Wrap(err, "message.sent_at")
					return false
				}
				message.SentAt = &sentAt
			}

		case "scheduled_at":
			if value.Type != gjson.Null {
				var scheduledAt time.Time
				if scheduledAt, err = time.Parse(time.RFC3339, value.String()); err != nil {
					err = eris.Wrap(err, "message.scheduled_at")
					return false
				}
				message.ScheduledAt = &scheduledAt
			}

		case "delivered_at":
			if value.Type != gjson.Null {
				var deliveredAt time.Time
				if deliveredAt, err = time.Parse(time.RFC3339, value.String()); err != nil {
					err = eris.Wrap(err, "message.delivered_at")
					return false
				}
				message.DeliveredAt = &deliveredAt
			}

		case "first_open_at":
			if value.Type != gjson.Null {
				var openAt time.Time
				if openAt, err = time.Parse(time.RFC3339, value.String()); err != nil {
					err = eris.Wrap(err, "message.first_open_at")
					return false
				}
				message.FirstOpenAt = &openAt
			}

		case "first_click_at":
			if value.Type != gjson.Null {
				var clickAt time.Time
				if clickAt, err = time.Parse(time.RFC3339, value.String()); err != nil {
					err = eris.Wrap(err, "message.first_click_at")
					return false
				}
				message.FirstClickAt = &clickAt
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
	if message.UpdatedAt == nil {
		message.UpdatedAt = &message.CreatedAt
	}

	// Validation

	if message.CreatedAt.IsZero() {
		return nil, eris.New("message.created_at is required")
	}

	if message.Channel == "" {
		return nil, eris.New("message.channel is required")
	}

	if message.Status != nil && message.StatusAt.IsZero() {
		return nil, eris.New("message.status_at is required with message.status")
	}

	return message, nil
}

func (s *Message) UpdateFieldTimestamp(field string, timestamp *time.Time) {
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
func (s *Message) SetCreatedAt(value time.Time) {
	// update if current value is older
	if value.Before(s.CreatedAt) {
		s.CreatedAt = value
	}
}

func (s *Message) GetFieldDate(field string) time.Time {
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

func (o *Message) SetDomainID(value *string, timestamp time.Time) (update *UpdatedField) {
	key := "domain_id"
	// value cant be null
	if value == nil {
		return nil
	}
	// abort if values are equal
	if FixedDeepEqual(o.DomainID, value) {
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
		o.DomainID = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: StringPointerToInterface(o.DomainID),
		NewValue:  StringPointerToInterface(value),
	}
	o.DomainID = value
	o.FieldsTimestamp[key] = timestamp
	return
}

func (o *Message) SetSessionID(value *string, timestamp time.Time) (update *UpdatedField) {
	key := "session_id"
	// value cant be null
	if value == nil {
		return nil
	}
	// abort if values are equal
	if FixedDeepEqual(o.SessionID, value) {
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
		o.SessionID = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: StringPointerToInterface(o.SessionID),
		NewValue:  StringPointerToInterface(value),
	}
	o.SessionID = value
	o.FieldsTimestamp[key] = timestamp
	return
}

func (o *Message) SetChannel(value string, timestamp time.Time) (update *UpdatedField) {
	key := "channel"
	// abort if values are equal
	if o.Channel == value {
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
		o.Channel = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: o.Channel,
		NewValue:  value,
	}
	o.Channel = value
	o.FieldsTimestamp[key] = timestamp
	return
}

func (o *Message) SetIsInbound(value bool, timestamp time.Time) (update *UpdatedField) {
	key := "is_inbound"
	// abort if values are equal
	if o.IsInbound == value {
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
		o.IsInbound = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: o.IsInbound,
		NewValue:  value,
	}

	o.IsInbound = value
	o.FieldsTimestamp[key] = timestamp
	return
}

// only update if value becomes true
func (u *Message) SetIsTransactional(value *bool, timestamp time.Time) (update *UpdatedField) {
	// value cant be null
	if value == nil || !*value {
		return nil
	}
	// only update if value becomes true
	if u.IsTransactional != nil && *u.IsTransactional {
		return nil
	}

	update = &UpdatedField{
		Field:     "is_transactional",
		PrevValue: false,
		NewValue:  true,
	}
	u.IsTransactional = value
	u.UpdateFieldTimestamp("is_transactional", &timestamp)
	return
}

// only update if value is set
func (o *Message) SetMessageTemplateID(value *string, timestamp time.Time) (update *UpdatedField) {
	key := "message_template_id"
	// value cant be null
	if value == nil {
		return nil
	}
	// abort if values are equal
	if FixedDeepEqual(o.MessageTemplateID, value) {
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
		o.MessageTemplateID = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: StringPointerToInterface(o.MessageTemplateID),
		NewValue:  StringPointerToInterface(value),
	}
	o.MessageTemplateID = value
	o.FieldsTimestamp[key] = timestamp
	return
}

func (o *Message) SetMessageTemplateVersion(value *int64, timestamp time.Time) (update *UpdatedField) {
	key := "message_template_version"
	// value cant be null
	if value == nil {
		return nil
	}
	// abort if values are equal
	if FixedDeepEqual(o.MessageTemplateVersion, value) {
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
		o.MessageTemplateVersion = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: Int64PointerToInterface(o.MessageTemplateVersion),
		NewValue:  Int64PointerToInterface(value),
	}
	o.MessageTemplateVersion = value
	o.FieldsTimestamp[key] = timestamp
	return
}

func (o *Message) SetSubscriptionListID(value *string, timestamp time.Time) (update *UpdatedField) {
	key := "subscription_list_id"
	// value cant be null
	if value == nil {
		return nil
	}
	// abort if values are equal
	if FixedDeepEqual(o.SubscriptionListID, value) {
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
		o.SubscriptionListID = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: StringPointerToInterface(o.SubscriptionListID),
		NewValue:  StringPointerToInterface(value),
	}
	o.SubscriptionListID = value
	o.FieldsTimestamp[key] = timestamp
	return
}

// keep the biggest map
func (o *Message) SetData(value MapOfInterfaces, timestamp time.Time) (update *UpdatedField) {
	if value == nil {
		return nil
	}

	if o.Data == nil {
		o.Data = value
		return nil
	}

	if len(value) > len(o.Data) {
		o.Data = value
	}

	return
}

func (o *Message) SetStatus(value *int64, timestamp time.Time) (update *UpdatedField) {
	key := "status"
	// value cant be null
	if value == nil {
		return nil
	}
	// abort if values are equal
	if Int64Equal(o.Status, value) {
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
		o.Status = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: Int64PointerToInterface(o.Status),
		NewValue:  Int64PointerToInterface(value),
	}
	o.Status = value
	o.FieldsTimestamp[key] = timestamp
	return
}

func (o *Message) SetStatusAt(value *time.Time, timestamp time.Time) (update *UpdatedField) {
	key := "status_at"
	// abort if values are equal
	if o.StatusAt != nil && o.StatusAt.Equal(*value) {
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
		o.StatusAt = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: o.StatusAt,
		NewValue:  value,
	}
	o.StatusAt = value
	o.FieldsTimestamp[key] = timestamp
	return
}

func (o *Message) SetComment(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "comment"
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
		o.Comment = &NullableString{IsNull: value == nil, String: value.String}
		return
	}

	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(o.Comment),
		NewValue:  NullableStringToInterface(&NullableString{IsNull: value == nil, String: value.String}),
	}

	o.Comment = &NullableString{IsNull: value == nil, String: value.String}
	o.FieldsTimestamp[key] = timestamp
	return
}

func (o *Message) SetRetryCount(value int64, timestamp time.Time) (update *UpdatedField) {
	key := "retry_count"
	// abort if values are equal
	if o.RetryCount == value {
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
		o.RetryCount = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: o.RetryCount,
		NewValue:  value,
	}

	o.RetryCount = value
	o.FieldsTimestamp[key] = timestamp
	return
}

func (o *Message) SetIsSent(value bool, timestamp time.Time) (update *UpdatedField) {
	key := "is_sent"
	// abort if values are equal
	if o.IsSent == value {
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
		o.IsSent = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: o.IsSent,
		NewValue:  value,
	}

	o.IsSent = value
	o.FieldsTimestamp[key] = timestamp
	return
}

func (o *Message) SetSentAt(value *time.Time, timestamp time.Time) (update *UpdatedField) {
	key := "sent_at"
	// abort if values are equal
	if o.SentAt != nil && o.SentAt.Equal(*value) {
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
		o.SentAt = value
		return
	}

	update = &UpdatedField{
		Field:     key,
		PrevValue: o.SentAt,
		NewValue:  value,
	}

	o.SentAt = value
	o.FieldsTimestamp[key] = timestamp
	return
}

func (o *Message) SetScheduledAt(value *time.Time, timestamp time.Time) (update *UpdatedField) {
	key := "scheduled_at"
	// abort if values are equal
	if o.ScheduledAt != nil && o.ScheduledAt.Equal(*value) {
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
		o.ScheduledAt = value
		return
	}

	update = &UpdatedField{
		Field:     key,
		PrevValue: o.ScheduledAt,
		NewValue:  value,
	}

	o.ScheduledAt = value
	o.FieldsTimestamp[key] = timestamp
	return
}

func (o *Message) SetDeliveredAt(value *time.Time, timestamp time.Time) (update *UpdatedField) {
	key := "delivered_at"
	// abort if values are equal
	if o.DeliveredAt != nil && o.DeliveredAt.Equal(*value) {
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
		o.DeliveredAt = value
		return
	}

	update = &UpdatedField{
		Field:     key,
		PrevValue: o.DeliveredAt,
		NewValue:  value,
	}

	o.DeliveredAt = value
	o.FieldsTimestamp[key] = timestamp
	return
}

// keep the oldest date
func (u *Message) SetFirstOpenAt(value *time.Time) (update *UpdatedField) {
	if value == nil {
		return nil
	}

	if u.FirstOpenAt == nil {
		u.FirstOpenAt = value
		u.UpdateFieldTimestamp("first_open_at", value)
		return &UpdatedField{
			Field:     "first_open_at",
			PrevValue: nil,
			NewValue:  value.UTC(),
		}
	}

	// update if current value is older
	if value.After(*u.FirstOpenAt) || value.Equal(*u.FirstOpenAt) {
		return nil
	}

	update = &UpdatedField{
		Field:     "first_open_at",
		PrevValue: u.FirstOpenAt.UTC(),
		NewValue:  value.UTC(),
	}
	u.FirstOpenAt = value
	u.FieldsTimestamp["first_open_at"] = *value
	return
}

// keep the oldest date
func (u *Message) SetFirstClickAt(value *time.Time) (update *UpdatedField) {
	if value == nil {
		return nil
	}

	if u.FirstOpenAt == nil {
		u.FirstOpenAt = value
		u.UpdateFieldTimestamp("first_click_at", value)
		return &UpdatedField{
			Field:     "first_click_at",
			PrevValue: nil,
			NewValue:  value.UTC(),
		}
	}

	// update if current value is older
	if value.After(*u.FirstOpenAt) || value.Equal(*u.FirstOpenAt) {
		return nil
	}

	update = &UpdatedField{
		Field:     "first_click_at",
		PrevValue: u.FirstOpenAt.UTC(),
		NewValue:  value.UTC(),
	}
	u.FirstOpenAt = value
	u.FieldsTimestamp["first_click_at"] = *value
	return
}

// merges two subs and returns the list of updated fields
func (fromMessage *Message) MergeInto(toMessage *Message) (updatedFields []*UpdatedField) {

	updatedFields = []*UpdatedField{} // init

	if toMessage.FieldsTimestamp == nil {
		toMessage.FieldsTimestamp = FieldsTimestamp{}
	}

	if fieldUpdate := toMessage.SetDomainID(fromMessage.DomainID, fromMessage.GetFieldDate("domain_id")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	if fieldUpdate := toMessage.SetSessionID(fromMessage.SessionID, fromMessage.GetFieldDate("session_id")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	if fieldUpdate := toMessage.SetChannel(fromMessage.Channel, fromMessage.GetFieldDate("channel")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	if fieldUpdate := toMessage.SetIsInbound(fromMessage.IsInbound, fromMessage.GetFieldDate("is_inbound")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	if fieldUpdate := toMessage.SetIsTransactional(fromMessage.IsTransactional, fromMessage.GetFieldDate("is_transactional")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	if fieldUpdate := toMessage.SetMessageTemplateID(fromMessage.MessageTemplateID, fromMessage.GetFieldDate("message_template_id")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	if fieldUpdate := toMessage.SetMessageTemplateVersion(fromMessage.MessageTemplateVersion, fromMessage.GetFieldDate("message_template_version")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	if fieldUpdate := toMessage.SetSubscriptionListID(fromMessage.SubscriptionListID, fromMessage.GetFieldDate("subscription_list_id")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	if fieldUpdate := toMessage.SetData(fromMessage.Data, fromMessage.GetFieldDate("data")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	// status, status_at and comment are updated together with the status_at field
	if fieldUpdate := toMessage.SetStatus(fromMessage.Status, fromMessage.GetFieldDate("status_at")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	if fieldUpdate := toMessage.SetStatusAt(fromMessage.StatusAt, fromMessage.GetFieldDate("status_at")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	if fieldUpdate := toMessage.SetComment(fromMessage.Comment, fromMessage.GetFieldDate("status_at")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	if fieldUpdate := toMessage.SetRetryCount(fromMessage.RetryCount, fromMessage.GetFieldDate("retry_count")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	if fieldUpdate := toMessage.SetIsSent(fromMessage.IsSent, fromMessage.GetFieldDate("is_sent")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	if fieldUpdate := toMessage.SetSentAt(fromMessage.SentAt, fromMessage.GetFieldDate("sent_at")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	if fieldUpdate := toMessage.SetScheduledAt(fromMessage.ScheduledAt, fromMessage.GetFieldDate("scheduled_at")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	if fieldUpdate := toMessage.SetDeliveredAt(fromMessage.DeliveredAt, fromMessage.GetFieldDate("delivered_at")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	if fieldUpdate := toMessage.SetFirstOpenAt(fromMessage.FirstOpenAt); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	if fieldUpdate := toMessage.SetFirstClickAt(fromMessage.FirstClickAt); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	// UpdatedAt is the timeOfEvent for ITs
	toMessage.UpdatedAt = fromMessage.UpdatedAt
	// priority to oldest date
	toMessage.SetCreatedAt(fromMessage.CreatedAt)

	return
}

var MessageSchema = `CREATE TABLE IF NOT EXISTS message (	
  id VARCHAR(64) NOT NULL,
  external_id VARCHAR(255) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  domain_id VARCHAR(64),
  session_id VARCHAR(64),
  created_at DATETIME NOT NULL,
  created_at_trunc AS DATE_TRUNC('hour', created_at) PERSISTED DATETIME,
  db_created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  is_deleted BOOLEAN DEFAULT FALSE,
  merged_from_user_id VARCHAR(64),
  fields_timestamp JSON NOT NULL,

  channel VARCHAR(24) NOT NULL,
  is_inbound BOOLEAN NOT NULL,
  is_transactional BOOLEAN NOT NULL,
  message_template_id VARCHAR(64),
  message_template_version INT,
  subscription_list_id VARCHAR(64),
  data JSON NOT NULL,

  status INT NOT NULL DEFAULT 0,
  status_at DATETIME NOT NULL,
  comment VARCHAR(255),
  retry_count INT NOT NULL DEFAULT 0,
  is_sent BOOLEAN NOT NULL DEFAULT FALSE,
  sent_at DATETIME,
  scheduled_at DATETIME,
  delivered_at DATETIME,
  first_open_at DATETIME,
  first_click_at DATETIME,

  SORT KEY (created_at_trunc DESC),
  PRIMARY KEY (id, user_id),
  KEY (user_id) USING HASH, -- for merging
  SHARD KEY (user_id)
) COLLATE utf8mb4_general_ci`

var MessageSchemaMYSQL = `CREATE TABLE IF NOT EXISTS message (
  id VARCHAR(64) NOT NULL,
  external_id VARCHAR(255) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  domain_id VARCHAR(64),
  session_id VARCHAR(64),
  created_at DATETIME NOT NULL,
  created_at_trunc DATETIME GENERATED ALWAYS AS (DATE_FORMAT(created_at, '%Y-%m-%d %H:00:00')) STORED,
  db_created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  is_deleted BOOLEAN DEFAULT FALSE,
  merged_from_user_id VARCHAR(64),
  fields_timestamp JSON NOT NULL,

  channel VARCHAR(24) NOT NULL,
  is_inbound BOOLEAN NOT NULL,
  is_transactional BOOLEAN NOT NULL,
  message_template_id VARCHAR(64),
  message_template_version INT,
  subscription_list_id VARCHAR(64),
  data JSON NOT NULL,

  status INT NOT NULL DEFAULT 0,
  status_at DATETIME NOT NULL,
  comment VARCHAR(255),
  retry_count INT NOT NULL DEFAULT 0,
  is_sent BOOLEAN NOT NULL DEFAULT FALSE,
  sent_at DATETIME,
  scheduled_at DATETIME,
  delivered_at DATETIME,
  first_open_at DATETIME,
  first_click_at DATETIME,

  PRIMARY KEY (id, user_id)
) COLLATE utf8mb4_general_ci`
