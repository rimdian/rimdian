package entity

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/rimdian/rimdian/internal/api/common"
	"github.com/rotisserie/eris"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var (
	SessionExpirationTimeout = 30 * time.Minute

	// computed fields should be excluded from SELECT/INSERT while cloning rows
	SessionComputedFields []string = []string{
		"created_at_trunc",
		"year",
		"month",
		"month_day",
		"week_day",
		"hour",
		"bounced",
	}

	ErrSessionRequired           = eris.New("session is required")
	ErrSessionIDRequired         = eris.New("session id or external id is required")
	ErrSessionUserIDRequired     = eris.New("session user id is required")
	ErrSessionDomainIDRequired   = eris.New("session domain id is required")
	ErrSessionCreatedAtRequired  = eris.New("session created at is required")
	ErrSessionTimezoneInvalid    = eris.New("session timezone is not valid")
	ErrSessionLandingPageInvalid = eris.New("session landing_page is invalid")
	ErrSessionDomainIDInvalid    = eris.New("session domain_id is invalid")
	ErrSessionReferrerInvalid    = eris.New("session referrer is invalid")
)

type Session struct {
	ID               string          `db:"id" json:"id"`
	ExternalID       string          `db:"external_id" json:"external_id"`
	UserID           string          `db:"user_id" json:"user_id"`
	DomainID         string          `db:"domain_id" json:"domain_id"`
	DeviceID         *string         `db:"device_id" json:"device_id,omitempty"`
	CreatedAt        time.Time       `db:"created_at" json:"created_at"`
	CreatedAtTrunc   time.Time       `db:"created_at_trunc" json:"created_at_trunc"` // date truncated at hour, used to optimize columnstore storage and queries
	DBCreatedAt      time.Time       `db:"db_created_at" json:"db_created_at"`
	DBUpdatedAt      time.Time       `db:"db_updated_at" json:"db_updated_at"`
	IsDeleted        bool            `db:"is_deleted" json:"is_deleted,omitempty"` // deleting rows in transactions cause deadlocks in singlestore, we use an update
	MergedFromUserID *string         `db:"merged_from_user_id" json:"merged_from_user_id,omitempty"`
	FieldsTimestamp  FieldsTimestamp `db:"fields_timestamp" json:"fields_timestamp"`

	// localized datetime parts for cohorts purpose
	Timezone *string `db:"timezone" json:"timezone"`   // field set while repo upsert if missing
	Year     int     `db:"year" json:"year"`           // computed from created_at + timezone
	Month    int     `db:"month" json:"month"`         // computed from created_at + timezone
	MonthDay int     `db:"month_day" json:"month_day"` // computed from created_at + timezone
	WeekDay  int     `db:"week_day" json:"week_day"`   // computed from created_at + timezone
	Hour     int     `db:"hour" json:"hour"`           // computed from created_at + timezone

	// bounce fields:
	// ExpiresAt         time.Time      `db:"expires_at" json:"expires_at"`
	Duration          *NullableInt64 `db:"duration" json:"duration,omitempty"`
	Bounced           bool           `db:"bounced" json:"bounced"`
	PageviewsCount    *int64         `db:"pageviews_count" json:"pageviews_count"`       // used to compute avg pages/session
	InteractionsCount *int64         `db:"interactions_count" json:"interactions_count"` // used to compute if session has bounced

	// web fields:
	LandingPage     *NullableString `db:"landing_page" json:"landing_page,omitempty"`
	LandingPagePath *string         `db:"landing_page_path" json:"landing_page_path,omitempty"` // computed from landing page
	Referrer        *NullableString `db:"referrer" json:"referrer,omitempty"`
	ReferrerDomain  *string         `db:"referrer_domain" json:"referrer_domain,omitempty"` // computed from referrer
	ReferrerPath    *string         `db:"referrer_path" json:"referrer_path,omitempty"`     // computed from referrer

	// channel mapping:
	ChannelOriginID string `db:"channel_origin_id" json:"channel_origin_id"` // source / medium( / campaign)
	ChannelID       string `db:"channel_id" json:"channel_id"`               // field set while repo upsert with utm_ params
	ChannelGroupID  string `db:"channel_group_id" json:"channel_group_id"`   // field set while repo upsert with utm_ params

	// utm_ parameters equivalent:
	UTMSource   *NullableString `db:"utm_source" json:"utm_source"`               // utm_source
	UTMMedium   *NullableString `db:"utm_medium" json:"utm_medium"`               // utm_medium
	UTMID       *NullableString `db:"utm_id" json:"utm_id,omitempty"`             // utm_id
	UTMIDFrom   *NullableString `db:"utm_id_from" json:"utm_id_from,omitempty"`   // gclid | fbclid | cmid...
	UTMCampaign *NullableString `db:"utm_campaign" json:"utm_campaign,omitempty"` // utm_campaign
	UTMContent  *NullableString `db:"utm_content" json:"utm_content,omitempty"`   // utm_content
	UTMTerm     *NullableString `db:"utm_term" json:"utm_term,omitempty"`         // utm_term

	// "via utm" fields are used to store a copy of original utm_ parameters in case of overwrite from "data filters"
	ViaUTMSource   *NullableString `db:"via_utm_source" json:"via_utm_source,omitempty"`     // overwritten utm_source
	ViaUTMMedium   *NullableString `db:"via_utm_medium" json:"via_utm_medium,omitempty"`     // overwritten utm_medium
	ViaUTMID       *NullableString `db:"via_utm_id" json:"via_utm_id,omitempty"`             // overwritten utm_id
	ViaUTMIDFrom   *NullableString `db:"via_utm_id_from" json:"via_utm_id_from,omitempty"`   // overwritten utm_id from (gclid, fbclid...)
	ViaUTMCampaign *NullableString `db:"via_utm_campaign" json:"via_utm_campaign,omitempty"` // overwritten utm_campaign
	ViaUTMContent  *NullableString `db:"via_utm_content" json:"via_utm_content,omitempty"`   // overwritten utm_content
	ViaUTMTerm     *NullableString `db:"via_utm_term" json:"via_utm_term,omitempty"`         // overwritten utm_term

	// attribution, set when the session contributed to a conversion
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

func (s *Session) GetFieldDate(field string) time.Time {
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

// determines if object has already been persisted in DB
func (u *Session) IsNew() bool {
	return u.DBCreatedAt.IsZero()
}

func (s *Session) Validate(domains []*Domain) error {
	if s.ID == "" && s.ExternalID == "" {
		return ErrSessionIDRequired
	}
	if s.UserID == "" {
		return ErrSessionUserIDRequired
	}
	if s.DomainID == "" {
		return ErrSessionDomainIDRequired
	}
	// find domain
	var domain *Domain
	for _, d := range domains {
		if d.ID == s.DomainID {
			domain = d
			break
		}
	}
	if domain == nil {
		return ErrSessionDomainIDInvalid
	}
	if s.CreatedAt.IsZero() {
		return ErrSessionCreatedAtRequired
	}
	if s.Timezone != nil && !govalidator.IsIn(*s.Timezone, common.Timezones...) {
		return ErrSessionTimezoneInvalid
	}

	// compute LandingPagePath if LandingPage is set and not null
	if s.LandingPage != nil && !s.LandingPage.IsNull {
		if !govalidator.IsRequestURL(s.LandingPage.String) {
			return ErrSessionLandingPageInvalid
		}

		newURL, err := ParseAndCleanURL(s.LandingPage.String, domain.ParamsWhitelist)
		if err != nil {
			return ErrSessionLandingPageInvalid
		}

		// replace url with cleaned url (without frament and non-allowed parameters)
		s.LandingPage = &NullableString{IsNull: false, String: newURL.String()}

		// get path
		path := newURL.Path
		if path == "" {
			path = "/"
		}

		s.LandingPagePath = &path
	}

	// compute  ReferrerDomain & ReferrerPath if Referrer is set and not null
	if s.Referrer != nil && !s.Referrer.IsNull {
		if !govalidator.IsRequestURL(s.Referrer.String) {
			return ErrSessionReferrerInvalid
		}

		u, err := url.Parse(s.Referrer.String)
		if err != nil {
			return ErrSessionReferrerInvalid
		}

		s.ReferrerDomain = &u.Host

		// get path
		path := u.Path
		if path == "" {
			path = "/"
		}
		s.ReferrerPath = &path
	}

	return nil
}

// update a field timestamp to its most recent value
// func (s *Session) UpdateFieldTimestamp(field string, timestamp *time.Time) {
// 	if timestamp == nil {
// 		return
// 	}
// 	if previousTimestamp, exists := s.FieldsTimestamp[field]; exists {
// 		if previousTimestamp.Before(*timestamp) {
// 			s.FieldsTimestamp[field] = *timestamp
// 		}
// 	} else {
// 		s.FieldsTimestamp[field] = *timestamp
// 	}
// }

func (s *Session) SetDeviceID(value *string, timestamp time.Time) (update *UpdatedField) {
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
func (s *Session) SetCreatedAt(value time.Time) {
	// update if current value is older
	if value.Before(s.CreatedAt) {
		s.CreatedAt = value
		s.CreatedAtTrunc = value.Truncate(time.Hour)
	}
}
func (s *Session) SetTimezone(value *string, timestamp time.Time) (update *UpdatedField) {
	key := "timezone"
	// value cant be null
	if value == nil {
		return nil
	}
	// abort if values are equal
	if StringsEqual(s.Timezone, value) {
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
		s.Timezone = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: StringPointerToInterface(s.Timezone),
		NewValue:  StringPointerToInterface(value),
	}
	s.Timezone = value
	s.FieldsTimestamp[key] = timestamp
	return
}

func (s *Session) SetDuration(value *NullableInt64, timestamp time.Time) (update *UpdatedField) {
	key := "duration"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// set value to 10800 secs (60x60x3 hours) if value is above
	if value != nil && value.Int64 > 10800 {
		value.Int64 = 10800
	}
	// abort if values are equal
	if value != nil && s.Duration != nil && s.Duration.IsNull == value.IsNull && s.Duration.Int64 == value.Int64 {
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
		s.Duration = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableInt64ToInterface(s.Duration),
		NewValue:  NullableInt64ToInterface(value),
	}
	s.Duration = value
	s.FieldsTimestamp[key] = timestamp
	return
}

func (s *Session) SetPageviewsCount(value *int64, timestamp time.Time) (update *UpdatedField) {
	key := "pageviews_count"
	// value cant be null
	if value == nil {
		return nil
	}
	// abort if values are equal
	if Int64Equal(s.PageviewsCount, value) {
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
		s.PageviewsCount = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: Int64PointerToInterface(s.PageviewsCount),
		NewValue:  Int64PointerToInterface(value),
	}
	s.PageviewsCount = value
	s.FieldsTimestamp[key] = timestamp
	return
}

func (s *Session) SetInteractionsCount(value *int64, timestamp time.Time) (update *UpdatedField) {
	key := "interactions_count"
	// value cant be null
	if value == nil {
		return nil
	}
	// abort if values are equal
	if Int64Equal(s.InteractionsCount, value) {
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
		s.InteractionsCount = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: Int64PointerToInterface(s.InteractionsCount),
		NewValue:  Int64PointerToInterface(value),
	}
	s.InteractionsCount = value
	s.FieldsTimestamp[key] = timestamp
	return
}

func (s *Session) SetLandingPage(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "landing_page"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.LandingPage != nil && s.LandingPage.IsNull == value.IsNull && s.LandingPage.String == value.String {
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
		s.LandingPage = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.LandingPage),
		NewValue:  NullableStringToInterface(value),
	}
	s.LandingPage = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Session) SetLandingPagePath(value *string, timestamp time.Time) (update *UpdatedField) {
	key := "landing_page_path"
	// value cant be null
	if value == nil {
		return nil
	}
	// abort if values are equal
	if StringsEqual(value, s.LandingPagePath) {
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
		s.LandingPagePath = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: StringPointerToInterface(s.LandingPagePath),
		NewValue:  StringPointerToInterface(value),
	}
	s.LandingPagePath = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Session) SetReferrer(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "referrer"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.Referrer != nil && s.Referrer.IsNull == value.IsNull && s.Referrer.String == value.String {
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
		s.Referrer = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.Referrer),
		NewValue:  NullableStringToInterface(value),
	}
	s.Referrer = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Session) SetReferrerDomain(value *string, timestamp time.Time) (update *UpdatedField) {
	key := "referrer_domain"
	// value cant be null
	if value == nil {
		return nil
	}
	// abort if values are equal
	if StringsEqual(value, s.ReferrerDomain) {
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
		s.ReferrerDomain = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: StringPointerToInterface(s.ReferrerDomain),
		NewValue:  StringPointerToInterface(value),
	}
	s.ReferrerDomain = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Session) SetReferrerPath(value *string, timestamp time.Time) (update *UpdatedField) {
	key := "referrer_path"
	// value cant be null
	if value == nil {
		return nil
	}
	// abort if values are equal
	if StringsEqual(value, s.ReferrerPath) {
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
		s.ReferrerPath = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: StringPointerToInterface(s.ReferrerPath),
		NewValue:  StringPointerToInterface(value),
	}
	s.ReferrerPath = value
	s.FieldsTimestamp[key] = timestamp
	return
}

// field set by attribution algo, dont compare timestamps
func (s *Session) SetChannelOriginID(value string) (update *UpdatedField) {
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
func (s *Session) SetChannelID(value string) (update *UpdatedField) {
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
func (s *Session) SetChannelGroupID(value string) (update *UpdatedField) {
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

func (s *Session) SetUTMSource(value *NullableString, timestamp time.Time) (update *UpdatedField) {
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
func (s *Session) SetUTMMedium(value *NullableString, timestamp time.Time) (update *UpdatedField) {
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
func (s *Session) SetUTMID(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "utm_id"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.UTMID != nil && s.UTMID.IsNull == value.IsNull && s.UTMID.String == value.String {
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
		s.UTMID = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.UTMID),
		NewValue:  NullableStringToInterface(value),
	}
	s.UTMID = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Session) SetUTMIDFrom(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "utm_id_from"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.UTMIDFrom != nil && s.UTMIDFrom.IsNull == value.IsNull && s.UTMIDFrom.String == value.String {
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
		s.UTMIDFrom = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.UTMIDFrom),
		NewValue:  NullableStringToInterface(value),
	}
	s.UTMIDFrom = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Session) SetUTMCampaign(value *NullableString, timestamp time.Time) (update *UpdatedField) {
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
func (s *Session) SetUTMContent(value *NullableString, timestamp time.Time) (update *UpdatedField) {
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

func (s *Session) SetUTMTerm(value *NullableString, timestamp time.Time) (update *UpdatedField) {
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
func (s *Session) SetViaUTMSource(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "via_utm_source"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.ViaUTMSource != nil && s.ViaUTMSource.IsNull == value.IsNull && s.ViaUTMSource.String == value.String {
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
		s.ViaUTMSource = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.ViaUTMSource),
		NewValue:  NullableStringToInterface(value),
	}
	s.ViaUTMSource = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Session) SetViaUTMMedium(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "via_utm_medium"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.ViaUTMMedium != nil && s.ViaUTMMedium.IsNull == value.IsNull && s.ViaUTMMedium.String == value.String {
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
		s.ViaUTMMedium = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.ViaUTMMedium),
		NewValue:  NullableStringToInterface(value),
	}
	s.ViaUTMMedium = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Session) SetViaUTMID(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "via_utm_id"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.ViaUTMID != nil && s.ViaUTMID.IsNull == value.IsNull && s.ViaUTMID.String == value.String {
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
		s.ViaUTMID = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.ViaUTMID),
		NewValue:  NullableStringToInterface(value),
	}
	s.ViaUTMID = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Session) SetViaUTMIDFrom(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "via_utm_id_from"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.ViaUTMIDFrom != nil && s.ViaUTMIDFrom.IsNull == value.IsNull && s.ViaUTMIDFrom.String == value.String {
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
		s.ViaUTMIDFrom = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.ViaUTMIDFrom),
		NewValue:  NullableStringToInterface(value),
	}
	s.ViaUTMIDFrom = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Session) SetViaUTMCampaign(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "via_utm_campaign"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.ViaUTMCampaign != nil && s.ViaUTMCampaign.IsNull == value.IsNull && s.ViaUTMCampaign.String == value.String {
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
		s.ViaUTMCampaign = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.ViaUTMCampaign),
		NewValue:  NullableStringToInterface(value),
	}
	s.ViaUTMCampaign = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Session) SetViaUTMContent(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "via_utm_content"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.ViaUTMContent != nil && s.ViaUTMContent.IsNull == value.IsNull && s.ViaUTMContent.String == value.String {
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
		s.ViaUTMContent = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.ViaUTMContent),
		NewValue:  NullableStringToInterface(value),
	}
	s.ViaUTMContent = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Session) SetViaUTMTerm(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "via_utm_term"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.ViaUTMTerm != nil && s.ViaUTMTerm.IsNull == value.IsNull && s.ViaUTMTerm.String == value.String {
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
		s.ViaUTMTerm = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.ViaUTMTerm),
		NewValue:  NullableStringToInterface(value),
	}
	s.ViaUTMTerm = value
	s.FieldsTimestamp[key] = timestamp
	return
}

// fields set by attribution algorithm, dont compare timestamps
func (s *Session) SetConversionType(value *string) (update *UpdatedField) {
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
func (s *Session) SetConversionID(value *string) (update *UpdatedField) {
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
func (s *Session) SetConversionExternalID(value *string) (update *UpdatedField) {
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
func (s *Session) SetConversionAt(value *time.Time) (update *UpdatedField) {
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
func (s *Session) SetConversionAmount(value *int64) (update *UpdatedField) {
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
func (s *Session) SetLinearAmountAttributed(value *int64) (update *UpdatedField) {
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
func (s *Session) SetLinearPercentageAttributed(value *int64) (update *UpdatedField) {
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
func (s *Session) SetTimeToConversion(value *int64) (update *UpdatedField) {
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
func (s *Session) SetIsFirstConversion(value *bool) (update *UpdatedField) {
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
func (s *Session) SetRole(value *int64) (update *UpdatedField) {
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

func (s *Session) SetExtraColumns(field string, value *AppItemField, timestamp time.Time) (update *UpdatedField) {

	if s.ExtraColumns == nil {
		s.ExtraColumns = AppItemFields{}
	}

	// abort if field doesnt start with "app_"
	if !strings.HasPrefix(field, "app_") && !strings.HasPrefix(field, "appx_") {
		log.Printf("session field %s doesnt start with app_", field)
		return nil
	}

	// ignore if value is not provided
	if value == nil {
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

// merges two sessions and returns the list of updated fields
func (fromSession *Session) MergeInto(toSession *Session, workspace *Workspace) (updatedFields []*UpdatedField) {

	updatedFields = []*UpdatedField{} // init

	if toSession.FieldsTimestamp == nil {
		toSession.FieldsTimestamp = FieldsTimestamp{}
	}

	if fieldUpdate := toSession.SetDeviceID(fromSession.DeviceID, fromSession.GetFieldDate("device_id")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toSession.SetTimezone(fromSession.Timezone, fromSession.GetFieldDate("timezone")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toSession.SetDuration(fromSession.Duration, fromSession.GetFieldDate("duration")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toSession.SetPageviewsCount(fromSession.PageviewsCount, fromSession.GetFieldDate("pageviews_count")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toSession.SetInteractionsCount(fromSession.InteractionsCount, fromSession.GetFieldDate("interactions_count")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	if fieldUpdate := toSession.SetLandingPage(fromSession.LandingPage, fromSession.GetFieldDate("landing_page")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toSession.SetLandingPagePath(fromSession.LandingPagePath, fromSession.GetFieldDate("landing_page_path")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toSession.SetReferrer(fromSession.Referrer, fromSession.GetFieldDate("referrer")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toSession.SetReferrerDomain(fromSession.ReferrerDomain, fromSession.GetFieldDate("referrer_domain")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toSession.SetReferrerPath(fromSession.ReferrerPath, fromSession.GetFieldDate("referrer_path")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toSession.SetUTMSource(fromSession.UTMSource, fromSession.GetFieldDate("utm_source")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toSession.SetUTMMedium(fromSession.UTMMedium, fromSession.GetFieldDate("utm_medium")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toSession.SetUTMID(fromSession.UTMID, fromSession.GetFieldDate("utm_id")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toSession.SetUTMIDFrom(fromSession.UTMIDFrom, fromSession.GetFieldDate("utm_id_from")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toSession.SetUTMCampaign(fromSession.UTMCampaign, fromSession.GetFieldDate("utm_campaign")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toSession.SetUTMContent(fromSession.UTMContent, fromSession.GetFieldDate("utm_content")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toSession.SetUTMTerm(fromSession.UTMTerm, fromSession.GetFieldDate("utm_term")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	// channel mapping
	channelOriginID := ""
	channelID := ChannelNotMapped
	channelGroupID := ChannelNotMapped

	if channel, origin := toSession.FindChannelFromOrigin(workspace.Channels); channel != nil {
		channelOriginID = origin.ID
		channelID = channel.ID
		channelGroupID = channel.GroupID
	}

	if fieldUpdate := toSession.SetChannelOriginID(channelOriginID); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toSession.SetChannelID(channelID); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toSession.SetChannelGroupID(channelGroupID); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	// via
	if fieldUpdate := toSession.SetViaUTMSource(fromSession.ViaUTMSource, fromSession.GetFieldDate("via_utm_source")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toSession.SetViaUTMMedium(fromSession.ViaUTMMedium, fromSession.GetFieldDate("via_utm_medium")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toSession.SetViaUTMID(fromSession.ViaUTMID, fromSession.GetFieldDate("via_utm_id")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toSession.SetViaUTMIDFrom(fromSession.ViaUTMIDFrom, fromSession.GetFieldDate("via_utm_id_from")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toSession.SetViaUTMCampaign(fromSession.ViaUTMCampaign, fromSession.GetFieldDate("via_utm_campaign")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toSession.SetViaUTMContent(fromSession.ViaUTMContent, fromSession.GetFieldDate("via_utm_content")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toSession.SetViaUTMTerm(fromSession.ViaUTMTerm, fromSession.GetFieldDate("via_utm_term")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	// note: conversion fields are set by attribution algo, not by import

	for key, value := range fromSession.ExtraColumns {
		if fieldUpdate := toSession.SetExtraColumns(key, value, fromSession.GetFieldDate(key)); fieldUpdate != nil {
			updatedFields = append(updatedFields, fieldUpdate)
		}
	}

	// UpdatedAt is the timeOfEvent for ITs
	toSession.UpdatedAt = fromSession.UpdatedAt
	// priority to oldest date
	toSession.SetCreatedAt(fromSession.CreatedAt)

	return
}

// overwrite json marshaller, to convert map of extra columns into "app_xxx" fields
func (s *Session) MarshalJSON() ([]byte, error) {

	type Alias Session

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
			return nil, eris.Errorf("set session custom dimension err: %v", err)
		}
	}

	return []byte(jsonValue), nil
}

func (s *Session) FindChannelFromOrigin(channels []*Channel) (channel *Channel, origin *ChannelOrigin) {
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
	return FindChannelFromOrigin(channels, source, medium, campaign)
}

func ComputeSessionID(externalID string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(externalID)))
}

func NewSession(externalID string, userID string, domainID string, createdAt time.Time, updatedAt time.Time) *Session {
	return &Session{
		ID:              ComputeSessionID(externalID),
		ExternalID:      externalID,
		UserID:          userID,
		DomainID:        domainID,
		CreatedAt:       createdAt,
		FieldsTimestamp: FieldsTimestamp{},
		ExtraColumns:    AppItemFields{},
		UpdatedAt:       &updatedAt,
	}
}

func NewSessionFromDataLog(dataLog *DataLog, clockDifference time.Duration, workspace *Workspace) (session *Session, err error) {

	result := gjson.Get(dataLog.Item, "session")
	if !result.Exists() {
		return nil, eris.New("item has no session object")
	}

	if result.Type != gjson.JSON {
		return nil, eris.New("item session is not an object")
	}

	extraColumns := workspace.FindExtraColumnsForItemKind("session")

	// init
	session = &Session{
		UserID:          dataLog.UserID,
		FieldsTimestamp: FieldsTimestamp{},
		ExtraColumns:    AppItemFields{},
	}

	// loop over fields
	result.ForEach(func(key, value gjson.Result) bool {

		fieldName := key.String()
		switch fieldName {

		case "external_id":
			session.ExternalID = value.String()
			session.ID = ComputeSessionID(session.ExternalID)

		case "domain_id":
			session.DomainID = value.String()

		case "created_at":
			if session.CreatedAt, err = time.Parse(time.RFC3339, value.String()); err != nil {
				err = eris.Wrap(err, "session.created_at")
				return false
			}

			// apply clock difference
			if session.CreatedAt.After(time.Now()) {

				session.CreatedAt = session.CreatedAt.Add(clockDifference)
				if session.CreatedAt.After(time.Now()) {
					err = eris.New("session.created_at cannot be in the future")
					return false
				}
			}

			session.CreatedAtTrunc = session.CreatedAt.Truncate(time.Hour)

		case "updated_at":
			var updatedAt time.Time
			if updatedAt, err = time.Parse(time.RFC3339, value.String()); err != nil {
				err = eris.Wrap(err, "session.updated_at")
				return false
			}

			if updatedAt.After(time.Now()) {
				// apply clock difference
				updatedAtModified := updatedAt.Add(clockDifference)
				updatedAt = updatedAtModified
				if updatedAt.After(time.Now()) {
					err = eris.New("session.updated_at cannot be in the future")
					return false
				}
			}

			session.UpdatedAt = &updatedAt

		case "timezone":
			if value.Type == gjson.Null {
				session.Timezone = &workspace.DefaultUserTimezone
			} else {
				session.Timezone = StringPtr(value.String())
			}

		case "device_external_id":
			if value.Type == gjson.Null {
				session.DeviceID = nil
			} else {
				session.DeviceID = StringPtr(ComputeDeviceID(value.String()))
			}

		case "landing_page":
			if value.Type == gjson.Null {
				session.LandingPage = NewNullableString(nil)
			} else {
				landingPage := value.String()
				session.LandingPage = NewNullableString(&landingPage)

				if !govalidator.IsRequestURL(landingPage) {
					err = eris.New("session.landing_page is invalid")
					return false
				}

				u, err := url.Parse(landingPage)
				if err != nil {
					err = eris.Wrap(err, "session.landing_page")
					return false
				}

				// get path
				path := u.Path
				if path == "" {
					path = "/"
				}
				session.LandingPagePath = &path
			}

		case "referrer":
			if value.Type == gjson.Null {
				session.Referrer = NewNullableString(nil)
			} else {
				referrer := value.String()
				session.Referrer = NewNullableString(&referrer)

				if !govalidator.IsRequestURL(referrer) {
					err = eris.New("session referrer is invalid")
					return false
				}

				u, err := url.Parse(referrer)
				if err != nil {
					err = eris.Wrap(err, "session referrer")
					return false
				}

				session.ReferrerDomain = &u.Host

				// get path
				path := u.Path
				if path == "" {
					path = "/"
				}
				session.ReferrerPath = &path
			}

		case "duration":
			if value.Type == gjson.Null {
				session.Duration = NewNullableInt64(nil)
			} else {
				duration := value.Int()
				if duration < 0 {
					err = eris.New("session.duration cannot be negative")
					return false
				}
				session.Duration = NewNullableInt64(&duration)
			}

		case "pageviews_count":
			if value.Type == gjson.Null {
				session.PageviewsCount = nil
			} else {
				count := value.Int()
				if count < 0 {
					err = eris.New("session.pageviews_count cannot be negative")
					return false
				}
				session.PageviewsCount = &count
			}

		case "interactions_count":
			if value.Type == gjson.Null {
				session.InteractionsCount = nil
			} else {
				count := value.Int()
				if count < 0 {
					err = eris.New("session.interactions_count cannot be negative")
					return false
				}
				session.InteractionsCount = &count
			}

		case "utm_source":
			if value.Type == gjson.Null {
				session.UTMSource = NewNullableString(nil)
			} else {
				session.UTMSource = NewNullableString(StringPtr(value.String()))
			}

		case "utm_medium":
			if value.Type == gjson.Null {
				session.UTMMedium = NewNullableString(nil)
			} else {
				session.UTMMedium = NewNullableString(StringPtr(value.String()))
			}

		case "utm_id":
			if value.Type == gjson.Null {
				session.UTMID = NewNullableString(nil)
			} else {
				session.UTMID = NewNullableString(StringPtr(value.String()))
			}

		case "utm_id_from":
			if value.Type == gjson.Null {
				session.UTMIDFrom = NewNullableString(nil)
			} else {
				session.UTMIDFrom = NewNullableString(StringPtr(value.String()))
			}

		case "utm_campaign":
			if value.Type == gjson.Null {
				session.UTMCampaign = NewNullableString(nil)
			} else {
				session.UTMCampaign = NewNullableString(StringPtr(value.String()))
			}

		case "utm_content":
			if value.Type == gjson.Null {
				session.UTMContent = NewNullableString(nil)
			} else {
				session.UTMContent = NewNullableString(StringPtr(value.String()))
			}

		case "utm_term":
			if value.Type == gjson.Null {
				session.UTMTerm = NewNullableString(nil)
			} else {
				session.UTMTerm = NewNullableString(StringPtr(value.String()))
			}

		case "via_utm_source":
			if value.Type == gjson.Null {
				session.ViaUTMSource = NewNullableString(nil)
			} else {
				session.ViaUTMSource = NewNullableString(StringPtr(value.String()))
			}

		case "via_utm_medium":
			if value.Type == gjson.Null {
				session.ViaUTMMedium = NewNullableString(nil)
			} else {
				session.ViaUTMMedium = NewNullableString(StringPtr(value.String()))
			}

		case "via_utm_id":
			if value.Type == gjson.Null {
				session.ViaUTMID = NewNullableString(nil)
			} else {
				session.ViaUTMID = NewNullableString(StringPtr(value.String()))
			}

		case "via_utm_id_from":
			if value.Type == gjson.Null {
				session.ViaUTMIDFrom = NewNullableString(nil)
			} else {
				session.ViaUTMIDFrom = NewNullableString(StringPtr(value.String()))
			}

		case "via_utm_campaign":
			if value.Type == gjson.Null {
				session.ViaUTMCampaign = NewNullableString(nil)
			} else {
				session.ViaUTMCampaign = NewNullableString(StringPtr(value.String()))
			}

		case "via_utm_content":
			if value.Type == gjson.Null {
				session.ViaUTMContent = NewNullableString(nil)
			} else {
				session.ViaUTMContent = NewNullableString(StringPtr(value.String()))
			}

		case "via_utm_term":
			if value.Type == gjson.Null {
				session.ViaUTMTerm = NewNullableString(nil)
			} else {
				session.ViaUTMTerm = NewNullableString(StringPtr(value.String()))
			}

		default:
			// ignore other fields and handle app_ extra columns
			if (strings.HasPrefix(fieldName, "app_") || strings.HasPrefix(fieldName, "appx_")) && extraColumns != nil {

				// check if column exists in app table
				for _, col := range extraColumns {

					if col.Name == fieldName {

						fieldValue, errExtract := ExtractFieldValueFromGJSON(col, value, 0)
						if errExtract != nil {
							err = eris.Wrapf(errExtract, "extract field %s", col.Name)
							return false
						}
						session.ExtraColumns[col.Name] = fieldValue
					}
				}
			}
		}

		return true
	})

	if err != nil {
		return nil, err
	}

	if session.DomainID == "" && dataLog.DomainID != nil {
		session.DomainID = *dataLog.DomainID
	}

	// use data import createdAt as updatedAt if not provided
	if session.UpdatedAt == nil {
		session.UpdatedAt = &session.CreatedAt
	}

	if session.Timezone == nil {
		session.Timezone = &workspace.DefaultUserTimezone
	}

	// enrich session with device
	if dataLog.UpsertedDevice != nil && session.DeviceID == nil {
		session.DeviceID = &dataLog.UpsertedDevice.ID
	}

	// Validation

	// verify that domainID exists
	found := false
	for _, domain := range workspace.Domains {
		if domain.ID == session.DomainID {
			found = true
			break
		}
	}

	if !found {
		return nil, eris.New("order domain_id invalid")
	}

	if session.ExternalID == "" {
		return nil, eris.New("session.external_id is required")
	}

	if session.CreatedAt.IsZero() {
		return nil, eris.New("session.created_at is required")
	}

	if !govalidator.IsIn(*session.Timezone, common.Timezones...) {
		return nil, eris.Errorf("session.timezone %s is invalid", *session.Timezone)
	}

	// error if utm_source is empty string or nil
	if session.UTMSource == nil || session.UTMSource.IsNull || session.UTMSource.String == "" {
		return nil, eris.New("utm_source is required")
	}

	if session.UTMMedium == nil || session.UTMMedium.IsNull || session.UTMMedium.String == "" {
		return nil, eris.New("utm_medium is required")
	}

	// find channel from mapping
	if channel, origin := session.FindChannelFromOrigin(workspace.Channels); channel != nil {
		session.SetChannelOriginID(origin.ID)
		session.SetChannelID(channel.ID)
		session.SetChannelGroupID(channel.GroupID)
	} else {
		// default values
		session.SetChannelID(ChannelNotMapped)
		session.SetChannelGroupID(ChannelNotMapped)
	}

	return session, nil
}

// index on utm_id: it contains gclid value, which will be enriched by the GoogleAds integration
var SessionSchema string = `CREATE TABLE IF NOT EXISTS session (
  id VARCHAR(64) NOT NULL,
  external_id VARCHAR(255) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  domain_id VARCHAR(64) NOT NULL,
  device_id VARCHAR(64),
  created_at DATETIME NOT NULL,
  created_at_trunc AS DATE_TRUNC('hour', created_at) PERSISTED DATETIME,
  db_created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  is_deleted BOOLEAN DEFAULT FALSE,
  merged_from_user_id VARCHAR(64),
  fields_timestamp JSON NOT NULL,

  timezone VARCHAR(30) NOT NULL,
  year AS YEAR(CONVERT_TZ(created_at, 'UTC', timezone)) PERSISTED SMALLINT UNSIGNED,
  month AS MONTH(CONVERT_TZ(created_at, 'UTC', timezone)) PERSISTED TINYINT UNSIGNED,
  month_day AS DAY(CONVERT_TZ(created_at, 'UTC', timezone)) PERSISTED TINYINT UNSIGNED,
  week_day AS WEEKDAY(CONVERT_TZ(created_at, 'UTC', timezone)) PERSISTED TINYINT UNSIGNED,
  hour AS HOUR(CONVERT_TZ(created_at, 'UTC', timezone)) PERSISTED TINYINT UNSIGNED,

  -- expires_at DATETIME NOT NULL,
  duration SMALLINT UNSIGNED DEFAULT 0,
  bounced AS IF(duration >= 15 OR interactions_count > 1, false, true) PERSISTED BOOLEAN,
  pageviews_count SMALLINT UNSIGNED DEFAULT 0,
  interactions_count SMALLINT UNSIGNED DEFAULT 0,

  landing_page VARCHAR(2083),
  landing_page_path VARCHAR(2083),
  referrer VARCHAR(2083),
  referrer_domain VARCHAR(255),
  referrer_path VARCHAR(255),

  channel_origin_id VARCHAR(255) NOT NULL,
  channel_id VARCHAR(255) NOT NULL,
  channel_group_id VARCHAR(255) NOT NULL,

  utm_source VARCHAR(255) NOT NULL,
  utm_medium VARCHAR(255) NOT NULL,
  utm_campaign VARCHAR(255),
  utm_id VARCHAR(255),
  utm_id_from VARCHAR(20),
  utm_content VARCHAR(255),
  utm_term VARCHAR(255),

  via_utm_medium VARCHAR(255),
  via_utm_source VARCHAR(255),
  via_utm_id VARCHAR(255),
  via_utm_id_from VARCHAR(255),
  via_utm_campaign VARCHAR(255),
  via_utm_content VARCHAR(255),
  via_utm_term VARCHAR(255),
  
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

  SORT KEY (created_at_trunc DESC),
  PRIMARY KEY (id, user_id),
  KEY (conversion_id) USING HASH,
  KEY (utm_id) USING HASH,
  KEY (user_id) USING HASH, -- for merging
  SHARD KEY (user_id)
) COLLATE utf8mb4_general_ci;`

var SessionSchemaMYSQL string = `CREATE TABLE IF NOT EXISTS session (
  id VARCHAR(64) NOT NULL,
  external_id VARCHAR(255) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  domain_id VARCHAR(64) NOT NULL,
  device_id VARCHAR(64),
  created_at DATETIME NOT NULL,
  created_at_trunc DATETIME GENERATED ALWAYS AS (DATE_FORMAT(created_at, '%Y-%m-%d %H:00:00')) STORED,
  db_created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  is_deleted BOOLEAN DEFAULT FALSE,
  merged_from_user_id VARCHAR(64),
  fields_timestamp JSON NOT NULL,

  timezone VARCHAR(30) NOT NULL,
  year SMALLINT UNSIGNED GENERATED ALWAYS AS (YEAR(CONVERT_TZ(created_at, 'UTC', timezone))) STORED,
  month TINYINT UNSIGNED GENERATED ALWAYS AS (MONTH(CONVERT_TZ(created_at, 'UTC', timezone))) STORED,
  month_day TINYINT UNSIGNED GENERATED ALWAYS AS (DAY(CONVERT_TZ(created_at, 'UTC', timezone))) STORED,
  week_day TINYINT UNSIGNED GENERATED ALWAYS AS (WEEKDAY(CONVERT_TZ(created_at, 'UTC', timezone))) STORED,
  hour TINYINT UNSIGNED GENERATED ALWAYS AS (HOUR(CONVERT_TZ(created_at, 'UTC', timezone))) STORED,

  -- expires_at DATETIME NOT NULL,
  duration SMALLINT UNSIGNED DEFAULT 0,
  bounced BOOLEAN GENERATED ALWAYS AS (IF(duration >= 15 OR interactions_count > 1, false, true)) STORED,
  pageviews_count SMALLINT UNSIGNED DEFAULT 0,
  interactions_count SMALLINT UNSIGNED DEFAULT 0,

  landing_page VARCHAR(2083),
  landing_page_path VARCHAR(2083),
  referrer VARCHAR(2083),
  referrer_domain VARCHAR(255),
  referrer_path VARCHAR(255),

  channel_origin_id VARCHAR(255) NOT NULL,
  channel_id VARCHAR(255) NOT NULL,
  channel_group_id VARCHAR(255) NOT NULL,

  utm_source VARCHAR(255) NOT NULL,
  utm_medium VARCHAR(255) NOT NULL,
  utm_campaign VARCHAR(255),
  utm_id VARCHAR(255),
  utm_id_from VARCHAR(20),
  utm_content VARCHAR(255),
  utm_term VARCHAR(255),

  via_utm_medium VARCHAR(255),
  via_utm_source VARCHAR(255),
  via_utm_id VARCHAR(255),
  via_utm_id_from VARCHAR(255),
  via_utm_campaign VARCHAR(255),
  via_utm_content VARCHAR(255),
  via_utm_term VARCHAR(255),
  
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
  KEY (utm_id) USING HASH,
  KEY (user_id) USING HASH -- for merging
  -- SHARD KEY (user_id)
) COLLATE utf8mb4_general_ci;`

func NewSessionCube() *CubeJSSchema {
	return &CubeJSSchema{
		Title:       "Sessions",
		Description: "Sessions",
		SQL:         "SELECT * FROM `session`",
		Joins: map[string]CubeJSSchemaJoin{
			"User": {
				Relationship: "many_to_one",
				SQL:          "${CUBE}.user_id = ${User}.id",
			},
			"Device": {
				Relationship: "many_to_one",
				SQL:          "${CUBE}.user_id = ${Device}.user_id AND ${CUBE}.device_id = ${Device}.id",
			},
			"Order": {
				Relationship: "many_to_one",
				SQL:          "${CUBE}.user_id = ${Order}.user_id AND ${CUBE}.conversion_id = ${Order}.id",
			},
		},
		Measures: map[string]CubeJSSchemaMeasure{
			"count": {
				Type:        "count",
				Title:       "Sessions",
				Description: "count of all sessions",
			},
			"unique_users": {
				Type:        "countDistinct",
				SQL:         "user_id",
				Title:       "Unique users",
				Description: "count of distinct user_id",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"users_last_24h": {
				Type:        "countDistinct",
				SQL:         "user_id",
				Title:       "Users last 24h",
				Description: "count of distinct user_id in last 24h",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "created_at > date_sub(now(), interval 24 hour)"},
				},
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"users_online": {
				Type:        "countDistinct",
				SQL:         "user_id",
				Title:       "Users online",
				Description: "count of distinct user_id in last 3 minutes",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "created_at > date_sub(now(), interval 3 minute)"},
				},
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"contributions_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "Contributions count",
				Description: "count of: conversion_id IS NOT NULL",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "${CUBE}.conversion_id IS NOT NULL"},
				},
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"orders_contributions": {
				Type:        "count",
				SQL:         "id",
				Title:       "Orders contributions",
				Description: "Sessions that contributed to an order. Count of: conversion_id IS NOT NULL AND conversion_type = 'order'",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "${CUBE}.conversion_id IS NOT NULL AND ${CUBE}.conversion_type = 'order'"},
				},
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"distinct_orders": {
				Type:        "countDistinct",
				SQL:         "conversion_id",
				Title:       "Distinct orders",
				Description: "count of distinct conversion_id, where conversion_type = 'order'",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "${CUBE}.conversion_type = 'order'"},
				},
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"distinct_conversions": {
				Type:        "countDistinct",
				SQL:         "conversion_id",
				Title:       "Distinct conversions",
				Description: "Count of distinct conversion_id",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"conversion_rate": {
				Type:        "number",
				SQL:         "(${distinct_conversions}) / ${count}",
				Title:       "Conversion rate",
				Description: "Conversion rate: distinct_conversions / count",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
					"rimdian_format":         "percentage",
				},
			},
			"bounce_rate": {
				Type:        "number",
				SQL:         "COALESCE(ROUND(AVG(bounced), 4), 0)",
				Title:       "Bounce rate",
				Description: "Bounce rate: AVG(bounced)",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
					"rimdian_format":         "percentage",
				},
			},
			"avg_duration": {
				Type:        "number",
				SQL:         "COALESCE(AVG(CASE WHEN duration > 0 THEN duration ELSE NULL END), 0)",
				Title:       "Avg duration",
				Description: "Avg duration: AVG(CASE WHEN duration > 0 THEN duration ELSE NULL END)",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
					"rimdian_format":         "duration",
				},
			},
			"pageviews_sum": {
				Type:        "number",
				SQL:         "COALESCE(SUM(pageviews_count), 0)",
				Title:       "Pageviews sum",
				Description: "Sum of pageviews_count",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"avg_pageviews_count": {
				Type:        "number",
				SQL:         "COALESCE(AVG(pageviews_count), 0)",
				Title:       "Avg pageviews count",
				Description: "Avg of pageviews_count",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"interactions_count": {
				Type:        "number",
				SQL:         "COALESCE(SUM(interactions_count), 0)",
				Title:       "Interactions count",
				Description: "Sum of interactions_count",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"linear_amount_attributed": {
				Type:        "number",
				SQL:         "COALESCE(SUM(linear_amount_attributed), 0)",
				Title:       "Linear amount attributed",
				Description: "Sum of linear_amount_attributed",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
					"rimdian_format":         "currency",
				},
			},
			"linear_percentage_attributed": {
				Type:        "number",
				SQL:         "COALESCE(ROUND(AVG(linear_percentage_attributed) / 10000, 2), 0)",
				Title:       "Linear percentage attributed",
				Description: "Avg of linear_percentage_attributed",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
					"rimdian_format":         "percentage",
				},
			},
			"linear_conversions_attributed": {
				Type:        "number",
				SQL:         "COALESCE(ROUND(SUM(linear_percentage_attributed) / 10000, 2), 0)",
				Title:       "Linear conversions attributed",
				Description: "Sum of linear_percentage_attributed",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			// "alone_count": {
			// 	Type:        "count",
			// 	SQL:         "id",
			// 	Title:       "Alone role count",
			// 	Description: "count of: role = 0 AND conversion_id IS NOT NULL",
			// 	Filters: []CubeJSSchemaMeasureFilter{
			// 		{SQL: "${CUBE}.role = 0 AND ${CUBE}.conversion_id IS NOT NULL"},
			// 	},
			// 	Meta: MapOfInterfaces{
			// 		"hide_from_segmentation": true,
			// 	},
			// },
			// "initiator_count": {
			// 	Type:        "count",
			// 	SQL:         "id",
			// 	Title:       "Initiator role count",
			// 	Description: "count of: role = 1 AND conversion_id IS NOT NULL",
			// 	Filters: []CubeJSSchemaMeasureFilter{
			// 		{SQL: "${CUBE}.role = 1 AND ${CUBE}.conversion_id IS NOT NULL"},
			// 	},
			// 	Meta: MapOfInterfaces{
			// 		"hide_from_segmentation": true,
			// 	},
			// },
			// "assisting_count": {
			// 	Type:        "count",
			// 	SQL:         "id",
			// 	Title:       "Assisting role count",
			// 	Description: "count of: role = 2 AND conversion_id IS NOT NULL",
			// 	Filters: []CubeJSSchemaMeasureFilter{
			// 		{SQL: "${CUBE}.role = 2 AND ${CUBE}.conversion_id IS NOT NULL"},
			// 	},
			// 	Meta: MapOfInterfaces{
			// 		"hide_from_segmentation": true,
			// 	},
			// },
			// "closer_count": {
			// 	Type:        "count",
			// 	SQL:         "id",
			// 	Title:       "Closer role count",
			// 	Description: "count of: role = 3 AND conversion_id IS NOT NULL",
			// 	Filters: []CubeJSSchemaMeasureFilter{
			// 		{SQL: "${CUBE}.role = 3 AND ${CUBE}.conversion_id IS NOT NULL"},
			// 	},
			// 	Meta: MapOfInterfaces{
			// 		"hide_from_segmentation": true,
			// 	},
			// },
			// "alone_ratio": {
			// 	Type:        "number",
			// 	SQL:         "COALESCE(${alone_count} / ${contributions_count}, 0)",
			// 	Title:       "Alone role ratio",
			// 	Description: "ratio of: alone_count / contributions_count",
			// 	Meta: MapOfInterfaces{
			// 		"hide_from_segmentation": true,
			// 	},
			// },

			// "initiator_ratio": {
			// 	Type:        "number",
			// 	SQL:         "COALESCE(${initiator_count} / ${contributions_count}, 0)",
			// 	Title:       "Initiator role ratio",
			// 	Description: "ratio of: initiator_count / contributions_count",
			// 	Meta: MapOfInterfaces{
			// 		"hide_from_segmentation": true,
			// 		"rimdian_format":         "percentage",
			// 	},
			// },
			// "assisting_ratio": {
			// 	Type:        "number",
			// 	SQL:         "COALESCE(${assisting_count} / ${contributions_count}, 0)",
			// 	Title:       "Assisting role ratio",
			// 	Description: "ratio of: assisting_count / contributions_count",
			// 	Meta: MapOfInterfaces{
			// 		"hide_from_segmentation": true,
			// 		"rimdian_format":         "percentage",
			// 	},
			// },
			// "closer_ratio": {
			// 	Type:        "number",
			// 	SQL:         "COALESCE(${closer_count} / ${contributions_count}, 0)",
			// 	Title:       "Closer role ratio",
			// 	Description: "ratio of: closer_count / contributions_count",
			// 	Meta: MapOfInterfaces{
			// 		"hide_from_segmentation": true,
			// 		"rimdian_format":         "percentage",
			// 	},
			// },
			// "alone_linear_conversions_attributed": {
			// 	Type:        "number",
			// 	SQL:         "COALESCE(ROUND(SUM(CASE WHEN ${CUBE}.role = 0 THEN ${CUBE}.linear_percentage_attributed ELSE 0 END) / 10000, 2), 0)",
			// 	Title:       "Alone linear conversions attributed",
			// 	Description: "Sum of: CASE WHEN role = 0 THEN linear_percentage_attributed ELSE 0 END",
			// 	Meta: MapOfInterfaces{
			// 		"hide_from_segmentation": true,
			// 	},
			// },
			// "initiator_linear_conversions_attributed": {
			// 	Type:        "number",
			// 	SQL:         "COALESCE(ROUND(SUM(CASE WHEN ${CUBE}.role = 1 THEN ${CUBE}.linear_percentage_attributed ELSE 0 END) / 10000, 2), 0)",
			// 	Title:       "Initiator linear conversions attributed",
			// 	Description: "Sum of: CASE WHEN role = 1 THEN linear_percentage_attributed ELSE 0 END",
			// 	Meta: MapOfInterfaces{
			// 		"hide_from_segmentation": true,
			// 	},
			// },
			// "assisting_linear_conversions_attributed": {
			// 	Type:        "number",
			// 	SQL:         "COALESCE(ROUND(SUM(CASE WHEN ${CUBE}.role = 2 THEN ${CUBE}.linear_percentage_attributed ELSE 0 END) / 10000, 2), 0)",
			// 	Title:       "Assisting linear conversions attributed",
			// 	Description: "Sum of: CASE WHEN role = 2 THEN linear_percentage_attributed ELSE 0 END",
			// 	Meta: MapOfInterfaces{
			// 		"hide_from_segmentation": true,
			// 	},
			// },
			// "closer_linear_conversions_attributed": {
			// 	Type:        "number",
			// 	SQL:         "COALESCE(ROUND(SUM(CASE WHEN ${CUBE}.role = 3 THEN ${CUBE}.linear_percentage_attributed ELSE 0 END) / 10000, 2), 0)",
			// 	Title:       "Closer linear conversions attributed",
			// 	Description: "Sum of: CASE WHEN role = 3 THEN linear_percentage_attributed ELSE 0 END",
			// 	Meta: MapOfInterfaces{
			// 		"hide_from_segmentation": true,
			// 	},
			// },
			// "alone_linear_amount_attributed": {
			// 	Type:        "number",
			// 	SQL:         "COALESCE(SUM(CASE WHEN ${CUBE}.role = 0 THEN ${CUBE}.linear_amount_attributed ELSE 0 END), 0)",
			// 	Title:       "Alone linear amount attributed",
			// 	Description: "Sum of: CASE WHEN role = 0 THEN linear_amount_attributed ELSE 0 END",
			// 	Meta: MapOfInterfaces{
			// 		"hide_from_segmentation": true,
			// 		"rimdian_format":         "currency",
			// 	},
			// },
			// "initiator_linear_amount_attributed": {
			// 	Type:        "number",
			// 	SQL:         "COALESCE(SUM(CASE WHEN ${CUBE}.role = 1 THEN ${CUBE}.linear_amount_attributed ELSE 0 END), 0)",
			// 	Title:       "Initiator linear amount attributed",
			// 	Description: "Sum of: CASE WHEN role = 1 THEN linear_amount_attributed ELSE 0 END",
			// 	Meta: MapOfInterfaces{
			// 		"hide_from_segmentation": true,
			// 		"rimdian_format":         "currency",
			// 	},
			// },
			// "assisting_linear_amount_attributed": {
			// 	Type:        "number",
			// 	SQL:         "COALESCE(SUM(CASE WHEN ${CUBE}.role = 2 THEN ${CUBE}.linear_amount_attributed ELSE 0 END), 0)",
			// 	Title:       "Assisting linear amount attributed",
			// 	Description: "Sum of: CASE WHEN role = 2 THEN linear_amount_attributed ELSE 0 END",
			// 	Meta: MapOfInterfaces{
			// 		"hide_from_segmentation": true,
			// 		"rimdian_format":         "currency",
			// 	},
			// },
			// "closer_linear_amount_attributed": {
			// 	Type:        "number",
			// 	SQL:         "COALESCE(SUM(CASE WHEN ${CUBE}.role = 3 THEN ${CUBE}.linear_amount_attributed ELSE 0 END), 0)",
			// 	Title:       "Closer linear amount attributed",
			// 	Description: "Sum of: CASE WHEN role = 3 THEN linear_amount_attributed ELSE 0 END",
			// 	Meta: MapOfInterfaces{
			// 		"hide_from_segmentation": true,
			// 		"rimdian_format":         "currency",
			// 	},
			// },
			// Acquisition
			"acquisition_contributions_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "Acquisition: contributions",
				Description: "count of: conversion_id IS NOT NULL AND is_first_conversion = 1",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "${CUBE}.conversion_id IS NOT NULL AND ${CUBE}.is_first_conversion = 1"},
				},
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"acquisition_orders_contributions": {
				Type:        "count",
				SQL:         "id",
				Title:       "Acquisition: orders contributions",
				Description: "Sessions that contributed to a 1st order. Count of: conversion_id IS NOT NULL AND conversion_type = 'order' AND is_first_conversion = 1",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "${CUBE}.conversion_id IS NOT NULL AND ${CUBE}.conversion_type = 'order' AND  ${CUBE}.is_first_conversion = 1"},
				},
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"acquisition_alone_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "Acquisition: alone role count",
				Description: "count of: role = 0 AND conversion_id IS NOT NULL AND is_first_conversion = 1",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "${CUBE}.role = 0 AND ${CUBE}.conversion_id IS NOT NULL AND ${CUBE}.is_first_conversion = 1"},
				},
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"acquisition_initiator_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "Acquisition: initiator role count",
				Description: "count of: role = 1 AND conversion_id IS NOT NULL AND is_first_conversion = 1",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "${CUBE}.role = 1 AND ${CUBE}.conversion_id IS NOT NULL AND ${CUBE}.is_first_conversion = 1"},
				},
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"acquisition_assisting_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "Acquisition: assisting role count",
				Description: "count of: role = 2 AND conversion_id IS NOT NULL AND is_first_conversion = 1",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "${CUBE}.role = 2 AND ${CUBE}.conversion_id IS NOT NULL AND ${CUBE}.is_first_conversion = 1"},
				},
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"acquisition_closer_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "Acquisition: closer role count",
				Description: "count of: role = 3 AND conversion_id IS NOT NULL AND is_first_conversion = 1",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "${CUBE}.role = 3 AND ${CUBE}.conversion_id IS NOT NULL AND ${CUBE}.is_first_conversion = 1"},
				},
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"acquisition_alone_ratio": {
				Type:        "number",
				Title:       "Acquisition: alone role ratio",
				SQL:         "COALESCE(${acquisition_alone_count} / ${acquisition_contributions_count}, 0)",
				Description: "ratio of: acquisition_alone_count / acquisition_contributions_count",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},

			"acquisition_initiator_ratio": {
				Type:        "number",
				SQL:         "COALESCE(${acquisition_initiator_count} / ${acquisition_contributions_count}, 0)",
				Title:       "Acquisition: initiator role ratio",
				Description: "ratio of: acquisition_initiator_count / acquisition_contributions_count",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
					"rimdian_format":         "percentage",
				},
			},
			"acquisition_assisting_ratio": {
				Type:        "number",
				SQL:         "COALESCE(${acquisition_assisting_count} / ${acquisition_contributions_count}, 0)",
				Title:       "Acquisition: assisting role ratio",
				Description: "ratio of: acquisition_assisting_count / acquisition_contributions_count",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
					"rimdian_format":         "percentage",
				},
			},
			"acquisition_closer_ratio": {
				Type:        "number",
				SQL:         "COALESCE(${acquisition_closer_count} / ${acquisition_contributions_count}, 0)",
				Title:       "Acquisition: closer role ratio",
				Description: "ratio of: acquisition_closer_count / acquisition_contributions_count",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
					"rimdian_format":         "percentage",
				},
			},
			"acquisition_alone_linear_conversions_attributed": {
				Type:        "number",
				SQL:         "COALESCE(ROUND(SUM(CASE WHEN ${CUBE}.role = 0 THEN ${CUBE}.linear_percentage_attributed ELSE 0 END) / 10000, 2), 0)",
				Title:       "Acquisition: Alone linear conversions attributed",
				Description: "Sum of: CASE WHEN role = 0 THEN linear_percentage_attributed ELSE 0 END, WHERE is_first_conversion = 1",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "${CUBE}.is_first_conversion = 1"},
				},
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"acquisition_initiator_linear_conversions_attributed": {
				Type:        "number",
				SQL:         "COALESCE(ROUND(SUM(CASE WHEN (${CUBE}.role = 1 AND ${CUBE}.is_first_conversion = 1) THEN ${CUBE}.linear_percentage_attributed ELSE 0 END) / 10000, 2), 0)",
				Title:       "Acquisition: Initiator linear conversions attributed",
				Description: "Sum of: CASE WHEN role = 1 THEN linear_percentage_attributed ELSE 0 END, WHERE is_first_conversion = 1",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"acquisition_assisting_linear_conversions_attributed": {
				Type:        "number",
				SQL:         "COALESCE(ROUND(SUM(CASE WHEN (${CUBE}.role = 2 AND ${CUBE}.is_first_conversion = 1) THEN ${CUBE}.linear_percentage_attributed ELSE 0 END) / 10000, 2), 0)",
				Title:       "Acquisition: Assisting linear conversions attributed",
				Description: "Sum of: CASE WHEN role = 2 THEN linear_percentage_attributed ELSE 0 END, WHERE is_first_conversion = 1",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"acquisition_closer_linear_conversions_attributed": {
				Type:        "number",
				SQL:         "COALESCE(ROUND(SUM(CASE WHEN (${CUBE}.role = 3 AND ${CUBE}.is_first_conversion = 1) THEN ${CUBE}.linear_percentage_attributed ELSE 0 END) / 10000, 2), 0)",
				Title:       "Acquisition: Closer linear conversions attributed",
				Description: "Sum of: CASE WHEN role = 3 THEN linear_percentage_attributed ELSE 0 END, WHERE is_first_conversion = 1",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"acquisition_alone_linear_amount_attributed": {
				Type:        "number",
				SQL:         "COALESCE(SUM(CASE WHEN (${CUBE}.role = 0 AND ${CUBE}.is_first_conversion = 1) THEN ${CUBE}.linear_amount_attributed ELSE 0 END), 0)",
				Title:       "Acquisition: alone linear amount attributed",
				Description: "Sum of: CASE WHEN role = 0 THEN linear_amount_attributed ELSE 0 END, WHERE is_first_conversion = 1",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
					"rimdian_format":         "currency",
				},
			},
			"acquisition_initiator_linear_amount_attributed": {
				Type:        "number",
				SQL:         "COALESCE(SUM(CASE WHEN (${CUBE}.role = 1 AND ${CUBE}.is_first_conversion = 1) THEN ${CUBE}.linear_amount_attributed ELSE 0 END), 0)",
				Title:       "Acquisition: initiator linear amount attributed",
				Description: "Sum of: CASE WHEN role = 1 THEN linear_amount_attributed ELSE 0 END, WHERE is_first_conversion = 1",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
					"rimdian_format":         "currency",
				},
			},
			"acquisition_assisting_linear_amount_attributed": {
				Type:        "number",
				SQL:         "COALESCE(SUM(CASE WHEN (${CUBE}.role = 2 AND ${CUBE}.is_first_conversion = 1) THEN ${CUBE}.linear_amount_attributed ELSE 0 END), 0)",
				Title:       "Acquisition: assisting linear amount attributed",
				Description: "Sum of: CASE WHEN role = 2 THEN linear_amount_attributed ELSE 0 END, WHERE is_first_conversion = 1",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
					"rimdian_format":         "currency",
				},
			},
			"acquisition_closer_linear_amount_attributed": {
				Type:        "number",
				SQL:         "COALESCE(SUM(CASE WHEN (${CUBE}.role = 3 AND ${CUBE}.is_first_conversion = 1) THEN ${CUBE}.linear_amount_attributed ELSE 0 END), 0)",
				Title:       "Acquisition: closer linear amount attributed",
				Description: "Sum of: CASE WHEN role = 3 THEN linear_amount_attributed ELSE 0 END, WHERE is_first_conversion = 1",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
					"rimdian_format":         "currency",
				},
			},
			"acquisition_linear_amount_attributed": {
				Type:        "number",
				SQL:         "COALESCE(SUM(CASE WHEN ${CUBE}.is_first_conversion = 1 THEN ${CUBE}.linear_amount_attributed ELSE 0 END), 0)",
				Title:       "Acquisition: Linear amount attributed",
				Description: "Sum of linear_amount_attributed WHERE is_first_conversion = 1",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
					"rimdian_format":         "currency",
				},
			},
			"acquisition_linear_percentage_attributed": {
				Type:        "number",
				SQL:         "COALESCE(ROUND(AVG(linear_percentage_attributed) / 10000, 2), 0)",
				Title:       "Acquisition: Linear percentage attributed",
				Description: "Avg of linear_percentage_attributed WHERE is_first_conversion = 1",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "${CUBE}.is_first_conversion = 1"},
				},
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
					"rimdian_format":         "percentage",
				},
			},
			"acquisition_linear_conversions_attributed": {
				Type:        "number",
				SQL:         "COALESCE(ROUND(SUM(CASE WHEN ${CUBE}.is_first_conversion = 1 THEN ${CUBE}.linear_percentage_attributed ELSE 0 END) / 10000, 2), 0)",
				Title:       "Acquisition: Linear conversions attributed",
				Description: "Sum of linear_percentage_attributed WHERE is_first_conversion = 1",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},

			// RETENTION
			"retention_contributions_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "Retention: Contributions",
				Description: "count of: conversion_id IS NOT NULL AND is_first_conversion = 0",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "${CUBE}.conversion_id IS NOT NULL AND ${CUBE}.is_first_conversion = 0"},
				},
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"retention_orders_contributions": {
				Type:        "count",
				SQL:         "id",
				Title:       "Retention: orders contributions",
				Description: "Sessions that contributed to a 1st order. Count of: conversion_id IS NOT NULL AND conversion_type = 'order' AND is_first_conversion = 1",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "${CUBE}.conversion_id IS NOT NULL AND ${CUBE}.conversion_type = 'order' AND  ${CUBE}.is_first_conversion = 0"},
				},
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"retention_alone_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "Retention: alone role count",
				Description: "count of: role = 0 AND conversion_id IS NOT NULL AND is_first_conversion = 0",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "${CUBE}.role = 0 AND ${CUBE}.conversion_id IS NOT NULL AND ${CUBE}.is_first_conversion = 0"},
				},
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"retention_initiator_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "Retention: initiator role count",
				Description: "count of: role = 1 AND conversion_id IS NOT NULL AND is_first_conversion = 0",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "${CUBE}.role = 1 AND ${CUBE}.conversion_id IS NOT NULL AND ${CUBE}.is_first_conversion = 0"},
				},
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"retention_assisting_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "Retention: assisting role count",
				Description: "count of: role = 2 AND conversion_id IS NOT NULL AND is_first_conversion = 0",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "${CUBE}.role = 2 AND ${CUBE}.conversion_id IS NOT NULL AND ${CUBE}.is_first_conversion = 0"},
				},
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"retention_closer_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "Retention: closer role count",
				Description: "count of: role = 3 AND conversion_id IS NOT NULL AND is_first_conversion = 0",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "${CUBE}.role = 3 AND ${CUBE}.conversion_id IS NOT NULL AND ${CUBE}.is_first_conversion = 0"},
				},
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"retention_alone_ratio": {
				Type:        "number",
				Title:       "Retention: alone role ratio",
				SQL:         "COALESCE(${retention_alone_count} / ${retention_contributions_count}, 0)",
				Description: "ratio of: retention_alone_count / retention_contributions_count",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
					"rimdian_format":         "percentage",
				},
			},

			"retention_initiator_ratio": {
				Type:        "number",
				SQL:         "COALESCE(${retention_initiator_count} / ${retention_contributions_count}, 0)",
				Title:       "Retention: initiator role ratio",
				Description: "ratio of: retention_initiator_count / retention_contributions_count",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"retention_assisting_ratio": {
				Type:        "number",
				SQL:         "COALESCE(${retention_assisting_count} / ${retention_contributions_count}, 0)",
				Title:       "Retention: assisting role ratio",
				Description: "ratio of: retention_assisting_count / retention_contributions_count",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
					"rimdian_format":         "percentage",
				},
			},
			"retention_closer_ratio": {
				Type:        "number",
				SQL:         "COALESCE(${retention_closer_count} / ${retention_contributions_count}, 0)",
				Title:       "Retention: closer role ratio",
				Description: "ratio of: retention_closer_count / retention_contributions_count",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
					"rimdian_format":         "percentage",
				},
			},
			"retention_alone_linear_conversions_attributed": {
				Type:        "number",
				SQL:         "COALESCE(ROUND(SUM(CASE WHEN (${CUBE}.role = 0 AND ${CUBE}.is_first_conversion = 0) THEN ${CUBE}.linear_percentage_attributed ELSE 0 END) / 10000, 2), 0)",
				Title:       "Retention: Alone linear conversions attributed",
				Description: "Sum of: CASE WHEN role = 0 THEN linear_percentage_attributed ELSE 0 END, WHERE is_first_conversion = 0",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"retention_initiator_linear_conversions_attributed": {
				Type:        "number",
				SQL:         "COALESCE(ROUND(SUM(CASE WHEN (${CUBE}.role = 1 AND ${CUBE}.is_first_conversion = 0) THEN ${CUBE}.linear_percentage_attributed ELSE 0 END) / 10000, 2), 0)",
				Title:       "Retention: Initiator linear conversions attributed",
				Description: "Sum of: CASE WHEN role = 1 THEN linear_percentage_attributed ELSE 0 END, WHERE is_first_conversion = 0",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"retention_assisting_linear_conversions_attributed": {
				Type:        "number",
				SQL:         "COALESCE(ROUND(SUM(CASE WHEN (${CUBE}.role = 2 AND ${CUBE}.is_first_conversion = 0) THEN ${CUBE}.linear_percentage_attributed ELSE 0 END) / 10000, 2), 0)",
				Title:       "Retention: Assisting linear conversions attributed",
				Description: "Sum of: CASE WHEN role = 2 THEN linear_percentage_attributed ELSE 0 END, WHERE is_first_conversion = 0",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"retention_closer_linear_conversions_attributed": {
				Type:        "number",
				SQL:         "COALESCE(ROUND(SUM(CASE WHEN (${CUBE}.role = 3 AND ${CUBE}.is_first_conversion = 0) THEN ${CUBE}.linear_percentage_attributed ELSE 0 END) / 10000, 2), 0)",
				Title:       "Retention: Closer linear conversions attributed",
				Description: "Sum of: CASE WHEN role = 3 THEN linear_percentage_attributed ELSE 0 END, WHERE is_first_conversion = 0",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"retention_alone_linear_amount_attributed": {
				Type:        "number",
				SQL:         "COALESCE(SUM(CASE WHEN (${CUBE}.role = 0 AND ${CUBE}.is_first_conversion = 0) THEN ${CUBE}.linear_amount_attributed ELSE 0 END), 0)",
				Title:       "Retention: Alone linear amount attributed",
				Description: "Sum of: CASE WHEN role = 0 THEN linear_amount_attributed ELSE 0 END, WHERE is_first_conversion = 0",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
					"rimdian_format":         "currency",
				},
			},
			"retention_initiator_linear_amount_attributed": {
				Type:        "number",
				SQL:         "COALESCE(SUM(CASE WHEN (${CUBE}.role = 1 AND ${CUBE}.is_first_conversion = 0) THEN ${CUBE}.linear_amount_attributed ELSE 0 END), 0)",
				Title:       "Retention: Initiator linear amount attributed",
				Description: "Sum of: CASE WHEN role = 1 THEN linear_amount_attributed ELSE 0 END, WHERE is_first_conversion = 0",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
					"rimdian_format":         "currency",
				},
			},
			"retention_assisting_linear_amount_attributed": {
				Type:        "number",
				SQL:         "COALESCE(SUM(CASE WHEN (${CUBE}.role = 2 AND ${CUBE}.is_first_conversion = 0) THEN ${CUBE}.linear_amount_attributed ELSE 0 END), 0)",
				Title:       "Retention: Assisting linear amount attributed",
				Description: "Sum of: CASE WHEN role = 2 THEN linear_amount_attributed ELSE 0 END, WHERE is_first_conversion = 0",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
					"rimdian_format":         "currency",
				},
			},
			"retention_closer_linear_amount_attributed": {
				Type:        "number",
				SQL:         "COALESCE(SUM(CASE WHEN (${CUBE}.role = 3 AND ${CUBE}.is_first_conversion = 0) THEN ${CUBE}.linear_amount_attributed ELSE 0 END), 0)",
				Title:       "Retention: Closer linear amount attributed",
				Description: "Sum of: CASE WHEN role = 3 THEN linear_amount_attributed ELSE 0 END, WHERE is_first_conversion = 0",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
					"rimdian_format":         "currency",
				},
			},
			"retention_linear_amount_attributed": {
				Type:        "number",
				SQL:         "COALESCE(SUM(CASE WHEN ${CUBE}.is_first_conversion = 0 THEN ${CUBE}.linear_amount_attributed ELSE 0 END), 0)",
				Title:       "Retention: Linear amount attributed",
				Description: "Sum of linear_amount_attributed WHERE is_first_conversion = 0",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
					"rimdian_format":         "currency",
				},
			},
			"retention_linear_percentage_attributed": {
				Type:        "number",
				SQL:         "COALESCE(ROUND(AVG(linear_percentage_attributed) / 10000, 2), 0)",
				Title:       "Retention: Linear percentage attributed",
				Description: "Avg of linear_percentage_attributed WHERE is_first_conversion = 0",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "${CUBE}.is_first_conversion = 0"},
				},
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
					"rimdian_format":         "percentage",
				},
			},
			"retention_linear_conversions_attributed": {
				Type:        "number",
				SQL:         "COALESCE(ROUND(SUM(CASE WHEN ${CUBE}.is_first_conversion = 0 THEN ${CUBE}.linear_percentage_attributed ELSE 0 END) / 10000, 2), 0)",
				Title:       "Retention: Linear conversions attributed",
				Description: "Sum of linear_percentage_attributed WHERE is_first_conversion = 0",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
		},
		Dimensions: map[string]CubeJSSchemaDimension{

			"id": {
				SQL:         "id",
				Type:        "string",
				PrimaryKey:  true,
				Title:       "Session ID",
				Description: "field: id",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},

			"user_id": {
				Type:        "string",
				SQL:         "user_id",
				Title:       "User ID",
				Description: "field: user_id",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"domain_id": {
				Type:        "string",
				SQL:         "domain_id",
				Title:       "Domain ID",
				Description: "field: domain_id",
			},
			"device_id": {
				Type:        "string",
				SQL:         "device_id",
				Title:       "Device ID",
				Description: "field: device_id",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"created_at": {
				Type:        "time",
				SQL:         "created_at",
				Title:       "Created at",
				Description: "field: created_at",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"created_at_trunc": {
				Type:        "time",
				SQL:         "created_at_trunc",
				Title:       "Created at (truncated to hour)",
				Description: "field: created_at_trunc",
			},
			"timezone": {
				Type:        "string",
				SQL:         "timezone",
				Title:       "Timezone",
				Description: "field: timezone",
			},
			"year": {
				Type:        "number",
				SQL:         "year",
				Title:       "Year",
				Description: "field: year",
			},
			"month": {
				Type:        "number",
				SQL:         "month",
				Title:       "Month",
				Description: "field: month",
			},
			"month_day": {
				Type:        "number",
				SQL:         "month_day",
				Title:       "Month day",
				Description: "field: month_day",
			},
			"week_day": {
				Type:        "number",
				SQL:         "week_day",
				Title:       "Week day",
				Description: "field: week_day",
			},
			"hour": {
				Type:        "number",
				SQL:         "hour",
				Title:       "Hour",
				Description: "field: hour",
			},
			"duration": {
				Type:        "number",
				SQL:         "duration",
				Title:       "Duration",
				Description: "field: duration",
			},
			"bounced": {
				Type:        "number",
				SQL:         "bounced",
				Title:       "Bounced",
				Description: "field: bounced",
			},
			"landing_page": {
				Type:        "string",
				SQL:         "landing_page",
				Title:       "Landing page URL",
				Description: "field: landing_page",
			},
			"landing_page_path": {
				Type:        "string",
				SQL:         "landing_page_path",
				Title:       "Landing page URL /path",
				Description: "field: landing_page_path",
			},
			"referrer": {
				Type:        "string",
				SQL:         "referrer",
				Title:       "Referrer URL",
				Description: "field: referrer",
			},
			"referrer_domain": {
				Type:        "string",
				SQL:         "referrer_domain",
				Title:       "Referrer URL domain",
				Description: "field: referrer_domain",
			},
			"referrer_path": {
				Type:        "string",
				SQL:         "referrer_path",
				Title:       "Referrer URL /path",
				Description: "field: referrer_path",
			},
			"channel_origin_id": {
				Type:        "string",
				SQL:         "channel_origin_id",
				Title:       "Channel origin ID",
				Description: "field: channel_origin_id",
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
			"utm_id": {
				Type:        "string",
				SQL:         "utm_id",
				Title:       "UTM ID",
				Description: "field: utm_id",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"utm_id_from": {
				Type:        "string",
				SQL:         "utm_id_from",
				Title:       "UTM ID from",
				Description: "field: utm_id_from",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
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

			"via_utm_medium": {
				Type:        "string",
				SQL:         "via_utm_medium",
				Title:       "Via UTM medium",
				Description: "field: via_utm_medium",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"via_utm_source": {
				Type:        "string",
				SQL:         "via_utm_source",
				Title:       "Via UTM source",
				Description: "field: via_utm_source",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"via_utm_id": {
				Type:        "string",
				SQL:         "via_utm_id",
				Title:       "Via UTM ID",
				Description: "field: via_utm_id",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"via_utm_id_from": {
				Type:        "string",
				SQL:         "via_utm_id_from",
				Title:       "Via UTM ID from",
				Description: "field: via_utm_id_from",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"via_utm_campaign": {
				Type:        "string",
				SQL:         "via_utm_campaign",
				Title:       "Via UTM campaign",
				Description: "field: via_utm_campaign",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"via_utm_content": {
				Type:        "string",
				SQL:         "via_utm_content",
				Title:       "Via UTM content",
				Description: "field: via_utm_content",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"via_utm_term": {
				Type:        "string",
				SQL:         "via_utm_term",
				Title:       "Via UTM term",
				Description: "field: via_utm_term",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},

			"conversion_type": {
				Type:        "string",
				SQL:         "conversion_type",
				Title:       "Conversion type",
				Description: "field: conversion_type",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"conversion_external_id": {
				Type:        "string",
				SQL:         "conversion_external_id",
				Title:       "Conversion external ID",
				Description: "field: conversion_external_id",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"conversion_id": {
				Type:        "string",
				SQL:         "conversion_id",
				Title:       "Conversion ID",
				Description: "field: conversion_id",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"conversion_at": {
				Type:        "time",
				SQL:         "conversion_at",
				Title:       "Conversion at",
				Description: "field: conversion_at",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"conversion_amount": {
				Type:        "number",
				SQL:         "conversion_amount",
				Title:       "Conversion amount",
				Description: "field: conversion_amount",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"time_to_conversion": {
				Type:        "number",
				SQL:         "time_to_conversion",
				Title:       "Time to conversion",
				Description: "field: time_to_conversion",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"is_first_conversion": {
				Type:        "number",
				SQL:         "is_first_conversion",
				Title:       "Is first conversion",
				Description: "field: is_first_conversion",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
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
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
		},
	}
}
