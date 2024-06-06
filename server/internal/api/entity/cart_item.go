package entity

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"strings"
	"time"
)

var (
	// computed fields should be excluded from SELECT/INSERT while cloning rows
	CartItemComputedFields []string = []string{
		"converted_price",
		"created_at_trunc",
	}
)

type CartItems []*CartItem

func (c CartItems) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c CartItems) ToInterface() interface{} {
	v, _ := json.Marshal(c)
	return v
}

func (x *CartItems) Scan(val interface{}) error {

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

func (x CartItems) Equals(y CartItems) bool {

	if len(x) != len(y) {
		return false
	}

	// compare each item
	for _, item := range x {
		// find the same item in the other slice by id
		var found bool
		for _, cartItem := range y {
			if item.ID == cartItem.ID {
				found = true
				if !item.Equals(*cartItem) {
					return false
				}
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}

type CartItem struct {
	ID                string `db:"id" json:"id"`
	ExternalID        string `db:"external_id" json:"external_id"`
	CartID            string `db:"cart_id" json:"cart_id"`
	UserID            string `db:"user_id" json:"user_id"`
	ProductExternalID string `db:"product_external_id" json:"product_external_id"`
	Name              string `db:"name" json:"name"`
	// OrderID           *string         `db:"order_id" json:"order_id,omitempty"`
	SKU               *NullableString `db:"sku" json:"sku,omitempty"`
	Brand             *NullableString `db:"brand" json:"brand,omitempty"`
	Category          *NullableString `db:"category" json:"category,omitempty"`
	VariantExternalID *NullableString `db:"variant_external_id" json:"variant_external_id,omitempty"`
	VariantTitle      *NullableString `db:"variant_title" json:"variant_title,omitempty"`
	ImageURL          *NullableString `db:"image_url" json:"image_url,omitempty"`
	Quantity          int64           `db:"quantity" json:"quantity,omitempty"`
	Price             int64           `db:"price" json:"price,omitempty"`
	Currency          *string         `db:"currency" json:"currency"`                         // currency of the order, otherwise use workspace currency as default
	FxRate            float64         `db:"fx_rate" json:"fx_rate"`                           // fx rate used to convert the order to the workspace currency
	ConvertedPrice    int64           `db:"converted_price" json:"converted_price,omitempty"` // price converted into the workspace currency, computed field
	// DiscountCodes     DiscountCodes   `json:"discount_codes,omitempty"`
	CreatedAt        time.Time       `db:"created_at" json:"created_at"`
	CreatedAtTrunc   time.Time       `db:"created_at_trunc" json:"created_at_trunc"` // date truncated at hour, used to optimize columnstore storage and queries
	DBCreatedAt      time.Time       `db:"db_created_at" json:"db_created_at"`
	DBUpdatedAt      time.Time       `db:"db_updated_at" json:"db_updated_at"`
	IsDeleted        bool            `db:"is_deleted" json:"is_deleted,omitempty"` // deleting rows in transactions cause deadlocks in singlestore, we use an update
	MergedFromUserID *string         `db:"merged_from_user_id" json:"merged_from_user_id,omitempty"`
	FieldsTimestamp  FieldsTimestamp `db:"fields_timestamp" json:"fields_timestamp"`

	// Not persisted in DB:
	UpdatedAt    *time.Time    `db:"-" json:"-"` // used to merge fields and append item_timeline at the right time
	ExtraColumns AppItemFields `db:"-" json:"-"` // converted into "app_xxx" fields when marshaling JSON
}

func (o CartItem) Equals(x CartItem) bool {

	if updatedFields := o.MergeInto(&x); len(updatedFields) > 0 {
		return false
	}

	return true
}

func (o *CartItem) GetFieldDate(field string) time.Time {
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
func (o *CartItem) UpdateFieldTimestamp(field string, timestamp *time.Time) {
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

//	func (o *CartItem) SetOrderID(value *string, timestamp *time.Time) (update *UpdatedField) {
//		key := "order_id"
//		// value cant be null
//		if value == nil {
//			return nil
//		}
//		// abort if values are equal
//		if FixedDeepEqual(o.OrderID, value) {
//			return nil
//		}
//		// compare timestamp with mergeable fields timestamp and mutate value if timestamp is newer
//		// init field update
//		update = &UpdatedField{
//			Field:     key,
//			PrevValue: StringPointerToInterface(o.OrderID),
//		}
//		previousTimestamp, exists := o.FieldsTimestamp[key]
//		if !exists {
//			o.OrderID = value
//			o.UpdateFieldTimestamp(key, timestamp)
//			update.NewValue = StringPointerToInterface(value)
//			return
//		}
//		// abort if a previous timestamp exists, and the current one is not provided
//		if timestamp == nil {
//			return nil
//		}
//		// abort if the current timestamp is older than the previous one
//		if timestamp.Before(previousTimestamp) {
//			return nil
//		}
//		o.OrderID = value
//		o.UpdateFieldTimestamp(key, timestamp)
//		update.NewValue = StringPointerToInterface(value)
//		return
//	}

func (o *CartItem) SetName(value string, timestamp time.Time) (update *UpdatedField) {
	key := "name"
	// abort if values are equal
	if value == o.Name {
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
		o.Name = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: o.Name,
		NewValue:  value,
	}
	o.Name = value
	o.FieldsTimestamp[key] = timestamp
	return
}

func (o *CartItem) SetSKU(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "sku"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && o.SKU != nil && o.SKU.IsNull == value.IsNull && o.SKU.String == value.String {
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
		o.SKU = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(o.SKU),
		NewValue:  NullableStringToInterface(value),
	}
	o.SKU = value
	o.FieldsTimestamp[key] = timestamp
	return
}
func (o *CartItem) SetBrand(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "brand"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && o.Brand != nil && o.Brand.IsNull == value.IsNull && o.Brand.String == value.String {
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
		o.Brand = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(o.Brand),
		NewValue:  NullableStringToInterface(value),
	}
	o.Brand = value
	o.FieldsTimestamp[key] = timestamp
	return
}
func (o *CartItem) SetCategory(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "category"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && o.Category != nil && o.Category.IsNull == value.IsNull && o.Category.String == value.String {
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
		o.Category = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(o.Category),
		NewValue:  NullableStringToInterface(value),
	}
	o.Category = value
	o.FieldsTimestamp[key] = timestamp
	return
}

func (o *CartItem) SetVariantExternalID(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "variant_external_id"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && o.VariantExternalID != nil && o.VariantExternalID.IsNull == value.IsNull && o.VariantExternalID.String == value.String {
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
		o.VariantExternalID = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(o.VariantExternalID),
		NewValue:  NullableStringToInterface(value),
	}
	o.VariantExternalID = value
	o.FieldsTimestamp[key] = timestamp
	return
}

func (o *CartItem) SetVariantTitle(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "variant_title"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && o.VariantTitle != nil && o.VariantTitle.IsNull == value.IsNull && o.VariantTitle.String == value.String {
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
		o.VariantTitle = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(o.VariantTitle),
		NewValue:  NullableStringToInterface(value),
	}
	o.VariantTitle = value
	o.FieldsTimestamp[key] = timestamp
	return
}

func (o *CartItem) SetImageURL(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "image_url"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && o.ImageURL != nil && o.ImageURL.IsNull == value.IsNull && o.ImageURL.String == value.String {
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
		o.ImageURL = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(o.ImageURL),
		NewValue:  NullableStringToInterface(value),
	}
	o.ImageURL = value
	o.FieldsTimestamp[key] = timestamp
	return
}

func (o *CartItem) SetCurrency(value *string, timestamp time.Time) (update *UpdatedField) {
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
func (o *CartItem) SetFxRate(value float64, timestamp time.Time) (update *UpdatedField) {
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

// SetPrice should only be compared if there is no OriginalPrice set
func (o *CartItem) SetPrice(value int64, timestamp time.Time) (update *UpdatedField) {
	key := "price"
	// abort if values are equal
	if o.Price == value {
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
		o.Price = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: o.Price,
		NewValue:  value,
	}
	o.Price = value
	o.FieldsTimestamp[key] = timestamp
	return
}

func (o *CartItem) SetQuantity(value int64, timestamp time.Time) (update *UpdatedField) {
	key := "quantity"
	// abort if values are equal
	if o.Quantity == value {
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
		o.Quantity = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: o.Quantity,
		NewValue:  value,
	}
	o.Quantity = value
	o.FieldsTimestamp[key] = timestamp
	return
}

func (s *CartItem) SetExtraColumns(field string, value *AppItemField, timestamp time.Time) (update *UpdatedField) {
	if s.ExtraColumns == nil {
		s.ExtraColumns = AppItemFields{}
	}

	// abort if field doesnt start with "app_" or "appx_"
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

// merges two orders and returns the list of updated fields
func (fromCartItem *CartItem) MergeInto(toCartItem *CartItem) (updatedFields []*UpdatedField) {
	updatedFields = []*UpdatedField{} // init

	if toCartItem.FieldsTimestamp == nil {
		toCartItem.FieldsTimestamp = FieldsTimestamp{}
	}

	// toCartItem.SetOrderID(item.OrderID, nil)

	if fieldUpdate := toCartItem.SetName(fromCartItem.Name, fromCartItem.GetFieldDate("name")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toCartItem.SetSKU(fromCartItem.SKU, fromCartItem.GetFieldDate("sku")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toCartItem.SetBrand(fromCartItem.Brand, fromCartItem.GetFieldDate("brand")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toCartItem.SetCategory(fromCartItem.Category, fromCartItem.GetFieldDate("category")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toCartItem.SetVariantExternalID(fromCartItem.VariantExternalID, fromCartItem.GetFieldDate("variant_external_id")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toCartItem.SetVariantTitle(fromCartItem.VariantTitle, fromCartItem.GetFieldDate("variant_title")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toCartItem.SetImageURL(fromCartItem.ImageURL, fromCartItem.GetFieldDate("image_url")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toCartItem.SetPrice(fromCartItem.Price, fromCartItem.GetFieldDate("price")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toCartItem.SetQuantity(fromCartItem.Quantity, fromCartItem.GetFieldDate("quantity")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	// if fieldUpdate := toCartItem.SetOriginalPrice(item.OriginalPrice, TimePtr(item.GetFieldDate("original_price"))); fieldUpdate != nil {
	// 	updatedFields = append(updatedFields, fieldUpdate)
	// }

	for key, value := range fromCartItem.ExtraColumns {
		if fieldUpdate := toCartItem.SetExtraColumns(key, value, fromCartItem.GetFieldDate(key)); fieldUpdate != nil {
			updatedFields = append(updatedFields, fieldUpdate)
		}
	}

	// UpdatedAt is the timeOfEvent for ITs
	toCartItem.UpdatedAt = fromCartItem.UpdatedAt
	// priority to oldest date
	toCartItem.CreatedAt = fromCartItem.CreatedAt

	return
}

func NewCartItem(externalID string, name string, userID string, cartID string, productExternalID string, createdAt time.Time, updatedAt *time.Time) *CartItem {
	return &CartItem{
		ID:                ComputeOrderID(externalID),
		ExternalID:        externalID,
		UserID:            userID,
		CartID:            cartID,
		ProductExternalID: productExternalID,
		Name:              name,
		CreatedAt:         createdAt,
		CreatedAtTrunc:    createdAt.Truncate(time.Hour),
		FieldsTimestamp:   FieldsTimestamp{},

		UpdatedAt:    updatedAt,
		ExtraColumns: AppItemFields{},
	}
}

var CartItemSchema string = `CREATE TABLE IF NOT EXISTS cart_item (
	id VARCHAR(64) NOT NULL,
	external_id VARCHAR(255) NOT NULL,
	cart_id VARCHAR(64) NOT NULL,
	user_id VARCHAR(64) NOT NULL,
	product_external_id VARCHAR(128) NOT NULL,
	-- order_id VARCHAR(64),
	name VARCHAR(128) NOT NULL,
	sku VARCHAR(128),
	brand VARCHAR(128),
	category VARCHAR(128),
	variant_external_id VARCHAR(128),
	variant_title VARCHAR(128),
	image_url VARCHAR(2083),
	quantity INT UNSIGNED DEFAULT 0,
	price INT UNSIGNED DEFAULT 0,
	currency VARCHAR(3) NOT NULL,
	fx_rate FLOAT DEFAULT 1,
	converted_price AS FLOOR(price*fx_rate) PERSISTED INT,
	created_at DATETIME NOT NULL,
	created_at_trunc AS DATE_TRUNC('hour', created_at) PERSISTED DATETIME,
	db_created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	is_deleted BOOLEAN DEFAULT FALSE,
	merged_from_user_id VARCHAR(64),
	fields_timestamp JSON NOT NULL,

	SORT KEY (created_at_trunc),
	PRIMARY KEY (id, cart_id, user_id),
	KEY (external_id) USING HASH,
	-- KEY (order_id) USING HASH,
	KEY (sku) USING HASH,
	KEY (user_id) USING HASH, -- for user merging
	SHARD KEY (user_id)
) COLLATE utf8mb4_general_ci;`

var CartItemSchemaMYSQL string = `CREATE TABLE IF NOT EXISTS cart_item (
	id VARCHAR(64) NOT NULL,
	external_id VARCHAR(255) NOT NULL,
	cart_id VARCHAR(64) NOT NULL,
	user_id VARCHAR(64) NOT NULL,
	product_external_id VARCHAR(64) NOT NULL,
	-- order_id VARCHAR(64),
	name VARCHAR(128) NOT NULL,
	sku VARCHAR(128),
	brand VARCHAR(128),
	category VARCHAR(128),
	variant_external_id VARCHAR(128),
	variant_title VARCHAR(128),
	image_url VARCHAR(2083),
	quantity INT UNSIGNED DEFAULT 0,
	price INT UNSIGNED DEFAULT 0,
	currency VARCHAR(3) NOT NULL,
	fx_rate FLOAT DEFAULT 1,
	converted_price INT GENERATED ALWAYS AS (FLOOR(price*fx_rate)) STORED,
	created_at DATETIME NOT NULL,
    created_at_trunc DATETIME GENERATED ALWAYS AS (DATE_FORMAT(created_at, '%Y-%m-%d %H:00:00')) STORED,
	db_created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	is_deleted BOOLEAN DEFAULT FALSE,
	merged_from_user_id VARCHAR(64),
	fields_timestamp JSON NOT NULL,

	-- SORT KEY (created_at_trunc),
	PRIMARY KEY (id, cart_id, user_id),
	KEY (external_id) USING HASH,
	-- KEY (order_id) USING HASH,
	KEY (sku) USING HASH,
	KEY (user_id) USING HASH -- for user merging
	-- SHARD KEY (user_id)
) COLLATE utf8mb4_general_ci;`
