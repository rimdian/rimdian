package entity

import (
	"crypto/sha1"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/rimdian/rimdian/internal/api/common"
	"github.com/rimdian/rimdian/internal/common/dto"
	"github.com/rotisserie/eris"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var (
	ErrOrderRequired = eris.New("order is required")

	// computed fields should be excluded from SELECT/INSERT while cloning rows
	OrderComputedFields []string = []string{
		"converted_subtotal_price",
		"converted_total_price",
		"created_at_trunc",
	}
)

type ConversionFunnel []*ConversionTouchpoint

func (x *ConversionFunnel) Scan(val interface{}) error {
	return json.Unmarshal(val.([]byte), &x)
}

func (x ConversionFunnel) Value() (driver.Value, error) {
	return json.Marshal(x)
}

type ConversionTouchpoint struct {
	ChannelOriginID string `json:"channel_origin_id"`
	Postview        bool   `json:"postview"` // is ad impression postview
	Position        int    `json:"position"`
	Role            int    `json:"role"`
	Count           int    `json:"count"` // when many session with same origin_id are side by side
	ChannelID       string `json:"channel_id"`
	ChannelGroupID  string `json:"channel_group_id"`
}

type DiscountCodes []string

func (x *DiscountCodes) Scan(val interface{}) error {
	return json.Unmarshal(val.([]byte), &x)
}

func (x DiscountCodes) Value() (driver.Value, error) {
	return json.Marshal(x)
}

func (x DiscountCodes) ToInterface() interface{} {
	v, _ := json.Marshal(x)
	return v
}

type Order struct {
	ID               string          `db:"id" json:"id"`
	ExternalID       string          `db:"external_id" json:"external_id"`
	UserID           string          `db:"user_id" json:"user_id"`
	DomainID         string          `db:"domain_id" json:"domain_id"`
	SessionID        *string         `db:"session_id" json:"session_id,omitempty"` // historical sessions might not have session recorded
	CreatedAt        time.Time       `db:"created_at" json:"created_at"`
	CreatedAtTrunc   time.Time       `db:"created_at_trunc" json:"created_at_trunc"` // date truncated at hour, used to optimize columnstore storage and queries
	DBCreatedAt      time.Time       `db:"db_created_at" json:"db_created_at"`
	DBUpdatedAt      time.Time       `db:"db_updated_at" json:"db_updated_at"`
	IsDeleted        bool            `db:"is_deleted" json:"is_deleted,omitempty"` // deleting rows in transactions cause deadlocks in singlestore, we use an update
	MergedFromUserID *string         `db:"merged_from_user_id" json:"merged_from_user_id,omitempty"`
	FieldsTimestamp  FieldsTimestamp `db:"fields_timestamp" json:"fields_timestamp"`

	DiscountCodes          *DiscountCodes `db:"discount_codes" json:"discount_codes"`                               // used for channels mapping (=voucher attribution)
	SubtotalPrice          *NullableInt64 `db:"subtotal_price" json:"subtotal_price"`                               // current price in cents without shipping and taxes
	TotalPrice             *NullableInt64 `db:"total_price" json:"total_price"`                                     // current price in cents with shipping and taxes
	Currency               *string        `db:"currency" json:"currency"`                                           // currency of the order, otherwise use workspace currency as default
	FxRate                 float64        `db:"fx_rate" json:"fx_rate"`                                             // fx rate used to convert the order to the workspace currency
	ConvertedSubtotalPrice *NullableInt64 `db:"converted_subtotal_price" json:"converted_subtotal_price,omitempty"` // in local currency in cents
	ConvertedTotalPrice    *NullableInt64 `db:"converted_total_price" json:"converted_total_price,omitempty"`       // in local currency in cents
	// ConversionRateError *string         `db:"conversion_rate_error" json:"conversion_rate_error,omitempty"` // eventual error while converting local price into workspace price
	CancelledAt  *NullableTime   `db:"cancelled_at" json:"cancelled_at,omitempty"`
	CancelReason *NullableString `db:"cancel_reason" json:"cancel_reason,omitempty"`
	IP           *NullableString `db:"ip" json:"ip,omitempty"`

	// attribution fields
	IsFirstConversion    bool             `db:"is_first_conversion" json:"is_first_conversion"`         // used to separate acquisition stats from retention stats
	TimeToConversion     *int64           `db:"time_to_conversion" json:"time_to_conversion,omitempty"` // in seconds
	DevicesFunnel        *string          `db:"devices_funnel" json:"devices_funnel"`                   // used to do cross-device reports
	DevicesTypeCount     int64            `db:"devices_type_count" json:"devices_type_count"`           // used to compute an average of cross-device touchpoints
	DomainsFunnel        *string          `db:"domains_funnel" json:"domains_funnel"`                   // "~" separated domain ids, acts as a hash: website~ios~amazon used to do cross-domain reports (web-to-store...)
	DomainsTypeFunnel    *string          `db:"domains_type_funnel" json:"domains_type_funnel"`         // "~" separated domain kinds, acts as a hash: web~app~marketplace. used to do cross-domain-kind reports (web-to-store...)
	DomainsCount         int64            `db:"domains_count" json:"domains_count"`                     // used to compute an average of cross-domain visits
	Funnel               ConversionFunnel `db:"funnel" json:"funnel"`                                   // summary of conversion path touchpoints
	FunnelHash           *string          `db:"funnel_hash" json:"funnel_hash"`                         // hash of funnel to GROUP BY for conversion paths reports
	AttributionUpdatedAt *time.Time       `db:"attribution_updated_at" json:"attribution_updated_at,omitempty"`

	// Not persisted in DB:
	Items        OrderItems    `db:"-" json:"items"`
	ExtraColumns AppItemFields `db:"-" json:"extra_columns"` // converted into "app_xxx" fields when marshaling JSON
	UpdatedAt    *time.Time    `db:"-" json:"-"`             // used to merge fields and append item_timeline at the right time
}

func (o *Order) GetFieldDate(field string) time.Time {
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
func (o *Order) UpdateFieldTimestamp(field string, timestamp *time.Time) {
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

func (o *Order) SetSessionID(value *string, timestamp time.Time) (update *UpdatedField) {
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
func (s *Order) SetCreatedAt(value time.Time) {
	// update if current value is older
	if value.Before(s.CreatedAt) {
		s.CreatedAt = value
		s.CreatedAtTrunc = value.Truncate(time.Hour)
	}
}
func (o *Order) SetCurrency(value *string, timestamp time.Time) (update *UpdatedField) {
	key := "currency"
	// value cant be null
	if value == nil {
		return nil
	}
	// abort if values are equal
	if FixedDeepEqual(o.Currency, value) {
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
		o.Currency = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: StringPointerToInterface(o.Currency),
		NewValue:  StringPointerToInterface(value),
	}
	o.Currency = value
	o.FieldsTimestamp[key] = timestamp
	return
}

func (o *Order) SetDiscountCodes(value *DiscountCodes, timestamp time.Time) (update *UpdatedField) {
	key := "discount_codes"
	// abort if values are equal
	if value == nil && o.DiscountCodes == nil || FixedDeepEqual(o.DiscountCodes, value) {
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
		o.DiscountCodes = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: o.DiscountCodes.ToInterface(),
		NewValue:  value.ToInterface(),
	}
	o.FieldsTimestamp[key] = timestamp
	o.DiscountCodes = value
	return
}

func (o *Order) SetFxRate(value float64, timestamp time.Time) (update *UpdatedField) {
	key := "fx_rate"
	// abort if values are equal
	if o.FxRate == value {
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
		o.FxRate = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: o.FxRate,
		NewValue:  value,
	}
	o.FxRate = value
	o.FieldsTimestamp[key] = timestamp
	return
}
func (o *Order) SetSubtotalPrice(value *NullableInt64, timestamp time.Time) (update *UpdatedField) {
	key := "subtotal_price"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && o.SubtotalPrice != nil && o.SubtotalPrice.IsNull == value.IsNull && o.SubtotalPrice.Int64 == value.Int64 {
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
		o.SubtotalPrice = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableInt64ToInterface(o.SubtotalPrice),
		NewValue:  NullableInt64ToInterface(value),
	}
	o.SubtotalPrice = value
	o.FieldsTimestamp[key] = timestamp
	return
}
func (o *Order) SetTotalPrice(value *NullableInt64, timestamp time.Time) (update *UpdatedField) {
	key := "total_price"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && o.TotalPrice != nil && o.TotalPrice.IsNull == value.IsNull && o.TotalPrice.Int64 == value.Int64 {
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
		o.TotalPrice = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableInt64ToInterface(o.TotalPrice),
		NewValue:  NullableInt64ToInterface(value),
	}
	o.TotalPrice = value
	o.FieldsTimestamp[key] = timestamp
	return
}
func (o *Order) SetCancelledAt(value *NullableTime, timestamp time.Time) (update *UpdatedField) {
	key := "cancelled_at"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && o.CancelledAt != nil && o.CancelledAt.IsNull == value.IsNull && o.CancelledAt.Time.Equal(value.Time) {
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
		o.CancelledAt = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableTimeToInterface(o.CancelledAt),
		NewValue:  NullableTimeToInterface(value),
	}
	o.CancelledAt = value
	o.FieldsTimestamp[key] = timestamp
	return
}
func (o *Order) SetCancelReason(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "cancel_reason"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && o.CancelReason != nil && o.CancelReason.IsNull == value.IsNull && o.CancelReason.String == value.String {
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
		o.CancelReason = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(o.CancelReason),
		NewValue:  NullableStringToInterface(value),
	}
	o.CancelReason = value
	o.FieldsTimestamp[key] = timestamp
	return
}

func (o *Order) SetIP(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "ip"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && o.IP != nil && o.IP.IsNull == value.IsNull && o.IP.String == value.String {
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
		o.IP = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(o.IP),
		NewValue:  NullableStringToInterface(value),
	}
	o.IP = value
	o.FieldsTimestamp[key] = timestamp
	return
}

// func (o *Order) SetItems(value OrderItems, timestamp time.Time) (update *UpdatedField) {
// 	key := "items"
// 	// abort if values are equal
// 	if value.Equals(o.Items) {
// 		return nil
// 	}
// 	existingValueTimestamp := o.GetFieldDate(key)
// 	// abort if existing value is newer
// 	if existingValueTimestamp.After(timestamp) {
// 		return nil
// 	}
// 	// the value might be set for the first time
// 	// so we set the value without producing a field update
// 	if existingValueTimestamp.Equal(timestamp) {
// 		o.Items = value
// 		return
// 	}
// 	update = &UpdatedField{
// 		Field:     key,
// 		PrevValue: o.Items.ToInterface(),
// 		NewValue:  value.ToInterface(),
// 	}
// 	o.Items = value
// 	o.FieldsTimestamp[key] = timestamp
// 	return
// }

// attribution fields are not "watched" for changes as they are not controlled by API data import
func (o *Order) SetIsFirstConversion(value bool) {
	o.IsFirstConversion = value
}
func (o *Order) SetTimeToConversion(value *int64) {
	o.TimeToConversion = value
}
func (o *Order) SetDevicesFunnel(value string) {
	o.DevicesFunnel = &value
}
func (o *Order) SetDevicesTypeCount(value int64) {
	o.DevicesTypeCount = value
}
func (o *Order) SetDomainsFunnel(value string) {
	o.DomainsFunnel = &value
}
func (o *Order) SetDomainsTypeFunnel(value string) {
	o.DomainsTypeFunnel = &value
}
func (o *Order) SetDomainsCount(value int64) {
	o.DomainsCount = value
}
func (o *Order) SetFunnel(value ConversionFunnel) {
	o.Funnel = value
}
func (o *Order) SetFunnelHash(value string) {
	o.FunnelHash = &value
}
func (o *Order) SetAttributionUpdatedAt(value *time.Time) {
	o.AttributionUpdatedAt = value
}

func (s *Order) SetExtraColumns(field string, value *AppItemField, timestamp time.Time) (update *UpdatedField) {

	if s.ExtraColumns == nil {
		s.ExtraColumns = AppItemFields{}
	}

	// abort if field doesnt start with "app_"
	if !strings.HasPrefix(field, "app_") && !strings.HasPrefix(field, "appx_") {
		log.Printf("order field %s doesnt start with app_", field)
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

// merges two orders and returns the list of updated fields
func (fromOrder *Order) MergeInto(toOrder *Order) (updatedFields []*UpdatedField) {
	updatedFields = []*UpdatedField{} // init

	if toOrder.FieldsTimestamp == nil {
		toOrder.FieldsTimestamp = FieldsTimestamp{}
	}

	if fieldUpdate := toOrder.SetSessionID(fromOrder.SessionID, fromOrder.GetFieldDate("session_id")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toOrder.SetCurrency(fromOrder.Currency, fromOrder.GetFieldDate("currency")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	// if fieldUpdate := toOrder.SetItems(fromOrder.Items, fromOrder.GetFieldDate("items")); fieldUpdate != nil {
	// 	updatedFields = append(updatedFields, fieldUpdate)
	// }
	if fieldUpdate := toOrder.SetDiscountCodes(fromOrder.DiscountCodes, fromOrder.GetFieldDate("discount_codes")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toOrder.SetSubtotalPrice(fromOrder.SubtotalPrice, fromOrder.GetFieldDate("subtotal_price")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toOrder.SetTotalPrice(fromOrder.TotalPrice, fromOrder.GetFieldDate("total_price")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toOrder.SetCancelledAt(fromOrder.CancelledAt, fromOrder.GetFieldDate("cancelled_at")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toOrder.SetCancelReason(fromOrder.CancelReason, fromOrder.GetFieldDate("cancel_reason")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toOrder.SetIP(fromOrder.IP, fromOrder.GetFieldDate("ip")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	for key, value := range fromOrder.ExtraColumns {
		if fieldUpdate := toOrder.SetExtraColumns(key, value, fromOrder.GetFieldDate(key)); fieldUpdate != nil {
			updatedFields = append(updatedFields, fieldUpdate)
		}
	}

	// UpdatedAt is the timeOfEvent for ITs
	toOrder.UpdatedAt = fromOrder.UpdatedAt
	// priority to oldest date
	toOrder.SetCreatedAt(fromOrder.CreatedAt)

	return
}

// overwrite json marshaller, to convert map of extra columns into "app_xxx" fields
func (s *Order) MarshalJSON() ([]byte, error) {

	type Alias Order

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
			return nil, eris.Errorf("set order custom dimension err: %v", err)
		}
	}

	return []byte(jsonValue), nil
}

func NewOrder(externalID string, userID string, domainID string, createdAt time.Time, updatedAt *time.Time, items []*OrderItem) *Order {

	// default empty items
	if items == nil {
		items = []*OrderItem{}
	}

	return &Order{
		ID:              ComputeOrderID(externalID),
		ExternalID:      externalID,
		UserID:          userID,
		DomainID:        domainID,
		CreatedAt:       createdAt,
		CreatedAtTrunc:  createdAt.Truncate(time.Hour),
		FieldsTimestamp: FieldsTimestamp{},

		Items: items,

		UpdatedAt:    updatedAt,
		ExtraColumns: AppItemFields{},
	}
}

func NewOrderFromDataLog(dataLog *DataLog, clockDifference time.Duration, workspace *Workspace) (order *Order, err error) {

	result := gjson.Get(dataLog.Item, "order")
	if !result.Exists() {
		return nil, eris.New("item has no order object")
	}

	if result.Type != gjson.JSON {
		return nil, eris.New("item order is not an object")
	}

	extraColumns := workspace.FindExtraColumnsForItemKind("order")

	// init
	order = &Order{
		UserID:          dataLog.UserID,
		FieldsTimestamp: FieldsTimestamp{},
		Items:           []*OrderItem{},
		ExtraColumns:    AppItemFields{},
	}

	// loop over order fields
	result.ForEach(func(key, value gjson.Result) bool {

		keyString := key.String()
		switch keyString {

		case "external_id":
			order.ExternalID = value.String()
			order.ID = ComputeCartID(order.ExternalID)

		case "domain_id":
			order.DomainID = value.String()

		case "session_external_id":
			if value.Type == gjson.Null {
				order.SessionID = nil
			} else {
				order.SessionID = StringPtr(ComputeSessionID(value.String()))
			}

		case "created_at":
			if order.CreatedAt, err = time.Parse(time.RFC3339, value.String()); err != nil {
				err = eris.Wrap(err, "order.created_at")
				return false
			}

			// apply clock difference
			if order.CreatedAt.After(time.Now()) {

				order.CreatedAt = order.CreatedAt.Add(clockDifference)
				if order.CreatedAt.After(time.Now()) {
					err = eris.New("order.created_at cannot be in the future")
					return false
				}
			}

			order.CreatedAtTrunc = order.CreatedAt.Truncate(time.Hour)

		case "updated_at":
			var updatedAt time.Time
			if updatedAt, err = time.Parse(time.RFC3339, value.String()); err != nil {
				err = eris.Wrap(err, "order.updated_at")
				return false
			}

			if updatedAt.After(time.Now()) {
				// apply clock difference
				updatedAtModified := updatedAt.Add(clockDifference)
				updatedAt = updatedAtModified
				if updatedAt.After(time.Now()) {
					err = eris.New("order.updated_at cannot be in the future")
					return false
				}
			}

			order.UpdatedAt = &updatedAt

		case "currency":
			if value.Type == gjson.Null {
				err = eris.New("order.currency is required")
				return false
			}
			currency := value.String()
			order.Currency = &currency

		case "discount_codes":
			if value.Type == gjson.Null {
				order.DiscountCodes = &DiscountCodes{}
			} else {
				discountCodes := DiscountCodes{}
				// loop over discount codes
				value.ForEach(func(keyCode, discountCode gjson.Result) bool {
					discountCodes = append(discountCodes, discountCode.String())
					return true
				})
				order.DiscountCodes = &discountCodes
			}

		case "subtotal_price":
			if value.Type == gjson.Null {
				order.SubtotalPrice = NewNullableInt64(nil)
			} else {
				subtotalPrice := value.Int()
				if subtotalPrice < 0 {
					err = eris.New("order.subtotal_price cannot be negative")
					return false
				}
				order.SubtotalPrice = NewNullableInt64(&subtotalPrice)
			}

		case "total_price":
			if value.Type == gjson.Null {
				order.TotalPrice = NewNullableInt64(nil)
			} else {
				totalPrice := value.Int()
				if totalPrice < 0 {
					err = eris.New("order.subtotal_price cannot be negative")
					return false
				}
				order.TotalPrice = NewNullableInt64(&totalPrice)
			}

		case "cancelled_at":
			if value.Type == gjson.Null {
				order.CancelledAt = NewNullableTime(nil)
			} else {
				cancelledAt, errParse := time.Parse(time.RFC3339, value.String())
				if errParse != nil {
					err = eris.Wrap(errParse, "order.cancelled_at")
					return false
				}
				order.CancelledAt = NewNullableTime(&cancelledAt)
			}

		case "cancel_reason":
			if value.Type == gjson.Null {
				order.CancelReason = NewNullableString(nil)
			} else {
				order.CancelReason = NewNullableString(StringPtr(value.String()))
			}

		case "ip":
			if value.Type == gjson.Null {
				order.IP = NewNullableString(nil)
			} else {
				order.IP = NewNullableString(StringPtr(value.String()))
			}

		case "items":

			if value.Type != gjson.JSON {
				err = eris.New("order.items is not an array")
				return false
			}

			// loop over items
			value.ForEach(func(key, item gjson.Result) bool {

				if value.Type != gjson.JSON {
					err = eris.New("order.items is not an array of objects")
					return false
				}

				// init
				orderItem := &OrderItem{
					FieldsTimestamp: FieldsTimestamp{},
				}

				// loop over item fields
				item.ForEach(func(itemKey, itemField gjson.Result) bool {

					itemKeyString := itemKey.String()

					switch itemKeyString {

					case "external_id":
						orderItem.ExternalID = strings.TrimSpace(itemField.String())
						orderItem.ID = ComputeOrderID(orderItem.ExternalID)

					case "product_external_id":
						orderItem.ProductExternalID = strings.TrimSpace(itemField.String())

					case "name":
						orderItem.Name = strings.TrimSpace(itemField.String())

					case "sku":
						if itemField.Type == gjson.Null {
							orderItem.SKU = NewNullableString(nil)
						} else {
							orderItem.SKU = NewNullableString(StringPtr(strings.TrimSpace(itemField.String())))
						}

					case "brand":
						if itemField.Type == gjson.Null {
							orderItem.Brand = NewNullableString(nil)
						} else {
							orderItem.Brand = NewNullableString(StringPtr(strings.TrimSpace(itemField.String())))
						}

					case "category":
						if itemField.Type == gjson.Null {
							orderItem.Category = NewNullableString(nil)
						} else {
							orderItem.Category = NewNullableString(StringPtr(strings.TrimSpace(itemField.String())))
						}

					case "variant_external_id":
						if itemField.Type == gjson.Null {
							orderItem.VariantExternalID = NewNullableString(nil)
						} else {
							orderItem.VariantExternalID = NewNullableString(StringPtr(strings.TrimSpace(itemField.String())))
						}

					case "variant_title":
						if itemField.Type == gjson.Null {
							orderItem.VariantTitle = NewNullableString(nil)
						} else {
							orderItem.VariantTitle = NewNullableString(StringPtr(strings.TrimSpace(itemField.String())))
						}

					case "image_url":
						if itemField.Type == gjson.Null {
							orderItem.ImageURL = NewNullableString(nil)
						} else {
							orderItem.ImageURL = NewNullableString(StringPtr(strings.TrimSpace(itemField.String())))
						}

					case "price":
						if itemField.Type != gjson.Number {
							err = eris.New("order.items.price is not an integer")
							return false
						}

						price := itemField.Int()

						if price < 0 {
							err = eris.New("order.items.price cannot be negative")
							return false
						}

						orderItem.Price = price

					case "quantity":
						if itemField.Type != gjson.Number {
							err = eris.New("order.items.quantity is not an integer")
							return false
						}

						quantity := itemField.Int()

						if itemField.Int() < 0 {
							err = eris.New("order.items.quantity cannot be negative")
							return false
						}

						orderItem.Quantity = quantity

					default:
						// TODO: handle app_ extra columns
					}

					return true
				})

				if err != nil {
					return false
				}

				// validate item
				if orderItem.ExternalID == "" {
					err = eris.New("order.external_id is required")
					return false
				}
				if orderItem.Name == "" {
					err = eris.New("order.name is required")
					return false
				}
				if orderItem.ProductExternalID == "" {
					err = eris.New("order.product_external_id is required")
					return false
				}

				order.Items = append(order.Items, orderItem)

				return true
			})

			if err != nil {
				return false
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
						order.ExtraColumns[col.Name] = fieldValue
					}
				}
			}
		}

		return true
	})

	if err != nil {
		return nil, err
	}
	if order.DomainID == "" && dataLog.DomainID != nil {
		order.DomainID = *dataLog.DomainID
	}

	if order.Currency == nil || *order.Currency == "" {
		order.Currency = &workspace.Currency
	}

	// set fx_rate if needed
	if *order.Currency != workspace.Currency {
		if fxRate, err := workspace.GetFxRateForCurrency(*order.Currency); err == nil {
			order.FxRate = fxRate
		} else {
			return nil, err
		}
	}

	// default rate is 1.0 if currency is the same as workspace
	if order.FxRate == 0 {
		noRate := 1.0
		order.FxRate = noRate
	}

	// use data import createdAt as updatedAt if not provided
	if order.UpdatedAt == nil {
		order.UpdatedAt = &order.CreatedAt
	}

	// enrich order with session and domain
	if dataLog.UpsertedSession != nil {
		if order.SessionID == nil {
			order.SessionID = &dataLog.UpsertedSession.ID
		}
		if order.DomainID == "" {
			order.DomainID = dataLog.UpsertedSession.DomainID
		}
	}

	// set ip field for Client origins
	if dataLog.Origin == dto.DataLogOriginClient && dataLog.Context.IP != "" {
		order.IP = NewNullableString(StringPtr(dataLog.Context.IP))
	}

	// set order_items fields
	for _, item := range order.Items {
		item.OrderID = order.ID
		item.UserID = order.UserID
		item.CreatedAt = order.CreatedAt
		item.CreatedAtTrunc = order.CreatedAtTrunc
		item.UpdatedAt = order.UpdatedAt
		item.Currency = order.Currency
		item.FxRate = order.FxRate
	}

	// Validation
	if order.ExternalID == "" {
		return nil, eris.New("order.external_id is required")
	}

	// verify that domainID exists
	found := false
	for _, domain := range workspace.Domains {
		if domain.ID == order.DomainID {
			found = true
			break
		}
	}

	if !found {
		return nil, eris.New("order domain_id invalid")
	}

	if order.CreatedAt.IsZero() {
		return nil, eris.New("order.created_at is required")
	}

	if !govalidator.IsIn(*order.Currency, common.CurrenciesCodes...) {
		return nil, eris.New("order.currency invalid")
	}

	// reject if user is not authenticated
	if !dataLog.UpsertedUser.IsAuthenticated {
		return nil, eris.New("user.is_authenticated should be true")
	}

	// default JSON empty array is required
	if order.Funnel == nil {
		order.SetFunnel(ConversionFunnel{})
	}

	return order, nil
}

func ComputeOrderID(externalID string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(externalID)))
}

var OrderSchema string = `CREATE TABLE IF NOT EXISTS ` + "`order`" + ` (
	id VARCHAR(64) NOT NULL,
	external_id VARCHAR(255) NOT NULL,
	user_id VARCHAR(64) NOT NULL,
	domain_id VARCHAR(64) NOT NULL,
	session_id VARCHAR(64),
	created_at DATETIME NOT NULL,
	created_at_trunc AS DATE_TRUNC('hour', created_at) PERSISTED DATETIME,
	db_created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	is_deleted BOOLEAN DEFAULT FALSE,
	merged_from_user_id VARCHAR(64),
	fields_timestamp JSON NOT NULL,

	-- public_url VARCHAR(2083),
	discount_codes JSON,
	subtotal_price INT DEFAULT 0,
	total_price INT DEFAULT 0,
	currency VARCHAR(3) NOT NULL,
	fx_rate FLOAT DEFAULT 1 NOT NULL,
  	converted_subtotal_price AS FLOOR(subtotal_price*fx_rate) PERSISTED INT,
  	converted_total_price AS FLOOR(total_price*fx_rate) PERSISTED INT,
	cancelled_at DATETIME,
	cancel_reason VARCHAR(2083),
	ip VARCHAR(255),

	is_first_conversion BOOLEAN DEFAULT FALSE,
	time_to_conversion INT DEFAULT 0,
	devices_funnel VARCHAR(512),
	devices_type_count INT DEFAULT 1,
	domains_funnel VARCHAR(512),
	domains_type_funnel VARCHAR(512),
	domains_count INT DEFAULT 1,
	funnel JSON NOT NULL,
	funnel_hash VARCHAR(40),
	attribution_updated_at DATETIME,

	SORT KEY (created_at_trunc DESC, created_at DESC),
	PRIMARY KEY (id, user_id),
	KEY (external_id) USING HASH,
	KEY (session_id) USING HASH,
	KEY (user_id) USING HASH, -- for merging
	KEY (attribution_updated_at),
	SHARD KEY (user_id)
) COLLATE utf8mb4_general_ci;`

var OrderSchemaMYSQL string = `CREATE TABLE IF NOT EXISTS ` + "`order`" + ` (
	id VARCHAR(64) NOT NULL,
	external_id VARCHAR(255) NOT NULL,
	user_id VARCHAR(64) NOT NULL,
	domain_id VARCHAR(64) NOT NULL,
	session_id VARCHAR(64),
	created_at DATETIME NOT NULL,
  	created_at_trunc DATETIME GENERATED ALWAYS AS (DATE_FORMAT(created_at, '%Y-%m-%d %H:00:00')) STORED,
	db_created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	is_deleted BOOLEAN DEFAULT FALSE,
	merged_from_user_id VARCHAR(64),
	fields_timestamp JSON NOT NULL,

	-- public_url VARCHAR(2083),
	discount_codes JSON,
	subtotal_price INT DEFAULT 0,
	total_price INT DEFAULT 0,
	currency VARCHAR(3) NOT NULL,
	fx_rate FLOAT DEFAULT 1 NOT NULL,
  	converted_subtotal_price INT GENERATED ALWAYS AS (FLOOR(subtotal_price*fx_rate)) STORED,
  	converted_total_price INT GENERATED ALWAYS AS (FLOOR(total_price*fx_rate)) STORED,
	cancelled_at DATETIME,
	cancel_reason VARCHAR(2083),
	ip VARCHAR(255),

	is_first_conversion BOOLEAN DEFAULT FALSE,
	time_to_conversion INT DEFAULT 0,
	devices_funnel VARCHAR(512),
	devices_type_count INT DEFAULT 1,
	domains_funnel VARCHAR(512),
	domains_type_funnel VARCHAR(512),
	domains_count INT DEFAULT 1,
	funnel JSON,
	funnel_hash VARCHAR(40),
	attribution_updated_at DATETIME,

	-- SORT KEY (created_at_trunc),
	PRIMARY KEY (id, user_id),
	KEY (external_id) USING HASH,
	KEY (session_id) USING HASH,
	KEY (user_id) USING HASH, -- for merging
	KEY (attribution_updated_at)
	-- SHARD KEY (user_id)
) COLLATE utf8mb4_general_ci;`

func NewOrderCube() *CubeJSSchema {
	return &CubeJSSchema{
		Title:       "Orders",
		Description: "Orders",
		SQL:         "SELECT * FROM `order`",
		// https://cube.dev/docs/schema/reference/joins
		Joins: map[string]CubeJSSchemaJoin{
			"Session": {
				Relationship: "one_to_many",
				SQL:          "${CUBE}.user_id = ${Session}.user_id AND ${CUBE}.id = ${Session}.conversion_id",
			},
			"Postview": {
				Relationship: "one_to_many",
				SQL:          "${CUBE}.user_id = ${Postview}.user_id AND ${CUBE}.id = ${Postview}.conversion_id",
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
				Title:       "Orders",
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
			"orders_per_user": {
				Type:        "number",
				SQL:         "${count} / ${unique_users}",
				Title:       "Orders per user",
				Description: "count / unique_users",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"subtotal_per_user": {
				Type:        "number",
				SQL:         "${subtotal_price} / ${unique_users}",
				Title:       "Subtotal per user",
				Description: "subtotal_price / unique_users",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"subtotal_sum": {
				Type:        "number",
				SQL:         "COALESCE(SUM(subtotal_price), 0)",
				Title:       "Subtotal sum",
				Description: "SUM(subtotal_price)",
			},
			"acquisition_subtotal_sum": {
				Type:        "number",
				Title:       "Acquisition Subtotal sum",
				Description: "SUM(subtotal_price)",
				SQL:         "COALESCE(SUM(subtotal_price), 0)",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "is_first_conversion IS TRUE"},
				},
			},
			"retention_subtotal_sum": {
				Type:        "number",
				Title:       "Retention Subtotal sum",
				Description: "SUM(subtotal_price) WHERE is_first_conversion IS FALSE",
				SQL:         "COALESCE(SUM(subtotal_price), 0)",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "is_first_conversion IS FALSE"},
				},
			},
			"avg_cart": {
				Type:        "number",
				Title:       "Average cart",
				Description: "AVG(subtotal_price)",
				SQL:         "COALESCE(AVG(subtotal_price), 0)",
			},
			"acquisition_avg_cart": {
				Type:        "number",
				Title:       "Acquisition average cart",
				Description: "AVG(subtotal_price) WHERE is_first_conversion IS TRUE",
				SQL:         "COALESCE(AVG(subtotal_price), 0)",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "is_first_conversion IS TRUE"},
				},
			},
			"retention_avg_cart": {
				Type:        "number",
				Title:       "Retention average cart",
				Description: "AVG(subtotal_price) WHERE is_first_conversion IS FALSE",
				SQL:         "COALESCE(AVG(subtotal_price), 0)",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "is_first_conversion IS FALSE"},
				},
			},
			"avg_ttc": {
				Type:        "number",
				Title:       "Average time to conversion",
				Description: "AVG(time_to_conversion) WHERE time_to_conversion > 0",
				SQL:         "COALESCE(AVG(time_to_conversion), 0)",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "time_to_conversion > 0"},
				},
			},
			"acquisition_avg_ttc": {
				Type:        "number",
				Title:       "Acquisition Average time to conversion",
				Description: "AVG(time_to_conversion) WHERE time_to_conversion > 0 AND is_first_conversion IS TRUE",
				SQL:         "COALESCE(AVG(time_to_conversion), 0)",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "time_to_conversion > 0 AND is_first_conversion IS TRUE"},
				},
			},
			"retention_avg_ttc": {
				Type:        "number",
				Title:       "Retention Average time to conversion",
				Description: "AVG(time_to_conversion) WHERE time_to_conversion > 0 AND is_first_conversion IS FALSE",
				SQL:         "COALESCE(AVG(time_to_conversion), 0)",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "time_to_conversion > 0 AND is_first_conversion IS FALSE"},
				},
			},

			"acquisition_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "Acquisition orders",
				Description: "Count of id WHERE is_first_conversion IS TRUE",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "is_first_conversion IS TRUE"},
				},
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},

			"retention_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "Retention orders",
				Description: "Count of id WHERE is_first_conversion IS FALSE",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "is_first_conversion IS FALSE"},
				},
			},
			"avg_devices_type": {
				Type:        "number",
				Title:       "Average count of devices type",
				Description: "AVG(devices_type_count)",
				SQL:         "COALESCE(AVG(devices_type_count), 0)",
			},

			"avg_domains": {
				Type:        "number",
				Title:       "Average domains count",
				Description: "AVG(domains_count)",
				SQL:         "COALESCE(AVG(domains_count), 0)",
			},

			"cross_device_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "Cross device conversion paths",
				Description: "Count of id WHERE devices_type_count > 1",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "devices_type_count > 1"},
				},
			},

			"retention_ratio": {
				Type:        "number",
				Title:       "Retention ratio",
				Description: "Retention ratio (retention count / total count)",
				SQL:         "(${retention_count}) / ${count}",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},

			"cross_device_ratio": {
				Type:        "number",
				Title:       "Cross device ratio",
				Description: "Cross device ratio (cross device count / total count)",
				SQL:         "(${cross_device_count}) / ${count}",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},

			"desktop_device_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "Desktop device count",
				Description: "Count of: devices_funnel = 'desktop'",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "devices_funnel = 'desktop'"},
				},
			},

			"desktop_device_ratio": {
				Type:        "number",
				Title:       "Desktop device ratio",
				Description: "Desktop device ratio (desktop device count / total count)",
				SQL:         "(${desktop_device_count}) / ${count}",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},

			"mobile_device_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "Mobile device count",
				Description: "Count of: devices_funnel = 'mobile'",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "devices_funnel = 'mobile'"},
				},
			},

			"mobile_device_ratio": {
				Type:        "number",
				Title:       "Mobile device ratio",
				Description: "Mobile device ratio (mobile device count / total count)",
				SQL:         "(${mobile_device_count}) / ${count}",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},

			"tablet_device_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "Tablet device count",
				Description: "Count of: devices_funnel = 'tablet'",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "devices_funnel = 'tablet'"},
				},
			},

			"tablet_device_ratio": {
				Type:        "number",
				Title:       "Tablet device ratio",
				Description: "Tablet device ratio (tablet device count / total count)",
				SQL:         "(${tablet_device_count}) / ${count}",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},

			"non_cross_device_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "Non cross device count",
				Description: "Count of: devices_type_count = 1",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "devices_type_count = 1"},
				},
			},

			"web_domain_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "Web domain count",
				Description: "Count of: domains_type_funnel = 'web'",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "domains_type_funnel = 'web'"},
				},
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},

			"web_domain_ratio": {
				Type:        "number",
				Title:       "Web domain ratio",
				Description: "Web domain ratio (web domain count / total count)",
				SQL:         "(${web_domain_count}) / ${count}",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},

			"retail_domain_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "Retail domain count",
				Description: "Count of: domains_type_funnel = 'retail'",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "domains_type_funnel = 'retail'"},
				},
			},

			"retail_domain_ratio": {
				Type:        "number",
				Title:       "Retail domain ratio",
				Description: "Retail domain ratio (retail domain count / total count)",
				SQL:         "(${retail_domain_count}) / ${count}",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},

			"app_domain_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "App domain count",
				Description: "Count of: domains_type_funnel = 'app'",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "domains_type_funnel = 'app'"},
				},
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},

			"app_domain_ratio": {
				Type:        "number",
				Title:       "App domain ratio",
				Description: "App domain ratio (app domain count / total count)",
				SQL:         "(${app_domain_count}) / ${count}",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},

			"marketplace_domain_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "Marketplace domain count",
				Description: "Count of: domains_type_funnel = 'marketplace'",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "domains_type_funnel = 'marketplace'"},
				},
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},

			"marketplace_domain_ratio": {
				Type:        "number",
				Title:       "Marketplace domain ratio",
				Description: "Marketplace domain ratio (marketplace domain count / total count)",
				SQL:         "(${marketplace_domain_count}) / ${count}",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},

			"telephone_domain_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "Telephone domain count",
				Description: "Count of: domains_type_funnel = 'telephone'",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "domains_type_funnel = 'telephone'"},
				},
			},

			"telephone_domain_ratio": {
				Type:        "number",
				Title:       "Telephone domain ratio",
				Description: "Telephone domain ratio (telephone domain count / total count)",
				SQL:         "(${telephone_domain_count}) / ${count}",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},

			"web_to_retail_domain_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "Web to retail domain count",
				Description: "Count of: domains_type_funnel = 'web~retail'",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "domains_type_funnel = 'web~retail'"},
				},
			},

			"web_to_retail_domain_ratio": {
				Type:        "number",
				Title:       "Web to retail domain ratio",
				Description: "Web to retail domain ratio (web to retail domain count / total count)",
				SQL:         "(${web_to_retail_domain_count}) / ${count}",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},

			"retail_to_web_domain_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "Retail to web domain count",
				Description: "Count of: domains_type_funnel = 'retail~web'",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "domains_type_funnel = 'retail~web'"},
				},
			},

			"retail_to_web_domain_ratio": {
				Type:        "number",
				Title:       "Retail to web domain ratio",
				Description: "Retail to web domain ratio (retail to web domain count / total count)",
				SQL:         "(${retail_to_web_domain_count}) / ${count}",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},

			"cross_domain_count": {
				Type:        "count",
				SQL:         "id",
				Title:       "Cross domain count",
				Description: "Count of: domains_count > 1",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "domains_count > 1"},
				},
			},

			"cross_domain_ratio": {
				Type:        "number",
				Title:       "Cross domain ratio",
				Description: "Cross domain ratio (cross domain count / total count)",
				SQL:         "(${cross_domain_count}) / ${count}",
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
				Title:       "Order ID",
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

			"subtotal_price": {
				SQL:         "subtotal_price",
				Type:        "number",
				Title:       "Subtotal price",
				Description: "field: subtotal_price",
			},
			"total_price": {
				SQL:         "total_price",
				Type:        "number",
				Title:       "Total price",
				Description: "field: total_price",
			},
			"original_currency": {
				SQL:         "original_currency",
				Type:        "string",
				Title:       "Original currency",
				Description: "field: original_currency",
			},
			"original_subtotal_price": {
				SQL:         "original_subtotal_price",
				Type:        "number",
				Title:       "Original subtotal price",
				Description: "field: original_subtotal_price",
			},
			"original_total_price": {
				SQL:         "original_total_price",
				Type:        "number",
				Title:       "Original total price",
				Description: "field: original_total_price",
			},
			"cancelled_at": {
				SQL:         "created_at_trunc",
				Type:        "time",
				Title:       "Cancelled at",
				Description: "field: cancelled_at",
			},
			"is_first_conversion": {
				SQL:         "is_first_conversion",
				Type:        "number",
				Title:       "Is first conversion",
				Description: "field: is_first_conversion",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"time_to_conversion": {
				SQL:         "time_to_conversion",
				Type:        "number",
				Title:       "Time to conversion",
				Description: "field: time_to_conversion",
			},
			"devices_funnel": {
				SQL:         "devices_funnel",
				Type:        "string",
				Title:       "Devices funnel",
				Description: "field: devices_funnel",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"devices_type_count": {
				SQL:         "devices_type_count",
				Type:        "number",
				Title:       "Devices type count",
				Description: "field: devices_type_count",
			},
			"domains_funnel": {
				SQL:         "domains_funnel",
				Type:        "string",
				Title:       "Domains funnel",
				Description: "field: domains_funnel",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"domains_type_funnel": {
				SQL:         "domains_type_funnel",
				Type:        "string",
				Title:       "Domains type funnel",
				Description: "field: domains_type_funnel",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"domains_count": {
				SQL:         "domains_count",
				Type:        "number",
				Title:       "Domains count",
				Description: "field: domains_count",
			},
			"funnel": {
				SQL:         "funnel",
				Type:        "string",
				Title:       "Funnel",
				Description: "field: funnel",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"funnel_hash": {
				SQL:         "funnel_hash",
				Type:        "string",
				Title:       "Funnel hash",
				Description: "field: funnel_hash",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"attribution_updated_at": {
				SQL:         "attribution_updated_at",
				Type:        "time",
				Title:       "Attribution updated at",
				Description: "field: attribution_updated_at",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
		},
	}
}
