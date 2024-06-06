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
	OrderItemComputedFields []string = []string{
		"converted_price",
		"created_at_trunc",
	}
)

type OrderItems []*OrderItem

func (c OrderItems) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c OrderItems) ToInterface() interface{} {
	v, _ := json.Marshal(c)
	return v
}

func (x *OrderItems) Scan(val interface{}) error {

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

func (x OrderItems) Equals(y OrderItems) bool {

	if len(x) != len(y) {
		return false
	}

	// compare each item
	for _, item := range x {
		// find the same item in the other slice by id
		var found bool
		for _, otherItem := range y {
			if item.ID == otherItem.ID {
				found = true
				if !item.Equals(*otherItem) {
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

type OrderItem struct {
	ID                string          `db:"id" json:"id"`
	ExternalID        string          `db:"external_id" json:"external_id"`
	OrderID           string          `db:"order_id" json:"order_id"`
	UserID            string          `db:"user_id" json:"user_id"`
	ProductExternalID string          `db:"product_external_id" json:"product_external_id"`
	Name              string          `db:"name" json:"name"`
	SKU               *NullableString `db:"sku" json:"sku,omitempty"`
	Brand             *NullableString `db:"brand" json:"brand,omitempty"`
	Category          *NullableString `db:"category" json:"category,omitempty"`
	VariantExternalID *NullableString `db:"variant_external_id" json:"variant_external_id,omitempty"`
	VariantTitle      *NullableString `db:"variant_title" json:"variant_title,omitempty"`
	ImageURL          *NullableString `db:"image_url" json:"image_url,omitempty"`
	Price             int64           `db:"price" json:"price,omitempty"`
	Currency          *string         `db:"currency" json:"currency"` // currency of the order, otherwise use workspace currency as default
	FxRate            float64         `db:"fx_rate" json:"fx_rate"`   // fx rate used to convert the order to the workspace currency
	Quantity          int64           `db:"quantity" json:"quantity,omitempty"`
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

func (o OrderItem) Equals(x OrderItem) bool {

	if updatedFields := o.MergeInto(&x); len(updatedFields) > 0 {
		return false
	}

	return true
}

func (o *OrderItem) GetFieldDate(field string) time.Time {
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

func (s *OrderItem) SetCreatedAt(value time.Time) {
	// update if current value is older
	if value.Before(s.CreatedAt) {
		s.CreatedAt = value
		s.CreatedAtTrunc = value.Truncate(time.Hour)
	}
}

// update a field timestamp to its most recent value
func (o *OrderItem) UpdateFieldTimestamp(field string, timestamp *time.Time) {
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

func (o *OrderItem) SetName(value string, timestamp time.Time) (update *UpdatedField) {
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

func (o *OrderItem) SetSKU(value *NullableString, timestamp time.Time) (update *UpdatedField) {
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
func (o *OrderItem) SetBrand(value *NullableString, timestamp time.Time) (update *UpdatedField) {
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
func (o *OrderItem) SetCategory(value *NullableString, timestamp time.Time) (update *UpdatedField) {
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

func (o *OrderItem) SetVariantExternalID(value *NullableString, timestamp time.Time) (update *UpdatedField) {
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

func (o *OrderItem) SetVariantTitle(value *NullableString, timestamp time.Time) (update *UpdatedField) {
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

func (o *OrderItem) SetImageURL(value *NullableString, timestamp time.Time) (update *UpdatedField) {
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

func (o *OrderItem) SetCurrency(value *string, timestamp time.Time) (update *UpdatedField) {
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
func (o *OrderItem) SetFxRate(value float64, timestamp time.Time) (update *UpdatedField) {
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
func (o *OrderItem) SetPrice(value int64, timestamp time.Time) (update *UpdatedField) {
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

func (o *OrderItem) SetQuantity(value int64, timestamp time.Time) (update *UpdatedField) {
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

func (s *OrderItem) SetExtraColumns(field string, value *AppItemField, timestamp time.Time) (update *UpdatedField) {
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

// merges two orders and returns the list of updated fields
func (fromOrderItem *OrderItem) MergeInto(toOrderItem *OrderItem) (updatedFields []*UpdatedField) {
	updatedFields = []*UpdatedField{} // init

	if toOrderItem.FieldsTimestamp == nil {
		toOrderItem.FieldsTimestamp = FieldsTimestamp{}
	}

	// toOrderItem.SetOrderID(item.OrderID, nil)

	if fieldUpdate := toOrderItem.SetName(fromOrderItem.Name, fromOrderItem.GetFieldDate("name")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toOrderItem.SetSKU(fromOrderItem.SKU, fromOrderItem.GetFieldDate("sku")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toOrderItem.SetBrand(fromOrderItem.Brand, fromOrderItem.GetFieldDate("brand")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toOrderItem.SetCategory(fromOrderItem.Category, fromOrderItem.GetFieldDate("category")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toOrderItem.SetVariantExternalID(fromOrderItem.VariantExternalID, fromOrderItem.GetFieldDate("variant_external_id")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toOrderItem.SetVariantTitle(fromOrderItem.VariantTitle, fromOrderItem.GetFieldDate("variant_title")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toOrderItem.SetImageURL(fromOrderItem.ImageURL, fromOrderItem.GetFieldDate("image_url")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toOrderItem.SetPrice(fromOrderItem.Price, fromOrderItem.GetFieldDate("price")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toOrderItem.SetQuantity(fromOrderItem.Quantity, fromOrderItem.GetFieldDate("quantity")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	// if fieldUpdate := toOrderItem.SetOriginalPrice(item.OriginalPrice, TimePtr(item.GetFieldDate("original_price"))); fieldUpdate != nil {
	// 	updatedFields = append(updatedFields, fieldUpdate)
	// }

	for key, value := range fromOrderItem.ExtraColumns {
		if fieldUpdate := toOrderItem.SetExtraColumns(key, value, fromOrderItem.GetFieldDate(key)); fieldUpdate != nil {
			updatedFields = append(updatedFields, fieldUpdate)
		}
	}

	// UpdatedAt is the timeOfEvent for ITs
	toOrderItem.UpdatedAt = fromOrderItem.UpdatedAt
	// priority to oldest date
	toOrderItem.SetCreatedAt(fromOrderItem.CreatedAt)

	return
}

func NewOrderItem(externalID string, name string, userID string, orderID string, productExternalID string, createdAt time.Time, updatedAt *time.Time) *OrderItem {
	return &OrderItem{
		ID:                ComputeOrderID(externalID),
		ExternalID:        externalID,
		UserID:            userID,
		OrderID:           orderID,
		ProductExternalID: productExternalID,
		Name:              name,
		CreatedAt:         createdAt,
		CreatedAtTrunc:    createdAt.Truncate(time.Hour),
		FieldsTimestamp:   FieldsTimestamp{},

		UpdatedAt:    updatedAt,
		ExtraColumns: AppItemFields{},
	}
}

var OrderItemSchema string = `CREATE TABLE IF NOT EXISTS order_item (
	id VARCHAR(64) NOT NULL,
	external_id VARCHAR(255) NOT NULL,
	order_id VARCHAR(64) NOT NULL,
	user_id VARCHAR(64) NOT NULL,
	product_external_id VARCHAR(128) NOT NULL,
	name VARCHAR(128) NOT NULL,
	sku VARCHAR(128),
	brand VARCHAR(128),
	category VARCHAR(128),
	variant_external_id VARCHAR(128),
	variant_title VARCHAR(128),
	image_url VARCHAR(2083),
	price INT UNSIGNED DEFAULT 0,
	currency VARCHAR(3) NOT NULL,
	fx_rate FLOAT DEFAULT 1 NOT NULL,
	converted_price AS FLOOR(price*fx_rate) PERSISTED INT,
	quantity INT UNSIGNED DEFAULT 0,
	created_at DATETIME NOT NULL,
	created_at_trunc AS DATE_TRUNC('hour', created_at) PERSISTED DATETIME,
	db_created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	is_deleted BOOLEAN DEFAULT FALSE,
	merged_from_user_id VARCHAR(64),
	fields_timestamp JSON NOT NULL,

	SORT KEY (created_at_trunc DESC),
	PRIMARY KEY (id, order_id, user_id),
	KEY (external_id) USING HASH,
	KEY (order_id) USING HASH,
	KEY (product_external_id) USING HASH,
	KEY (user_id) USING HASH, -- for user merging
	SHARD KEY (user_id)
) COLLATE utf8mb4_general_ci;`

var OrderItemSchemaMYSQL string = `CREATE TABLE IF NOT EXISTS order_item (
	id VARCHAR(64) NOT NULL,
	external_id VARCHAR(255) NOT NULL,
	order_id VARCHAR(64) NOT NULL,
	user_id VARCHAR(64) NOT NULL,
	product_external_id VARCHAR(128) NOT NULL,
	name VARCHAR(128) NOT NULL,
	sku VARCHAR(128),
	brand VARCHAR(128),
	category VARCHAR(128),
	variant_external_id VARCHAR(128),
	variant_title VARCHAR(128),
	image_url VARCHAR(2083),
	price INT UNSIGNED DEFAULT 0,
	currency VARCHAR(3) NOT NULL,
	fx_rate FLOAT DEFAULT 1 NOT NULL,
	converted_price INT GENERATED ALWAYS AS (FLOOR(price*fx_rate)) STORED,
	quantity INT UNSIGNED DEFAULT 0,
	created_at DATETIME NOT NULL,
  	created_at_trunc DATETIME GENERATED ALWAYS AS (DATE_FORMAT(created_at, '%Y-%m-%d %H:00:00')) STORED,
	db_created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	is_deleted BOOLEAN DEFAULT FALSE,
	merged_from_user_id VARCHAR(64),
	fields_timestamp JSON NOT NULL,

	-- SORT KEY (created_at_trunc),
	PRIMARY KEY (id, order_id, user_id),
	KEY (external_id) USING HASH,
	KEY (order_id) USING HASH,
	KEY (product_external_id) USING HASH,
	KEY (user_id) USING HASH -- for user merging
	-- SHARD KEY (user_id)
) COLLATE utf8mb4_general_ci;`

func NewOrderItemCube() *CubeJSSchema {
	return &CubeJSSchema{
		Title:       "Order items",
		Description: "Order items",
		SQL:         "SELECT * FROM `order_item`",
		// https://cube.dev/docs/schema/reference/joins
		Joins: map[string]CubeJSSchemaJoin{
			"Order": {
				Relationship: "many_to_one",
				SQL:          "${CUBE}.order_id = ${Order}.id",
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
		},
		Dimensions: map[string]CubeJSSchemaDimension{
			"id": {
				SQL:         "id",
				Type:        "string",
				PrimaryKey:  true,
				Title:       "Order item ID",
				Description: "field: id",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"order_id": {
				SQL:         "order_id",
				Type:        "string",
				Title:       "Order ID",
				Description: "field: order_id",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"user_id": {
				SQL:         "user_id",
				Type:        "string",
				Title:       "Session ID",
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
			"product_external_id": {
				SQL:         "product_external_id",
				Type:        "string",
				Title:       "Product external ID",
				Description: "field: product_external_id",
			},
			"name": {
				SQL:         "name",
				Type:        "string",
				Title:       "Product name",
				Description: "field: name",
			},
			"sku": {
				SQL:         "sku",
				Type:        "string",
				Title:       "SKU",
				Description: "field: sku",
			},
			"brand": {
				SQL:         "brand",
				Type:        "string",
				Title:       "Brand",
				Description: "field: brand",
			},
			"category": {
				SQL:         "category",
				Type:        "string",
				Title:       "Category",
				Description: "field: category",
			},
			"variant_external_id": {
				SQL:         "variant_external_id",
				Type:        "string",
				Title:       "Variant external ID",
				Description: "field: variant_external_id",
			},
			"variant_title": {
				SQL:         "variant_title",
				Type:        "string",
				Title:       "Variant title",
				Description: "field: variant_title",
			},
			"price": {
				SQL:         "price",
				Type:        "number",
				Title:       "Price",
				Description: "field: price",
			},
			"fx_rate": {
				SQL:         "fx_rate",
				Type:        "number",
				Title:       "Fx rate",
				Description: "field: fx_rate",
			},
			"converted_price": {
				SQL:         "converted_price",
				Type:        "number",
				Title:       "Converted price",
				Description: "field: converted_price (price * fx_rate)",
			},
			"quantity": {
				SQL:         "quantity",
				Type:        "number",
				Title:       "Quantity",
				Description: "field: quantity",
			},
			"currency": {
				SQL:         "currency",
				Type:        "string",
				Title:       "Currency",
				Description: "field: currency",
			},
		},
	}
}
