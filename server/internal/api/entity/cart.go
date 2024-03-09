package entity

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/rimdian/rimdian/internal/api/common"
	"github.com/rotisserie/eris"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var (
	ErrCartRequired = eris.New("cart is required")

	// computed fields should be excluded from SELECT/INSERT while cloning rows
	CartComputedFields []string = []string{
		"created_at_trunc",
	}
)

type CartStatusType = int

const (
	CartStatusAbandoned CartStatusType = iota
	CartStatusConverted
	CartStatusRecovered
)

type Cart struct {
	ID         string  `db:"id" json:"id"`
	ExternalID string  `db:"external_id" json:"external_id"`
	UserID     string  `db:"user_id" json:"user_id"`
	DomainID   string  `db:"domain_id" json:"domain_id"`
	SessionID  *string `db:"session_id" json:"session_id,omitempty"`
	// OrderID          *string         `db:"order_id" json:"order_id,omitempty"`
	CreatedAt        time.Time       `db:"created_at" json:"created_at"`
	CreatedAtTrunc   time.Time       `db:"created_at_trunc" json:"created_at_trunc"` // date truncated at hour, used to optimize columnstore storage and queries
	DBCreatedAt      time.Time       `db:"db_created_at" json:"db_created_at"`
	DBUpdatedAt      time.Time       `db:"db_updated_at" json:"db_updated_at"`
	IsDeleted        bool            `db:"is_deleted" json:"is_deleted,omitempty"` // deleting rows in transactions cause deadlocks in singlestore, we use an update
	MergedFromUserID *string         `db:"merged_from_user_id" json:"merged_from_user_id,omitempty"`
	FieldsTimestamp  FieldsTimestamp `db:"fields_timestamp" json:"fields_timestamp"`

	Currency  *string         `db:"currency" json:"currency"`
	FxRate    float64         `db:"fx_rate" json:"fx_rate"`                 // fx rate used to convert the order to the workspace currency
	PublicURL *NullableString `db:"public_url" json:"public_url,omitempty"` // url sent in emails to retrieve the shopping cart in case of abandoned cart
	Status    *NullableInt64  `db:"status" json:"status"`                   // 0: abandoned, 1: converted, 2: recovered

	// Not persisted in DB:
	Items        CartItems     `db:"-" json:"items"`
	ExtraColumns AppItemFields `db:"-" json:"extra_columns"` // converted into "app_xxx" fields when marshaling JSON
	UpdatedAt    *time.Time    `db:"-" json:"-"`             // used to merge fields and append item_timeline at the right time
}

func (o *Cart) GetFieldDate(field string) time.Time {
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
func (o *Cart) UpdateFieldTimestamp(field string, timestamp *time.Time) {
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

func (s *Cart) SetCreatedAt(value time.Time) {
	// update if current value is older
	if value.Before(s.CreatedAt) {
		s.CreatedAt = value
		s.CreatedAtTrunc = value.Truncate(time.Hour)
	}
}
func (o *Cart) SetCurrency(value *string, timestamp time.Time) (update *UpdatedField) {
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
func (o *Cart) SetFxRate(value float64, timestamp time.Time) (update *UpdatedField) {
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
func (o *Cart) SetSessionID(value *string, timestamp time.Time) (update *UpdatedField) {
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

// TODO: revoir les SET avec nouveau format, comme dans order

func (o *Cart) SetPublicURL(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "public_url"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && o.PublicURL != nil && o.PublicURL.IsNull == value.IsNull && o.PublicURL.String == value.String {
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
		o.PublicURL = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(o.PublicURL),
		NewValue:  NullableStringToInterface(value),
	}
	o.PublicURL = value
	o.FieldsTimestamp[key] = timestamp
	return
}

// func (o *Cart) SetItems(value CartItems, timestamp time.Time) (update *UpdatedField) {
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

func (o *Cart) SetStatus(value *NullableInt64, timestamp time.Time) (update *UpdatedField) {
	key := "total_price"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && o.Status != nil && o.Status.IsNull == value.IsNull && o.Status.Int64 == value.Int64 {
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
		PrevValue: NullableInt64ToInterface(o.Status),
		NewValue:  NullableInt64ToInterface(value),
	}
	o.Status = value
	o.FieldsTimestamp[key] = timestamp
	return
}

func (s *Cart) SetExtraColumns(field string, value *AppItemField, timestamp time.Time) (update *UpdatedField) {

	if s.ExtraColumns == nil {
		s.ExtraColumns = AppItemFields{}
	}

	// abort if field doesnt start with "app_" or "appx_"
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

func (fromCart *Cart) MergeInto(toCart *Cart) (updatedFields []*UpdatedField) {
	updatedFields = []*UpdatedField{} // init

	if toCart.FieldsTimestamp == nil {
		toCart.FieldsTimestamp = FieldsTimestamp{}
	}

	if fieldUpdate := toCart.SetSessionID(fromCart.SessionID, fromCart.GetFieldDate("session_id")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toCart.SetCurrency(fromCart.Currency, fromCart.GetFieldDate("currency")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	// if fieldUpdate := toCart.SetItems(fromCart.Items, fromCart.GetFieldDate("items")); fieldUpdate != nil {
	// 	updatedFields = append(updatedFields, fieldUpdate)
	// }
	if fieldUpdate := toCart.SetPublicURL(fromCart.PublicURL, fromCart.GetFieldDate("public_url")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toCart.SetStatus(fromCart.Status, fromCart.GetFieldDate("status")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	for key, value := range fromCart.ExtraColumns {
		if fieldUpdate := toCart.SetExtraColumns(key, value, fromCart.GetFieldDate(key)); fieldUpdate != nil {
			updatedFields = append(updatedFields, fieldUpdate)
		}
	}

	// UpdatedAt is the timeOfEvent for ITs
	toCart.UpdatedAt = fromCart.UpdatedAt
	// priority to oldest date
	toCart.SetCreatedAt(fromCart.CreatedAt)

	return
}

// overwrite json marshaller, to convert map of extra columns into "app_xxx" fields
func (s *Cart) MarshalJSON() ([]byte, error) {

	type Alias Cart

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
			return nil, eris.Errorf("set cart custom dimension err: %v", err)
		}
	}

	return []byte(jsonValue), nil
}

func NewCart(externalID string, userID string, domainID string, createdAt time.Time, updatedAt *time.Time, items []*CartItem) *Cart {

	// default empty items
	if items == nil {
		items = []*CartItem{}
	}

	return &Cart{
		ID:              ComputeCartID(externalID),
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

func NewCartFromDataLog(dataLog *DataLog, clockDifference time.Duration, workspace *Workspace) (cart *Cart, err error) {

	result := gjson.Get(dataLog.Item, "cart")
	if !result.Exists() {
		return nil, eris.New("item has no cart object")
	}

	if result.Type != gjson.JSON {
		return nil, eris.New("item cart is not an object")
	}

	extraColumns := workspace.FindExtraColumnsForItemKind("cart")

	// init
	cart = &Cart{
		UserID:          dataLog.UserID,
		FieldsTimestamp: FieldsTimestamp{},
		Items:           []*CartItem{},
		ExtraColumns:    AppItemFields{},
	}

	// loop over cart fields
	result.ForEach(func(key, value gjson.Result) bool {

		keyString := key.String()
		switch keyString {

		case "external_id":
			if value.Type == gjson.Null {
				err = eris.New("external_id is required")
				return false
			}

			cart.ExternalID = value.String()
			cart.ID = ComputeCartID(cart.ExternalID)

		case "domain_id":
			if value.Type == gjson.Null {
				err = eris.New("domain_id is required")
				return false
			}

			cart.DomainID = value.String()

		case "session_external_id":
			if value.Type == gjson.Null {
				cart.SessionID = nil
			} else {
				cart.SessionID = StringPtr(ComputeSessionID(value.String()))
			}

		case "created_at":
			if cart.CreatedAt, err = time.Parse(time.RFC3339, value.String()); err != nil {
				err = eris.Wrap(err, "cart.created_at")
				return false
			}

			// apply clock difference
			if cart.CreatedAt.After(time.Now()) {

				cart.CreatedAt = cart.CreatedAt.Add(clockDifference)
				if cart.CreatedAt.After(time.Now()) {
					err = eris.New("cart.created_at cannot be in the future")
					return false
				}
			}

			cart.CreatedAtTrunc = cart.CreatedAt.Truncate(time.Hour)

		case "updated_at":
			var updatedAt time.Time
			if updatedAt, err = time.Parse(time.RFC3339, value.String()); err != nil {
				err = eris.Wrap(err, "cart.updated_at")
				return false
			}

			if updatedAt.After(time.Now()) {
				// apply clock difference
				updatedAtModified := updatedAt.Add(clockDifference)
				updatedAt = updatedAtModified
				if updatedAt.After(time.Now()) {
					err = eris.New("cart.updated_at cannot be in the future")
					return false
				}
			}

			cart.UpdatedAt = &updatedAt

		case "currency":
			cart.Currency = StringPtr(value.String())

		case "public_url":
			if value.Type == gjson.Null {
				cart.PublicURL = NewNullableString(nil)
			} else {
				cart.PublicURL = NewNullableString(StringPtr(strings.TrimSpace(value.String())))
			}

		case "status":
			if value.Type == gjson.Null {
				cart.Status = NewNullableInt64(nil)
			} else {
				cart.Status = NewNullableInt64(Int64Ptr(value.Int()))
			}

		case "items":

			if value.Type != gjson.JSON {
				err = eris.New("cart.items is not an array")
				return false
			}

			// loop over items
			value.ForEach(func(key, item gjson.Result) bool {

				if value.Type != gjson.JSON {
					err = eris.New("cart.items is not an array of objects")
					return false
				}

				// init
				cartItem := &CartItem{
					FieldsTimestamp: FieldsTimestamp{},
				}

				// loop over item fields
				item.ForEach(func(itemKey, itemField gjson.Result) bool {

					itemKeyString := itemKey.String()

					switch itemKeyString {

					case "external_id":
						cartItem.ExternalID = strings.TrimSpace(itemField.String())
						cartItem.ID = ComputeOrderID(cartItem.ExternalID)

					case "product_external_id":
						cartItem.ProductExternalID = strings.TrimSpace(itemField.String())

					case "name":
						cartItem.Name = strings.TrimSpace(itemField.String())

					case "sku":
						if itemField.Type == gjson.Null {
							cartItem.SKU = NewNullableString(nil)
						} else {
							cartItem.SKU = NewNullableString(StringPtr(strings.TrimSpace(itemField.String())))
						}

					case "brand":
						if itemField.Type == gjson.Null {
							cartItem.Brand = NewNullableString(nil)
						} else {
							cartItem.Brand = NewNullableString(StringPtr(strings.TrimSpace(itemField.String())))
						}

					case "category":
						if itemField.Type == gjson.Null {
							cartItem.Category = NewNullableString(nil)
						} else {
							cartItem.Category = NewNullableString(StringPtr(strings.TrimSpace(itemField.String())))
						}

					case "variant_external_id":
						if itemField.Type == gjson.Null {
							cartItem.VariantExternalID = NewNullableString(nil)
						} else {
							cartItem.VariantExternalID = NewNullableString(StringPtr(strings.TrimSpace(itemField.String())))
						}

					case "variant_title":
						if itemField.Type == gjson.Null {
							cartItem.VariantTitle = NewNullableString(nil)
						} else {
							cartItem.VariantTitle = NewNullableString(StringPtr(strings.TrimSpace(itemField.String())))
						}

					case "image_url":
						if itemField.Type == gjson.Null {
							cartItem.ImageURL = NewNullableString(nil)
						} else {
							cartItem.ImageURL = NewNullableString(StringPtr(strings.TrimSpace(itemField.String())))
						}

					case "price":
						if itemField.Type != gjson.Number {
							err = eris.New("cart.items.price is not an integer")
							return false
						}

						price := itemField.Int()

						if price < 0 {
							err = eris.New("cart.items.price cannot be negative")
							return false
						}

						cartItem.Price = price

					case "quantity":
						if itemField.Type != gjson.Number {
							err = eris.New("cart.items.quantity is not an integer")
							return false
						}

						quantity := itemField.Int()

						if quantity < 0 {
							err = eris.New("cart.items.quantity cannot be negative")
							return false
						}

						cartItem.Quantity = quantity

					default:
						// TODO: handle app_ extra columns
					}

					return true
				})

				if err != nil {
					return false
				}

				// validate item
				if cartItem.ExternalID == "" {
					err = eris.New("cart.external_id is required")
					return false
				}
				if cartItem.Name == "" {
					err = eris.New("cart.name is required")
					return false
				}
				if cartItem.ProductExternalID == "" {
					err = eris.New("cart.product_external_id is required")
					return false
				}

				cart.Items = append(cart.Items, cartItem)

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
						cart.ExtraColumns[col.Name] = fieldValue
					}
				}
			}
		}

		return true
	})

	if err != nil {
		return nil, err
	}

	if cart.DomainID == "" && dataLog.DomainID != nil {
		cart.DomainID = *dataLog.DomainID
	}

	if cart.Currency == nil || *cart.Currency == "" {
		cart.Currency = &workspace.Currency
	}

	// set fx_rate if needed
	if *cart.Currency != workspace.Currency {
		if fxRate, err := workspace.GetFxRateForCurrency(*cart.Currency); err == nil {
			cart.FxRate = fxRate
		} else {
			return nil, err
		}
	}

	// default rate is 1.0 if currency is the same as workspace
	if cart.FxRate == 0 {
		noRate := 1.0
		cart.FxRate = noRate
	}

	// use data import createdAt as updatedAt if not provided
	if cart.UpdatedAt == nil {
		cart.UpdatedAt = &cart.CreatedAt
	}

	// enrich cart with session and domain
	if dataLog.UpsertedSession != nil {
		if cart.SessionID == nil {
			cart.SessionID = &dataLog.UpsertedSession.ID
		}
		if cart.DomainID == "" {
			cart.DomainID = dataLog.UpsertedSession.DomainID
		}
	}

	// set cart_items fields
	for _, item := range cart.Items {
		item.CartID = cart.ID
		item.UserID = cart.UserID
		item.CreatedAt = cart.CreatedAt
		item.CreatedAtTrunc = cart.CreatedAtTrunc
		item.UpdatedAt = cart.UpdatedAt
		item.Currency = cart.Currency
		item.FxRate = cart.FxRate
	}

	// Validation

	// verify that domainID exists
	found := false
	for _, domain := range workspace.Domains {
		if domain.ID == cart.DomainID {
			found = true
			break
		}
	}

	if !found {
		return nil, eris.New("cart domain_id invalid")
	}

	if cart.CreatedAt.IsZero() {
		return nil, eris.New("cart.created_at is required")
	}

	if !govalidator.IsIn(*cart.Currency, common.CurrenciesCodes...) {
		return nil, eris.New("cart.currency invalid")
	}
	if cart.PublicURL != nil && !cart.PublicURL.IsNull && !govalidator.IsRequestURL(cart.PublicURL.String) {
		return nil, eris.New("cart.public_url is not an URL")
	}
	if cart.Status != nil && !cart.Status.IsNull && cart.Status.Int64 < 0 && cart.Status.Int64 > 2 {
		return nil, eris.New("cart.status invalid")
	}

	return cart, nil
}

func ComputeCartID(externalID string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(externalID)))
}

var CartSchema string = `CREATE TABLE IF NOT EXISTS cart (
	id VARCHAR(64) NOT NULL,
	external_id VARCHAR(255) NOT NULL,
	user_id VARCHAR(64) NOT NULL,
	domain_id VARCHAR(64) NOT NULL,
	session_id VARCHAR(64),
	-- order_id VARCHAR(64),
	created_at DATETIME NOT NULL,
	created_at_trunc AS DATE_TRUNC('hour', created_at) PERSISTED DATETIME,
	db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	is_deleted BOOLEAN DEFAULT FALSE,
	merged_from_user_id VARCHAR(64),
	fields_timestamp JSON NOT NULL,

	currency VARCHAR(3) NOT NULL,
	fx_rate FLOAT DEFAULT 1,

	public_url VARCHAR(2083),
	status TINYINT,

	SORT KEY (created_at_trunc DESC),
	PRIMARY KEY (id, user_id),
	KEY (session_id) USING HASH,
	-- KEY (order_id) USING HASH,
	KEY (user_id) USING HASH, -- for merging
	SHARD KEY (user_id)
) COLLATE utf8mb4_general_ci;`

var CartSchemaMYSQL string = `CREATE TABLE IF NOT EXISTS cart (
	id VARCHAR(64) NOT NULL,
	external_id VARCHAR(255) NOT NULL,
	user_id VARCHAR(64) NOT NULL,
	domain_id VARCHAR(64) NOT NULL,
	session_id VARCHAR(64),
	-- order_id VARCHAR(64),
	created_at DATETIME NOT NULL,
	created_at_trunc DATETIME GENERATED ALWAYS AS (DATE_FORMAT(created_at, '%Y-%m-%d %H:00:00')) STORED,
	db_created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	is_deleted BOOLEAN DEFAULT FALSE,
	merged_from_user_id VARCHAR(64),
	fields_timestamp JSON NOT NULL,

	currency VARCHAR(3) NOT NULL,
	fx_rate FLOAT DEFAULT 1,

	public_url VARCHAR(2083),
	status TINYINT,

	-- SORT KEY (created_at_trunc),
	PRIMARY KEY (id, user_id),
	KEY (session_id) USING HASH,
	-- KEY (order_id) USING HASH,
	KEY (user_id) USING HASH -- for merging
	-- SHARD KEY (user_id)
) COLLATE utf8mb4_general_ci;`

func NewCartCube() *CubeJSSchema {
	return &CubeJSSchema{
		Title:       "Carts",
		Description: "Carts",
		SQL:         "SELECT * FROM `cart`",
		Joins: map[string]CubeJSSchemaJoin{

			"Session": {
				Relationship: "many_to_one",
				SQL:          "${CUBE}.user_id = ${Session}.user_id AND ${CUBE}.session_id = ${Session}.id",
			},
			"Order": {
				Relationship: "many_to_one",
				SQL:          "${CUBE}.user_id = ${Order}.user_id AND ${CUBE}.order_id = ${Order}.id",
			},
			"User": {
				Relationship: "many_to_one",
				SQL:          "${CUBE}.user_id = ${User}.id",
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
			"abandoned_count": {
				Type: "count",
				SQL:  "",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "status = 0"},
				},
				Title:       "Abandoned carts",
				Description: "Count carts WHERE status = 0",
			},
			"converted_count": {
				Type: "count",
				SQL:  "",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "status = 1"},
				},
				Title:       "Converted carts",
				Description: "Count carts WHERE status = 1",
			},
			"recovered_count": {
				Type: "count",
				SQL:  "",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "status = 2"},
				},
				Title:       "Recovered carts",
				Description: "Count carts WHERE status = 2",
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

			"status": {
				SQL:         "status",
				Type:        "number",
				Title:       "Status",
				Description: "field: status",
			},
		},
	}
}

var CubeSchemaCarts = `cube('Cart', {
  sql: 'SELECT * FROM cart',

  joins: {
    Session: {
      relationship: 'many_to_one',
      sql: CUBE + '.user_id = ' + Session + '.user_id AND ' + CUBE + '.session_id = ' + Session + '.id'
    },
    Order: {
      relationship: 'many_to_one',
      sql: CUBE + '.user_id = ' + Order + '.user_id AND ' + CUBE + '.order_id = ' + Order + '.id'
    },
    User: {
      relationship: 'many_to_one',
      sql: CUBE + '.user_id = ' + User + '.id'
    },
  },

  measures: {
    count: {
      type: 'count'
    },

    unique_users: {
      type: 'countDistinct',
      sql: 'user_id',
    },

    abandoned_count: {
      type: 'count',
      filters: [
        { sql: 'status = 0' }
      ]
    },

    converted_count: {
      type: 'count',
      filters: [
        { sql: 'status = 1' }
      ]
    },

    recovered_count: {
      type: 'count',
      filters: [
        { sql: 'status = 2' }
      ]
    },
  },


  dimensions: {
    id: {
      sql: 'id',
      type: 'string',
      primaryKey: true
    },

    created_at: { sql: 'created_at', type: 'time' },
    created_at_trunc: { sql: 'created_at_trunc', type: 'time' },

    status: { sql: 'status', type: 'number' },
  }

});`
