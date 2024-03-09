package entity

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/rotisserie/eris"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var (
	// computed fields should be excluded from SELECT/INSERT while cloning rows
	DeviceComputedFields = []string{
		"created_at_trunc",
	}

	ErrParseUserAgentNetwork = eris.New("ParseUserAgent: network error")
)

type UserAgentResult struct {
	Ua      string `json:"ua"`
	Browser struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Major   string `json:"major"`
	} `json:"browser"`
	OS struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"os"`
	Device struct {
		Model  string `json:"model"`
		Type   string `json:"type"`
		Vendor string `json:"vendor"`
	} `json:"device"`
	// we dont care about the rest...
	// Engine struct {
	// 	Name    string `json:"name"`
	// 	Version string `json:"version"`
	// } `json:"engine"`
	// CPU struct {
	// 	Architecture string `json:"architecture"`
	// } `json:"cpu"`
}

type Device struct {
	ID               string          `db:"id" json:"id"`
	ExternalID       string          `db:"external_id" json:"external_id"`
	UserID           string          `db:"user_id" json:"user_id"`
	CreatedAt        time.Time       `db:"created_at" json:"created_at"`
	CreatedAtTrunc   time.Time       `db:"created_at_trunc" json:"created_at_trunc"` // date truncated at hour, used to optimize columnstore storage and queries
	DBCreatedAt      time.Time       `db:"db_created_at" json:"db_created_at"`
	DBUpdatedAt      time.Time       `db:"db_updated_at" json:"db_updated_at"`
	MergedFromUserID *string         `db:"merged_from_user_id" json:"merged_from_user_id,omitempty"`
	FieldsTimestamp  FieldsTimestamp `db:"fields_timestamp" json:"fields_timestamp"`

	UserAgent           *NullableString `db:"user_agent" json:"user_agent,omitempty"`           // user agent
	UserAgentHash       *string         `db:"user_agent_hash" json:"user_agent_hash,omitempty"` // Computed field
	Browser             *NullableString `db:"browser" json:"browser,omitempty"`
	BrowserVersion      *NullableString `db:"browser_version" json:"browser_version,omitempty"`
	BrowserVersionMajor *NullableString `db:"browser_version_major" json:"browser_version_major,omitempty"`
	OS                  *NullableString `db:"os" json:"os,omitempty"`
	OSVersion           *NullableString `db:"os_version" json:"os_version,omitempty"`
	DeviceType          *NullableString `db:"device_type" json:"device_type,omitempty"`
	Resolution          *NullableString `db:"resolution" json:"resolution,omitempty"` // resolution
	Language            *NullableString `db:"language" json:"language,omitempty"`     // language
	AdBlocker           *NullableBool   `db:"ad_blocker" json:"ad_blocker,omitempty"` // has ad block
	InWebview           *NullableBool   `db:"in_webview" json:"in_webview,omitempty"` // was loaded in a webview

	// Not persisted in DB:
	UpdatedAt    *time.Time    `db:"-" json:"-"` // used to merge fields and append item_timeline at the right time
	ExtraColumns AppItemFields `db:"-" json:"-"` // converted into "app_xxx" fields when marshaling JSON
}

// compute and set the user agent hash
func (device *Device) ComputeUserAgentHash() {
	if device.UserAgent == nil || device.UserAgent.IsNull {
		device.UserAgentHash = nil
		return
	}
	device.UserAgentHash = StringPtr(fmt.Sprintf("%x", sha1.Sum([]byte(strings.ToLower(device.UserAgent.String)))))
}

func (s *Device) ShouldParseUserAgent() bool {
	if s.UserAgent == nil || s.UserAgent.IsNull {
		return false
	}
	// if browser is already set, we don't need to parse the user agent
	if s.Browser != nil && !s.Browser.IsNull {
		return false
	}
	return true
}

func (s *Device) GetFieldDate(field string) time.Time {
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

// update a field timestamp to its most recent value
func (s *Device) UpdateFieldTimestamp(field string, timestamp *time.Time) {
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
func (s *Device) SetCreatedAt(value time.Time) {
	// update if current value is older
	if value.Before(s.CreatedAt) {
		s.CreatedAt = value
		s.CreatedAtTrunc = value.Truncate(time.Hour)
	}
}
func (s *Device) SetUserAgent(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "user_agent"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.UserAgent != nil && s.UserAgent.IsNull == value.IsNull && s.UserAgent.String == value.String {
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
		s.UserAgent = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.UserAgent),
		NewValue:  NullableStringToInterface(value),
	}
	s.UserAgent = value
	s.ComputeUserAgentHash()
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Device) SetBrowser(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "browser"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if s.Browser != nil && s.Browser.IsNull == value.IsNull && s.Browser.String == value.String {
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
		s.Browser = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.Browser),
		NewValue:  NullableStringToInterface(value),
	}
	s.Browser = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Device) SetBrowserVersion(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "browser_version"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.BrowserVersion != nil && s.BrowserVersion.IsNull == value.IsNull && s.BrowserVersion.String == value.String {
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
		s.BrowserVersion = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.BrowserVersion),
		NewValue:  NullableStringToInterface(value),
	}
	s.BrowserVersion = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Device) SetBrowserVersionMajor(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "browser_version_major"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.BrowserVersionMajor != nil && s.BrowserVersionMajor.IsNull == value.IsNull && s.BrowserVersionMajor.String == value.String {
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
		s.BrowserVersionMajor = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.BrowserVersionMajor),
		NewValue:  NullableStringToInterface(value),
	}
	s.BrowserVersionMajor = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Device) SetOS(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "os"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.OS != nil && s.OS.IsNull == value.IsNull && s.OS.String == value.String {
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
		s.OS = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.OS),
		NewValue:  NullableStringToInterface(value),
	}
	s.OS = value
	s.FieldsTimestamp[key] = timestamp
	return
}

func (s *Device) SetOSVersion(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "os_version"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.OSVersion != nil && s.OSVersion.IsNull == value.IsNull && s.OSVersion.String == value.String {
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
		s.OSVersion = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.OSVersion),
		NewValue:  NullableStringToInterface(value),
	}
	s.OSVersion = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Device) SetDeviceType(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "device_type"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.DeviceType != nil && s.DeviceType.IsNull == value.IsNull && s.DeviceType.String == value.String {
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
		s.DeviceType = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.DeviceType),
		NewValue:  NullableStringToInterface(value),
	}
	s.DeviceType = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Device) SetResolution(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "resolution"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.Resolution != nil && s.Resolution.IsNull == value.IsNull && s.Resolution.String == value.String {
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
		s.Resolution = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.Resolution),
		NewValue:  NullableStringToInterface(value),
	}
	s.Resolution = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Device) SetLanguage(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "language"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.Language != nil && s.Language.IsNull == value.IsNull && s.Language.String == value.String {
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
		s.Language = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.Language),
		NewValue:  NullableStringToInterface(value),
	}
	s.Language = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Device) SetAdBlocker(value *NullableBool, timestamp time.Time) (update *UpdatedField) {
	key := "ad_blocker"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.AdBlocker != nil && s.AdBlocker.IsNull == value.IsNull && s.AdBlocker.Bool == value.Bool {
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
		s.AdBlocker = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableBoolToInterface(s.AdBlocker),
		NewValue:  NullableBoolToInterface(value),
	}
	s.AdBlocker = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Device) SetInWebview(value *NullableBool, timestamp time.Time) (update *UpdatedField) {
	key := "in_webview"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.InWebview != nil && s.InWebview.IsNull == value.IsNull && s.InWebview.Bool == value.Bool {
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
		s.InWebview = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableBoolToInterface(s.InWebview),
		NewValue:  NullableBoolToInterface(value),
	}
	s.InWebview = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Device) SetExtraColumns(field string, value *AppItemField, timestamp time.Time) (update *UpdatedField) {

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

// overwrite json marshaller, to convert map of extra columns into "app_xxx" fields
func (s *Device) MarshalJSON() ([]byte, error) {

	type Alias Device

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
			return nil, eris.Errorf("set device custom dimension err: %v", err)
		}
	}

	return []byte(jsonValue), nil
}

func (device *Device) ComputeInWebview() {

	if device.UserAgent == nil || device.UserAgent.IsNull {
		return
	}

	userAgent := strings.ToLower(device.UserAgent.String)

	// if it says it's a webview, let's go with that
	if strings.Contains(userAgent, "webview") {
		device.InWebview = &NullableBool{Bool: true}
		return
	}

	// iOS webview will be the same as safari but missing "Safari"
	if (strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "ipod") || strings.Contains(userAgent, "ipad")) && !strings.Contains(userAgent, "safari") {
		device.InWebview = &NullableBool{Bool: true}
		return
	}

	if strings.Contains(userAgent, "android") && (strings.Contains(userAgent, "wv") || strings.Contains(userAgent, ".0.0.0")) {
		device.InWebview = &NullableBool{Bool: true}
		return
	}

	if strings.Contains(userAgent, "Linux; U; Android") {
		device.InWebview = &NullableBool{Bool: true}
		return
	}
}

func ComputeDeviceID(externalID string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(externalID)))
}

func NewDevice(externalID string, userID string, createdAt time.Time, updatedAt time.Time) *Device {
	return &Device{
		ID:              ComputeDeviceID(externalID),
		ExternalID:      externalID,
		UserID:          userID,
		CreatedAt:       createdAt,
		FieldsTimestamp: FieldsTimestamp{},
		ExtraColumns:    AppItemFields{},
		UpdatedAt:       &updatedAt,
	}
}

func NewDeviceFromDataLog(dataLog *DataLog, clockDifference time.Duration, workspace *Workspace) (device *Device, err error) {

	result := gjson.Get(dataLog.Item, "device")
	if !result.Exists() {
		return nil, eris.New("item has no device object")
	}

	if result.Type != gjson.JSON {
		return nil, eris.New("item device is not an object")
	}

	extraColumns := workspace.FindExtraColumnsForItemKind("device")

	// init
	device = &Device{
		UserID:          dataLog.UserID,
		FieldsTimestamp: FieldsTimestamp{},
		ExtraColumns:    AppItemFields{},
	}

	// loop over fields
	result.ForEach(func(key, value gjson.Result) bool {

		keyString := key.String()
		switch keyString {

		case "external_id":
			if value.Type == gjson.Null {
				err = eris.New("external_id is required")
				return false
			}

			device.ExternalID = value.String()
			device.ID = ComputeDeviceID(device.ExternalID)

		case "created_at":
			if device.CreatedAt, err = time.Parse(time.RFC3339, value.String()); err != nil {
				err = eris.Wrap(err, "device.created_at")
				return false
			}

			// apply clock difference
			if device.CreatedAt.After(time.Now()) {

				device.CreatedAt = device.CreatedAt.Add(clockDifference)
				if device.CreatedAt.After(time.Now()) {
					err = eris.New("device.created_at cannot be in the future")
					return false
				}
			}

			device.CreatedAtTrunc = device.CreatedAt.Truncate(time.Hour)

		case "updated_at":
			var updatedAt time.Time
			if updatedAt, err = time.Parse(time.RFC3339, value.String()); err != nil {
				err = eris.Wrap(err, "device.updated_at")
				return false
			}

			if updatedAt.After(time.Now()) {
				// apply clock difference
				updatedAtModified := updatedAt.Add(clockDifference)
				updatedAt = updatedAtModified
				if updatedAt.After(time.Now()) {
					err = eris.New("device.updated_at cannot be in the future")
					return false
				}
			}

			device.UpdatedAt = &updatedAt

		case "user_agent":
			if value.Type == gjson.Null {
				device.UserAgent = NewNullableString(nil)
			} else {
				device.UserAgent = NewNullableString(StringPtr(value.String()))
			}

		case "device_type":
			if value.Type == gjson.Null {
				device.DeviceType = NewNullableString(nil)
			} else {
				deviceType := value.String()

				if !govalidator.IsIn(deviceType, DeviceTypes...) {
					err = eris.Errorf("device_type %s is not valid", deviceType)
					return false
				}

				device.DeviceType = NewNullableString(&deviceType)
			}

		case "browser":
			if value.Type == gjson.Null {
				device.Browser = NewNullableString(nil)
			} else {
				device.Browser = NewNullableString(StringPtr(value.String()))
			}

		case "browser_version":
			if value.Type == gjson.Null {
				device.BrowserVersion = NewNullableString(nil)
			} else {
				device.BrowserVersion = NewNullableString(StringPtr(value.String()))
			}

		case "browser_version_major":
			if value.Type == gjson.Null {
				device.BrowserVersionMajor = NewNullableString(nil)
			} else {
				device.BrowserVersionMajor = NewNullableString(StringPtr(value.String()))
			}

		case "os":
			if value.Type == gjson.Null {
				device.OS = NewNullableString(nil)
			} else {
				device.OS = NewNullableString(StringPtr(value.String()))
			}

		case "os_version":
			if value.Type == gjson.Null {
				device.OSVersion = NewNullableString(nil)
			} else {
				device.OSVersion = NewNullableString(StringPtr(value.String()))
			}

		case "resolution":
			if value.Type == gjson.Null {
				device.Resolution = NewNullableString(nil)
			} else {
				device.Resolution = NewNullableString(StringPtr(value.String()))
			}

		case "language":
			if value.Type == gjson.Null {
				device.Language = NewNullableString(nil)
			} else {
				device.Language = NewNullableString(StringPtr(value.String()))
			}

		case "ad_blocker":
			if value.Type == gjson.Null {
				device.AdBlocker = NewNullableBool(nil)
			} else {
				device.AdBlocker = NewNullableBool(BoolPtr(value.Bool()))
			}

		case "in_webview":
			if value.Type == gjson.Null {
				device.InWebview = NewNullableBool(nil)
			} else {
				device.InWebview = NewNullableBool(BoolPtr(value.Bool()))
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
						device.ExtraColumns[col.Name] = fieldValue
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
	if device.UpdatedAt == nil {
		device.UpdatedAt = &device.CreatedAt
	}

	// Validation

	if device.ExternalID == "" {
		return nil, eris.New("device.external_id is required")
	}

	if device.CreatedAt.IsZero() {
		return nil, eris.New("device.created_at is required")
	}

	return device, nil
}

// takes a device context from a data import batch
// and returns a device object
// func NewDeviceFromImport(deviceCtx *DataImportDevice, userID string) *Device {

// 	if deviceCtx.UpdatedAt == nil {
// 		deviceCtx.UpdatedAt = &deviceCtx.CreatedAt
// 	}

// 	device := NewDevice(deviceCtx.ExternalID, userID, deviceCtx.CreatedAt, *deviceCtx.UpdatedAt)

// 	// convert values to nullable fields
// 	if deviceCtx.UserAgent != nil {
// 		device.UserAgent = NewNullableString(deviceCtx.UserAgent)
// 	}
// 	if deviceCtx.DeviceType != nil {
// 		device.DeviceType = NewNullableString(deviceCtx.DeviceType)
// 	}

// 	// Browser, OS, Platform, DeviceType are extracted from User Agent
// 	// however, a server-side data import can set them if they are already known
// 	device.Browser = deviceCtx.Browser
// 	device.BrowserVersion = deviceCtx.BrowserVersion
// 	device.BrowserVersionMajor = deviceCtx.BrowserVersionMajor
// 	device.OS = deviceCtx.OS
// 	device.OSVersion = deviceCtx.OSVersion
// 	device.Resolution = deviceCtx.Resolution
// 	device.Language = deviceCtx.Language
// 	device.AdBlocker = deviceCtx.AdBlocker

// 	// if deviceCtx.ExtraColumns != nil {
// 	// 	for key, value := range deviceCtx.ExtraColumns {
// 	// 		device.SetCustomColumns(key, value, deviceCtx.UpdatedAt)
// 	// 	}
// 	// }

// 	return device
// }

// parses the useragent and set updated fields
func (device *Device) ProcessUserAgent(result *UserAgentResult) {

	updatedAt := device.CreatedAt

	if device.UpdatedAt != nil {
		updatedAt = *device.UpdatedAt
	}

	if device.Browser == nil {

		if result.Browser.Name != "" {
			device.SetBrowser(&NullableString{IsNull: false, String: result.Browser.Name}, updatedAt)
		}

		if result.Browser.Version != "" {
			device.SetBrowserVersion(&NullableString{IsNull: false, String: result.Browser.Version}, updatedAt)
		}

		if result.Browser.Major != "" {
			device.SetBrowserVersionMajor(&NullableString{IsNull: false, String: result.Browser.Major}, updatedAt)
		}
	}

	if device.OS == nil {
		if result.OS.Name != "" {
			device.SetOS(&NullableString{IsNull: false, String: result.OS.Name}, updatedAt)
		}

		if result.OS.Version != "" {
			device.SetOSVersion(&NullableString{IsNull: false, String: result.OS.Version}, updatedAt)
		}
	}

	if device.DeviceType == nil {
		device.SetDeviceType(&NullableString{IsNull: false, String: result.Device.Type}, updatedAt)
	}

	device.ComputeInWebview()
	device.ComputeUserAgentHash()
}

// merges two devices and returns the list of updated fields
func (fromDevice *Device) MergeInto(toDevice *Device) (updatedFields []*UpdatedField) {
	updatedFields = []*UpdatedField{} // init

	if toDevice.FieldsTimestamp == nil {
		toDevice.FieldsTimestamp = FieldsTimestamp{}
	}

	if fieldUpdate := toDevice.SetUserAgent(fromDevice.UserAgent, fromDevice.GetFieldDate("user_agent")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toDevice.SetBrowser(fromDevice.Browser, fromDevice.GetFieldDate("browser")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toDevice.SetBrowserVersion(fromDevice.BrowserVersion, fromDevice.GetFieldDate("browser_version")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toDevice.SetBrowserVersionMajor(fromDevice.BrowserVersionMajor, fromDevice.GetFieldDate("browser_version_major")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toDevice.SetOS(fromDevice.OS, fromDevice.GetFieldDate("os")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toDevice.SetOSVersion(fromDevice.OSVersion, fromDevice.GetFieldDate("os_version")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toDevice.SetDeviceType(fromDevice.DeviceType, fromDevice.GetFieldDate("device_type")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toDevice.SetResolution(fromDevice.Resolution, fromDevice.GetFieldDate("resolution")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toDevice.SetLanguage(fromDevice.Language, fromDevice.GetFieldDate("language")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toDevice.SetAdBlocker(fromDevice.AdBlocker, fromDevice.GetFieldDate("ad_blocker")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toDevice.SetInWebview(fromDevice.InWebview, fromDevice.GetFieldDate("in_webview")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	for key, value := range fromDevice.ExtraColumns {
		if fieldUpdate := toDevice.SetExtraColumns(key, value, fromDevice.GetFieldDate(key)); fieldUpdate != nil {
			updatedFields = append(updatedFields, fieldUpdate)
		}
	}

	// UpdatedAt is the timeOfEvent for ITs
	toDevice.UpdatedAt = fromDevice.UpdatedAt

	// priority to oldest date
	toDevice.SetCreatedAt(fromDevice.CreatedAt)

	return
}

// timezone VARCHAR(64),
// timezone_offset SMALLINT,

var DeviceSchema string = `CREATE TABLE IF NOT EXISTS device (
  id VARCHAR(64) NOT NULL,
  external_id VARCHAR(255) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  created_at DATETIME NOT NULL,
  created_at_trunc AS DATE_TRUNC('hour', created_at) PERSISTED DATETIME,
  db_created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  merged_from_user_id VARCHAR(64),
  fields_timestamp JSON NOT NULL,

  user_agent TEXT,
  user_agent_hash VARCHAR(40),
  browser VARCHAR(20),
  browser_version VARCHAR(30),
  browser_version_major VARCHAR(30),
  os VARCHAR(20),
  os_version VARCHAR(20),
  device_type VARCHAR(20),
  resolution VARCHAR(20),
  language VARCHAR(10),
  ad_blocker BOOLEAN DEFAULT FALSE,
  in_webview BOOLEAN DEFAULT FALSE,

  SORT KEY (created_at_trunc DESC),
  PRIMARY KEY (id, user_id),
  SHARD KEY (user_id)
) COLLATE utf8mb4_general_ci;`

var DeviceSchemaMYSQL string = `CREATE TABLE IF NOT EXISTS device (
  id VARCHAR(64) NOT NULL,
  external_id VARCHAR(255) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  created_at DATETIME NOT NULL,
  created_at_trunc DATETIME GENERATED ALWAYS AS (DATE_FORMAT(created_at, '%Y-%m-%d %H:00:00')) STORED,
  db_created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  merged_from_user_id VARCHAR(64),
  fields_timestamp JSON NOT NULL,

  user_agent TEXT,
  user_agent_hash VARCHAR(40),
  browser VARCHAR(20),
  browser_version VARCHAR(30),
  browser_version_major VARCHAR(30),
  os VARCHAR(20),
  os_version VARCHAR(20),
  device_type VARCHAR(20),
  resolution VARCHAR(20),
  language VARCHAR(10),
  ad_blocker BOOLEAN DEFAULT FALSE,
  in_webview BOOLEAN DEFAULT FALSE,

  -- SORT KEY (created_at_trunc),
  PRIMARY KEY (id, user_id)
  -- SHARD KEY (user_id)
) COLLATE utf8mb4_general_ci;`

var DeviceTypes = []string{
	"desktop", // doesnt exist in the ua-parser-js lib, added when device is unknown
	"console",
	"mobile",
	"tablet",
	"smarttv",
	"wearable",
	"embedded",
}

func NewDeviceCube() *CubeJSSchema {
	return &CubeJSSchema{
		Title:       "Devices",
		Description: "Devices",
		SQL:         "SELECT * FROM `device`",
		Joins: map[string]CubeJSSchemaJoin{
			"Session": {
				Relationship: "one_to_many",
				SQL:          "${CUBE}.user_id = ${Session}.user_id AND ${CUBE}.id = ${Session}.device_id",
			},
		},
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
				Description: "count distinct user_id",
			},
		},
		Dimensions: map[string]CubeJSSchemaDimension{
			"id": {
				SQL:         "id",
				Type:        "string",
				PrimaryKey:  true,
				Title:       "Order ID",
				Description: "field: id",
			},

			"created_at": {
				SQL:         "created_at",
				Type:        "time",
				Title:       "Created at",
				Description: "field: created_at",
			},
			"created_at_trunc": {
				SQL:         "created_at_trunc",
				Type:        "time",
				Title:       "Created at (truncated to hour)",
				Description: "field: created_at_trunc",
			},
			"user_agent": {
				SQL:         "user_agent",
				Type:        "string",
				Title:       "User agent",
				Description: "field: user_agent",
			},
			"user_agent_hash": {
				SQL:         "user_agent_hash",
				Type:        "string",
				Title:       "User agent hash",
				Description: "field: user_agent_hash",
			},
			"browser": {
				SQL:         "browser",
				Type:        "string",
				Title:       "Browser",
				Description: "field: browser",
			},
			"browser_version": {
				SQL:         "browser_version",
				Type:        "string",
				Title:       "Browser version",
				Description: "field: browser_version",
			},
			"browser_version_major": {
				SQL:         "browser_version_major",
				Type:        "number",
				Title:       "Browser version major",
				Description: "field: browser_version_major",
			},
			"os": {
				SQL:         "os",
				Type:        "string",
				Title:       "OS",
				Description: "field: os",
			},
			"os_version": {
				SQL:         "os_version",
				Type:        "string",
				Title:       "OS version",
				Description: "field: os_version",
			},
			"device_type": {
				SQL:         "device_type",
				Type:        "string",
				Title:       "Device type",
				Description: "field: device_type",
			},
			"resolution": {
				SQL:         "resolution",
				Type:        "string",
				Title:       "Resolution",
				Description: "field: resolution",
			},
			"language": {
				SQL:         "language",
				Type:        "string",
				Title:       "Language",
				Description: "field: language",
			},
			"ad_blocker": {
				SQL:         "ad_blocker",
				Type:        "number",
				Title:       "Ad blocker",
				Description: "field: ad_blocker",
			},
			"in_webview": {
				SQL:         "in_webview",
				Type:        "number",
				Title:       "In webview",
				Description: "field: in_webview",
			},
		},
	}
}

// var BrowserNames = []string{
// 	"Unknown",
// 	"Chrome",
// 	"IE",
// 	"Safari",
// 	"Firefox",
// 	"Android",
// 	"Opera",
// 	"Blackberry",
// 	"UC",
// 	"Browser",
// 	"Silk",
// 	"Nokia",
// 	"NetFront",
// 	"QQ",
// 	"Maxthon",
// 	"SogouExplorer",
// 	"Spotify",
// 	"Nintendo",
// 	"Samsung",
// 	"Yandex",
// 	"CocCoc",
// 	"Bot",
// 	"AppleBot",
// 	"BaiduBot",
// 	"BingBot",
// 	"DuckDuckGoBot",
// 	"FacebookBot",
// 	"GoogleBot",
// 	"LinkedInBot",
// 	"MsnBot",
// 	"PingdomBot",
// 	"TwitterBot",
// 	"YandexBot",
// 	"CocCocBot",
// 	"YahooBot",
// }

// var OSNames = []string{
// 	"Unknown",
// 	"WindowsPhone",
// 	"Windows",
// 	"MacOSX",
// 	"iOS",
// 	"Android",
// 	"Blackberry",
// 	"ChromeOS",
// 	"Kindle",
// 	"WebOS",
// 	"Linux",
// 	"Playstation",
// 	"Xbox",
// 	"Nintendo",
// 	"Bot",
// }
