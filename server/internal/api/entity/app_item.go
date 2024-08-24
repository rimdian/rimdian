package entity

import (
	"bytes"
	"crypto/sha1"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/golang-module/carbon/v2"
	"github.com/rotisserie/eris"
	"github.com/tidwall/gjson"
)

var (
	ErrAppItemRequired = eris.New("app item is required")

	None = "none"

	// computed fields should be excluded from SELECT/INSERT while cloning rows
	AppItemComputedFields []string = []string{
		"created_at_trunc",
	}
)

type AppItemFields map[string]*AppItemField

func (c AppItemFields) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (x *AppItemFields) Scan(val interface{}) error {

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

func (fields AppItemFields) GetExternalID() *string {
	field, ok := fields["external_id"]
	if !ok || field.StringValue.IsNull || field.StringValue.String == "" {
		return nil
	}
	return &field.StringValue.String
}

func (fields AppItemFields) GetCreatedAt() *time.Time {
	field, ok := fields["created_at"]
	if !ok || field.TimeValue.IsNull || field.TimeValue.Time.IsZero() {
		return nil
	}
	return &field.TimeValue.Time
}

func (fields AppItemFields) GetUpdatedAt() *time.Time {
	field, ok := fields["updated_at"]
	if !ok || field.TimeValue.IsNull || field.TimeValue.Time.IsZero() {
		return nil
	}
	return &field.TimeValue.Time
}

func (fields AppItemFields) GetUserID() *string {
	field, ok := fields["user_id"]
	if !ok || field.StringValue.IsNull || field.StringValue.String == "" {
		return nil
	}
	return &field.StringValue.String
}

type AppItemField struct {
	Name         string          `json:"name"`
	Type         string          `json:"type"`
	BoolValue    NullableBool    `json:"bool_value,omitempty"`
	Float64Value NullableFloat64 `json:"float64_value,omitempty"`
	StringValue  NullableString  `json:"string_value,omitempty"`
	TimeValue    NullableTime    `json:"time_value,omitempty"`
	JSONValue    NullableJSON    `json:"json_value,omitempty"`
}

func (field *AppItemField) Equals(other *AppItemField) bool {
	if field == nil && other == nil {
		return true
	}
	if field == nil || other == nil {
		return false
	}
	if field.Type != other.Type {
		return false
	}
	switch field.Type {
	case ColumnTypeNumber:
		if field.Float64Value.IsNull && other.Float64Value.IsNull {
			return true
		}
		if field.Float64Value.IsNull || other.Float64Value.IsNull {
			return false
		}
		return field.Float64Value.Float64 == other.Float64Value.Float64

	case ColumnTypeVarchar, ColumnTypeLongText:
		if field.StringValue.IsNull && other.StringValue.IsNull {
			return true
		}
		if field.StringValue.IsNull || other.StringValue.IsNull {
			return false
		}
		return field.StringValue.String == other.StringValue.String

	case ColumnTypeBoolean:
		if field.BoolValue.IsNull && other.BoolValue.IsNull {
			return true
		}
		if field.BoolValue.IsNull || other.BoolValue.IsNull {
			return false
		}
		return field.BoolValue.Bool == other.BoolValue.Bool

	case ColumnTypeDatetime:
		if field.TimeValue.IsNull && other.TimeValue.IsNull {
			return true
		}
		if field.TimeValue.IsNull || other.TimeValue.IsNull {
			return false
		}
		return field.TimeValue.Time.Equal(other.TimeValue.Time)

	case ColumnTypeJSON:
		// if field.JSONValue == nil && other.JSONValue == nil {
		// 	return true
		// }
		// if field.JSONValue == nil || other.JSONValue == nil {
		// 	return false
		// }
		if field.JSONValue.IsNull && other.JSONValue.IsNull {
			return true
		}
		if field.JSONValue.IsNull || other.JSONValue.IsNull {
			return false
		}
		return AreEqualJSON(field.JSONValue.JSON, other.JSONValue.JSON)

	default:
		return false
	}
}

// ToInterface() is called for API responses
func (field *AppItemField) ToInterface() interface{} {
	if field == nil {
		return nil
	}
	switch field.Type {
	case ColumnTypeNumber:
		if field.Float64Value.IsNull {
			return nil
		}
		return field.Float64Value.Float64

	case ColumnTypeVarchar, ColumnTypeLongText:
		if field.StringValue.IsNull {
			return nil
		}
		return field.StringValue.String

	case ColumnTypeBoolean:
		if field.BoolValue.IsNull {
			return nil
		}
		return field.BoolValue.Bool

	case ColumnTypeDatetime:
		if field.TimeValue.IsNull {
			return nil
		}
		return field.TimeValue.Time.Format(time.RFC3339)

	case ColumnTypeJSON:
		if field.JSONValue.IsNull {
			return nil
		}
		var value interface{}
		if err := json.Unmarshal(field.JSONValue.JSON, &value); err != nil {
			return nil
		}
		return value

	default:
		return false
	}
}

func (field *AppItemField) InitForBool() {
	field.Type = ColumnTypeBoolean
	field.BoolValue = NullableBool{IsNull: true}
}

func (field *AppItemField) InitForFloat64() {
	field.Type = ColumnTypeNumber
	field.Float64Value = NullableFloat64{IsNull: true}
}

func (field *AppItemField) InitForString() {
	field.Type = ColumnTypeVarchar
	field.StringValue = NullableString{IsNull: true}
}

func (field *AppItemField) InitForTime() {
	field.Type = ColumnTypeDatetime
	field.TimeValue = NullableTime{IsNull: true}
}

func (field *AppItemField) InitForJSON() {
	field.Type = ColumnTypeJSON
	field.JSONValue = NullableJSON{IsNull: true}
}

// init null values on marshal json
func (field *AppItemField) MarshalJSON() ([]byte, error) {

	// if field.Name == "fields_timestamp" {
	// 	log.Printf("AppItemField.MarshalJSON: %+v\n", string(field.JSONValue.JSON))
	// }

	// use an alias to avoid recursion
	type Alias AppItemField
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(field),
	}

	// make sure null values are initialized
	switch aux.Type {
	case ColumnTypeNumber:
		// if aux.Float64Value == nil {
		// 	aux.Float64Value = &NullableFloat64{IsNull: true}
		// }

	case ColumnTypeVarchar, ColumnTypeLongText:
		// if aux.StringValue == nil {
		// 	aux.StringValue = &NullableString{IsNull: true}
		// }

	case ColumnTypeBoolean:
		// if aux.BoolValue == nil {
		// 	aux.BoolValue = &NullableBool{IsNull: true}
		// }

	case ColumnTypeDate, ColumnTypeDatetime, ColumnTypeTimestamp:
		// if aux.TimeValue == nil {
		// 	aux.TimeValue = &NullableTime{IsNull: true}
		// }

	case ColumnTypeJSON:
		// if aux.JSONValue == nil {
		// 	aux.JSONValue = &NullableJSON{IsNull: true}
		// }

	default:
		return nil, eris.New("AppItemField.MarshalJSON: unknown custom field type")
	}

	return json.Marshal(aux)
}

// unmarshal
func (field *AppItemField) UnmarshalJSON(data []byte) error {
	type Alias AppItemField
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(field),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return eris.Wrapf(err, "data %v", string(data))
	}

	// make sure null values are initialized
	switch field.Type {
	case ColumnTypeNumber:
		// if field.Float64Value == nil {
		// 	field.Float64Value = &NullableFloat64{IsNull: true}
		// }

	case ColumnTypeVarchar, ColumnTypeLongText:
		// if field.StringValue == nil {
		// 	field.StringValue = &NullableString{IsNull: true}
		// }

	case ColumnTypeBoolean:
		// if field.BoolValue == nil {
		// 	field.BoolValue = &NullableBool{IsNull: true}
		// }

	case ColumnTypeDate, ColumnTypeDatetime, ColumnTypeTimestamp:
		// if field.TimeValue == nil {
		// 	field.TimeValue = &NullableTime{IsNull: true}
		// }

	case ColumnTypeJSON:
		// if field.JSONValue == nil {
		// 	field.JSONValue = &NullableJSON{IsNull: true}
		// }

	default:
		return eris.New("AppItemField.UnmarshalJSON: unknown custom field type")
	}

	return nil
}

func (field *AppItemField) Validate(fieldDefinition *TableColumn) error {
	switch fieldDefinition.Type {
	case ColumnTypeNumber:
		if fieldDefinition.IsRequired && field.Float64Value.IsNull {
			return fmt.Errorf("%v is required", field.Name)
		}

	case ColumnTypeVarchar, ColumnTypeLongText:
		if fieldDefinition.IsRequired && field.StringValue.IsNull {
			return fmt.Errorf("%v is required", field.Name)
		}

	case ColumnTypeBoolean:
		if fieldDefinition.IsRequired && field.BoolValue.IsNull {
			return fmt.Errorf("%v is required", field.Name)
		}

	case ColumnTypeDate, ColumnTypeDatetime, ColumnTypeTimestamp:
		if fieldDefinition.IsRequired && field.TimeValue.IsNull {
			return fmt.Errorf("%v is required", field.Name)
		}

	case ColumnTypeJSON:
		// TEST
		// if fieldDefinition.IsRequired && (field.JSONValue == nil || field.JSONValue.IsNull) {
		if fieldDefinition.IsRequired && field.JSONValue.IsNull {
			return fmt.Errorf("%v is required", field.Name)
		}

	default:
	}

	if field.Name == "external_id" && fieldDefinition.Type != ColumnTypeVarchar {
		return fmt.Errorf("external_id must be a varchar")
	}

	if field.Name == "created_at" && fieldDefinition.Type != ColumnTypeDatetime {
		return fmt.Errorf("created_at must be a datetime")
	}

	if field.Name == "updated_at" && fieldDefinition.Type != ColumnTypeDatetime {
		return fmt.Errorf("updated_at must be a datetime")
	}

	return nil
}

// value interface for sql driver
func (field *AppItemField) Value() (driver.Value, error) {
	switch field.Type {
	case ColumnTypeNumber:
		if field.Float64Value.IsNull {
			return nil, nil
		}
		return field.Float64Value.Float64, nil

	case ColumnTypeVarchar, ColumnTypeLongText:
		if field.StringValue.IsNull {
			return nil, nil
		}

		return field.StringValue.String, nil

	case ColumnTypeBoolean:
		if field.BoolValue.IsNull {
			return nil, nil
		}

		return field.BoolValue.Bool, nil

	case ColumnTypeDate, ColumnTypeDatetime, ColumnTypeTimestamp:
		if field.TimeValue.IsNull {
			return nil, nil
		}

		return field.TimeValue.Time, nil

	case ColumnTypeJSON:
		if field.JSONValue.IsNull {
			return nil, nil
		}

		return field.JSONValue.JSON, nil

	default:
		return nil, eris.New("AppItemField.Value: unknown custom field type")
	}
}

type AppItem struct {
	ID               string          `db:"id" json:"id"`
	ExternalID       string          `db:"external_id" json:"external_id"`
	UserID           string          `db:"user_id" json:"user_id,omitempty"`
	CreatedAt        time.Time       `db:"created_at" json:"created_at"`
	CreatedAtTrunc   time.Time       `db:"created_at_trunc" json:"created_at_trunc"` // date truncated at hour, used to optimize columnstore storage and queries
	DBCreatedAt      time.Time       `db:"db_created_at" json:"db_created_at"`
	DBUpdatedAt      time.Time       `db:"db_updated_at" json:"db_updated_at"`
	MergedFromUserID *string         `db:"merged_from_user_id" json:"merged_from_user_id,omitempty"`
	FieldsTimestamp  FieldsTimestamp `db:"fields_timestamp" json:"fields_timestamp"`

	Fields AppItemFields `db:"-" json:"app_item_fields"`
	// optional reserved fields
	// DomainID  *string `db:"domain_id" json:"domain_id,omitempty"`
	// SessionID *string `db:"session_id" json:"session_id,omitempty"`

	// Not persisted in DB:
	Kind      string     `db:"-" json:"kind"`
	UpdatedAt *time.Time `db:"-" json:"-"` // used to merge fields and append item_timeline at the right time
}

// DB items (app_item, session, user, app item...) that are sent back
// to the client should be flattened, like they are stored in the DB
func (item *AppItem) ToFlatInterface() (object map[string]interface{}) {
	object = map[string]interface{}{
		"id":               item.ID,
		"external_id":      item.ExternalID,
		"kind":             item.Kind,
		"created_at":       item.CreatedAt,
		"created_at_trunc": item.CreatedAtTrunc,
		"db_created_at":    item.DBCreatedAt,
		"db_updated_at":    item.DBUpdatedAt,
		"fields_timestamp": item.FieldsTimestamp,
	}

	if item.UserID != None {
		object["user_id"] = item.UserID

		if item.MergedFromUserID != nil && *item.MergedFromUserID != "" {
			object["merged_from_user_id"] = *item.MergedFromUserID
		}
	}

	for _, field := range item.Fields {
		value := field.ToInterface()
		if value != nil {
			object[field.Name] = value
		}
	}

	return object
}

func (s *AppItem) GetFieldDate(field string) time.Time {
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

func (s *AppItem) SetCreatedAt(value time.Time) {
	// update if current value is older
	if value.Before(s.CreatedAt) {
		s.CreatedAt = value
		s.CreatedAtTrunc = value.Truncate(time.Hour)
	}
}

// merges two items and returns the list of updated fields
func (fromAppItem *AppItem) MergeInto(toAppItem *AppItem) (updatedFields []*UpdatedField) {

	updatedFields = []*UpdatedField{} // init

	for fieldName, field := range fromAppItem.Fields {

		// ignore updated_at is not persisted
		if fieldName == "updated_at" {
			continue
		}

		// find field
		var existingField *AppItemField

		if _, ok := toAppItem.Fields[fieldName]; ok {
			existingField = toAppItem.Fields[fieldName]
		}

		// if field didnt exist yet, add it
		if existingField == nil {
			toAppItem.Fields[field.Name] = field

			updatedFields = append(updatedFields, &UpdatedField{
				Field:     field.Name,
				PrevValue: nil,
				NewValue:  field.ToInterface(),
			})

			toAppItem.FieldsTimestamp[field.Name] = fromAppItem.GetFieldDate(field.Name)
		} else {
			// field already exists
			existingValueTimestamp := toAppItem.GetFieldDate(field.Name)
			timestamp := fromAppItem.GetFieldDate(field.Name)

			// check if values are equal
			switch field.Type {
			case ColumnTypeBoolean:
				// if existingField.BoolValue == nil {
				// 	log.Printf("app item field %v should not be nil", field.Name)
				// 	continue
				// }
				// abort if values are equal
				if existingField.BoolValue.IsNull == field.BoolValue.IsNull && existingField.BoolValue.Bool == field.BoolValue.Bool {
					continue
				}
				// abort if existing value is newer
				if existingValueTimestamp.After(timestamp) {
					continue
				}
				// the value might be set for the first time
				// so we set the value without producing a field update
				if existingValueTimestamp.Equal(timestamp) {
					existingField.BoolValue = field.BoolValue
					continue
				}

				updatedFields = append(updatedFields, &UpdatedField{
					Field:     field.Name,
					PrevValue: existingField.ToInterface(),
					NewValue:  field.ToInterface(),
				})
				existingField.BoolValue = field.BoolValue
				toAppItem.FieldsTimestamp[field.Name] = timestamp
				continue

			case ColumnTypeVarchar, ColumnTypeLongText:
				// if existingField.StringValue == nil {
				// 	continue
				// }
				if existingField.StringValue.IsNull == field.StringValue.IsNull && existingField.StringValue.String == field.StringValue.String {
					continue
				}
				// abort if existing value is newer
				if existingValueTimestamp.After(timestamp) {
					continue
				}
				// the value might be set for the first time
				// so we set the value without producing a field update
				if existingValueTimestamp.Equal(timestamp) {
					existingField.StringValue = field.StringValue
					continue
				}

				updatedFields = append(updatedFields, &UpdatedField{
					Field:     field.Name,
					PrevValue: existingField.ToInterface(),
					NewValue:  field.ToInterface(),
				})
				existingField.StringValue = field.StringValue
				toAppItem.FieldsTimestamp[field.Name] = timestamp
				continue

			case ColumnTypeNumber:

				// if existingField.Float64Value == nil {
				// 	log.Printf("app item field %v should not be nil", field.Name)
				// 	continue
				// }
				// abort if values are equal
				if existingField.Float64Value.IsNull == field.Float64Value.IsNull && existingField.Float64Value.Float64 == field.Float64Value.Float64 {
					continue
				}
				// abort if existing value is newer
				if existingValueTimestamp.After(timestamp) {
					continue
				}
				// the value might be set for the first time
				// so we set the value without producing a field update
				if existingValueTimestamp.Equal(timestamp) {
					existingField.Float64Value = field.Float64Value
					continue
				}

				updatedFields = append(updatedFields, &UpdatedField{
					Field:     field.Name,
					PrevValue: existingField.ToInterface(),
					NewValue:  field.ToInterface(),
				})
				existingField.Float64Value = field.Float64Value
				toAppItem.FieldsTimestamp[field.Name] = timestamp
				continue

			case ColumnTypeDate, ColumnTypeDatetime, ColumnTypeTimestamp:
				// if existingField.TimeValue == nil {
				// 	log.Printf("app item field %v should not be nil", field.Name)
				// 	continue
				// }
				// abort if values are equal
				if existingField.TimeValue.IsNull == field.TimeValue.IsNull && existingField.TimeValue.Time.Equal(field.TimeValue.Time) {
					continue
				}
				// abort if existing value is newer
				if existingValueTimestamp.After(timestamp) {
					continue
				}
				// the value might be set for the first time
				// so we set the value without producing a field update
				if existingValueTimestamp.Equal(timestamp) {
					existingField.TimeValue = field.TimeValue
					continue
				}

				updatedFields = append(updatedFields, &UpdatedField{
					Field:     field.Name,
					PrevValue: existingField.ToInterface(),
					NewValue:  field.ToInterface(),
				})
				existingField.TimeValue = field.TimeValue
				toAppItem.FieldsTimestamp[field.Name] = timestamp

				continue

			case ColumnTypeJSON:
				// if existingField.JSONValue == nil {
				// 	log.Printf("app item field %v should not be nil", field.Name)
				// 	continue
				// }
				// abort if values are equal
				if existingField.JSONValue.IsNull == field.JSONValue.IsNull && AreEqualJSON(existingField.JSONValue.JSON, field.JSONValue.JSON) {
					continue
				}
				// abort if existing value is newer
				if existingValueTimestamp.After(timestamp) {
					continue
				}
				// the value might be set for the first time
				// so we set the value without producing a field update
				if existingValueTimestamp.Equal(timestamp) {
					existingField.JSONValue = field.JSONValue
					continue
				}

				updatedFields = append(updatedFields, &UpdatedField{
					Field:     field.Name,
					PrevValue: existingField.ToInterface(),
					NewValue:  field.ToInterface(),
				})
				existingField.JSONValue = field.JSONValue
				toAppItem.FieldsTimestamp[field.Name] = timestamp

				continue
			default:
			}
		}
	}

	// UpdatedAt is the timeOfEvent for RTs
	toAppItem.UpdatedAt = fromAppItem.UpdatedAt
	// priority to oldest date
	toAppItem.SetCreatedAt(fromAppItem.CreatedAt)

	return
}

func ComputeAppItemID(externalID string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(externalID)))
}

func NewAppItem(kind string, externalID string, userID string, createdAt time.Time, updatedAt *time.Time) *AppItem {
	item := &AppItem{
		ID:              ComputeAppItemID(externalID),
		ExternalID:      externalID,
		Kind:            kind,
		UserID:          userID,
		CreatedAt:       createdAt,
		CreatedAtTrunc:  createdAt.Truncate(time.Hour),
		Fields:          AppItemFields{},
		FieldsTimestamp: FieldsTimestamp{},
		UpdatedAt:       updatedAt,
	}

	return item
}

func NewAppItemFromDataLog(dataLog *DataLog, clockDifference time.Duration, workspace *Workspace) (appItem *AppItem, err error) {

	result := gjson.Get(dataLog.Item, dataLog.Kind)
	if !result.Exists() {
		return nil, eris.Errorf("item %v is required", dataLog.Kind)
	}

	if result.Type != gjson.JSON {
		return nil, eris.Errorf("item %v must be a JSON object", dataLog.Kind)
	}

	// check if app table exists
	tableDefinition := workspace.FindAppTableDefinitionForItem(dataLog.Kind)

	if tableDefinition == nil {
		return nil, eris.Errorf("table %v not found", dataLog.Kind)
	}

	if tableDefinition.HasUserColumn() && dataLog.UpsertedUser == nil {
		return nil, eris.Errorf("user is required for table %v", dataLog.Kind)
	}

	// extract app item fields from JSON
	fields := AppItemFields{}
	var updatedAt *time.Time

	// loop over raw item fields, find their definition and extract the value
	result.ForEach(func(key, value gjson.Result) bool {

		fieldName := key.String()

		// updated_at is not in the definition
		if fieldName == "updated_at" {
			// parse time
			parsed := carbon.Parse(value.String())
			if parsed.Error != nil {
				err = fmt.Errorf("field updated_at is not a valid date, got %v, err: %v", value.String(), err)
				return false
			}

			t := parsed.StdTime()

			// t, errParse := time.Parse(time.RFC3339Nano, value.String())
			// if errParse != nil {
			// 	err = fmt.Errorf("field updated_at is not a valid date, got %v, err: %v", value.String(), err)
			// 	return false
			// }

			// apply clock difference on system fields
			// apply clockDifference if date in future
			if t.After(time.Now()) {
				t = t.Add(clockDifference)
			}

			updatedAt = &t
		}

		// find field definition
		for _, fieldDefinition := range tableDefinition.Columns {

			if fieldDefinition.Name == fieldName {

				fieldValue, errExtract := ExtractFieldValueFromGJSON(fieldDefinition, value, clockDifference)

				if errExtract != nil {
					log.Printf("ExtractAppItemAndValidate, err: %v.%v: %v\n", dataLog.Kind, fieldName, errExtract)
					err = eris.Wrapf(errExtract, "%v.%v", dataLog.Kind, fieldName)
					return false
				}

				fields[fieldName] = fieldValue
			}
		}

		return true
	})

	if err != nil {
		return nil, err
	}

	externalID := fields.GetExternalID()

	if externalID == nil {
		return nil, eris.Errorf("%v.external_id is required", dataLog.Kind)
	}

	createdAt := fields.GetCreatedAt()

	if createdAt == nil {
		log.Printf("fields: %+v\n", fields)
		return nil, eris.Errorf("%v.created_at is required (nil)", dataLog.Kind)
	}

	if createdAt.IsZero() {
		return nil, eris.Errorf("%v.created_at is required (zero)", dataLog.Kind)
	}

	// use createdAt as updatedAt if not provided
	if updatedAt == nil {
		updatedAt = createdAt
	}

	userID := None
	if dataLog.UpsertedUser != nil {
		userID = dataLog.UpsertedUser.ID
	}

	appItem = &AppItem{
		ID:              ComputeAppItemID(*externalID),
		ExternalID:      *externalID,
		Kind:            dataLog.Kind,
		UserID:          userID,
		CreatedAt:       *createdAt,
		CreatedAtTrunc:  createdAt.Truncate(time.Hour),
		Fields:          fields,
		FieldsTimestamp: FieldsTimestamp{},
		UpdatedAt:       updatedAt,
	}

	// validate fields
	for _, field := range fields {

		// ignore computed fields
		if govalidator.IsIn(field.Name, ComputedColumns...) {
			continue
		}

		// find field definition
		var fieldDefinition *TableColumn
		for _, column := range tableDefinition.Columns {
			if column.Name == field.Name {
				fieldDefinition = column
				break
			}
		}
		if fieldDefinition == nil {
			return nil, eris.Errorf("%v.%v is not a valid field", tableDefinition.Name, field.Name)
		}
		// validate field type
		if err := field.Validate(fieldDefinition); err != nil {
			return nil, eris.Wrapf(err, "%v.%v ", tableDefinition.Name, err.Error())
		}
	}

	// check that required fields are present
	for _, column := range tableDefinition.Columns {

		// ignore computed fields
		if govalidator.IsIn(column.Name, ComputedColumns...) {
			continue
		}

		if column.IsRequired {
			var found bool
			for _, field := range fields {
				if field.Name == column.Name {
					found = true
					break
				}
			}
			if !found {
				return nil, eris.Errorf("%v.%v is required", tableDefinition.Name, column.Name)
			}
		}
	}

	return appItem, nil
}
