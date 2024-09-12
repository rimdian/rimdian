package entity

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/rotisserie/eris"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var (
	// computed fields should be excluded from SELECT/INSERT while cloning rows
	PostviewComputedFields []string = []string{
		"created_at_trunc",
		"year",
		"month",
		"month_day",
		"week_day",
		"hour",
	}

	ErrPostviewRequired          = eris.New("postview is required")
	ErrPostviewIDRequired        = eris.New("postview id or external id is required")
	ErrPostviewUserIDRequired    = eris.New("postview user id is required")
	ErrPostviewCreatedAtRequired = eris.New("postview created at is required")
	ErrPostviewTimezoneInvalid   = eris.New("postview timezone is not valid")
)

type Postview struct {
	ID               string          `db:"id" json:"id"`
	ExternalID       string          `db:"external_id" json:"external_id"`
	UserID           string          `db:"user_id" json:"user_id"`
	DeviceID         *string         `db:"device_id" json:"device_id,omitempty"`
	CreatedAt        time.Time       `db:"created_at" json:"created_at"`
	CreatedAtTrunc   time.Time       `db:"created_at_trunc" json:"created_at_trunc"` // date truncated at hour, used to optimize columnstore storage and queries
	DBCreatedAt      time.Time       `db:"db_created_at" json:"db_created_at"`
	DBUpdatedAt      time.Time       `db:"db_updated_at" json:"db_updated_at"`
	IsDeleted        bool            `db:"is_deleted" json:"is_deleted,omitempty"` // deleting rows in transactions cause deadlocks in singlestore, we use an update
	MergedFromUserID *string         `db:"merged_from_user_id" json:"merged_from_user_id,omitempty"`
	FieldsTimestamp  FieldsTimestamp `db:"fields_timestamp" json:"fields_timestamp"`

	// localized datetime parts for cohorts purpose
	// Timezone *string `db:"timezone" json:"timezone"`   // field set while repo upsert if missing
	// Year     int     `db:"year" json:"year"`           // computed from created_at + timezone
	// Month    int     `db:"month" json:"month"`         // computed from created_at + timezone
	// MonthDay int     `db:"month_day" json:"month_day"` // computed from created_at + timezone
	// WeekDay  int     `db:"week_day" json:"week_day"`   // computed from created_at + timezone
	// Hour     int     `db:"hour" json:"hour"`           // computed from created_at + timezone

	// channel mapping, require for conversion paths:
	ChannelOriginID string `db:"channel_origin_id" json:"channel_origin_id"` // db computed field: source / medium( / campaign)
	ChannelID       string `db:"channel_id" json:"channel_id"`               // field set while repo upsert with utm_ params
	ChannelGroupID  string `db:"channel_group_id" json:"channel_group_id"`   // field set while repo upsert with utm_ params

	// utm_ parameters equivalent:
	UTMSource *NullableString `db:"utm_source" json:"utm_source"` // utm_source
	UTMMedium *NullableString `db:"utm_medium" json:"utm_medium"` // utm_medium
	// UTMID       *NullableString `db:"utm_id" json:"utm_id,omitempty"`             // utm_id
	// UTMIDFrom   *NullableString `db:"utm_id_from" json:"utm_id_from,omitempty"`   // gclid | fbclid | cmid...
	UTMCampaign *NullableString `db:"utm_campaign" json:"utm_campaign,omitempty"` // utm_campaign
	UTMContent  *NullableString `db:"utm_content" json:"utm_content,omitempty"`   // utm_content
	UTMTerm     *NullableString `db:"utm_term" json:"utm_term,omitempty"`         // utm_term

	Country *NullableString `db:"country" json:"country,omitempty"`
	Region  *NullableString `db:"region" json:"region,omitempty"`

	// attribution, set when the postview contributed to a conversion
	ConversionType             *string    `db:"conversion_type" json:"conversion_type,omitempty"` // order | lead | subscription
	ConversionID               *string    `db:"conversion_id" json:"conversion_id,omitempty"`
	ConversionExternalID       *string    `db:"conversion_external_id" json:"conversion_external_id,omitempty"`
	ConversionAt               *time.Time `db:"conversion_at" json:"conversion_at,omitempty"`
	ConversionAmount           *int64     `db:"conversion_amount" json:"conversion_amount,omitempty"` // integer stored in cents. 1,500.21 -> 150021
	LinearAmountAttributed     *int64     `db:"linear_amount_attributed" json:"linear_amount_attributed,omitempty"`
	LinearPercentageAttributed *int64     `db:"linear_percentage_attributed" json:"linear_percentage_attributed,omitempty"`
	TimeToConversion           *int64     `db:"time_to_conversion" json:"time_to_conversion,omitempty"`
	IsFirstConversion          *bool      `db:"is_first_conversion" json:"is_first_conversion,omitempty"` // used to compute acquisition | retention stats
	Role                       *int64     `db:"role" json:"role,omitempty"`                               // 0: alone, 1: initiator, 2: assisting, 3: closer

	// Not persisted in DB:
	UpdatedAt    *time.Time    `db:"-" json:"-"` // used to merge fields and append item_timeline at the right time
	ExtraColumns AppItemFields `db:"-" json:"-"` // converted into "app_xxx" fields when marshaling JSON
}

func (s *Postview) GetFieldDate(field string) time.Time {
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

func (x *Postview) Validate() error {
	if x.ID == "" && x.ExternalID == "" {
		return ErrPostviewIDRequired
	}
	if x.UserID == "" {
		return ErrPostviewUserIDRequired
	}
	if x.CreatedAt.IsZero() {
		return ErrPostviewCreatedAtRequired
	}
	// if x.Timezone != nil && !govalidator.IsIn(*x.Timezone, common.Timezones...) {
	// 	return ErrPostviewTimezoneInvalid
	// }

	return nil
}

func (s *Postview) UpdateFieldTimestamp(field string, timestamp *time.Time) {
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
func (s *Postview) SetCreatedAt(value time.Time) {
	// update if current value is older
	if value.Before(s.CreatedAt) {
		s.CreatedAt = value
		s.CreatedAtTrunc = value.Truncate(time.Hour)
	}
}

func (s *Postview) SetDeviceID(value *string, timestamp time.Time) (update *UpdatedField) {
	key := "device_id"
	// value cant be null
	if value == nil {
		return nil
	}
	// abort if values are equal
	if StringsEqual(s.DeviceID, value) {
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
		s.DeviceID = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: StringPointerToInterface(s.DeviceID),
		NewValue:  StringPointerToInterface(value),
	}
	s.DeviceID = value
	s.FieldsTimestamp[key] = timestamp
	return
}

// field set by attribution algo, dont compare timestamps
func (s *Postview) SetChannelOriginID(value string) (update *UpdatedField) {
	key := "channel_origin_id"
	// abort if values are equal
	if s.ChannelOriginID == value {
		return nil
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: s.ChannelOriginID,
		NewValue:  value,
	}

	s.ChannelOriginID = value
	return
}
func (s *Postview) SetChannelID(value string) (update *UpdatedField) {
	key := "channel_id"
	// abort if values are equal
	if s.ChannelID == value {
		return nil
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: s.ChannelID,
		NewValue:  value,
	}

	s.ChannelID = value
	return
}
func (s *Postview) SetChannelGroupID(value string) (update *UpdatedField) {
	key := "channel_group_id"
	// abort if values are equal
	if s.ChannelGroupID == value {
		return nil
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: s.ChannelGroupID,
		NewValue:  value,
	}

	s.ChannelGroupID = value
	return
}

func (s *Postview) SetUTMSource(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "utm_source"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.UTMSource != nil && s.UTMSource.IsNull == value.IsNull && s.UTMSource.String == value.String {
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
		s.UTMSource = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.UTMSource),
		NewValue:  NullableStringToInterface(value),
	}
	s.UTMSource = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Postview) SetUTMMedium(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "utm_medium"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.UTMMedium != nil && s.UTMMedium.IsNull == value.IsNull && s.UTMMedium.String == value.String {
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
		s.UTMMedium = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.UTMMedium),
		NewValue:  NullableStringToInterface(value),
	}
	s.UTMMedium = value
	s.FieldsTimestamp[key] = timestamp
	return
}

func (s *Postview) SetUTMCampaign(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "utm_campaign"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.UTMCampaign != nil && s.UTMCampaign.IsNull == value.IsNull && s.UTMCampaign.String == value.String {
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
		s.UTMCampaign = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.UTMCampaign),
		NewValue:  NullableStringToInterface(value),
	}
	s.UTMCampaign = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Postview) SetUTMContent(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "utm_content"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.UTMContent != nil && s.UTMContent.IsNull == value.IsNull && s.UTMContent.String == value.String {
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
		s.UTMContent = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.UTMContent),
		NewValue:  NullableStringToInterface(value),
	}
	s.UTMContent = value
	s.FieldsTimestamp[key] = timestamp
	return
}

func (s *Postview) SetUTMTerm(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "utm_term"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.UTMTerm != nil && s.UTMTerm.IsNull == value.IsNull && s.UTMTerm.String == value.String {
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
		s.UTMTerm = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.UTMTerm),
		NewValue:  NullableStringToInterface(value),
	}
	s.UTMTerm = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Postview) SetCountry(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "country"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.Country != nil && s.Country.IsNull == value.IsNull && s.Country.String == value.String {
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
		s.Country = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.Country),
		NewValue:  NullableStringToInterface(value),
	}
	s.Country = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Postview) SetRegion(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "region"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.Region != nil && s.Region.IsNull == value.IsNull && s.Region.String == value.String {
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
		s.Region = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.Region),
		NewValue:  NullableStringToInterface(value),
	}
	s.Region = value
	s.FieldsTimestamp[key] = timestamp
	return
}

// fields set by attribution algorithm, dont compare timestamps
func (s *Postview) SetConversionType(value *string) (update *UpdatedField) {
	key := "conversion_type"
	// abort if values are equal
	if StringsEqual(value, s.ConversionType) {
		return nil
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: StringPointerToInterface(s.ConversionType),
		NewValue:  StringPointerToInterface(value),
	}
	s.ConversionType = value
	return
}
func (s *Postview) SetConversionID(value *string) (update *UpdatedField) {
	key := "conversion_id"
	// abort if values are equal
	if StringsEqual(value, s.ConversionID) {
		return nil
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: StringPointerToInterface(s.ConversionID),
		NewValue:  StringPointerToInterface(value),
	}
	s.ConversionID = value
	return
}
func (s *Postview) SetConversionExternalID(value *string) (update *UpdatedField) {
	key := "conversion_external_id"
	// abort if values are equal
	if StringsEqual(value, s.ConversionExternalID) {
		return nil
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: StringPointerToInterface(s.ConversionExternalID),
		NewValue:  StringPointerToInterface(value),
	}
	s.ConversionExternalID = value
	return
}
func (s *Postview) SetConversionAt(value *time.Time) (update *UpdatedField) {
	key := "conversion_at"
	// abort if values are equal
	if TimeEqual(value, s.ConversionAt) {
		return nil
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: TimePointerToInterface(s.ConversionAt),
		NewValue:  TimePointerToInterface(value),
	}
	s.ConversionAt = value
	return
}
func (s *Postview) SetConversionAmount(value *int64) (update *UpdatedField) {
	key := "conversion_amount"
	// abort if values are equal
	if Int64Equal(value, s.ConversionAmount) {
		return nil
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: Int64PointerToInterface(s.ConversionAmount),
		NewValue:  Int64PointerToInterface(value),
	}
	s.ConversionAmount = value
	return
}
func (s *Postview) SetLinearAmountAttributed(value *int64) (update *UpdatedField) {
	key := "linear_amount_attributed"
	// abort if values are equal
	if Int64Equal(value, s.LinearAmountAttributed) {
		return nil
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: Int64PointerToInterface(s.LinearAmountAttributed),
		NewValue:  Int64PointerToInterface(value),
	}
	s.LinearAmountAttributed = value
	return
}
func (s *Postview) SetLinearPercentageAttributed(value *int64) (update *UpdatedField) {
	key := "linear_percentage_attributed"
	// abort if values are equal
	if Int64Equal(value, s.LinearPercentageAttributed) {
		return nil
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: Int64PointerToInterface(s.LinearPercentageAttributed),
		NewValue:  Int64PointerToInterface(value),
	}
	s.LinearPercentageAttributed = value
	return
}
func (s *Postview) SetTimeToConversion(value *int64) (update *UpdatedField) {
	key := "time_to_conversion"
	// abort if values are equal
	if Int64Equal(value, s.TimeToConversion) {
		return nil
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: Int64PointerToInterface(s.TimeToConversion),
		NewValue:  Int64PointerToInterface(value),
	}
	s.TimeToConversion = value
	return
}
func (s *Postview) SetIsFirstConversion(value *bool) (update *UpdatedField) {
	key := "time_to_conversion"
	// abort if values are equal
	if BoolEqual(value, s.IsFirstConversion) {
		return nil
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: BoolPointerToInterface(s.IsFirstConversion),
		NewValue:  BoolPointerToInterface(value),
	}
	s.IsFirstConversion = value
	return
}
func (s *Postview) SetRole(value *int64) (update *UpdatedField) {
	key := "time_to_conversion"
	// abort if values are equal
	if Int64Equal(value, s.Role) {
		return nil
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: Int64PointerToInterface(s.Role),
		NewValue:  Int64PointerToInterface(value),
	}
	s.Role = value
	return
}

func (s *Postview) SetExtraColumns(field string, value *AppItemField, timestamp time.Time) (update *UpdatedField) {
	if s.ExtraColumns == nil {
		s.ExtraColumns = AppItemFields{}
	}

	// abort if field doesnt start with "app_"
	if !strings.HasPrefix(field, "app_") && !strings.HasPrefix(field, "appx_") {
		return nil
	}

	var prevValueInterface interface{}
	previousValue, previousValueExists := s.ExtraColumns[field]

	// abort if values are equal
	if previousValueExists {
		if previousValue.Equals(value) {
			return nil
		}
		prevValueInterface = previousValue.ToInterface()
	}

	existingValueTimestamp := s.GetFieldDate(field)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		s.ExtraColumns[field] = value
		return
	}
	update = &UpdatedField{
		Field:     field,
		PrevValue: prevValueInterface,
		NewValue:  value.ToInterface(),
	}
	s.ExtraColumns[field] = value
	s.FieldsTimestamp[field] = timestamp

	return
}

// merges two postviews and returns the list of updated fields
func (fromPostview *Postview) MergeInto(toPostview *Postview, workspace *Workspace) (updatedFields []*UpdatedField) {

	updatedFields = []*UpdatedField{} // init

	if toPostview.FieldsTimestamp == nil {
		toPostview.FieldsTimestamp = FieldsTimestamp{}
	}

	if fieldUpdate := toPostview.SetDeviceID(fromPostview.DeviceID, fromPostview.GetFieldDate("device_id")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	// if fieldUpdate := toPostview.SetTimezone(fromPostview.Timezone, fromPostview.GetFieldDate("timezone")); fieldUpdate != nil {
	// 	updatedFields = append(updatedFields, fieldUpdate)
	// }

	if fieldUpdate := toPostview.SetChannelOriginID(fromPostview.ChannelOriginID); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPostview.SetChannelID(fromPostview.ChannelID); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPostview.SetChannelGroupID(fromPostview.ChannelGroupID); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPostview.SetUTMSource(fromPostview.UTMSource, fromPostview.GetFieldDate("utm_source")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPostview.SetUTMMedium(fromPostview.UTMMedium, fromPostview.GetFieldDate("utm_medium")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	// if fieldUpdate := toPostview.SetUTMID(fromPostview.UTMID, fromPostview.GetFieldDate("utm_id")); fieldUpdate != nil {
	// 	updatedFields = append(updatedFields, fieldUpdate)
	// }
	// if fieldUpdate := toPostview.SetUTMIDFrom(fromPostview.UTMIDFrom, fromPostview.GetFieldDate("utm_id_from")); fieldUpdate != nil {
	// 	updatedFields = append(updatedFields, fieldUpdate)
	// }
	if fieldUpdate := toPostview.SetUTMCampaign(fromPostview.UTMCampaign, fromPostview.GetFieldDate("utm_campaign")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPostview.SetUTMContent(fromPostview.UTMContent, fromPostview.GetFieldDate("utm_content")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPostview.SetUTMTerm(fromPostview.UTMTerm, fromPostview.GetFieldDate("utm_term")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	// channel mapping
	channelOriginID := ""
	channelID := ChannelNotMapped
	channelGroupID := ChannelNotMapped

	if channel, origin := toPostview.FindChannelFromOrigin(workspace.Channels); channel != nil {
		channelOriginID = origin.ID
		channelID = channel.ID
		channelGroupID = channel.GroupID
	}

	if fieldUpdate := toPostview.SetChannelOriginID(channelOriginID); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPostview.SetChannelID(channelID); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPostview.SetChannelGroupID(channelGroupID); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	// attribution
	if fieldUpdate := toPostview.SetConversionType(fromPostview.ConversionType); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPostview.SetConversionID(fromPostview.ConversionID); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPostview.SetConversionExternalID(fromPostview.ConversionExternalID); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPostview.SetConversionAt(fromPostview.ConversionAt); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPostview.SetConversionAmount(fromPostview.ConversionAmount); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPostview.SetLinearAmountAttributed(fromPostview.LinearAmountAttributed); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPostview.SetLinearPercentageAttributed(fromPostview.LinearPercentageAttributed); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPostview.SetTimeToConversion(fromPostview.TimeToConversion); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPostview.SetIsFirstConversion(fromPostview.IsFirstConversion); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPostview.SetRole(fromPostview.Role); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	for key, value := range fromPostview.ExtraColumns {
		if fieldUpdate := toPostview.SetExtraColumns(key, value, fromPostview.GetFieldDate(key)); fieldUpdate != nil {
			updatedFields = append(updatedFields, fieldUpdate)
		}
	}

	// UpdatedAt is the timeOfEvent for ITs
	toPostview.UpdatedAt = fromPostview.UpdatedAt
	// priority to oldest date
	toPostview.SetCreatedAt(fromPostview.CreatedAt)

	return
}

// overwrite json marshaller, to convert map of extra columns into "app_xxx" fields
func (s *Postview) MarshalJSON() ([]byte, error) {

	type Alias Postview

	result, err := json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	})

	if err != nil {
		return nil, err
	}

	if s.ExtraColumns == nil || len(s.ExtraColumns) == 0 {
		return result, nil
	}

	jsonValue := string(result)

	// convert extra columns into "app_xxx" fields
	for key, value := range s.ExtraColumns {
		jsonValue, err = sjson.Set(jsonValue, key, value)

		if err != nil {
			return nil, eris.Errorf("set impression custom dimension err: %v", err)
		}
	}

	return []byte(jsonValue), nil
}

func (s *Postview) FindChannelFromOrigin(channels []*Channel) (channel *Channel, origin *ChannelOrigin) {
	// find channel from mapping
	source := "direct"
	medium := "none"
	var campaign *string

	// if source and medium are provided, we will look for a matching channel
	if s.UTMSource != nil && !s.UTMSource.IsNull && s.UTMMedium != nil && !s.UTMMedium.IsNull {
		source = s.UTMSource.String
		medium = s.UTMMedium.String

		// add eventual campaign
		if s.UTMCampaign != nil && !s.UTMCampaign.IsNull {
			campaign = &s.UTMCampaign.String
		}
	}

	// find eventual matching channel
	return FindChannelFromOrigin(channels, false, source, medium, campaign)
}

func ComputePostviewID(externalID string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(externalID)))
}

func NewPostview(externalID string, userID string, createdAt time.Time, updatedAt time.Time) *Postview {
	return &Postview{
		ID:              ComputePostviewID(externalID),
		ExternalID:      externalID,
		UserID:          userID,
		CreatedAt:       createdAt,
		FieldsTimestamp: FieldsTimestamp{},
		ExtraColumns:    AppItemFields{},
		UpdatedAt:       &updatedAt,
	}
}

func NewPostviewFromDataLog(dataLog *DataLog, clockDifference time.Duration, workspace *Workspace) (postview *Postview, err error) {

	result := gjson.Get(dataLog.Item, "postview")
	if !result.Exists() {
		return nil, eris.New("item has no postview object")
	}

	if result.Type != gjson.JSON {
		return nil, eris.New("item postview is not an object")
	}

	extraColumns := workspace.FindExtraColumnsForItemKind("postview")

	// init
	postview = &Postview{
		UserID:          dataLog.UserID,
		FieldsTimestamp: FieldsTimestamp{},
		ExtraColumns:    AppItemFields{},
	}

	// loop over fields
	result.ForEach(func(key, value gjson.Result) bool {

		keyString := key.String()
		switch keyString {

		case "external_id":
			postview.ExternalID = value.String()
			postview.ID = ComputePostviewID(postview.ExternalID)

		case "created_at":
			if postview.CreatedAt, err = time.Parse(time.RFC3339, value.String()); err != nil {
				err = eris.Wrap(err, "postview.created_at")
				return false
			}

			// apply clock difference
			if postview.CreatedAt.After(time.Now()) {

				postview.CreatedAt = postview.CreatedAt.Add(clockDifference)
				if postview.CreatedAt.After(time.Now()) {
					err = eris.New("postview.created_at cannot be in the future")
					return false
				}
			}

			postview.CreatedAtTrunc = postview.CreatedAt.Truncate(time.Hour)

		case "updated_at":
			var updatedAt time.Time
			if updatedAt, err = time.Parse(time.RFC3339, value.String()); err != nil {
				err = eris.Wrap(err, "postview.updated_at")
				return false
			}

			if updatedAt.After(time.Now()) {
				// apply clock difference
				updatedAtModified := updatedAt.Add(clockDifference)
				updatedAt = updatedAtModified
				if updatedAt.After(time.Now()) {
					err = eris.New("postview.updated_at cannot be in the future")
					return false
				}
			}

			postview.UpdatedAt = &updatedAt

		case "device_external_id":
			if value.Type == gjson.Null {
				postview.DeviceID = nil
			} else {
				postview.DeviceID = StringPtr(ComputeDeviceID(value.String()))
			}

		case "utm_source":
			if value.Type == gjson.Null {
				postview.UTMSource = NewNullableString(nil)
			} else {
				postview.UTMSource = NewNullableString(StringPtr(value.String()))
			}

		case "utm_medium":
			if value.Type == gjson.Null {
				postview.UTMMedium = NewNullableString(nil)
			} else {
				postview.UTMMedium = NewNullableString(StringPtr(value.String()))
			}

		case "utm_campaign":
			if value.Type == gjson.Null {
				postview.UTMCampaign = NewNullableString(nil)
			} else {
				postview.UTMCampaign = NewNullableString(StringPtr(value.String()))
			}

		case "utm_content":
			if value.Type == gjson.Null {
				postview.UTMContent = NewNullableString(nil)
			} else {
				postview.UTMContent = NewNullableString(StringPtr(value.String()))
			}

		case "utm_term":
			if value.Type == gjson.Null {
				postview.UTMTerm = NewNullableString(nil)
			} else {
				postview.UTMTerm = NewNullableString(StringPtr(value.String()))
			}

		case "country":
			if value.Type == gjson.Null {
				postview.Country = NewNullableString(nil)
			} else {
				country := value.String()
				postview.Country = NewNullableString(&country)

				// check if country is valid
				if !govalidator.IsISO3166Alpha2(country) {
					err = eris.Errorf("country %s is invalid", country)
					return false
				}
			}

		case "region":
			if value.Type == gjson.Null {
				postview.Region = NewNullableString(nil)
			} else {
				region := value.String()
				postview.Region = NewNullableString(&region)
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
						postview.ExtraColumns[col.Name] = fieldValue
					}
				}
			}
		}

		return true
	})

	if err != nil {
		return nil, err
	}

	// use data import createdAt as updatedAt if not provided
	if postview.UpdatedAt == nil {
		postview.UpdatedAt = &postview.CreatedAt
	}

	// enrich postview with device
	if dataLog.UpsertedDevice != nil && postview.DeviceID == nil {
		postview.DeviceID = &dataLog.UpsertedDevice.ID
	}

	// Validation

	if postview.ExternalID == "" {
		return nil, eris.New("postview.external_id is required")
	}

	if postview.CreatedAt.IsZero() {
		return nil, eris.New("postview.created_at is required")
	}

	// error if utm_source is empty string or nil
	if postview.UTMSource == nil || postview.UTMSource.IsNull || postview.UTMSource.String == "" {
		return nil, eris.New("utm_source is required")
	}

	if postview.UTMMedium == nil || postview.UTMMedium.IsNull || postview.UTMMedium.String == "" {
		return nil, eris.New("utm_medium is required")
	}

	// find channel from mapping
	if channel, origin := postview.FindChannelFromOrigin(workspace.Channels); channel != nil {
		postview.SetChannelOriginID(origin.ID)
		postview.SetChannelID(channel.ID)
		postview.SetChannelGroupID(channel.GroupID)
	} else {
		// default values
		postview.SetChannelID(ChannelNotMapped)
		postview.SetChannelGroupID(ChannelNotMapped)
	}

	return postview, nil
}

var PostviewSchema string = `CREATE TABLE IF NOT EXISTS postview (
  id VARCHAR(64) NOT NULL,
  external_id VARCHAR(255) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  device_id VARCHAR(64),
  created_at DATETIME NOT NULL,
  created_at_trunc AS DATE_TRUNC('hour', created_at) PERSISTED DATETIME,
  db_created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  is_deleted BOOLEAN DEFAULT FALSE,
  merged_from_user_id VARCHAR(64),
  fields_timestamp JSON NOT NULL,

  channel_origin_id VARCHAR(255) NOT NULL,
  channel_id VARCHAR(255) NOT NULL,
  channel_group_id VARCHAR(255) NOT NULL,

  utm_source VARCHAR(255) NOT NULL,
  utm_medium VARCHAR(255) NOT NULL,
  utm_campaign VARCHAR(255),
  -- utm_id VARCHAR(255),
  -- utm_id_from VARCHAR(20),
  utm_content VARCHAR(255),
  utm_term VARCHAR(255),

  country VARCHAR(2),
  region VARCHAR(255),

  conversion_type VARCHAR(20),
  conversion_external_id VARCHAR(64),
  conversion_id VARCHAR(64),
  conversion_at DATETIME,
  conversion_amount INT DEFAULT 0,
  linear_amount_attributed INT DEFAULT 0,
  linear_percentage_attributed INT DEFAULT 0,
  time_to_conversion INT DEFAULT 0,
  is_first_conversion BOOLEAN DEFAULT FALSE,
  role TINYINT UNSIGNED,

  SORT KEY (created_at_trunc),
  PRIMARY KEY (id, user_id),
  KEY (conversion_id) USING HASH,
  KEY (user_id) USING HASH, -- for merging
  SHARD KEY (user_id)
) COLLATE utf8mb4_general_ci;`

var PostviewSchemaMYSQL string = `CREATE TABLE IF NOT EXISTS postview (
  id VARCHAR(64) NOT NULL,
  external_id VARCHAR(255) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  device_id VARCHAR(64),
  created_at DATETIME NOT NULL,
  created_at_trunc DATETIME GENERATED ALWAYS AS (DATE_FORMAT(created_at, '%Y-%m-%d %H:00:00')) STORED,
  db_created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  is_deleted BOOLEAN DEFAULT FALSE,
  merged_from_user_id VARCHAR(64),
  fields_timestamp JSON NOT NULL,

  channel_origin_id VARCHAR(255) NOT NULL,
  channel_id VARCHAR(255) NOT NULL,
  channel_group_id VARCHAR(255) NOT NULL,

  utm_source VARCHAR(255) NOT NULL,
  utm_medium VARCHAR(255) NOT NULL,
  utm_campaign VARCHAR(255),
  -- utm_id VARCHAR(255),
  -- utm_id_from VARCHAR(20),
  utm_content VARCHAR(255),
  utm_term VARCHAR(255),

  country VARCHAR(2),
  region VARCHAR(255),

  conversion_type VARCHAR(20),
  conversion_external_id VARCHAR(64),
  conversion_id VARCHAR(64),
  conversion_at DATETIME,
  conversion_amount INT DEFAULT 0,
  linear_amount_attributed INT DEFAULT 0,
  linear_percentage_attributed INT DEFAULT 0,
  time_to_conversion INT DEFAULT 0,
  is_first_conversion BOOLEAN DEFAULT FALSE,
  role TINYINT UNSIGNED,

  -- SORT KEY (created_at_trunc),
  PRIMARY KEY (id, user_id),
  KEY (conversion_id) USING HASH,
  KEY (user_id) USING HASH -- for merging
  -- SHARD KEY (user_id)
) COLLATE utf8mb4_general_ci;`

func NewPostviewCube() *CubeJSSchema {
	return &CubeJSSchema{
		Title:       "Postviews",
		Description: "Postviews",
		SQL:         "SELECT * FROM `postview`",
		Joins: map[string]CubeJSSchemaJoin{
			"User": {
				Relationship: "hasOne",
				SQL:          "${CUBE}.user_id = ${User}.id",
			},
			// "Device": {
			// 	Relationship: "hasOne",
			// 	SQL:          "${CUBE}.user_id = ${Device}.user_id AND ${CUBE}.device_id = ${Device}.id",
			// },
		},
		Measures: map[string]CubeJSSchemaMeasure{
			"count": {
				Type:        "count",
				Title:       "Count all",
				Description: "count of all postviews",
			},
			"unique_users": {
				Type:        "countDistinct",
				SQL:         "user_id",
				Title:       "Unique users",
				Description: "count of distinct user_id",
			},
			// "users_last_24h": {
			// 	Type:        "countDistinct",
			// 	SQL:         "user_id",
			// 	Title:       "Users last 24h",
			// 	Description: "count of distinct user_id in last 24h",
			// 	Filters: []CubeJSSchemaMeasureFilter{
			// 		{SQL: "created_at > date_sub(now(), interval 24 hour)"},
			// 	},
			// },
			// "users_online": {
			// 	Type:        "countDistinct",
			// 	SQL:         "user_id",
			// 	Title:       "Users online",
			// 	Description: "count of distinct user_id in last 3 minutes",
			// 	Filters: []CubeJSSchemaMeasureFilter{
			// 		{SQL: "created_at > date_sub(now(), interval 3 minute)"},
			// 	},
			// },
			"orders_contributions": {
				Type:        "count",
				SQL:         "id",
				Title:       "Orders contributions",
				Description: "Postviews that contributed to an order. Count of: conversion_id IS NOT NULL AND conversion_type = 'order'",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "${CUBE}.conversion_id IS NOT NULL AND ${CUBE}.conversion_type = 'order'"},
				},
			},
			"distinct_orders": {
				Type:        "countDistinct",
				SQL:         "conversion_id",
				Title:       "Distinct orders",
				Description: "Count of distinct conversion_id, where conversion_type = 'order'",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "${CUBE}.conversion_type = 'order'"},
				},
			},
			"contributions_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "Contributions count",
				Description: "Count of: conversion_id IS NOT NULL",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "${CUBE}.conversion_id IS NOT NULL"},
				},
			},
			"distinct_conversions": {
				Type:        "countDistinct",
				SQL:         "conversion_id",
				Title:       "Distinct conversions",
				Description: "Count of distinct conversion_id",
			},
			"conversion_rate": {
				Type:        "number",
				SQL:         "(${distinct_conversions}) / ${count}",
				Title:       "Conversion rate",
				Description: "Conversion rate: distinct_conversions / count",
			},
			"linear_amount_attributed": {
				Type:        "number",
				SQL:         "COALESCE(SUM(linear_amount_attributed), 0)",
				Title:       "Linear amount attributed",
				Description: "Sum of linear_amount_attributed",
			},
			"linear_percentage_attributed": {
				Type:        "number",
				SQL:         "COALESCE(ROUND(AVG(linear_percentage_attributed) / 100, 2), 0)",
				Title:       "Linear percentage attributed",
				Description: "Avg of linear_percentage_attributed",
			},
			"linear_conversions_attributed": {
				Type:        "number",
				SQL:         "COALESCE(ROUND(SUM(linear_percentage_attributed) / 10000, 2), 0)",
				Title:       "Linear conversions attributed",
				Description: "Sum of linear_percentage_attributed",
			},
			"alone_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "Alone role count",
				Description: "count of: role = 0 AND conversion_id IS NOT NULL",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "${CUBE}.role = 0 AND ${CUBE}.conversion_id IS NOT NULL"},
				},
			},
			"initiator_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "Initiator role count",
				Description: "count of: role = 1 AND conversion_id IS NOT NULL",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "${CUBE}.role = 1 AND ${CUBE}.conversion_id IS NOT NULL"},
				},
			},
			"assisting_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "Assisting role count",
				Description: "count of: role = 2 AND conversion_id IS NOT NULL",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "${CUBE}.role = 2 AND ${CUBE}.conversion_id IS NOT NULL"},
				},
			},

			"closer_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "Closer role count",
				Description: "count of: role = 3 AND conversion_id IS NOT NULL",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "${CUBE}.role = 3 AND ${CUBE}.conversion_id IS NOT NULL"},
				},
			},
			"alone_ratio": {
				Type:        "number",
				SQL:         "COALESCE(${alone_count} / ${contributions_count}, 0)",
				Title:       "Alone role ratio",
				Description: "ratio of: alone_count / contributions_count",
			},

			"initiator_ratio": {
				Type:        "number",
				SQL:         "COALESCE(${initiator_count} / ${contributions_count}, 0)",
				Title:       "Initiator role ratio",
				Description: "ratio of: initiator_count / contributions_count",
			},
			"assisting_ratio": {
				Type:        "number",
				SQL:         "COALESCE(${assisting_count} / ${contributions_count}, 0)",
				Title:       "Assisting role ratio",
				Description: "ratio of: assisting_count / contributions_count",
			},
			"closer_ratio": {
				Type:        "number",
				SQL:         "COALESCE(${closer_count} / ${contributions_count}, 0)",
				Title:       "Closer role ratio",
				Description: "ratio of: closer_count / contributions_count",
			},
			"alone_linear_conversions_attributed": {
				Type:        "number",
				SQL:         "COALESCE(ROUND(SUM(CASE WHEN ${CUBE}.role = 0 THEN ${CUBE}.linear_percentage_attributed ELSE 0 END) / 10000, 2), 0)",
				Title:       "Alone linear conversions attributed",
				Description: "Sum of: CASE WHEN role = 0 THEN linear_percentage_attributed ELSE 0 END",
			},
			"initiator_linear_conversions_attributed": {
				Type:        "number",
				SQL:         "COALESCE(ROUND(SUM(CASE WHEN ${CUBE}.role = 1 THEN ${CUBE}.linear_percentage_attributed ELSE 0 END) / 10000, 2), 0)",
				Title:       "Initiator linear conversions attributed",
				Description: "Sum of: CASE WHEN role = 1 THEN linear_percentage_attributed ELSE 0 END",
			},
			"assisting_linear_conversions_attributed": {
				Type:        "number",
				SQL:         "COALESCE(ROUND(SUM(CASE WHEN ${CUBE}.role = 2 THEN ${CUBE}.linear_percentage_attributed ELSE 0 END) / 10000, 2), 0)",
				Title:       "Assisting linear conversions attributed",
				Description: "Sum of: CASE WHEN role = 2 THEN linear_percentage_attributed ELSE 0 END",
			},
			"closer_linear_conversions_attributed": {
				Type:        "number",
				SQL:         "COALESCE(ROUND(SUM(CASE WHEN ${CUBE}.role = 3 THEN ${CUBE}.linear_percentage_attributed ELSE 0 END) / 10000, 2), 0)",
				Title:       "Closer linear conversions attributed",
				Description: "Sum of: CASE WHEN role = 3 THEN linear_percentage_attributed ELSE 0 END",
			},
			"alone_linear_amount_attributed": {
				Type:        "number",
				SQL:         "COALESCE(SUM(CASE WHEN ${CUBE}.role = 0 THEN ${CUBE}.linear_amount_attributed ELSE 0 END), 0)",
				Title:       "Alone linear amount attributed",
				Description: "Sum of: CASE WHEN role = 0 THEN linear_amount_attributed ELSE 0 END",
			},
			"initiator_linear_amount_attributed": {
				Type:        "number",
				SQL:         "COALESCE(SUM(CASE WHEN ${CUBE}.role = 1 THEN ${CUBE}.linear_amount_attributed ELSE 0 END), 0)",
				Title:       "Initiator linear amount attributed",
				Description: "Sum of: CASE WHEN role = 1 THEN linear_amount_attributed ELSE 0 END",
			},
			"assisting_linear_amount_attributed": {
				Type:        "number",
				SQL:         "COALESCE(SUM(CASE WHEN ${CUBE}.role = 2 THEN ${CUBE}.linear_amount_attributed ELSE 0 END), 0)",
				Title:       "Assisting linear amount attributed",
				Description: "Sum of: CASE WHEN role = 2 THEN linear_amount_attributed ELSE 0 END",
			},
			"closer_linear_amount_attributed": {
				Type:        "number",
				SQL:         "COALESCE(SUM(CASE WHEN ${CUBE}.role = 3 THEN ${CUBE}.linear_amount_attributed ELSE 0 END), 0)",
				Title:       "Closer linear amount attributed",
				Description: "Sum of: CASE WHEN role = 3 THEN linear_amount_attributed ELSE 0 END",
			},
		},
		Dimensions: map[string]CubeJSSchemaDimension{

			"id": {
				SQL:         "id",
				Type:        "string",
				PrimaryKey:  true,
				Title:       "Postview ID",
				Description: "field: id",
			},

			"user_id": {
				Type:        "string",
				SQL:         "user_id",
				Title:       "User ID",
				Description: "field: user_id",
			},
			// "device_id": {
			// 	Type:        "string",
			// 	SQL:         "device_id",
			// 	Title:       "Device ID",
			// 	Description: "field: device_id",
			// },
			"created_at": {
				Type:        "time",
				SQL:         "created_at",
				Title:       "Created at",
				Description: "field: created_at",
			},
			"created_at_trunc": {
				Type:        "time",
				SQL:         "created_at_trunc",
				Title:       "Created at (truncated to hour)",
				Description: "field: created_at_trunc",
			},
			"channel_origin_id": {
				Type:        "string",
				SQL:         "channel_origin_id",
				Title:       "Channel origin ID",
				Description: "field: channel_origin_id",
			},
			"country": {
				Type:        "string",
				SQL:         "country",
				Title:       "Country code",
				Description: "field: country",
			},
			"region": {
				Type:        "string",
				SQL:         "region",
				Title:       "Region/State",
				Description: "field: region/state",
			},
			"channel_id": {
				Type:        "string",
				SQL:         "channel_id",
				Title:       "Channel ID",
				Description: "field: channel_id",
			},
			"channel_group_id": {
				Type:        "string",
				SQL:         "channel_group_id",
				Title:       "Channel group ID",
				Description: "field: channel_group_id",
			},
			"utm_source": {
				Type:        "string",
				SQL:         "utm_source",
				Title:       "UTM source",
				Description: "field: utm_source",
			},
			"utm_medium": {
				Type:        "string",
				SQL:         "utm_medium",
				Title:       "UTM medium",
				Description: "field: utm_medium",
			},
			"utm_campaign": {
				Type:        "string",
				SQL:         "utm_campaign",
				Title:       "UTM campaign",
				Description: "field: utm_campaign",
			},
			"utm_content": {
				Type:        "string",
				SQL:         "utm_content",
				Title:       "UTM content",
				Description: "field: utm_content",
			},
			"utm_term": {
				Type:        "string",
				SQL:         "utm_term",
				Title:       "UTM term",
				Description: "field: utm_term",
			},
			"conversion_type": {
				Type:        "string",
				SQL:         "conversion_type",
				Title:       "Conversion type",
				Description: "field: conversion_type",
			},
			"conversion_external_id": {
				Type:        "string",
				SQL:         "conversion_external_id",
				Title:       "Conversion external ID",
				Description: "field: conversion_external_id",
			},
			"conversion_id": {
				Type:        "string",
				SQL:         "conversion_id",
				Title:       "Conversion ID",
				Description: "field: conversion_id",
			},
			"conversion_at": {
				Type:        "time",
				SQL:         "conversion_at",
				Title:       "Conversion at",
				Description: "field: conversion_at",
			},
			"conversion_amount": {
				Type:        "number",
				SQL:         "conversion_amount",
				Title:       "Conversion amount",
				Description: "field: conversion_amount",
			},
			"time_to_conversion": {
				Type:        "number",
				SQL:         "time_to_conversion",
				Title:       "Time to conversion",
				Description: "field: time_to_conversion",
			},
			"is_first_conversion": {
				Type:        "number",
				SQL:         "is_first_conversion",
				Title:       "Is first conversion",
				Description: "field: is_first_conversion",
			},
			"role": {
				Title:       "Channel role",
				Description: "field: role",
				Type:        "string",
				Case: &CubeJSSchemaCase{
					When: []CubeJSSchemaCaseWhen{
						{SQL: "role = 1", Label: "initiator"},
						{SQL: "role = 2", Label: "assisting"},
						{SQL: "role = 3", Label: "closer"},
					},
					Else: CubeJSSchemaCaseElse{Label: "alone"},
				},
			},
		},
	}
}
