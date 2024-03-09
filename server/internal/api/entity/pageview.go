package entity

import (
	"crypto/sha1"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/rimdian/rimdian/internal/api/common"
	"github.com/rotisserie/eris"
	"github.com/tidwall/gjson"
)

var (
	ErrPageviewRequired = eris.New("pageview is required")

	// computed fields should be excluded from SELECT/INSERT while cloning rows
	PageviewComputedFields []string = []string{
		"product_converted_price",
		"created_at_trunc",
	}
)

type Pageview struct {
	ID               string          `db:"id" json:"id"`
	ExternalID       string          `db:"external_id" json:"external_id"`
	UserID           string          `db:"user_id" json:"user_id"`
	DomainID         string          `db:"domain_id" json:"domain_id"`
	SessionID        string          `db:"session_id" json:"session_id"`
	CreatedAt        time.Time       `db:"created_at" json:"created_at"`
	CreatedAtTrunc   time.Time       `db:"created_at_trunc" json:"created_at_trunc"` // date truncated at hour, used to optimize columnstore storage and queries
	DBCreatedAt      time.Time       `db:"db_created_at" json:"db_created_at"`
	DBUpdatedAt      time.Time       `db:"db_updated_at" json:"db_updated_at"`
	IsDeleted        bool            `db:"is_deleted" json:"is_deleted,omitempty"` // deleting rows in transactions cause deadlocks in singlestore, we use an update
	MergedFromUserID *string         `db:"merged_from_user_id" json:"merged_from_user_id,omitempty"`
	FieldsTimestamp  FieldsTimestamp `db:"fields_timestamp" json:"fields_timestamp"`

	PageID         string          `db:"page_id" json:"page_id,omitempty"`
	Title          string          `db:"title" json:"title,omitempty"`
	Referrer       *NullableString `db:"referrer" json:"referrer,omitempty"`
	ReferrerDomain *string         `db:"referrer_domain" json:"referrer_domain,omitempty"`
	ReferrerPath   *string         `db:"referrer_path" json:"referrer_path,omitempty"`
	Duration       *int64          `db:"duration" json:"duration,omitempty"`
	ImageURL       *NullableString `db:"image_url" json:"image_url,omitempty"`

	ProductExternalID        *NullableString `db:"product_external_id" json:"product_external_id,omitempty"`
	ProductSKU               *NullableString `db:"product_sku" json:"product_sku,omitempty"`
	ProductName              *NullableString `db:"product_name" json:"product_name,omitempty"`
	ProductBrand             *NullableString `db:"product_brand" json:"product_brand,omitempty"`
	ProductCategory          *NullableString `db:"product_category" json:"product_category,omitempty"`
	ProductVariantExternalID *NullableString `db:"product_variant_external_id" json:"product_variant_external_id,omitempty"`
	ProductVariantTitle      *NullableString `db:"product_variant_title" json:"product_variant_title,omitempty"`
	ProductPrice             *NullableInt64  `db:"product_price" json:"product_price,omitempty"`
	ProductCurrency          *NullableString `db:"product_currency" json:"product_currency,omitempty"`
	ProductFxRate            *float64        `db:"product_fx_rate" json:"product_fx_rate"`                           // fx rate used to convert the order to the workspace currency
	ProductConvertedPrice    *NullableInt64  `db:"product_converted_price" json:"product_converted_price,omitempty"` // price before currency conversion
	// ProductConversionRateError *NullableString `db:"product_conversion_rate_error" json:"product_conversion_rate_error,omitempty"`

	// Article *PageviewArticleData `json:"article,omitempty"`

	// Not persisted in DB:
	UpdatedAt    *time.Time    `db:"-" json:"-"` // used to merge fields and append item_timeline at the right time
	ExtraColumns AppItemFields `db:"-" json:"-"` // converted into "app_xxx" fields when marshaling JSON
}

func (s *Pageview) GetFieldDate(field string) time.Time {
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
func (s *Pageview) UpdateFieldTimestamp(field string, timestamp *time.Time) {
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

func (s *Pageview) SetExtraColumns(field string, value *AppItemField, timestamp *time.Time) (update *UpdatedField) {
	if s.ExtraColumns == nil {
		s.ExtraColumns = AppItemFields{}
	}

	// abort if field doesnt start with "app_"
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

func (s *Pageview) SetCreatedAt(value time.Time) {
	// update if current value is older
	if value.Before(s.CreatedAt) {
		s.CreatedAt = value
		s.CreatedAtTrunc = value.Truncate(time.Hour)
	}
}
func (p *Pageview) SetPageID(value string, timestamp time.Time) (update *UpdatedField) {
	key := "page_id"
	// value cant be null
	if value == "" {
		return nil
	}
	// abort if values are equal
	if p.PageID == value {
		return nil
	}
	existingValueTimestamp := p.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		p.PageID = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: p.PageID,
		NewValue:  value,
	}
	p.PageID = value
	p.FieldsTimestamp[key] = timestamp
	return
}
func (p *Pageview) SetTitle(value string, timestamp time.Time) (update *UpdatedField) {
	key := "title"
	// value cant be null
	if value == "" {
		return nil
	}
	// abort if values are equal
	if p.Title == value {
		return nil
	}
	existingValueTimestamp := p.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		p.Title = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: p.Title,
		NewValue:  value,
	}
	p.Title = value
	p.FieldsTimestamp[key] = timestamp
	return
}
func (s *Pageview) SetReferrer(value *NullableString, timestamp time.Time) (update *UpdatedField) {
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
func (s *Pageview) SetReferrerDomain(value *string, timestamp time.Time) (update *UpdatedField) {
	key := "referrer_domain"
	// value cant be null
	if value == nil {
		return nil
	}
	// abort if values are equal
	if StringsEqual(s.ReferrerDomain, value) {
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
func (s *Pageview) SetReferrerPath(value *string, timestamp time.Time) (update *UpdatedField) {
	key := "referrer_path"
	// value cant be null
	if value == nil {
		return nil
	}
	// abort if values are equal
	if StringsEqual(s.ReferrerPath, value) {
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
func (s *Pageview) SetDuration(value *int64, timestamp time.Time) (update *UpdatedField) {
	key := "duration"
	// value cant be null
	if value == nil {
		return nil
	}
	// set value to 10800 secs (60x60x3 hours) if value is above
	if value != nil && *value > 10800 {
		value = Int64Ptr(10800)
	}
	// abort if values are equal
	if Int64Equal(s.Duration, value) {
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
		PrevValue: Int64PointerToInterface(s.Duration),
		NewValue:  Int64PointerToInterface(value),
	}
	s.Duration = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Pageview) SetImageURL(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "image_url"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.ImageURL != nil && s.ImageURL.IsNull == value.IsNull && s.ImageURL.String == value.String {
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
		s.ImageURL = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.ImageURL),
		NewValue:  NullableStringToInterface(value),
	}
	s.ImageURL = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Pageview) SetProductExternalID(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "product_external_id"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.ProductExternalID != nil && s.ProductExternalID.IsNull == value.IsNull && s.ProductExternalID.String == value.String {
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
		s.ProductExternalID = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.ProductExternalID),
		NewValue:  NullableStringToInterface(value),
	}
	s.ProductExternalID = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Pageview) SetProductSKU(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "product_sku"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.ProductSKU != nil && s.ProductSKU.IsNull == value.IsNull && s.ProductSKU.String == value.String {
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
		s.ProductSKU = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.ProductSKU),
		NewValue:  NullableStringToInterface(value),
	}
	s.ProductSKU = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Pageview) SetProductName(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "product_name"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.ProductName != nil && s.ProductName.IsNull == value.IsNull && s.ProductName.String == value.String {
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
		s.ProductName = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.ProductName),
		NewValue:  NullableStringToInterface(value),
	}
	s.ProductName = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Pageview) SetProductBrand(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "product_brand"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.ProductBrand != nil && s.ProductBrand.IsNull == value.IsNull && s.ProductBrand.String == value.String {
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
		s.ProductBrand = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.ProductBrand),
		NewValue:  NullableStringToInterface(value),
	}
	s.ProductBrand = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Pageview) SetProductCategory(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "product_category"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.ProductCategory != nil && s.ProductCategory.IsNull == value.IsNull && s.ProductCategory.String == value.String {
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
		s.ProductCategory = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.ProductCategory),
		NewValue:  NullableStringToInterface(value),
	}
	s.ProductCategory = value
	s.FieldsTimestamp[key] = timestamp
	return
}

func (s *Pageview) SetProductVariantExternalID(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "product_variant_external_id"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.ProductVariantExternalID != nil && s.ProductVariantExternalID.IsNull == value.IsNull && s.ProductVariantExternalID.String == value.String {
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
		s.ProductVariantExternalID = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.ProductVariantExternalID),
		NewValue:  NullableStringToInterface(value),
	}
	s.ProductVariantExternalID = value
	s.FieldsTimestamp[key] = timestamp
	return
}
func (s *Pageview) SetProductVariantTitle(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "product_variant_title"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.ProductVariantTitle != nil && s.ProductVariantTitle.IsNull == value.IsNull && s.ProductVariantTitle.String == value.String {
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
		s.ProductVariantTitle = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.ProductVariantTitle),
		NewValue:  NullableStringToInterface(value),
	}
	s.ProductVariantTitle = value
	s.FieldsTimestamp[key] = timestamp
	return
}

func (s *Pageview) SetProductPrice(value *NullableInt64, timestamp time.Time) (update *UpdatedField) {
	key := "product_price"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.ProductPrice != nil && s.ProductPrice.IsNull == value.IsNull && s.ProductPrice.Int64 == value.Int64 {
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
		s.ProductPrice = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableInt64ToInterface(s.ProductPrice),
		NewValue:  NullableInt64ToInterface(value),
	}
	s.ProductPrice = value
	s.FieldsTimestamp[key] = timestamp
	return
}

func (s *Pageview) SetProductCurrency(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "product_currency"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && s.ProductCurrency != nil && s.ProductCurrency.IsNull == value.IsNull && s.ProductCurrency.String == value.String {
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
		s.ProductCurrency = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(s.ProductCurrency),
		NewValue:  NullableStringToInterface(value),
	}
	s.ProductCurrency = value
	s.FieldsTimestamp[key] = timestamp
	return
}

func (o *Pageview) SetProductFxRate(value *float64, timestamp time.Time) (update *UpdatedField) {
	key := "product_fx_rate"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if value != nil && o.ProductFxRate != nil && o.ProductFxRate == value {
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
		o.ProductFxRate = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: Float64PointerToInterface(o.ProductFxRate),
		NewValue:  Float64PointerToInterface(value),
	}
	o.ProductFxRate = value
	o.FieldsTimestamp[key] = timestamp
	return
}

// merges two pageviews and returns the list of updated fields
func (fromPageview *Pageview) MergeInto(toPageview *Pageview, workspace *Workspace) (updatedFields []*UpdatedField) {
	updatedFields = []*UpdatedField{} // init

	if toPageview.FieldsTimestamp == nil {
		toPageview.FieldsTimestamp = FieldsTimestamp{}
	}

	if fieldUpdate := toPageview.SetPageID(fromPageview.PageID, fromPageview.GetFieldDate("page_id")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPageview.SetTitle(fromPageview.Title, fromPageview.GetFieldDate("title")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPageview.SetReferrer(fromPageview.Referrer, fromPageview.GetFieldDate("referrer")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPageview.SetReferrerDomain(fromPageview.ReferrerDomain, fromPageview.GetFieldDate("referrer_domain")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPageview.SetReferrerPath(fromPageview.ReferrerPath, fromPageview.GetFieldDate("referrer_path")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPageview.SetDuration(fromPageview.Duration, fromPageview.GetFieldDate("duration")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPageview.SetImageURL(fromPageview.ImageURL, fromPageview.GetFieldDate("image_url")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPageview.SetProductExternalID(fromPageview.ProductExternalID, fromPageview.GetFieldDate("product_external_id")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPageview.SetProductSKU(fromPageview.ProductSKU, fromPageview.GetFieldDate("product_sku")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPageview.SetProductName(fromPageview.ProductName, fromPageview.GetFieldDate("product_name")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPageview.SetProductBrand(fromPageview.ProductBrand, fromPageview.GetFieldDate("product_brand")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPageview.SetProductCategory(fromPageview.ProductCategory, fromPageview.GetFieldDate("product_category")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPageview.SetProductVariantExternalID(fromPageview.ProductVariantExternalID, fromPageview.GetFieldDate("product_variant_external_id")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPageview.SetProductVariantTitle(fromPageview.ProductVariantTitle, fromPageview.GetFieldDate("product_variant_title")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPageview.SetProductPrice(fromPageview.ProductPrice, fromPageview.GetFieldDate("product_price")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toPageview.SetProductCurrency(fromPageview.ProductCurrency, fromPageview.GetFieldDate("product_original_currency")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}

	for key, value := range fromPageview.ExtraColumns {
		if fieldUpdate := toPageview.SetExtraColumns(key, value, TimePtr(fromPageview.GetFieldDate(key))); fieldUpdate != nil {
			updatedFields = append(updatedFields, fieldUpdate)
		}
	}

	// UpdatedAt is the timeOfEvent for RTs
	toPageview.UpdatedAt = fromPageview.UpdatedAt
	// priority to oldest date
	toPageview.SetCreatedAt(fromPageview.CreatedAt)

	return
}

func ComputePageviewID(externalID string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(externalID)))
}

func NewPageview(externalID string, userID string, domainID string, sessionExternalID string, createdAt time.Time, updatedAt *time.Time, pageID string, title string) *Pageview {
	return &Pageview{
		ID:              ComputePageviewID(externalID),
		ExternalID:      externalID,
		UserID:          userID,
		DomainID:        domainID,
		SessionID:       ComputeSessionID(sessionExternalID),
		CreatedAt:       createdAt,
		CreatedAtTrunc:  createdAt.Truncate(time.Hour),
		FieldsTimestamp: FieldsTimestamp{},

		PageID: pageID,
		Title:  title,

		UpdatedAt:    updatedAt,
		ExtraColumns: AppItemFields{},
	}
}

func NewPageviewFromDataLog(dataLog *DataLog, clockDifference time.Duration, workspace *Workspace) (pageview *Pageview, err error) {

	result := gjson.Get(dataLog.Item, "pageview")
	if !result.Exists() {
		return nil, eris.New("item has no pageview object")
	}

	if result.Type != gjson.JSON {
		return nil, eris.New("item pageview is not an object")
	}

	extraColumns := workspace.FindExtraColumnsForItemKind("pageview")

	// init
	pageview = &Pageview{
		UserID:          dataLog.UserID,
		FieldsTimestamp: FieldsTimestamp{},
		ExtraColumns:    AppItemFields{},
	}

	// loop over pageview fields
	result.ForEach(func(key, value gjson.Result) bool {

		keyString := key.String()
		switch keyString {

		case "external_id":
			if value.Type == gjson.Null {
				err = eris.New("pageview.external_id is required")
				return false
			}
			pageview.ExternalID = value.String()
			pageview.ID = ComputeCartID(pageview.ExternalID)

		case "domain_id":
			if value.Type == gjson.Null {
				err = eris.New("pageview.domain_id is required")
				return false
			}
			domain := value.String()
			pageview.DomainID = domain

		case "session_external_id":
			if value.Type != gjson.Null {
				pageview.SessionID = ComputeSessionID(value.String())
			}

		case "created_at":
			if pageview.CreatedAt, err = time.Parse(time.RFC3339, value.String()); err != nil {
				err = eris.Wrap(err, "pageview.created_at")
				return false
			}

			// apply clock difference
			if pageview.CreatedAt.After(time.Now()) {

				pageview.CreatedAt = pageview.CreatedAt.Add(clockDifference)
				if pageview.CreatedAt.After(time.Now()) {
					err = eris.New("pageview.created_at cannot be in the future")
					return false
				}
			}

			pageview.CreatedAtTrunc = pageview.CreatedAt.Truncate(time.Hour)

		case "updated_at":
			var updatedAt time.Time
			if updatedAt, err = time.Parse(time.RFC3339, value.String()); err != nil {
				err = eris.Wrap(err, "pageview.updated_at")
				return false
			}

			if updatedAt.After(time.Now()) {
				// apply clock difference
				updatedAtModified := updatedAt.Add(clockDifference)
				updatedAt = updatedAtModified
				if updatedAt.After(time.Now()) {
					err = eris.New("pageview.updated_at cannot be in the future")
					return false
				}
			}

			pageview.UpdatedAt = &updatedAt

		case "page_id":
			if value.Type == gjson.Null {
				err = eris.New("pageview.page_id is required")
				return false
			}

			pageview.PageID = strings.TrimSpace(value.String())

			// clean pageID URL
			if govalidator.IsRequestURL(pageview.PageID) {
				u, urlError := url.Parse(pageview.PageID)
				allowedValues := url.Values{}
				// extract allowed values
				if urlError == nil {
					for param, val := range u.Query() {
						for _, domain := range workspace.Domains {
							if domain.DeletedAt == nil {
								for _, p := range domain.ParamsWhitelist {
									if param == p {
										for _, v := range val {
											allowedValues.Add(param, v)
										}
									}
								}
							}
						}
					}
					// reconstruct URL
					newUrl := url.URL{
						Scheme:   u.Scheme,
						Host:     u.Host,
						Path:     u.Path,
						RawQuery: allowedValues.Encode(),
					}
					pageview.PageID = newUrl.String()
				}
			}

			// truncate pageId if it is above 512 characters
			if len(pageview.PageID) > 512 {
				pageview.PageID = pageview.PageID[:512]
			}

		case "title":
			if value.Type == gjson.Null {
				err = eris.New("pageview.title is required")
				return false
			}

			pageview.Title = strings.TrimSpace(value.String())

		case "referrer":
			if value.Type == gjson.Null {
				pageview.Referrer = NewNullableString(nil)
			} else {
				referrer := strings.TrimSpace(value.String())
				pageview.Referrer = NewNullableString(&referrer)

				if !govalidator.IsRequestURL(pageview.Referrer.String) {
					err = eris.New("pageview.referrer is not a valid URL")
					return false
				}

				u, err := url.Parse(referrer)
				if err != nil {
					err = eris.Wrap(err, "pageview.referrer")
				}

				pageview.ReferrerDomain = &u.Host

				// get path
				path := u.Path
				if path == "" {
					path = "/"
				}
				pageview.ReferrerPath = &path
			}

		case "image_url":
			if value.Type == gjson.Null {
				pageview.ImageURL = NewNullableString(nil)
			} else {
				imageURL := strings.TrimSpace(value.String())
				pageview.ImageURL = NewNullableString(&imageURL)

				if !govalidator.IsRequestURL(pageview.ImageURL.String) {
					err = eris.New("pageview.image_url is not a valid URL")
					return false
				}
			}

		case "duration":
			if value.Type == gjson.Null {
				pageview.Duration = nil
			} else {
				duration := value.Int()
				pageview.Duration = &duration

				if duration < 0 {
					err = eris.New("pageview.duration cannot be negative")
					return false
				}

				// set value to 10800 secs (60x60x3 hours) if value is above
				if pageview.Duration != nil && *pageview.Duration > 10800 {
					pageview.Duration = Int64Ptr(10800)
				}
			}

		case "product_external_id":
			if value.Type == gjson.Null {
				pageview.ProductExternalID = NewNullableString(nil)
			} else {
				productExternalID := strings.TrimSpace(value.String())
				pageview.ProductExternalID = NewNullableString(&productExternalID)
			}

		case "product_sku":
			if value.Type == gjson.Null {
				pageview.ProductSKU = NewNullableString(nil)
			} else {
				productSKU := strings.TrimSpace(value.String())
				pageview.ProductSKU = NewNullableString(&productSKU)
			}

		case "product_name":
			if value.Type == gjson.Null {
				pageview.ProductName = NewNullableString(nil)
			} else {
				productName := strings.TrimSpace(value.String())
				pageview.ProductName = NewNullableString(&productName)
			}

		case "product_brand":
			if value.Type == gjson.Null {
				pageview.ProductBrand = NewNullableString(nil)
			} else {
				productBrand := strings.TrimSpace(value.String())
				pageview.ProductBrand = NewNullableString(&productBrand)
			}

		case "product_category":
			if value.Type == gjson.Null {
				pageview.ProductCategory = NewNullableString(nil)
			} else {
				productCategory := strings.TrimSpace(value.String())
				pageview.ProductCategory = NewNullableString(&productCategory)
			}

		case "product_variant_external_id":
			if value.Type == gjson.Null {
				pageview.ProductVariantExternalID = NewNullableString(nil)
			} else {
				productVariantExternalID := strings.TrimSpace(value.String())
				pageview.ProductVariantExternalID = NewNullableString(&productVariantExternalID)
			}

		case "product_variant_title":
			if value.Type == gjson.Null {
				pageview.ProductVariantTitle = NewNullableString(nil)
			} else {
				productVariantTitle := strings.TrimSpace(value.String())
				pageview.ProductVariantTitle = NewNullableString(&productVariantTitle)
			}

		case "product_price":
			if value.Type == gjson.Null {
				pageview.ProductPrice = NewNullableInt64(nil)
			} else {
				productPrice := value.Int()
				pageview.ProductPrice = NewNullableInt64(&productPrice)
			}

		case "product_currency":
			if value.Type == gjson.Null {
				pageview.ProductCurrency = NewNullableString(nil)
			} else {
				productCurrency := strings.TrimSpace(value.String())
				pageview.ProductCurrency = NewNullableString(&productCurrency)

				if !govalidator.IsIn(productCurrency, common.CurrenciesCodes...) {
					err = eris.New("pageview.product_currency invalid")
					return false
				}
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
						pageview.ExtraColumns[col.Name] = fieldValue
					}
				}
			}
		}

		return true
	})

	if err != nil {
		return nil, err
	}
	if pageview.DomainID == "" && dataLog.DomainID != nil {
		pageview.DomainID = *dataLog.DomainID
	}

	// use data import createdAt as updatedAt if not provided
	if pageview.UpdatedAt == nil {
		pageview.UpdatedAt = &pageview.CreatedAt
	}

	// enrich pageview with session and domain
	if dataLog.UpsertedSession != nil {
		if pageview.SessionID == "" {
			pageview.SessionID = dataLog.UpsertedSession.ID
		}
		if pageview.DomainID == "" {
			pageview.DomainID = dataLog.UpsertedSession.DomainID
		}
	}

	// Validation
	if pageview.ExternalID == "" {
		return nil, eris.New("pageview.external_id is required")
	}

	if pageview.PageID == "" {
		return nil, eris.New("pageview.page_id is required")
	}

	if pageview.Title == "" {
		return nil, eris.New("pageview.title is required")
	}

	// verify that domainID exists
	found := false
	for _, domain := range workspace.Domains {
		if domain.ID == pageview.DomainID {
			found = true
			break
		}
	}

	if !found {
		return nil, eris.New("pageview domain_id invalid")
	}

	if pageview.CreatedAt.IsZero() {
		return nil, eris.New("pageview.created_at is required")
	}

	// convert price
	if pageview.ProductPrice != nil && !pageview.ProductPrice.IsNull && (pageview.ProductCurrency == nil || (pageview.ProductCurrency.IsNull || pageview.ProductCurrency.String == "")) {

		pageview.ProductCurrency = NewNullableString(&workspace.Currency)

		// set fx_rate if needed
		if pageview.ProductCurrency != nil && pageview.ProductCurrency.String != "" && pageview.ProductCurrency.String != workspace.Currency {
			if fxRate, err := workspace.GetFxRateForCurrency(pageview.ProductCurrency.String); err == nil {
				pageview.ProductFxRate = &fxRate
			} else {
				err = eris.Wrapf(err, "pageview.product_currency %s", pageview.ProductCurrency.String)
				return nil, err
			}
		}
	}

	return pageview, nil
}

// returns a URL without non-allowed parameters
func ParseAndCleanURL(value string, allowedParameters []string) (newUrl url.URL, err error) {

	u, parseErr := url.Parse(value)
	if parseErr != nil {
		return newUrl, parseErr
	}

	allowedValues := url.Values{}

	// extract allowed values
	for param, val := range u.Query() {
		for _, p := range allowedParameters {
			if param == p {
				for _, v := range val {
					allowedValues.Add(param, v)
				}
			}
		}
	}

	// reconstruct URL
	newUrl = url.URL{
		Scheme:   u.Scheme,
		Host:     u.Host,
		Path:     u.Path,
		RawQuery: allowedValues.Encode(),
	}

	return
}

var PageviewSchema string = `CREATE TABLE IF NOT EXISTS pageview (
  id VARCHAR(64) NOT NULL,
  external_id VARCHAR(255) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  domain_id VARCHAR(64) NOT NULL,
  session_id VARCHAR(64) NOT NULL,
  created_at DATETIME NOT NULL,
  created_at_trunc AS DATE_TRUNC('hour', created_at) PERSISTED DATETIME,
  db_created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  is_deleted BOOLEAN DEFAULT FALSE,
  merged_from_user_id VARCHAR(64),
  fields_timestamp JSON NOT NULL,

  page_id VARCHAR(512) CHARACTER SET utf8 COLLATE utf8_bin,
  title VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin,
  referrer VARCHAR(2083),
  referrer_domain VARCHAR(255),
  referrer_path VARCHAR(255),
  duration SMALLINT UNSIGNED DEFAULT 0,
  image_url VARCHAR(2083),

  product_external_id VARCHAR(255),
  product_sku VARCHAR(255),
  product_name VARCHAR(255),
  product_brand VARCHAR(255),
  product_category VARCHAR(255),
  product_variant_external_id VARCHAR(255),
  product_variant_title VARCHAR(255),
  product_price INT DEFAULT 0,
  product_currency VARCHAR(3),
  product_fx_rate FLOAT DEFAULT 1,
  product_converted_price AS FLOOR(product_price*product_fx_rate) PERSISTED INT,

  SORT KEY (created_at_trunc DESC),
  PRIMARY KEY (id, user_id),
  KEY (session_id) USING HASH,
  KEY (page_id) USING HASH,
  KEY (user_id) USING HASH, -- for merging
  SHARD KEY (user_id)
) COLLATE utf8mb4_general_ci;`

var PageviewSchemaMYSQL string = `CREATE TABLE IF NOT EXISTS pageview (
  id VARCHAR(64) NOT NULL,
  external_id VARCHAR(255) NOT NULL,
  user_id VARCHAR(64) NOT NULL,
  domain_id VARCHAR(64) NOT NULL,
  session_id VARCHAR(64) NOT NULL,
  created_at DATETIME NOT NULL,
  created_at_trunc DATETIME GENERATED ALWAYS AS (DATE_FORMAT(created_at, '%Y-%m-%d %H:00:00')) STORED,
  db_created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  is_deleted BOOLEAN DEFAULT FALSE,
  merged_from_user_id VARCHAR(64),
  fields_timestamp JSON NOT NULL,

  page_id VARCHAR(512) CHARACTER SET utf8 COLLATE utf8_bin,
  title VARCHAR(255) CHARACTER SET utf8 COLLATE utf8_bin,
  referrer VARCHAR(2083),
  referrer_domain VARCHAR(255),
  referrer_path VARCHAR(255),
  duration SMALLINT UNSIGNED DEFAULT 0,
  image_url VARCHAR(2083),

  product_external_id VARCHAR(255),
  product_sku VARCHAR(255),
  product_name VARCHAR(255),
  product_brand VARCHAR(255),
  product_category VARCHAR(255),
  product_variant_external_id VARCHAR(255),
  product_variant_title VARCHAR(255),
  product_price INT DEFAULT 0,
  product_currency VARCHAR(3),
  product_fx_rate FLOAT DEFAULT 1,
  product_converted_price INT GENERATED ALWAYS AS (FLOOR(product_price*product_fx_rate)) STORED,

  -- SORT KEY (created_at_trunc),
  PRIMARY KEY (id, user_id),
  KEY (session_id) USING HASH,
  KEY (page_id) USING HASH,
  KEY (user_id) USING HASH -- for merging
  -- SHARD KEY (user_id)
) COLLATE utf8mb4_general_ci;`

func NewPageviewCube() *CubeJSSchema {
	return &CubeJSSchema{
		Title:       "Pageviews",
		Description: "Pageviews",
		SQL:         "SELECT * FROM `pageview`",
		// https://cube.dev/docs/schema/reference/joins
		Joins: map[string]CubeJSSchemaJoin{
			"User": {
				Relationship: "many_to_one",
				SQL:          "${CUBE}.user_id = ${User}.id",
			},
			"Session": {
				Relationship: "many_to_one",
				SQL:          "${CUBE}.session_id = ${Session}.id",
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
			"unique_sessions": {
				Type:        "countDistinct",
				SQL:         "session_id",
				Title:       "Unique sessions",
				Description: "Count distinct session_id",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"unique_pages": {
				Type:        "countDistinct",
				SQL:         "page_id",
				Title:       "Unique pages",
				Description: "Count distinct page_id (=page URL)",
			},
			"unique_referrer_domains": {
				Type:        "countDistinct",
				SQL:         "referrer_domain",
				Title:       "Unique referrer domains",
				Description: "Count distinct referrer_domain",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"pageviews_per_user": {
				Type:        "number",
				SQL:         "${count} / ${unique_users}",
				Title:       "Pageviews per user",
				Description: "count / unique_users",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"pageviews_per_session": {
				Type:        "number",
				SQL:         "${count} / ${unique_sessions}",
				Title:       "Pageviews per session",
				Description: "count / unique_sessions",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"avg_duration": {
				Type:        "number",
				Title:       "Average duration",
				Description: "AVG(duration)",
				SQL:         "COALESCE(AVG(duration), 0)",
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
				Title:       "Pageview ID",
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
			"page_id": {
				SQL:         "page_id",
				Type:        "string",
				Title:       "Page ID",
				Description: "field: page_id (=page URL)",
			},
			"title": {
				SQL:         "title",
				Type:        "string",
				Title:       "Title",
				Description: "field: title",
			},
			"referrer": {
				SQL:         "referrer",
				Type:        "string",
				Title:       "Referrer",
				Description: "field: referrer (=referrer URL)",
			},
			"referrer_domain": {
				SQL:         "referrer_domain",
				Type:        "string",
				Title:       "Referrer domain",
				Description: "field: referrer_domain",
			},
			"referrer_path": {
				SQL:         "referrer_path",
				Type:        "string",
				Title:       "Referrer path",
				Description: "field: referrer_path",
			},
			"duration": {
				SQL:         "duration",
				Type:        "number",
				Title:       "Duration",
				Description: "field: duration",
			},
			"image_url": {
				SQL:         "image_url",
				Type:        "string",
				Title:       "Image URL",
				Description: "field: image_url",
			},
			"product_external_id": {
				SQL:         "product_external_id",
				Type:        "string",
				Title:       "Product external ID",
				Description: "field: product_external_id",
			},
			"product_sku": {
				SQL:         "product_sku",
				Type:        "string",
				Title:       "Product SKU",
				Description: "field: product_sku",
			},
			"product_name": {
				SQL:         "product_name",
				Type:        "string",
				Title:       "Product name",
				Description: "field: product_name",
			},
			"product_brand": {
				SQL:         "product_brand",
				Type:        "string",
				Title:       "Product brand",
				Description: "field: product_brand",
			},
			"product_category": {
				SQL:         "product_category",
				Type:        "string",
				Title:       "Product category",
				Description: "field: product_category",
			},
			"product_variant_external_id": {
				SQL:         "product_variant_external_id",
				Type:        "string",
				Title:       "Product variant external ID",
				Description: "field: product_variant_external_id",
			},
			"product_variant_title": {
				SQL:         "product_variant_title",
				Type:        "string",
				Title:       "Product variant title",
				Description: "field: product_variant_title",
			},
			"product_price": {
				SQL:         "product_price",
				Type:        "number",
				Title:       "Product price",
				Description: "field: product_price",
			},
			"product_currency": {
				SQL:         "product_currency",
				Type:        "string",
				Title:       "Product currency",
				Description: "field: product_currency",
			},
			"product_fx_rate": {
				SQL:         "product_fx_rate",
				Type:        "number",
				Title:       "Product FX rate",
				Description: "field: product_fx_rate",
			},
			"product_converted_price": {
				SQL:         "product_converted_price",
				Type:        "number",
				Title:       "Product converted price",
				Description: "field: product_converted_price",
			},
		},
	}
}
