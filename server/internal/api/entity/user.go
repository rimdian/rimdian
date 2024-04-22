package entity

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/rimdian/rimdian/internal/api/common"
	"github.com/rimdian/rimdian/internal/common/dto"
	"github.com/rotisserie/eris"
	"github.com/tidwall/gjson"
	"github.com/ttacon/libphonenumber"
)

var (
	Male   = "male"
	Female = "female"

	NativeReconciliationKeys = []string{
		"email",
		"email_md5",
		"email_sha1",
		"email_sha256",
		"telephone",
	}

	DefaultUserReconciliationKeys = []string{
		"email",
		"email_md5",
		"email_sha1",
		"email_sha256",
	}

	// computed fields should be excluded from SELECT/INSERT while cloning rows
	UserComputedFields = []string{
		"created_at_trunc",
		"geo",
		"email_md5",
		"email_sha1",
		"email_sha256",
		"cart_items_count",
		"cart_updated_at",
		"wishlist_items_count",
		"wishlist_updated_at",
	}
)

func ComputeUserID(externalID string) string {
	return fmt.Sprintf("%x", sha1.Sum([]byte(externalID)))
}

func NewUser(externalID string, isAuthenticated bool, createdAt time.Time, updatedAt time.Time, defaultTimezone string, defaultLanguage string, defaultCountry string, signedUpAt *time.Time) *User {
	user := &User{
		// user ID is a SHA1(externalID)
		ID:                ComputeUserID(externalID),
		IsAuthenticated:   isAuthenticated,
		ExternalID:        externalID,
		CreatedAt:         createdAt,
		UpdatedAt:         &updatedAt,
		Timezone:          &defaultTimezone,
		Language:          &defaultLanguage,
		Country:           &defaultCountry,
		LastInteractionAt: createdAt,
		FieldsTimestamp:   FieldsTimestamp{},
		// CustomColumns:     MapOfInterfaces{},
	}

	// signed_up_at is mandatory for authenticated users
	// it is a mergeable field too
	if user.IsAuthenticated {

		// use user created_at at default signed_up_at if missing
		if signedUpAt == nil || signedUpAt.IsZero() {
			user.SignedUpAt = &createdAt
		} else {
			user.SignedUpAt = signedUpAt
		}
	}

	if isAuthenticated && (signedUpAt == nil || signedUpAt.IsZero()) {
		user.SignedUpAt = &createdAt
	}

	return user
}

func NewUserFromDataLog(dataLog *DataLog, clockDifference time.Duration, workspace *Workspace, origin int, secretKey string) (user *User, err error) {

	result := gjson.Get(dataLog.Item, "user")
	if !result.Exists() {
		return nil, eris.New("item has no user object")
	}

	if result.Type != gjson.JSON {
		return nil, eris.New("item user is not an object")
	}

	extraColumns := workspace.FindExtraColumnsForItemKind("user")

	// init
	user = &User{
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

			user.ExternalID = value.String()
			user.ID = ComputeSessionID(user.ExternalID)

			// check wrong implementation of dynamic user_id
			userIDlower := strings.ToLower(user.ExternalID)
			forbiddenUserIDs := []string{
				"id",
				"userid",
				"user_id",
				"user_id()",
				"{{user_id}}",
				"{{userid}}",
				"{{ user_id }}",
				"{{ userid }}",
				"{{id}}",
				"{{ id }}",
				"$user_id",
				"$userid",
			}

			if govalidator.IsIn(userIDlower, forbiddenUserIDs...) {
				err = eris.Errorf("user.external_id %s is forbidden", user.ExternalID)
				return false
			}

		case "hmac":
			user.HMAC = StringPtr(value.String())

		case "created_at":
			if user.CreatedAt, err = time.Parse(time.RFC3339, value.String()); err != nil {
				err = eris.Wrap(err, "user.created_at")
				return false
			}

			// apply clock difference
			if user.CreatedAt.After(time.Now()) {

				user.CreatedAt = user.CreatedAt.Add(clockDifference)
				if user.CreatedAt.After(time.Now()) {
					err = eris.New("user.created_at cannot be in the future")
					return false
				}
			}

			user.CreatedAtTrunc = user.CreatedAt.Truncate(time.Hour)

		case "updated_at":
			var updatedAt time.Time
			if updatedAt, err = time.Parse(time.RFC3339, value.String()); err != nil {
				err = eris.Wrap(err, "user.updated_at")
				return false
			}

			if updatedAt.After(time.Now()) {
				// apply clock difference
				updatedAtModified := updatedAt.Add(clockDifference)
				updatedAt = updatedAtModified
				if updatedAt.After(time.Now()) {
					err = eris.New("user.updated_at cannot be in the future")
					return false
				}
			}

			user.UpdatedAt = &updatedAt

		case "timezone":
			if value.Type == gjson.Null {
				user.Timezone = &workspace.DefaultUserTimezone
			} else {
				tz := value.String()
				user.Timezone = &tz

				if !govalidator.IsIn(*user.Timezone, common.Timezones...) {
					err = eris.Errorf("user.timezone %s is invalid", *user.Timezone)
					return false
				}
			}

		case "language":
			if value.Type == gjson.Null {
				user.Language = &workspace.DefaultUserLanguage
			} else {
				language := value.String()
				user.Language = &language

				if !govalidator.IsIn(*user.Language, common.LanguageCodes...) {
					err = eris.Errorf("user.language %s is invalid", *user.Language)
					return false
				}
			}

		case "country":
			if value.Type == gjson.Null {
				user.Country = &workspace.DefaultUserCountry
			} else {
				country := value.String()
				user.Country = &country

				if !govalidator.IsIn(*user.Country, common.CountriesCodes...) {
					err = eris.Errorf("user.country %s is invalid", *user.Country)
					return false
				}
			}

		case "last_interaction_at":
			var lastInteractionAt time.Time
			if lastInteractionAt, err = time.Parse(time.RFC3339, value.String()); err != nil {
				err = eris.Wrap(err, "user.last_interaction_at")
				return false
			}

			lastInteractionAt = lastInteractionAt.Truncate(time.Second)

			if lastInteractionAt.After(time.Now()) {
				// apply clock difference
				lastInteractionAtModified := lastInteractionAt.Add(clockDifference)
				lastInteractionAt = lastInteractionAtModified
				if lastInteractionAt.After(time.Now()) {
					err = eris.New("user.last_interaction_at cannot be in the future")
					return false
				}
			}

			user.LastInteractionAt = lastInteractionAt

		case "is_authenticated":
			if value.Type == gjson.Null {
				err = eris.New("user.is_authenticated is required")
				return false
			} else {
				user.IsAuthenticated = value.Bool()
			}

		case "signed_up_at":
			if value.Type != gjson.Null {
				var signedUpAt time.Time
				if signedUpAt, err = time.Parse(time.RFC3339, value.String()); err != nil {
					err = eris.Wrap(err, "user.signed_up_at")
					return false
				}

				if signedUpAt.After(time.Now()) {
					// apply clock difference
					signedUpAtModified := signedUpAt.Add(clockDifference)
					signedUpAt = signedUpAtModified
					if signedUpAt.After(time.Now()) {
						err = eris.New("user.signed_up_at cannot be in the future")
						return false
					}
				}

				user.SignedUpAt = &signedUpAt
			}

		// user_centric_consent legacy of consent_all
		case "user_centric_consent":
			if value.Type == gjson.Null {
				user.ConsentAll = NewNullableBool(nil)
			} else {
				consentAll := value.Bool()
				user.ConsentAll = NewNullableBool(&consentAll)
			}

		case "consent_all":
			if value.Type == gjson.Null {
				user.ConsentAll = NewNullableBool(nil)
			} else {
				consentAll := value.Bool()
				user.ConsentAll = NewNullableBool(&consentAll)
			}

		case "consent_personalization":
			if value.Type == gjson.Null {
				user.ConsentPersonalization = NewNullableBool(nil)
			} else {
				consentPersonalization := value.Bool()
				user.ConsentPersonalization = NewNullableBool(&consentPersonalization)
			}

		case "consent_marketing":
			if value.Type == gjson.Null {
				user.ConsentMarketing = NewNullableBool(nil)
			} else {
				consentMarketing := value.Bool()
				user.ConsentMarketing = NewNullableBool(&consentMarketing)
			}

		case "last_ip":
			if value.Type == gjson.Null {
				user.LastIP = NewNullableString(nil)
			} else {
				lastIP := value.String()
				user.LastIP = NewNullableString(&lastIP)
			}

		case "longitude":
			if value.Type == gjson.Null {
				user.Longitude = NewNullableFloat64(nil)
				user.Latitude = NewNullableFloat64(nil)
			} else {
				longitude := value.Float()
				user.Longitude = NewNullableFloat64(&longitude)
			}

		case "latitude":
			if value.Type == gjson.Null {
				user.Longitude = NewNullableFloat64(nil)
				user.Latitude = NewNullableFloat64(nil)
			} else {
				latitude := value.Float()
				user.Latitude = NewNullableFloat64(&latitude)
			}

		case "first_name":
			if value.Type == gjson.Null {
				user.FirstName = NewNullableString(nil)
			} else {
				firstName := value.String()
				user.FirstName = NewNullableString(&firstName)
			}

		case "last_name":
			if value.Type == gjson.Null {
				user.LastName = NewNullableString(nil)
			} else {
				lastName := value.String()
				user.LastName = NewNullableString(&lastName)
			}

		case "gender":
			if value.Type == gjson.Null {
				user.Gender = NewNullableString(nil)
			} else {
				gender := value.String()
				user.Gender = NewNullableString(&gender)

				if gender != Male && gender != Female {
					err = eris.New("gender should be male or female")
				}
			}

		case "birthday":
			if value.Type == gjson.Null {
				user.Birthday = NewNullableString(nil)
			} else {
				birthday := value.String()

				date, errParse := time.Parse("2006-01-02", birthday)
				if errParse != nil {
					err = eris.Wrap(errParse, "user.birthday")
					return false
				}

				birthday = date.Format("2006-01-02")
				user.Birthday = NewNullableString(&birthday)
			}

		case "photo_url":
			if value.Type == gjson.Null {
				user.PhotoURL = NewNullableString(nil)
			} else {
				photoURL := value.String()
				user.PhotoURL = NewNullableString(&photoURL)

				if !govalidator.IsURL(photoURL) {
					err = eris.New("user.photo_url is not a valid URL")
					return false
				}
			}

		case "email":
			if value.Type == gjson.Null {
				user.Email = NewNullableString(nil)
			} else {
				email := value.String()
				user.Email = NewNullableString(&email)

				if !govalidator.IsEmail(email) {
					err = eris.New("user.email is not a valid email")
					return false
				}

				emailMD5 := fmt.Sprintf("%x", md5.Sum([]byte(value.String())))
				user.EmailMD5 = NewNullableString(&emailMD5)

				emailSHA1 := fmt.Sprintf("%x", sha1.Sum([]byte(value.String())))
				user.EmailSHA1 = NewNullableString(&emailSHA1)

				emailSHA256 := fmt.Sprintf("%x", sha256.Sum256([]byte(value.String())))
				user.EmailSHA256 = NewNullableString(&emailSHA256)
			}

		case "email_md5":
			if value.Type == gjson.Null {
				user.EmailMD5 = NewNullableString(nil)
			} else {
				emailMD5 := value.String()
				user.EmailMD5 = NewNullableString(&emailMD5)

				if !govalidator.IsMD5(emailMD5) {
					err = eris.New("user.email_md5 is not a valid MD5")
					return false
				}
			}

		case "email_sha1":
			if value.Type == gjson.Null {
				user.EmailSHA1 = NewNullableString(nil)
			} else {
				emailSHA1 := value.String()
				user.EmailSHA1 = NewNullableString(&emailSHA1)

				if !govalidator.IsHash(emailSHA1, "sha1") {
					err = eris.New("user.email_sha1 is not a valid SHA1")
					return false
				}
			}

		case "email_sha256":
			if value.Type == gjson.Null {
				user.EmailSHA256 = NewNullableString(nil)
			} else {
				emailSHA256 := value.String()
				user.EmailSHA256 = NewNullableString(&emailSHA256)

				if !govalidator.IsHash(emailSHA256, "sha256") {
					err = eris.New("user.email_sha256 is not a valid SHA256")
					return false
				}
			}

		case "telephone":
			if value.Type == gjson.Null {
				user.Telephone = NewNullableString(nil)
			} else {
				telephone := value.String()
				user.Telephone = NewNullableString(&telephone)

				country := workspace.DefaultUserCountry

				if user.Country != nil {
					country = *user.Country
				} else {
					// try to extract country from json
					countryValue := gjson.Get(dataLog.Item, "user.country")

					if countryValue.Exists() && countryValue.Type == gjson.String && govalidator.IsISO3166Alpha2(countryValue.String()) {
						country = countryValue.String()
					}
				}

				num, errParse := libphonenumber.Parse(telephone, country)

				if errParse == nil {
					formattedPhone := libphonenumber.Format(num, libphonenumber.E164)
					user.Telephone = NewNullableString(&formattedPhone)
				} else {
					err = eris.Wrap(errParse, "user.telephone")
					return false
				}

			}

		case "address_line_1":
			if value.Type == gjson.Null {
				user.AddressLine1 = NewNullableString(nil)
			} else {
				addressLine1 := value.String()
				user.AddressLine1 = NewNullableString(&addressLine1)
			}

		case "address_line_2":
			if value.Type == gjson.Null {
				user.AddressLine2 = NewNullableString(nil)
			} else {
				addressLine2 := value.String()
				user.AddressLine2 = NewNullableString(&addressLine2)
			}

		case "city":
			if value.Type == gjson.Null {
				user.City = NewNullableString(nil)
			} else {
				city := value.String()
				user.City = NewNullableString(&city)
			}

		case "region":
			if value.Type == gjson.Null {
				user.Region = NewNullableString(nil)
			} else {
				region := value.String()
				user.Region = NewNullableString(&region)
			}

		case "postal_code":
			if value.Type == gjson.Null {
				user.PostalCode = NewNullableString(nil)
			} else {
				postalCode := value.String()
				user.PostalCode = NewNullableString(&postalCode)
			}

		case "state":
			if value.Type == gjson.Null {
				user.State = NewNullableString(nil)
			} else {
				state := value.String()
				user.State = NewNullableString(&state)
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
						user.ExtraColumns[col.Name] = fieldValue
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
	if user.UpdatedAt == nil {
		user.UpdatedAt = &user.CreatedAt
	}

	// Validation
	if user.ExternalID == "" {
		return nil, eris.New("user.external_id is required")
	}

	if user.CreatedAt.IsZero() {
		return nil, eris.New("user.created_at is required")
	}

	// only verify user HMAC for untrusted client data
	if origin == dto.DataLogOriginClient && workspace.UserIDSigning != UserIDSigningNone {

		if (workspace.UserIDSigning == UserIDSigningAuthenticated && user.IsAuthenticated) ||
			workspace.UserIDSigning == UserIDSigningAll {

			// fail on hmac missing
			if user.HMAC == nil {
				return nil, eris.Errorf("user.hmac is required for user %v", user.ExternalID)
			}

			// get workspace active secret key
			wsSecretKey, errKey := workspace.GetActiveSecretKey(secretKey)

			if errKey != nil || wsSecretKey == nil {
				// no key found, abort here, should not happend
				return nil, eris.Errorf("workspace has no more active secret keys %v", err)
			}

			// only compare the first 8 characters of signature, thats enough entropy
			if !common.VerifyHMAC(wsSecretKey.Key, []byte(user.ExternalID), *user.HMAC, 8) {
				return nil, eris.Errorf("user.hmac is invalid for user %v", user.ExternalID)
			}
		}
	}

	return user, nil
}

type User struct {
	ID                string          `db:"id" json:"id"`
	ExternalID        string          `db:"external_id" json:"external_id"`
	IsMerged          bool            `db:"is_merged" json:"is_merged"` // mergedTo is set when a profile is merged into another
	MergedTo          *string         `db:"merged_to" json:"merged_to,omitempty"`
	MergedAt          *time.Time      `db:"merged_at" json:"merged_at,omitempty"`
	IsAuthenticated   bool            `db:"is_authenticated" json:"is_authenticated"`
	SignedUpAt        *time.Time      `db:"signed_up_at" json:"signed_up_at,omitempty"` // external signup date for authenticated users
	CreatedAt         time.Time       `db:"created_at" json:"created_at"`
	CreatedAtTrunc    time.Time       `db:"created_at_trunc" json:"created_at_trunc"` // date truncated at hour, used to optimize columnstore storage and queries
	LastInteractionAt time.Time       `db:"last_interaction_at" json:"last_interaction_at"`
	Timezone          *string         `db:"timezone" json:"timezone"`
	Language          *string         `db:"language" json:"language"`
	Country           *string         `db:"country" json:"country"`
	DBCreatedAt       time.Time       `db:"db_created_at" json:"db_created_at"`
	DBUpdatedAt       time.Time       `db:"db_updated_at" json:"db_updated_at"`
	FieldsTimestamp   FieldsTimestamp `db:"fields_timestamp" json:"fields_timestamp"` // holds a dictionary of fields timestamps to keep the most recent during user merging

	// optional fields:
	ConsentAll             *NullableBool    `db:"consent_all" json:"consent_all"`
	ConsentPersonalization *NullableBool    `db:"consent_personalization" json:"consent_personalization"`
	ConsentMarketing       *NullableBool    `db:"consent_marketing" json:"consent_marketing"`
	LastIP                 *NullableString  `db:"last_ip" json:"last_ip,omitempty"`
	Longitude              *NullableFloat64 `db:"longitude" json:"longitude,omitempty"`
	Latitude               *NullableFloat64 `db:"latitude" json:"latitude,omitempty"`
	Geo                    interface{}      `db:"geo" json:"geo"` // computed field used by DB
	FirstName              *NullableString  `db:"first_name" json:"first_name,omitempty"`
	LastName               *NullableString  `db:"last_name" json:"last_name,omitempty"`
	Gender                 *NullableString  `db:"gender" json:"gender,omitempty"`
	Birthday               *NullableString  `db:"birthday" json:"birthday,omitempty"`
	PhotoURL               *NullableString  `db:"photo_url" json:"photo_url,omitempty"`
	Email                  *NullableString  `db:"email" json:"email,omitempty"`
	EmailMD5               *NullableString  `db:"email_md5" json:"email_md5,omitempty"`
	EmailSHA1              *NullableString  `db:"email_sha1" json:"email_sha1,omitempty"`
	EmailSHA256            *NullableString  `db:"email_sha256" json:"email_sha256,omitempty"`
	Telephone              *NullableString  `db:"telephone" json:"telephone,omitempty"`
	AddressLine1           *NullableString  `db:"address_line_1" json:"address_line_1,omitempty"`
	AddressLine2           *NullableString  `db:"address_line_2" json:"address_line_2,omitempty"`
	City                   *NullableString  `db:"city" json:"city,omitempty"`
	Region                 *NullableString  `db:"region" json:"region,omitempty"`
	PostalCode             *NullableString  `db:"postal_code" json:"postal_code,omitempty"`
	State                  *NullableString  `db:"state" json:"state,omitempty"`
	// Cart               Cart             `db:"cart" json:"cart"`
	// CartItemsCount     int64            `db:"cart_items_count" json:"cart_items_count"`
	// CartUpdatedAt      *time.Time       `db:"cart_updated_at" json:"cart_updated_at,omitempty"`
	// CartAbandoned      bool             `db:"cart_abandoned" json:"cart_abandoned"`
	// WishList           Cart             `db:"wishlist" json:"wishlist"`
	// WishListItemsCount int64            `db:"wishlist_items_count" json:"wishlist_items_count"`
	// WishListUpdatedAt  *time.Time       `db:"wishlist_updated_at" json:"wishlist_updated_at,omitempty"`

	// Computed field
	OrdersCount          int64      `db:"orders_count" json:"orders_count"`
	OrdersLTV            int64      `db:"orders_ltv" json:"orders_ltv"` // in cents
	OrdersAvgCart        int64      `db:"orders_avg_cart" json:"orders_avg_cart"`
	FirstOrderAt         *time.Time `db:"first_order_at" json:"first_order_at,omitempty"`
	FirstOrderSubtotal   int64      `db:"first_order_subtotal" json:"first_order_subtotal"`                 // in cents
	FirstOrderTTC        int64      `db:"first_order_ttc" json:"first_order_ttc"`                           // in secs
	FirstOrderDomainID   *string    `db:"first_order_domain_id" json:"first_order_domain_id,omitempty"`     // in secs
	FirstOrderDomainType *string    `db:"first_order_domain_type" json:"first_order_domain_type,omitempty"` // in secs
	LastOrderAt          *time.Time `db:"last_order_at" json:"last_order_at,omitempty"`
	AvgRepeatCart        int64      `db:"avg_repeat_cart" json:"avg_repeat_cart"`           // in cents
	AvgRepeatOrderTTC    int64      `db:"avg_repeat_order_ttc" json:"avg_repeat_order_ttc"` // in secs

	// Not persisted in DB:
	HMAC         *string       `db:"-" json:"hmac,omitempty"` // signature for web activity with users protection
	UpdatedAt    *time.Time    `db:"-" json:"-"`              // used to merge fields and append item_timeline at the right time
	ExtraColumns AppItemFields `db:"-" json:"-"`              // converted into "app_xxx" fields when marshaling JSON

	// eventually joined server-side
	Segments          []*UserSegment          `json:"segments,omitempty"`
	Devices           []*Device               `json:"devices,omitempty"`
	Aliases           []*UserAlias            `json:"aliases,omitempty"`
	SubscriptionLists []*SubscriptionListUser `json:"subscription_lists,omitempty"`
}

func (u *User) BeforeInsert(workspace *Workspace) {

	if u.Timezone == nil {
		u.Timezone = &workspace.DefaultUserTimezone
	}

	if u.Language == nil {
		u.Language = &workspace.DefaultUserLanguage
	}

	if u.Country == nil {
		u.Country = &workspace.DefaultUserCountry
	}

	// signed_up_at is mandatory for authenticated users
	if u.IsAuthenticated && (u.SignedUpAt == nil || u.SignedUpAt.IsZero()) {
		u.SignedUpAt = &u.CreatedAt
	}

	if u.LastInteractionAt.IsZero() {
		u.LastInteractionAt = u.CreatedAt.Truncate(time.Second)
	}
}

// update a field timestamp to its most recent value
func (u *User) UpdateFieldTimestamp(field string, timestamp *time.Time) {
	if timestamp == nil {
		return
	}
	if previousTimestamp, exists := u.FieldsTimestamp[field]; exists && previousTimestamp.After(*timestamp) {
		return
	}

	u.FieldsTimestamp[field] = *timestamp
}

func (s *User) SetExtraColumns(field string, value *AppItemField, timestamp *time.Time) (update *UpdatedField) {
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

func (u *User) SetSignedUpAt(value *time.Time, timestamp *time.Time) (update *UpdatedField) {
	key := "signed_up_at"
	// we cant unset this value
	if value == nil {
		return
	}

	// abort if previous value is older than new value
	if u.SignedUpAt != nil && (u.SignedUpAt.Before(*value) || u.SignedUpAt.Equal(*value)) {
		return nil
	}

	update = &UpdatedField{
		Field:     key,
		PrevValue: TimePointerToInterface(u.SignedUpAt),
		NewValue:  TimePointerToInterface(value),
	}

	u.SignedUpAt = value
	u.UpdateFieldTimestamp(key, timestamp)
	return
}

// created_at will always keep the oldest date
// it's possible that a user a upserted with the current date from a web hit
// while it's real creation date is older in the "source-of-truth" customers DB
// it doesn't produce an UpdateField event as it is meaningless to observe such update
func (u *User) SetCreatedAt(value time.Time) {
	// update if current value is older
	if value.Before(u.CreatedAt) {
		u.CreatedAt = value
		u.CreatedAtTrunc = value.Truncate(time.Hour)
	}
}

// keep the newest value
func (u *User) SetLastInteractionAt(value time.Time) (update *UpdatedField) {
	// truncate value at second
	value = value.Truncate(time.Second)

	// update if current value is older
	if value.Before(u.LastInteractionAt) || value.Equal(u.LastInteractionAt.Truncate(time.Second)) {
		return nil
	}

	update = &UpdatedField{
		Field:     "last_interaction_at",
		PrevValue: u.LastInteractionAt.UTC(),
		NewValue:  value.UTC(),
	}
	u.LastInteractionAt = value
	u.UpdateFieldTimestamp("last_interaction_at", &value)
	return
}

func (u *User) SetTimezone(value *string, timestamp time.Time) (update *UpdatedField) {
	key := "timezone"
	// timezone cant be null
	if value == nil {
		return nil
	}
	// abort if values are equal
	if StringsEqual(u.Timezone, value) {
		return nil
	}

	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.Timezone = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: StringPointerToInterface(u.Timezone),
		NewValue:  StringPointerToInterface(value),
	}
	u.Timezone = value
	u.FieldsTimestamp[key] = timestamp
	return
}
func (u *User) SetLanguage(value *string, timestamp time.Time) (update *UpdatedField) {
	key := "language"
	// language cant be null
	if value == nil {
		return nil
	}
	// abort if values are equal
	if StringsEqual(u.Language, value) {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.Language = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: StringPointerToInterface(u.Language),
		NewValue:  StringPointerToInterface(value),
	}
	u.Language = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetCountry(value *string, timestamp time.Time) (update *UpdatedField) {
	key := "country"
	// country cant be null
	if value == nil {
		return nil
	}
	// abort if values are equal
	if StringsEqual(u.Country, value) {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.Country = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: StringPointerToInterface(u.Country),
		NewValue:  StringPointerToInterface(value),
	}
	u.Country = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetConsentAll(value *NullableBool, timestamp time.Time) (update *UpdatedField) {
	key := "consent_all"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if u.ConsentAll != nil && u.ConsentAll.IsNull == value.IsNull && u.ConsentAll.Bool == value.Bool {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.ConsentAll = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableBoolToInterface(u.ConsentAll),
		NewValue:  NullableBoolToInterface(value),
	}
	u.ConsentAll = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetConsentPersonalization(value *NullableBool, timestamp time.Time) (update *UpdatedField) {
	key := "consent_personalization"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if u.ConsentPersonalization != nil && u.ConsentPersonalization.IsNull == value.IsNull && u.ConsentPersonalization.Bool == value.Bool {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.ConsentPersonalization = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableBoolToInterface(u.ConsentPersonalization),
		NewValue:  NullableBoolToInterface(value),
	}
	u.ConsentPersonalization = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetConsentMarketing(value *NullableBool, timestamp time.Time) (update *UpdatedField) {
	key := "consent_marketing"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if u.ConsentMarketing != nil && u.ConsentMarketing.IsNull == value.IsNull && u.ConsentMarketing.Bool == value.Bool {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.ConsentMarketing = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableBoolToInterface(u.ConsentMarketing),
		NewValue:  NullableBoolToInterface(value),
	}
	u.ConsentMarketing = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetLastIP(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "last_ip"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if u.LastIP != nil && u.LastIP.IsNull == value.IsNull && u.LastIP.String == value.String {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.LastIP = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(u.LastIP),
		NewValue:  NullableStringToInterface(value),
	}
	u.LastIP = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetLongitude(value *NullableFloat64, timestamp time.Time) (update *UpdatedField) {
	key := "longitude"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// truncate to 6 decimals if value is float64
	if !value.IsNull {
		value.Float64 = math.Trunc(value.Float64*1e6) / 1e6
	}

	// abort if values are equal
	if u.Longitude != nil && u.Longitude.IsNull == value.IsNull && u.Longitude.Float64 == value.Float64 {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.Longitude = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableFloat64ToInterface(u.Longitude),
		NewValue:  NullableFloat64ToInterface(value),
	}
	u.Longitude = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetLatitude(value *NullableFloat64, timestamp time.Time) (update *UpdatedField) {
	key := "latitude"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// truncate to 6 decimals if value is float64
	if !value.IsNull {
		value.Float64 = math.Trunc(value.Float64*1e6) / 1e6
	}
	// abort if values are equal
	if u.Latitude != nil && u.Latitude.IsNull == value.IsNull && u.Latitude.Float64 == value.Float64 {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.Latitude = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableFloat64ToInterface(u.Latitude),
		NewValue:  NullableFloat64ToInterface(value),
	}
	u.Latitude = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetFirstName(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "first_name"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if u.FirstName != nil && u.FirstName.IsNull == value.IsNull && u.FirstName.String == value.String {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.FirstName = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(u.FirstName),
		NewValue:  NullableStringToInterface(value),
	}
	u.FirstName = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetLastName(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "last_name"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if u.LastName != nil && u.LastName.IsNull == value.IsNull && u.LastName.String == value.String {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.LastName = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(u.LastName),
		NewValue:  NullableStringToInterface(value),
	}
	u.LastName = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetGender(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "gender"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if u.Gender != nil && u.Gender.IsNull == value.IsNull && u.Gender.String == value.String {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.Gender = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(u.Gender),
		NewValue:  NullableStringToInterface(value),
	}
	u.Gender = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetBirthday(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "birthday"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if u.Birthday != nil && u.Birthday.IsNull == value.IsNull && u.Birthday.String == value.String {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.Birthday = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(u.Birthday),
		NewValue:  NullableStringToInterface(value),
	}
	u.Birthday = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetPhotoURL(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "photo_url"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if u.PhotoURL != nil && u.PhotoURL.IsNull == value.IsNull && u.PhotoURL.String == value.String {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.PhotoURL = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(u.PhotoURL),
		NewValue:  NullableStringToInterface(value),
	}
	u.PhotoURL = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetEmail(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "email"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if u.Email != nil && u.Email.IsNull == value.IsNull && u.Email.String == value.String {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.Email = value

		// update email hashes too
		if !value.IsNull {
			// compute md5 of email
			u.SetEmailMD5(&NullableString{IsNull: false, String: fmt.Sprintf("%x", md5.Sum([]byte(value.String)))}, timestamp)
			// compute sha1 of email
			u.SetEmailSHA1(&NullableString{IsNull: false, String: fmt.Sprintf("%x", sha1.Sum([]byte(value.String)))}, timestamp)
			// compute sha256 of email
			u.SetEmailSHA256(&NullableString{IsNull: false, String: fmt.Sprintf("%x", sha256.Sum256([]byte(value.String)))}, timestamp)
		}
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(u.Email),
		NewValue:  NullableStringToInterface(value),
	}

	u.Email = value
	// update email hashes too
	if !value.IsNull {
		// compute md5 of email
		u.SetEmailMD5(&NullableString{IsNull: false, String: fmt.Sprintf("%x", md5.Sum([]byte(value.String)))}, timestamp)
		// compute sha1 of email
		u.SetEmailSHA1(&NullableString{IsNull: false, String: fmt.Sprintf("%x", sha1.Sum([]byte(value.String)))}, timestamp)
		// compute sha256 of email
		u.SetEmailSHA256(&NullableString{IsNull: false, String: fmt.Sprintf("%x", sha256.Sum256([]byte(value.String)))}, timestamp)
	}

	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetEmailMD5(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "email_md5"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if u.EmailMD5 != nil && u.EmailMD5.IsNull == value.IsNull && u.EmailMD5.String == value.String {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.EmailMD5 = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(u.EmailMD5),
		NewValue:  NullableStringToInterface(value),
	}
	u.EmailMD5 = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetEmailSHA1(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "email_sha1"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if u.EmailSHA1 != nil && u.EmailSHA1.IsNull == value.IsNull && u.EmailSHA1.String == value.String {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.EmailSHA1 = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(u.EmailSHA1),
		NewValue:  NullableStringToInterface(value),
	}
	u.EmailSHA1 = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetEmailSHA256(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "email_sha256"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if u.EmailSHA256 != nil && u.EmailSHA256.IsNull == value.IsNull && u.EmailSHA256.String == value.String {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.EmailSHA256 = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(u.EmailSHA256),
		NewValue:  NullableStringToInterface(value),
	}
	u.EmailSHA256 = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetTelephone(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "telephone"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if u.Telephone != nil && u.Telephone.IsNull == value.IsNull && u.Telephone.String == value.String {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.Telephone = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(u.Telephone),
		NewValue:  NullableStringToInterface(value),
	}
	u.Telephone = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetAddressLine1(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "address_line_1"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if u.AddressLine1 != nil && u.AddressLine1.IsNull == value.IsNull && u.AddressLine1.String == value.String {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.AddressLine1 = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(u.AddressLine1),
		NewValue:  NullableStringToInterface(value),
	}
	u.AddressLine1 = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetAddressLine2(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "address_line_2"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if u.AddressLine2 != nil && u.AddressLine2.IsNull == value.IsNull && u.AddressLine2.String == value.String {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.AddressLine2 = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(u.AddressLine2),
		NewValue:  NullableStringToInterface(value),
	}
	u.AddressLine2 = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetCity(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "city"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if u.City != nil && u.City.IsNull == value.IsNull && u.City.String == value.String {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.City = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(u.City),
		NewValue:  NullableStringToInterface(value),
	}
	u.City = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetRegion(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "region"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if u.Region != nil && u.Region.IsNull == value.IsNull && u.Region.String == value.String {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.Region = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(u.Region),
		NewValue:  NullableStringToInterface(value),
	}
	u.Region = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetPostalCode(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "postal_code"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if u.PostalCode != nil && u.PostalCode.IsNull == value.IsNull && u.PostalCode.String == value.String {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.PostalCode = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(u.PostalCode),
		NewValue:  NullableStringToInterface(value),
	}
	u.PostalCode = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetState(value *NullableString, timestamp time.Time) (update *UpdatedField) {
	key := "state"
	// ignore if value is not provided
	if value == nil {
		return nil
	}
	// abort if values are equal
	if u.State != nil && u.State.IsNull == value.IsNull && u.State.String == value.String {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.State = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: NullableStringToInterface(u.State),
		NewValue:  NullableStringToInterface(value),
	}
	u.State = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetOrdersCount(value int64, timestamp time.Time) (update *UpdatedField) {
	key := "orders_count"
	// abort if values are equal
	if value == u.OrdersCount {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.OrdersCount = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: u.OrdersCount,
		NewValue:  value,
	}
	u.OrdersCount = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetOrdersLTV(value int64, timestamp time.Time) (update *UpdatedField) {
	key := "orders_ltv"
	// abort if values are equal
	if value == u.OrdersLTV {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.OrdersLTV = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: u.OrdersLTV,
		NewValue:  value,
	}
	u.OrdersLTV = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetOrdersAvgCart(value int64, timestamp time.Time) (update *UpdatedField) {
	key := "orders_avg_cart"
	// abort if values are equal
	if value == u.OrdersAvgCart {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.OrdersAvgCart = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: u.OrdersAvgCart,
		NewValue:  value,
	}
	u.OrdersAvgCart = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetFirstOrderAt(value *time.Time, timestamp time.Time) (update *UpdatedField) {
	key := "first_order_at"
	// abort if values are equal
	if TimeEqual(u.FirstOrderAt, value) {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.FirstOrderAt = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: TimePointerToInterface(u.FirstOrderAt),
		NewValue:  TimePointerToInterface(value),
	}
	u.FirstOrderAt = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetLastOrderAt(value *time.Time, timestamp time.Time) (update *UpdatedField) {
	key := "last_order_at"
	// abort if values are equal
	if TimeEqual(u.LastOrderAt, value) {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.LastOrderAt = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: TimePointerToInterface(u.LastOrderAt),
		NewValue:  TimePointerToInterface(value),
	}
	u.LastOrderAt = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetFirstOrderSubtotal(value int64, timestamp time.Time) (update *UpdatedField) {
	key := "first_order_subtotal"
	// abort if values are equal
	if value == u.FirstOrderSubtotal {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.FirstOrderSubtotal = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: u.FirstOrderSubtotal,
		NewValue:  value,
	}
	u.FirstOrderSubtotal = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetFirstOrderTTC(value int64, timestamp time.Time) (update *UpdatedField) {
	key := "first_order_ttc"
	// abort if values are equal
	if value == u.FirstOrderTTC {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.FirstOrderTTC = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: u.FirstOrderTTC,
		NewValue:  value,
	}
	u.FirstOrderTTC = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetFirstOrderDomainID(value *string, timestamp time.Time) (update *UpdatedField) {
	key := "first_order_domain_id"
	// country cant be null
	if value == nil {
		return nil
	}
	// abort if values are equal
	if StringsEqual(u.FirstOrderDomainID, value) {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.FirstOrderDomainID = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: StringPointerToInterface(u.FirstOrderDomainID),
		NewValue:  StringPointerToInterface(value),
	}
	u.FirstOrderDomainID = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetFirstOrderDomainType(value *string, timestamp time.Time) (update *UpdatedField) {
	key := "first_order_domain_type"
	// country cant be null
	if value == nil {
		return nil
	}
	// abort if values are equal
	if StringsEqual(u.FirstOrderDomainType, value) {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.FirstOrderDomainType = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: StringPointerToInterface(u.FirstOrderDomainType),
		NewValue:  StringPointerToInterface(value),
	}
	u.FirstOrderDomainType = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetAvgRepeatCart(value int64, timestamp time.Time) (update *UpdatedField) {
	key := "avg_repeat_cart"
	// abort if values are equal
	if value == u.AvgRepeatCart {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.AvgRepeatCart = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: u.AvgRepeatCart,
		NewValue:  value,
	}
	u.AvgRepeatCart = value
	u.FieldsTimestamp[key] = timestamp
	return
}

func (u *User) SetAvgRepeatOrderTTC(value int64, timestamp time.Time) (update *UpdatedField) {
	key := "avg_repeat_order_ttc"
	// abort if values are equal
	if value == u.AvgRepeatOrderTTC {
		return nil
	}
	existingValueTimestamp := u.GetFieldDate(key)
	// abort if existing value is newer
	if existingValueTimestamp.After(timestamp) {
		return nil
	}
	// the value might be set for the first time
	// so we set the value without producing a field update
	if existingValueTimestamp.Equal(timestamp) {
		u.AvgRepeatOrderTTC = value
		return
	}
	update = &UpdatedField{
		Field:     key,
		PrevValue: u.AvgRepeatOrderTTC,
		NewValue:  value,
	}
	u.AvgRepeatOrderTTC = value
	u.FieldsTimestamp[key] = timestamp
	return
}

// determines if user has already been persisted in DB
func (u *User) IsNew() bool {
	return u.DBCreatedAt.IsZero()
}

func (u *User) GetFieldDate(field string) time.Time {
	// use updated_at if it has been passed in the API data import
	if u.UpdatedAt != nil && u.UpdatedAt.After(u.CreatedAt) {
		return *u.UpdatedAt
	}
	// or use the existing field timestamp
	if date, exists := u.FieldsTimestamp[field]; exists {
		return date
	}
	// or use the object creation date as a fallback
	return u.CreatedAt
}

func (u *User) ComputeEmailHashes() {
	if u.Email != nil && !u.Email.IsNull {
		u.EmailMD5 = NewNullableString(StringPtr(fmt.Sprintf("%x", md5.Sum([]byte(u.Email.String)))))
		u.EmailSHA1 = NewNullableString(StringPtr(fmt.Sprintf("%x", sha1.Sum([]byte(u.Email.String)))))
		u.EmailSHA256 = NewNullableString(StringPtr(fmt.Sprintf("%x", sha256.Sum256([]byte(u.Email.String)))))
	}
}

// merges two user profiles and returns the list of updated fields
func (fromUser *User) MergeInto(toUser *User, workspace *Workspace) (updatedFields []*UpdatedField) {

	updatedFields = []*UpdatedField{} // init

	if toUser.FieldsTimestamp == nil {
		toUser.FieldsTimestamp = FieldsTimestamp{}
	}

	if fieldUpdate := toUser.SetLastInteractionAt(fromUser.LastInteractionAt); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if toUser.IsAuthenticated && fromUser.IsAuthenticated {
		if fieldUpdate := toUser.SetSignedUpAt(fromUser.SignedUpAt, TimePtr(fromUser.GetFieldDate("signed_up_at"))); fieldUpdate != nil {
			updatedFields = append(updatedFields, fieldUpdate)
		}
	}
	if fieldUpdate := toUser.SetConsentAll(fromUser.ConsentAll, fromUser.GetFieldDate("consent_all")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toUser.SetConsentPersonalization(fromUser.ConsentPersonalization, fromUser.GetFieldDate("consent_personalization")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toUser.SetConsentMarketing(fromUser.ConsentMarketing, fromUser.GetFieldDate("consent_marketing")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toUser.SetTimezone(fromUser.Timezone, fromUser.GetFieldDate("timezone")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toUser.SetLanguage(fromUser.Language, fromUser.GetFieldDate("language")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toUser.SetCountry(fromUser.Country, fromUser.GetFieldDate("country")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toUser.SetLastIP(fromUser.LastIP, fromUser.GetFieldDate("last_ip")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toUser.SetLongitude(fromUser.Longitude, fromUser.GetFieldDate("longitude")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toUser.SetLatitude(fromUser.Latitude, fromUser.GetFieldDate("latitude")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toUser.SetFirstName(fromUser.FirstName, fromUser.GetFieldDate("first_name")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toUser.SetLastName(fromUser.LastName, fromUser.GetFieldDate("last_name")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toUser.SetGender(fromUser.Gender, fromUser.GetFieldDate("gender")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toUser.SetBirthday(fromUser.Birthday, fromUser.GetFieldDate("birthday")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toUser.SetPhotoURL(fromUser.PhotoURL, fromUser.GetFieldDate("photo_url")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toUser.SetEmail(fromUser.Email, fromUser.GetFieldDate("email")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toUser.SetEmailMD5(fromUser.EmailMD5, fromUser.GetFieldDate("email_md5")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toUser.SetEmailSHA1(fromUser.EmailSHA1, fromUser.GetFieldDate("email_sha1")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toUser.SetEmailSHA256(fromUser.EmailSHA256, fromUser.GetFieldDate("email_sha256")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toUser.SetTelephone(fromUser.Telephone, fromUser.GetFieldDate("telephone")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toUser.SetAddressLine1(fromUser.AddressLine1, fromUser.GetFieldDate("address_line_1")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toUser.SetAddressLine2(fromUser.AddressLine2, fromUser.GetFieldDate("address_line_2")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toUser.SetCity(fromUser.City, fromUser.GetFieldDate("city")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toUser.SetRegion(fromUser.Region, fromUser.GetFieldDate("region")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toUser.SetPostalCode(fromUser.PostalCode, fromUser.GetFieldDate("postal_code")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	if fieldUpdate := toUser.SetState(fromUser.State, fromUser.GetFieldDate("state")); fieldUpdate != nil {
		updatedFields = append(updatedFields, fieldUpdate)
	}
	// computed KPIs cant be merged, they are computed later

	for key, value := range fromUser.ExtraColumns {
		if fieldUpdate := toUser.SetExtraColumns(key, value, TimePtr(fromUser.GetFieldDate(key))); fieldUpdate != nil {
			updatedFields = append(updatedFields, fieldUpdate)
		}
	}

	// set created_at + updated_at after merging fields
	toUser.SetCreatedAt(fromUser.CreatedAt)
	toUser.UpdatedAt = fromUser.UpdatedAt

	return
}

// the SORT KEY uses the external user creation date "created_at" and the microsecond precision internal db_created_at
// this is used to perform pagination ordering at the most precision
var UserSchema string = `CREATE TABLE IF NOT EXISTS user (
	id VARCHAR(64) NOT NULL,
	external_id VARCHAR(255) NOT NULL,
	is_merged BOOLEAN DEFAULT FALSE,
	merged_to VARCHAR(64),
	merged_at DATETIME,
	is_authenticated BOOLEAN DEFAULT FALSE,
	signed_up_at DATETIME,
	created_at DATETIME(6) NOT NULL,
	created_at_trunc AS DATE_TRUNC('hour', created_at) PERSISTED DATETIME,
	last_interaction_at DATETIME NOT NULL,
	timezone VARCHAR(50) NOT NULL,
	language CHAR(2) NOT NULL,
	country VARCHAR(50) NOT NULL,
	db_created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	fields_timestamp JSON NOT NULL,
	
	consent_all BOOLEAN,
	consent_personalization BOOLEAN,
	consent_marketing BOOLEAN,
	last_ip VARCHAR(45),
	latitude DECIMAL(9,6),
	longitude DECIMAL(9,6),
	geo AS GEOGRAPHY_POINT(longitude, latitude) PERSISTED geographypoint,
	first_name VARCHAR(50),
	last_name VARCHAR(50),
	gender CHAR(6),
	birthday CHAR(10), -- format: 0000-00-00
	photo_url VARCHAR(2083),
	email VARCHAR(255),
	email_md5 VARCHAR(32),
	email_sha1 VARCHAR(40),
	email_sha256 VARCHAR(64),
	telephone VARCHAR(22),
	address_line_1 VARCHAR(255),
	address_line_2 VARCHAR(255),
	city VARCHAR(50),
	region VARCHAR(50),
	postal_code VARCHAR(50),
	state VARCHAR(50),
	-- cart JSON NOT NULL,
	-- cart_items_count AS JSON_LENGTH(cart::items) PERSISTED SMALLINT UNSIGNED,
	-- cart_updated_at AS TO_DATE(cart::$updatedAt, 'YYYY-MM-DDTHH24:MI:SS') PERSISTED DATETIME,
	-- cart_abandoned BOOLEAN DEFAULT FALSE,
	-- wishlist JSON NOT NULL,
	-- wishlist_items_count AS JSON_LENGTH(wishlist::items) PERSISTED INT UNSIGNED,
	-- wishlist_updated_at AS TO_DATE(wishlist::$updatedAt, 'YYYY-MM-DDTHH24:MI:SS') PERSISTED DATETIME,
	
	orders_count SMALLINT UNSIGNED NOT NULL DEFAULT 0,
	orders_ltv INT UNSIGNED NOT NULL DEFAULT 0, -- not null saves some space in DB
	orders_avg_cart INT UNSIGNED NOT NULL DEFAULT 0,
	first_order_at DATETIME,
	first_order_domain_id VARCHAR(64),
	first_order_domain_type VARCHAR(32),
	first_order_subtotal INT UNSIGNED NOT NULL DEFAULT 0,
	first_order_ttc INT UNSIGNED NOT NULL DEFAULT 0,
	last_order_at DATETIME,
	avg_repeat_cart INT UNSIGNED NOT NULL DEFAULT 0,
	avg_repeat_order_ttc INT UNSIGNED NOT NULL DEFAULT 0,
	
	SORT KEY (created_at_trunc DESC, created_at DESC),
	PRIMARY KEY (id),
	KEY (is_authenticated),
	KEY (external_id),
	SHARD KEY (id)
) COLLATE utf8mb4_general_ci;`

var UserSchemaMYSQL string = `CREATE TABLE IF NOT EXISTS user (
	id VARCHAR(64) NOT NULL,
	external_id VARCHAR(255) NOT NULL,
	is_merged BOOLEAN DEFAULT FALSE,
	merged_to VARCHAR(64),
	merged_at DATETIME,
	is_authenticated BOOLEAN DEFAULT FALSE,
	signed_up_at DATETIME,
	created_at DATETIME(6) NOT NULL,
	created_at_trunc DATETIME GENERATED ALWAYS AS (DATE_FORMAT(created_at, '%Y-%m-%d %H:00:00')) STORED,
	last_interaction_at DATETIME NOT NULL,
	timezone VARCHAR(50) NOT NULL,
	language CHAR(2) NOT NULL,
	country VARCHAR(50) NOT NULL,
	db_created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	fields_timestamp JSON NOT NULL,
	
	consent_all BOOLEAN,
	consent_personalization BOOLEAN,
	consent_marketing BOOLEAN,
	last_ip VARCHAR(45),
	latitude DECIMAL(9,6),
	longitude DECIMAL(9,6),
	geo POINT GENERATED ALWAYS AS (POINT(longitude, latitude)) STORED,
	first_name VARCHAR(50),
	last_name VARCHAR(50),
	gender CHAR(6),
	birthday CHAR(10), -- format: 0000-00-00
	photo_url VARCHAR(2083),
	email VARCHAR(255),
	email_md5 VARCHAR(32),
	email_sha1 VARCHAR(40),
	email_sha256 VARCHAR(64),
	telephone VARCHAR(22),
	address_line_1 VARCHAR(255),
	address_line_2 VARCHAR(255),
	city VARCHAR(50),
	region VARCHAR(50),
	postal_code VARCHAR(50),
	state VARCHAR(50),
	-- cart JSON NOT NULL,
	-- cart_items_count AS JSON_LENGTH(cart::items) PERSISTED SMALLINT UNSIGNED,
	-- cart_updated_at AS TO_DATE(cart::$updatedAt, 'YYYY-MM-DDTHH24:MI:SS') PERSISTED DATETIME,
	-- cart_abandoned BOOLEAN DEFAULT FALSE,
	-- wishlist JSON NOT NULL,
	-- wishlist_items_count AS JSON_LENGTH(wishlist::items) PERSISTED INT UNSIGNED,
	-- wishlist_updated_at AS TO_DATE(wishlist::$updatedAt, 'YYYY-MM-DDTHH24:MI:SS') PERSISTED DATETIME,
	
	orders_count SMALLINT UNSIGNED NOT NULL DEFAULT 0,
	orders_ltv INT UNSIGNED NOT NULL DEFAULT 0, -- not null saves some space in DB
	orders_avg_cart INT UNSIGNED NOT NULL DEFAULT 0,
	first_order_at DATETIME,
	first_order_domain_id VARCHAR(64),
	first_order_domain_type VARCHAR(32),
	first_order_subtotal INT UNSIGNED NOT NULL DEFAULT 0,
	first_order_ttc INT UNSIGNED NOT NULL DEFAULT 0,
	last_order_at DATETIME,
	avg_repeat_cart INT UNSIGNED NOT NULL DEFAULT 0,
	avg_repeat_order_ttc INT UNSIGNED NOT NULL DEFAULT 0,
	
	-- SORT KEY (created_at_trunc, db_created_at),
	PRIMARY KEY (id),
	KEY (is_authenticated),
	KEY (external_id)
	-- SHARD KEY (id)
) COLLATE utf8mb4_general_ci;`

func NewUserCube() *CubeJSSchema {
	return &CubeJSSchema{
		Title:       "Users",
		Description: "Users",
		SQL:         "SELECT * FROM `user` WHERE is_merged = FALSE",
		Joins:       map[string]CubeJSSchemaJoin{
			// "Order": {
			// 	Relationship: "hasMany",
			// 	SQL:          "${CUBE}.id = ${Order.user_id}",
			// },
			// "Order_item": {
			// 	Relationship: "hasMany",
			// 	SQL:          "${CUBE}.id = ${Order_item.user_id}",
			// },
			// "Session": {
			// 	Relationship: "hasMany",
			// 	SQL:          "${CUBE}.id = ${Session.user_id}",
			// },
			// "Pageview": {
			// 	Relationship: "hasMany",
			// 	SQL:          "${CUBE}.id = ${Pageview.user_id}",
			// },
			// "Cart": {
			// 	Relationship: "hasMany",
			// 	SQL:          "${CUBE}.id = ${Cart.user_id}",
			// },
			// "Cart_item": {
			// 	Relationship: "hasMany",
			// 	SQL:          "${CUBE}.id = ${Cart_item.user_id}",
			// },
		},
		Segments: map[string]CubeJSSchemaSegment{
			"authenticated": {
				SQL: "${CUBE}.is_authenticated = true",
			},
		},
		Measures: map[string]CubeJSSchemaMeasure{
			"count": {
				Type:        "count",
				Title:       "Count all",
				Description: "Count all",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"sign_ups": {
				Type:        "count",
				Title:       "Count sign ups",
				Description: "Count sign ups: signed_up_at IS NOT NULL",
				SQL:         "${CUBE}.signed_up_at IS NOT NULL",
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
				Title:       "User ID",
				Description: "field: id",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"external_id": {
				SQL:         "external_id",
				Type:        "string",
				Title:       "External ID",
				Description: "field: external_id",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"is_authenticated": {
				SQL:         "is_authenticated",
				Type:        "number",
				Title:       "Is authenticated",
				Description: "field: is_authenticated",
			},
			"signed_up_at": {
				SQL:         "signed_up_at",
				Type:        "time",
				Title:       "Signed up at",
				Description: "field: signed_up_at",
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
				Title:       "Created at (truncated at hour)",
				Description: "field: created_at_trunc",
			},
			"db_created_at": {
				SQL:         "db_created_at",
				Type:        "time",
				Title:       "DB created at",
				Description: "field: db_created_at",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"db_updated_at": {
				SQL:         "db_updated_at",
				Type:        "time",
				Title:       "DB updated at",
				Description: "field: db_updated_at",
				Meta: MapOfInterfaces{
					"hide_from_segmentation": true,
				},
			},
			"last_interaction_at": {
				SQL:         "last_interaction_at",
				Type:        "time",
				Title:       "Last interaction at",
				Description: "field: last_interaction_at",
			},
			"timezone": {
				SQL:         "timezone",
				Type:        "string",
				Title:       "Timezone",
				Description: "field: timezone",
			},
			"language": {
				SQL:         "language",
				Type:        "string",
				Title:       "Language",
				Description: "field: language",
			},
			"gender": {
				SQL:         "gender",
				Type:        "string",
				Title:       "Gender",
				Description: "field: gender",
			},
			"birthday": {
				SQL:         "birthday",
				Type:        "string",
				Title:       "Birthday",
				Description: "field: birthday",
			},
			"city": {
				SQL:         "city",
				Type:        "string",
				Title:       "City",
				Description: "field: city",
			},
			"region": {
				SQL:         "region",
				Type:        "string",
				Title:       "Region",
				Description: "field: region",
			},
			"postal_code": {
				SQL:         "postal_code",
				Type:        "string",
				Title:       "Postal code",
				Description: "field: postal_code",
			},
			"state": {
				SQL:         "state",
				Type:        "string",
				Title:       "State",
				Description: "field: state",
			},
			"country": {
				SQL:         "country",
				Type:        "string",
				Title:       "Country",
				Description: "field: country",
			},
			"orders_count": {
				SQL:         "orders_count",
				Type:        "number",
				Title:       "Orders count",
				Description: "field: orders_count",
			},
			"orders_ltv": {
				SQL:         "orders_ltv",
				Type:        "number",
				Title:       "Order LTV",
				Description: "field: orders_ltv",
			},
			"orders_avg_cart": {
				SQL:         "orders_avg_cart",
				Type:        "number",
				Title:       "Average cart",
				Description: "field: orders_avg_cart",
			},
			"first_order_at": {
				SQL:         "first_order_at",
				Type:        "time",
				Title:       "Date of first order",
				Description: "field: first_order_at",
			},
			"first_order_domain_id": {
				SQL:         "first_order_domain_id",
				Type:        "string",
				Title:       "First order domain ID",
				Description: "field: first_order_domain_id",
			},
			"first_order_domain_type": {
				SQL:         "first_order_domain_type",
				Type:        "string",
				Title:       "First order domain type",
				Description: "field: first_order_domain_type",
			},
			"first_order_subtotal": {
				SQL:         "first_order_subtotal",
				Type:        "number",
				Title:       "First order subtotal",
				Description: "field: first_order_subtotal",
			},
			"first_order_ttc": {
				SQL:         "first_order_ttc",
				Type:        "number",
				Title:       "First order Time To Conversion",
				Description: "field: first_order_ttc",
			},
			"last_order_at": {
				SQL:         "last_order_at",
				Type:        "time",
				Title:       "Date of last order",
				Description: "field: last_order_at",
			},
			"avg_repeat_cart": {
				SQL:         "avg_repeat_cart",
				Type:        "number",
				Title:       "Average repeat cart",
				Description: "field: avg_repeat_cart",
			},
			"avg_repeat_order_ttc": {
				SQL:         "avg_repeat_order_ttc",
				Type:        "number",
				Title:       "Average repeat order Time To Conversion",
				Description: "field: avg_repeat_order_ttc",
			},
		},
	}
}
