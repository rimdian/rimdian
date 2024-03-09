package entity

import (
	"crypto/sha1"
	"fmt"
	"strings"
	"time"

	"github.com/rotisserie/eris"
	"github.com/tidwall/gjson"
)

var (
	ErrCustomEventRequired = eris.New("custom_event is required")

	// computed fields should be excluded from SELECT/INSERT while cloning rows
	CustomEventComputedFields []string = []string{
		"created_at_trunc",
	}
)

type CustomEvent struct {
	ID               string          `db:"id" json:"id"`
	ExternalID       string          `db:"external_id" json:"external_id"`
	UserID           string          `db:"user_id" json:"user_id"`
	DomainID         *string         `db:"domain_id" json:"domain_id,omitempty"`
	SessionID        *string         `db:"session_id" json:"session_id,omitempty"` // historical sessions might not have session recorded
	CreatedAt        time.Time       `db:"created_at" json:"created_at"`
	CreatedAtTrunc   time.Time       `db:"created_at_trunc" json:"created_at_trunc"` // date truncated at hour, used to optimize columnstore storage and queries
	DBCreatedAt      time.Time       `db:"db_created_at" json:"db_created_at"`
	DBUpdatedAt      time.Time       `db:"db_updated_at" json:"db_updated_at"`
	IsDeleted        bool            `db:"is_deleted" json:"is_deleted,omitempty"` // deleting rows in transactions cause deadlocks in singlestore, we use an update
	MergedFromUserID *string         `db:"merged_from_user_id" json:"merged_from_user_id,omitempty"`
	FieldsTimestamp  FieldsTimestamp `db:"fields_timestamp" json:"fields_timestamp"`

	Label          string           `db:"label" json:"label"`
	StringValue    *NullableString  `db:"string_value" json:"string_value"`
	NumberValue    *NullableFloat64 `db:"number_value" json:"number_value"`
	BooleanValue   *NullableBool    `db:"boolean_value" json:"boolean_value"`
	NonInteractive bool             `db:"non_interactive" json:"non_interactive"`

	// Not persisted in DB:
	UpdatedAt    *time.Time    `db:"-" json:"-"` // used to merge fields and append item_timeline at the right time
	ExtraColumns AppItemFields `db:"-" json:"-"` // converted into "app_xxx" fields when marshaling JSON
}

func (o *CustomEvent) GetFieldDate(field string) time.Time {
	// use updated_at if it has been passed in the API data import
	if o.UpdatedAt != nil && o.UpdatedAt.After(o.CreatedAt) {
		return *o.UpdatedAt
	}
	// or use the existing field timestamp
	if date, exists := o.FieldsTimestamp[field]; exists {
		return date
	}
	// or use the object creation date as a fallback
	return o.CreatedAt
}

// update a field timestamp to its most recent value
func (o *CustomEvent) UpdateFieldTimestamp(field string, timestamp *time.Time) {
	if timestamp == nil {
		return
	}
	if previousTimestamp, exists := o.FieldsTimestamp[field]; exists {
		if previousTimestamp.Before(*timestamp) {
			o.FieldsTimestamp[field] = *timestamp
		}
	} else {
		o.FieldsTimestamp[field] = *timestamp
	}
}

func (o *CustomEvent) SetSessionID(value *string, timestamp *time.Time) (update *UpdatedField) {
	key := "session_id"
	// value cant be null
	if value == nil {
		return nil
	}
	// abort if values are equal
	if FixedDeepEqual(o.SessionID, value) {
		return nil
	}
	// compare timestamp with mergeable fields timestamp and mutate value if timestamp is newer
	// init field update
	update = &UpdatedField{
		Field:     key,
		PrevValue: StringPointerToInterface(o.SessionID),
	}
	previousTimestamp, exists := o.FieldsTimestamp[key]
	if !exists {
		o.SessionID = value
		o.UpdateFieldTimestamp(key, timestamp)
		update.NewValue = StringPointerToInterface(value)
		return
	}
	// abort if a previous timestamp exists, and the current one is not provided
	if timestamp == nil {
		return nil
	}
	// abort if the current timestamp is older than the previous one
	if timestamp.Before(previousTimestamp) {
		return nil
	}
	o.SessionID = value
	o.UpdateFieldTimestamp(key, timestamp)
	update.NewValue = StringPointerToInterface(value)
	return
}
func (s *CustomEvent) SetCreatedAt(value time.Time) {
	// update if current value is older
	if value.Before(s.CreatedAt) {
		s.CreatedAt = value
		s.CreatedAtTrunc = value.Truncate(time.Hour)
	}
}

func (s *CustomEvent) SetDomainID(value *string, timestamp *time.Time) (update *UpdatedField) {
	key := "domain_id"
	// value cant be null
	if value == nil {
		return nil
	}
	// abort if values are equal
	if StringsEqual(s.DomainID, value) {
		return nil
	}
	// compare timestamp with mergeable fields timestamp and mutate value if timestamp is newer
	// init field update
	update = &UpdatedField{
		Field:     key,
		PrevValue: StringPointerToInterface(s.DomainID),
	}
	previousTimestamp, exists := s.FieldsTimestamp[key]
	if !exists {
		s.DomainID = value
		s.UpdateFieldTimestamp(key, timestamp)
		update.NewValue = StringPointerToInterface(value)
		return
	}
	// abort if a previous timestamp exists, and the current one is not provided
	if timestamp == nil {
		return nil
	}
	// abort if the current timestamp is older than the previous one
	if timestamp.Before(previousTimestamp) {
		return nil
	}
	s.DomainID = value
	s.UpdateFieldTimestamp(key, timestamp)
	update.NewValue = StringPointerToInterface(value)
	return
}

func (s *CustomEvent) SetLabel(value string, timestamp *time.Time) (update *UpdatedField) {
	key := "label"
	// abort if values are equal
	if s.Label == value {
		return nil
	}
	// compare timestamp with mergeable fields timestamp and mutate value if timestamp is newer
	// init field update
	update = &UpdatedField{
		Field:     key,
		PrevValue: s.Label,
	}
	previousTimestamp, exists := s.FieldsTimestamp[key]
	if !exists {
		s.Label = value
		s.UpdateFieldTimestamp(key, timestamp)
		update.NewValue = value
		return
	}
	// abort if a previous timestamp exists, and the current one is not provided
	if timestamp == nil {
		return nil
	}
	// abort if the current timestamp is older than the previous one
	if timestamp.Before(previousTimestamp) {
		return nil
	}
	s.Label = value
	s.UpdateFieldTimestamp(key, timestamp)
	update.NewValue = value
	return
}
func (s *CustomEvent) SetStringValue(value *NullableString, timestamp *time.Time) (update *UpdatedField) {
	key := "string_value"
	// abort if values are equal
	if value == nil && s.StringValue == nil {
		return nil
	}
	if value != nil && s.StringValue != nil && s.StringValue.IsNull == value.IsNull && s.StringValue.String == value.String {
		return nil
	}
	// compare timestamp with mergeable fields timestamp and mutate value if timestamp is newer
	// init field update
	update = &UpdatedField{Field: key, PrevValue: NullableStringToInterface(s.StringValue)}
	previousTimestamp, exists := s.FieldsTimestamp[key]
	if !exists {
		s.StringValue = value
		s.UpdateFieldTimestamp(key, timestamp)
		update.NewValue = NullableStringToInterface(value)
		return
	}
	// abort if a previous timestamp exists, and the current one is not provided
	if timestamp == nil {
		return nil
	}
	// abort if the current timestamp is older than the previous one
	if timestamp.Before(previousTimestamp) {
		return nil
	}
	s.StringValue = value
	s.UpdateFieldTimestamp(key, timestamp)
	update.NewValue = NullableStringToInterface(value)
	return
}
func (s *CustomEvent) SetNumberValue(value *NullableFloat64, timestamp *time.Time) (update *UpdatedField) {
	key := "number_value"
	// abort if values are equal
	if value == nil && s.NumberValue == nil {
		return nil
	}
	if value != nil && s.NumberValue != nil && s.NumberValue.IsNull == value.IsNull && s.NumberValue.Float64 == value.Float64 {
		return nil
	}
	// compare timestamp with mergeable fields timestamp and mutate value if timestamp is newer
	// init field update
	update = &UpdatedField{Field: key, PrevValue: NullableFloat64ToInterface(s.NumberValue)}
	previousTimestamp, exists := s.FieldsTimestamp[key]
	if !exists {
		s.NumberValue = value
		s.UpdateFieldTimestamp(key, timestamp)
		update.NewValue = NullableFloat64ToInterface(value)
		return
	}
	// abort if a previous timestamp exists, and the current one is not provided
	if timestamp == nil {
		return nil
	}
	// abort if the current timestamp is older than the previous one
	if timestamp.Before(previousTimestamp) {
		return nil
	}
	s.NumberValue = value
	s.UpdateFieldTimestamp(key, timestamp)
	update.NewValue = NullableFloat64ToInterface(value)
	return
}
func (s *CustomEvent) SetBooleanValue(value *NullableBool, timestamp *time.Time) (update *UpdatedField) {
	key := "boolean_value"
	// abort if values are equal
	if value == nil && s.BooleanValue == nil {
		return nil
	}
	if value != nil && s.BooleanValue != nil && s.BooleanValue.IsNull == value.IsNull && s.BooleanValue.Bool == value.Bool {
		return nil
	}
	// compare timestamp with mergeable fields timestamp and mutate value if timestamp is newer
	// init field update
	update = &UpdatedField{Field: key, PrevValue: NullableBoolToInterface(s.BooleanValue)}
	previousTimestamp, exists := s.FieldsTimestamp[key]
	if !exists {
		s.BooleanValue = value
		s.UpdateFieldTimestamp(key, timestamp)
		update.NewValue = NullableBoolToInterface(value)
		return
	}
	// abort if a previous timestamp exists, and the current one is not provided
	if timestamp == nil {
		return nil
	}
	// abort if the current timestamp is older than the previous one
	if timestamp.Before(previousTimestamp) {
		return nil
	}
	s.BooleanValue = value
	s.UpdateFieldTimestamp(key, timestamp)
	update.NewValue = NullableBoolToInterface(value)
	return
}
func (s *CustomEvent) SetNonInteractive(value bool, timestamp *time.Time) (update *UpdatedField) {
	key := "label"
	// abort if values are equal
	if s.NonInteractive == value {
		return nil
	}
	// compare timestamp with mergeable fields timestamp and mutate value if timestamp is newer
	// init field update
	update = &UpdatedField{
		Field:     key,
		PrevValue: s.NonInteractive,
	}
	previousTimestamp, exists := s.FieldsTimestamp[key]
	if !exists {
		s.NonInteractive = value
		s.UpdateFieldTimestamp(key, timestamp)
		update.NewValue = value
		return
	}
	// abort if a previous timestamp exists, and the current one is not provided
	if timestamp == nil {
		return nil
	}
	// abort if the current timestamp is older than the previous one
	if timestamp.Before(previousTimestamp) {
		return nil
	}
	s.NonInteractive = value
	s.UpdateFieldTimestamp(key, timestamp)
	update.NewValue = value
	return
}

func (s *CustomEvent) SetExtraColumns(field string, value *AppItemField, timestamp *time.Time) (update *UpdatedField) {
	if s.ExtraColumns == nil {
		s.ExtraColumns = AppItemFields{}
	}

	// abort if field doesnt start with "app_" or "appx_"
	if !strings.HasPrefix(field, "app_") && !strings.HasPrefix(field, "appx_") {
		return nil
	}
	// check if field already exists
	previousValue, previousValueExists := s.ExtraColumns[field]

	if !previousValueExists {
		s.ExtraColumns[field] = value
		s.UpdateFieldTimestamp(field, timestamp)
		update = &UpdatedField{Field: field, PrevValue: previousValue, NewValue: value}
		return
	}

	// abort if values are equal
	if previousValue.Equals(value) {
		s.UpdateFieldTimestamp(field, timestamp)
		return nil
	}

	// compare timestamp with mergeable fields timestamp and mutate value if timestamp is newer
	// init field update
	update = &UpdatedField{Field: field, PrevValue: previousValue}
	previousTimestamp, previousTimestampExists := s.FieldsTimestamp[field]

	if !previousTimestampExists {
		s.ExtraColumns[field] = value
		s.UpdateFieldTimestamp(field, timestamp)
		update.NewValue = value
		return
	}
	// abort if a previous timestamp exists, and the current one is not provided
	if timestamp == nil {
		return nil
	}
	// abort if the current timestamp is older than the previous one
	if timestamp.Before(previousTimestamp) {
		return nil
	}
	s.ExtraColumns[field] = value
	s.UpdateFieldTimestamp(field, timestamp)
	update.NewValue = value
	return
}

// merges two sessions and returns the list of updated fields
func (fromCustomEvent *CustomEvent) MergeInto(toCustomEvent *CustomEvent, workspace *Workspace) (updatedFields []*UpdatedField) {

	updatedFields = []*UpdatedField{} // init

	if toCustomEvent.FieldsTimestamp == nil {
		toCustomEvent.FieldsTimestamp = FieldsTimestamp{}
	}

	if fieldUpdate := toCustomEvent.SetSessionID(fromCustomEvent.SessionID, TimePtr(fromCustomEvent.GetFieldDate("session_id"))); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toCustomEvent.SetDomainID(fromCustomEvent.DomainID, TimePtr(fromCustomEvent.GetFieldDate("domain_id"))); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toCustomEvent.SetLabel(fromCustomEvent.Label, TimePtr(fromCustomEvent.GetFieldDate("label"))); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toCustomEvent.SetStringValue(fromCustomEvent.StringValue, TimePtr(fromCustomEvent.GetFieldDate("string_value"))); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toCustomEvent.SetNumberValue(fromCustomEvent.NumberValue, TimePtr(fromCustomEvent.GetFieldDate("number_value"))); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toCustomEvent.SetBooleanValue(fromCustomEvent.BooleanValue, TimePtr(fromCustomEvent.GetFieldDate("boolean_value"))); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toCustomEvent.SetNonInteractive(fromCustomEvent.NonInteractive, TimePtr(fromCustomEvent.GetFieldDate("non_interactive"))); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	for key, value := range fromCustomEvent.ExtraColumns {
		if fieldUpdate := toCustomEvent.SetExtraColumns(key, value, TimePtr(fromCustomEvent.GetFieldDate(key))); fieldUpdate != nil {
			updatedFields = append(updatedFields, fieldUpdate)
		}
	}

	// set created_at + updated_at after merging fields
	toCustomEvent.SetCreatedAt(fromCustomEvent.CreatedAt)
	toCustomEvent.UpdatedAt = fromCustomEvent.UpdatedAt

	return
}

func NewCustomEvent(externalID string, userID string, label string, createdAt time.Time, updatedAt *time.Time) *CustomEvent {

	return &CustomEvent{
		ID:              ComputeCustomEventID(externalID),
		ExternalID:      externalID,
		UserID:          userID,
		CreatedAt:       createdAt,
		CreatedAtTrunc:  createdAt.Truncate(time.Hour),
		FieldsTimestamp: FieldsTimestamp{},

		Label: label,

		UpdatedAt:    updatedAt,
		ExtraColumns: AppItemFields{},
	}
}

func NewCustomEventFromDataLog(dataLog *DataLog, clockDifference time.Duration, workspace *Workspace) (customEvent *CustomEvent, err error) {

	result := gjson.Get(dataLog.Item, "custom_event")
	if !result.Exists() {
		return nil, eris.New("item has no custom_event object")
	}

	if result.Type != gjson.JSON {
		return nil, eris.New("item custom_event is not an object")
	}

	extraColumns := workspace.FindExtraColumnsForItemKind("custom_event")

	// init
	customEvent = &CustomEvent{
		UserID:          dataLog.UserID,
		FieldsTimestamp: FieldsTimestamp{},
		ExtraColumns:    AppItemFields{},
	}

	// loop over custom_event fields
	result.ForEach(func(key, value gjson.Result) bool {

		keyString := key.String()
		switch keyString {

		case "external_id":
			if value.Type == gjson.Null {
				err = eris.New("custom_event.external_id is required")
				return false
			}
			customEvent.ExternalID = value.String()
			customEvent.ID = ComputeCartID(customEvent.ExternalID)

		case "domain_id":
			if value.Type == gjson.Null {
				err = eris.New("custom_event.domain_id is required")
				return false
			}
			domain := value.String()
			customEvent.DomainID = &domain

		case "session_external_id":
			if value.Type == gjson.Null {
				customEvent.SessionID = nil
			} else {
				customEvent.SessionID = StringPtr(ComputeSessionID(value.String()))
			}

		case "created_at":
			if customEvent.CreatedAt, err = time.Parse(time.RFC3339, value.String()); err != nil {
				err = eris.Wrap(err, "custom_event.created_at")
				return false
			}

			// apply clock difference
			if customEvent.CreatedAt.After(time.Now()) {

				customEvent.CreatedAt = customEvent.CreatedAt.Add(clockDifference)
				if customEvent.CreatedAt.After(time.Now()) {
					err = eris.New("custom_event.created_at cannot be in the future")
					return false
				}
			}

			customEvent.CreatedAtTrunc = customEvent.CreatedAt.Truncate(time.Hour)

		case "updated_at":
			var updatedAt time.Time
			if updatedAt, err = time.Parse(time.RFC3339, value.String()); err != nil {
				err = eris.Wrap(err, "custom_event.updated_at")
				return false
			}

			if updatedAt.After(time.Now()) {
				// apply clock difference
				updatedAtModified := updatedAt.Add(clockDifference)
				updatedAt = updatedAtModified
				if updatedAt.After(time.Now()) {
					err = eris.New("custom_event.updated_at cannot be in the future")
					return false
				}
			}

			customEvent.UpdatedAt = &updatedAt

		case "label":
			if value.Type == gjson.Null {
				err = eris.New("custom_event.label is required")
				return false
			}

			customEvent.Label = strings.TrimSpace(value.String())

		case "string_value":
			if value.Type == gjson.Null {
				customEvent.StringValue = NewNullableString(nil)
			} else {
				stringValue := strings.TrimSpace(value.String())
				customEvent.StringValue = NewNullableString(&stringValue)
			}

		case "number_value":
			if value.Type == gjson.Null {
				customEvent.NumberValue = NewNullableFloat64(nil)
			} else {
				numberValue := value.Float()
				customEvent.NumberValue = NewNullableFloat64(&numberValue)
			}

		case "boolean_value":
			if value.Type == gjson.Null {
				customEvent.BooleanValue = NewNullableBool(nil)
			} else {
				booleanValue := value.Bool()
				customEvent.BooleanValue = NewNullableBool(&booleanValue)
			}

		case "non_interactive":
			if value.Type == gjson.Null {
				customEvent.NonInteractive = false
			} else {
				customEvent.NonInteractive = value.Bool()
			}

		default:
			// ignore other fields and handle app_ extra columns
			if (strings.HasPrefix(keyString, "app_") || strings.HasPrefix(keyString, "appx_")) && extraColumns != nil {

				// check if column exists in app table
				for _, col := range extraColumns {

					if col.Name == keyString {

						fieldValue, errExtract := ExtractFieldValueFromGJSON(col, value, 0)
						if errExtract != nil {
							err = eris.Wrapf(errExtract, "extract field %s", col.Name)
							return false
						}
						customEvent.ExtraColumns[col.Name] = fieldValue
					}
				}
			}
		}

		return true
	})

	if err != nil {
		return nil, err
	}
	if customEvent.DomainID == nil && dataLog.DomainID != nil {
		customEvent.DomainID = dataLog.DomainID
	}

	// use data import createdAt as updatedAt if not provided
	if customEvent.UpdatedAt == nil {
		customEvent.UpdatedAt = &customEvent.CreatedAt
	}

	// enrich custom_event with session and domain
	if dataLog.UpsertedSession != nil {
		if customEvent.SessionID == nil {
			customEvent.SessionID = &dataLog.UpsertedSession.ID
		}
		if customEvent.DomainID == nil {
			customEvent.DomainID = &dataLog.UpsertedSession.DomainID
		}
	}

	// Validation
	if customEvent.ExternalID == "" {
		return nil, eris.New("custom_event.external_id is required")
	}

	if customEvent.Label == "" {
		return nil, eris.New("custom_event.label is required")
	}

	// verify that domainID exists
	found := false
	for _, domain := range workspace.Domains {
		if domain.ID == *customEvent.DomainID {
			found = true
			break
		}
	}

	if !found {
		return nil, eris.New("custom_event domain_id invalid")
	}

	if customEvent.CreatedAt.IsZero() {
		return nil, eris.New("custom_event.created_at is required")
	}

	return customEvent, nil
}

func ComputeCustomEventID(externalID string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(externalID)))
}

var CustomEventSchema string = `CREATE TABLE IF NOT EXISTS custom_event (
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

	label VARCHAR(64) NOT NULL,
	string_value VARCHAR(512),
	number_value FLOAT,
	boolean_value BOOLEAN,
	non_interactive BOOLEAN NOT NULL DEFAULT FALSE,

	SORT KEY (created_at_trunc DESC),
	PRIMARY KEY (id, user_id),
	KEY (label) USING HASH,
	KEY (external_id) USING HASH,
	KEY (session_id) USING HASH,
	KEY (user_id) USING HASH, -- for merging
	SHARD KEY (user_id)
) COLLATE utf8mb4_general_ci;`

var CustomEventSchemaMYSQL string = `CREATE TABLE IF NOT EXISTS custom_event (
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

	label VARCHAR(64) NOT NULL,
	string_value VARCHAR(512),
	number_value FLOAT,
	boolean_value BOOLEAN,
	non_interactive BOOLEAN NOT NULL DEFAULT FALSE,

	-- SORT KEY (created_at_trunc),
	PRIMARY KEY (id, user_id),
	KEY (label) USING HASH,
	KEY (external_id) USING HASH,
	KEY (session_id) USING HASH,
	KEY (user_id) USING HASH -- for merging
	-- SHARD KEY (user_id)
) COLLATE utf8mb4_general_ci;`

func NewCustomEventCube() *CubeJSSchema {
	return &CubeJSSchema{
		Title:       "Custom events",
		Description: "Custom events",
		SQL:         "SELECT * FROM `custom_event`",
		// https://cube.dev/docs/schema/reference/joins
		Joins: map[string]CubeJSSchemaJoin{
			"Session": {
				Relationship: "many_to_one",
				SQL:          "${CUBE}.user_id = ${Session}.user_id AND ${CUBE}.session_id = ${Session}.id",
			},
			"User": {
				Relationship: "many_to_one",
				SQL:          "${CUBE}.user_id = ${User}.id",
			},
		},
		Segments: map[string]CubeJSSchemaSegment{},
		Measures: map[string]CubeJSSchemaMeasure{
			"count": {
				Type:        "count",
				Title:       "Count all",
				Description: "Count all",
			},
			"unique_users": {
				Type:        "countDistinct",
				SQL:         "user_id",
				Title:       "Unique users",
				Description: "Count distinct user_id",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"events_per_user": {
				Type:        "number",
				SQL:         "${count} / ${unique_users}",
				Title:       "CustomEvents per user",
				Description: "count / unique_users",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"number_value_sum": {
				Type:        "number",
				SQL:         "COALESCE(SUM(number_value), 0)",
				Title:       "Number value sum",
				Description: "SUM(number_value)",
			},
			"number_value_avg": {
				Type:        "number",
				SQL:         "COALESCE(AVG(number_value), 0)",
				Title:       "Number value average",
				Description: "AVG(number_value)",
			},
		},
		Dimensions: map[string]CubeJSSchemaDimension{
			"id": {
				SQL:         "id",
				Type:        "string",
				PrimaryKey:  true,
				Title:       "CustomEvent ID",
				Description: "field: id",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"domain_id": {
				SQL:         "domain_id",
				Type:        "string",
				Title:       "Domain ID",
				Description: "field: domain_id",
			},
			"session_id": {
				SQL:         "session_id",
				Type:        "string",
				Title:       "Session ID",
				Description: "field: session_id",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"user_id": {
				SQL:         "user_id",
				Type:        "string",
				Title:       "User ID",
				Description: "field: user_id",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"created_at": {
				SQL:         "created_at",
				Type:        "time",
				Title:       "Created at",
				Description: "field: created_at",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"created_at_trunc": {
				SQL:         "created_at_trunc",
				Type:        "time",
				Title:       "Created at (truncated to hour)",
				Description: "field: created_at_trunc",
			},
			"label": {
				SQL:         "label",
				Type:        "number",
				Title:       "Label",
				Description: "field: label",
			},
			"string_value": {
				SQL:         "string_value",
				Type:        "string",
				Title:       "String value",
				Description: "field: string_value",
			},
			"number_value": {
				SQL:         "number_value",
				Type:        "number",
				Title:       "Number value",
				Description: "field: number_value",
			},
			"boolean_value": {
				SQL:         "boolean_value",
				Type:        "number",
				Title:       "Boolean value",
				Description: "field: boolean_value",
			},
			"non_interactive": {
				SQL:         "non_interactive",
				Type:        "number",
				Title:       "Non-interactive",
				Description: "field: non_interactive",
			},
		},
	}
}
