package entity

import (
	"fmt"
	"testing"
	"time"
)

func TestEntity_User_MergeInto(t *testing.T) {

	cfgSecretKey := "test"
	orgID := "testing"
	workspaceID := fmt.Sprintf("%v_%v", orgID, "demoecommerce")

	demoWorkspace, err := GenerateDemoWorkspace(workspaceID, WorkspaceDemoOrder, orgID, cfgSecretKey)
	if err != nil {
		t.Fatalf("generate demo workspace err %v", err)
	}

	t.Run("should merge every field properly", func(t *testing.T) {

		// we merge two authenticated users to test the signed_up_at field
		createdAt := time.Now().AddDate(0, -1, 0)
		signedUpAt := createdAt.AddDate(0, 0, 15)

		oldestDate := createdAt
		recentDate := oldestDate.AddDate(0, 0, 1)

		// init with fake values to be sure they will be overwritten
		fromUser := NewUser("ext-id", true, createdAt, createdAt, "NONE", "NONE", "NONE", nil)
		toUser := NewUser("ext-id", true, createdAt, createdAt, "NONE", "NONE", "NONE", nil)

		fromUser.SetSignedUpAt(&signedUpAt, &recentDate)
		toUser.SetSignedUpAt(&createdAt, &oldestDate)

		fromUser.SetTimezone(StringPtr("Europe/Paris"), recentDate)
		toUser.SetTimezone(StringPtr("America/New_York"), oldestDate)

		fromUser.SetLanguage(StringPtr("fr"), recentDate)
		toUser.SetLanguage(StringPtr("en"), oldestDate)

		fromUser.SetCountry(StringPtr("FR"), recentDate)
		toUser.SetCountry(StringPtr("US"), oldestDate)

		fromUser.SetConsentAll(&NullableBool{Bool: true}, recentDate)
		toUser.SetConsentAll(nil, oldestDate)

		fromUser.SetLastIP(&NullableString{String: "xxx"}, recentDate)
		toUser.SetLastIP(nil, oldestDate)

		fromUser.SetLongitude(&NullableFloat64{Float64: 10.1234567891}, recentDate)
		toUser.SetLongitude(nil, oldestDate)

		fromUser.SetLatitude(&NullableFloat64{Float64: 10.1234567891}, recentDate)
		toUser.SetLatitude(nil, oldestDate)

		fromUser.SetFirstName(&NullableString{String: "xxx"}, recentDate)
		toUser.SetFirstName(nil, oldestDate)

		fromUser.SetLastName(&NullableString{String: "xxx"}, recentDate)
		toUser.SetLastName(nil, oldestDate)

		fromUser.SetGender(&NullableString{String: "xxx"}, recentDate)
		toUser.SetGender(nil, oldestDate)

		fromUser.SetBirthday(&NullableString{String: "xxx"}, recentDate)
		toUser.SetBirthday(nil, oldestDate)

		fromUser.SetPhotoURL(&NullableString{String: "xxx"}, recentDate)
		toUser.SetPhotoURL(nil, oldestDate)

		fromUser.SetEmail(&NullableString{String: "xxx"}, recentDate)
		toUser.SetEmail(nil, oldestDate)

		fromUser.SetEmailMD5(&NullableString{String: "xxx"}, recentDate)
		toUser.SetEmailMD5(nil, oldestDate)

		fromUser.SetEmailSHA1(&NullableString{String: "xxx"}, recentDate)
		toUser.SetEmailSHA1(nil, oldestDate)

		fromUser.SetEmailSHA256(&NullableString{String: "xxx"}, recentDate)
		toUser.SetEmailSHA256(nil, oldestDate)

		fromUser.SetTelephone(&NullableString{String: "xxx"}, recentDate)
		toUser.SetTelephone(nil, oldestDate)

		fromUser.SetAddressLine1(&NullableString{String: "xxx"}, recentDate)
		toUser.SetAddressLine1(nil, oldestDate)

		fromUser.SetAddressLine2(&NullableString{String: "xxx"}, recentDate)
		toUser.SetAddressLine2(nil, oldestDate)

		fromUser.SetCity(&NullableString{String: "xxx"}, recentDate)
		toUser.SetCity(nil, oldestDate)

		fromUser.SetRegion(&NullableString{String: "xxx"}, recentDate)
		toUser.SetRegion(nil, oldestDate)

		fromUser.SetPostalCode(&NullableString{String: "xxx"}, recentDate)
		toUser.SetPostalCode(nil, oldestDate)

		fromUser.SetState(&NullableString{String: "xxx"}, recentDate)
		toUser.SetState(nil, oldestDate)

		// TODO
		// fromUser.SetCustomColumns("app_test_num", 123, &recentDate)
		// fromUser.SetCustomColumns("app_test_replaced", "replaced", &recentDate)

		// toUser.SetCustomColumns("app_test_replaced", "not-yet-replaced", &oldestDate)

		// merge fields and compare them
		updatedFields := fromUser.MergeInto(toUser, demoWorkspace)

		if len(updatedFields) == 0 {
			t.Fatal("updatedFields should not be empty")
		}

		if !toUser.SignedUpAt.Equal(*fromUser.SignedUpAt) {
			t.Errorf("SignedUpAt should be equal, got %v, want %v", *toUser.SignedUpAt, *fromUser.SignedUpAt)
		}
		if !toUser.SignedUpAt.Equal(oldestDate) {
			t.Errorf("signed_up_at should get oldestDate value, got %v, want %v", toUser.SignedUpAt, oldestDate)
		}

		if *toUser.Timezone != *fromUser.Timezone {
			t.Errorf("Timezone should be equal, got %v, want %v", *toUser.Timezone, *fromUser.Timezone)
		}
		if !toUser.FieldsTimestamp["timezone"].Equal(recentDate) {
			t.Errorf("timezone should get recent value, got %v, want %v", toUser.FieldsTimestamp["timezone"], recentDate)
		}
		if *toUser.Timezone != "Europe/Paris" {
			t.Errorf("timezone is %v, want %v", *toUser.Timezone, "Europe/Paris")
		}

		if *toUser.Language != *fromUser.Language {
			t.Errorf("Language should be equal, got %v, want %v", *toUser.Language, *fromUser.Language)
		}
		if !toUser.FieldsTimestamp["language"].Equal(recentDate) {
			t.Errorf("language should get recent value, got %v, want %v", toUser.FieldsTimestamp["language"], recentDate)
		}

		if *toUser.Country != *fromUser.Country {
			t.Errorf("Country should be equal, got %v, want %v", *toUser.Country, *fromUser.Country)
		}
		if !toUser.FieldsTimestamp["country"].Equal(recentDate) {
			t.Errorf("country should get recent value, got %v, want %v", toUser.FieldsTimestamp["country"], recentDate)
		}

		if toUser.ConsentAll.Bool != fromUser.ConsentAll.Bool {
			t.Errorf("UserCentricConsent should be equal, got %v, want %v", toUser.ConsentAll.Bool, fromUser.ConsentAll.Bool)
		}
		if !toUser.FieldsTimestamp["consent_all"].Equal(recentDate) {
			t.Errorf("consent_all should get recent value, got %v, want %v", toUser.FieldsTimestamp["consent_all"], recentDate)
		}

		if toUser.LastIP.String != fromUser.LastIP.String {
			t.Errorf("LastIP should be equal, got %v, want %v", toUser.LastIP.String, fromUser.LastIP.String)
		}
		if !toUser.FieldsTimestamp["last_ip"].Equal(recentDate) {
			t.Errorf("last_ip should get recent value, got %v, want %v", toUser.FieldsTimestamp["last_ip"], recentDate)
		}

		if toUser.Longitude.Float64 != fromUser.Longitude.Float64 {
			t.Errorf("Longitude should be equal, got %v, want %v", toUser.Longitude.Float64, fromUser.Longitude.Float64)
		}
		if !toUser.FieldsTimestamp["longitude"].Equal(recentDate) {
			t.Errorf("longitude should get recent value, got %v, want %v", toUser.FieldsTimestamp["longitude"], recentDate)
		}

		if toUser.Latitude.Float64 != fromUser.Latitude.Float64 {
			t.Errorf("Latitude should be equal, got %v, want %v", toUser.Latitude.Float64, fromUser.Latitude.Float64)
		}
		if !toUser.FieldsTimestamp["latitude"].Equal(recentDate) {
			t.Errorf("latitude should get recent value, got %v, want %v", toUser.FieldsTimestamp["latitude"], recentDate)
		}

		if toUser.FirstName.String != fromUser.FirstName.String {
			t.Errorf("FirstName should be equal, got %v, want %v", toUser.FirstName.String, fromUser.FirstName.String)
		}
		if !toUser.FieldsTimestamp["first_name"].Equal(recentDate) {
			t.Errorf("first_name should get recent value, got %v, want %v", toUser.FieldsTimestamp["first_name"], recentDate)
		}

		if toUser.LastName.String != fromUser.LastName.String {
			t.Errorf("LastName should be equal, got %v, want %v", toUser.LastName.String, fromUser.LastName.String)
		}
		if !toUser.FieldsTimestamp["last_name"].Equal(recentDate) {
			t.Errorf("last_name should get recent value, got %v, want %v", toUser.FieldsTimestamp["last_name"], recentDate)
		}

		if toUser.Gender.String != fromUser.Gender.String {
			t.Errorf("Gender should be equal, got %v, want %v", toUser.Gender.String, fromUser.Gender.String)
		}
		if !toUser.FieldsTimestamp["gender"].Equal(recentDate) {
			t.Errorf("gender should get recent value, got %v, want %v", toUser.FieldsTimestamp["gender"], recentDate)
		}

		if toUser.Birthday.String != fromUser.Birthday.String {
			t.Errorf("Birthday should be equal, got %v, want %v", toUser.Birthday.String, fromUser.Birthday.String)
		}
		if !toUser.FieldsTimestamp["birthday"].Equal(recentDate) {
			t.Errorf("birthday should get recent value, got %v, want %v", toUser.FieldsTimestamp["birthday"], recentDate)
		}

		if toUser.PhotoURL.String != fromUser.PhotoURL.String {
			t.Errorf("PhotoURL should be equal, got %v, want %v", toUser.PhotoURL.String, fromUser.PhotoURL.String)
		}
		if !toUser.FieldsTimestamp["photo_url"].Equal(recentDate) {
			t.Errorf("photo_url should get recent value, got %v, want %v", toUser.FieldsTimestamp["photo_url"], recentDate)
		}

		if toUser.Email.String != fromUser.Email.String {
			t.Errorf("Email should be equal, got %v, want %v", toUser.Email.String, fromUser.Email.String)
		}
		if !toUser.FieldsTimestamp["email"].Equal(recentDate) {
			t.Errorf("email should get recent value, got %v, want %v", toUser.FieldsTimestamp["email"], recentDate)
		}

		if toUser.EmailMD5.String != fromUser.EmailMD5.String {
			t.Errorf("EmailMD5 should be equal, got %v, want %v", toUser.EmailMD5.String, fromUser.EmailMD5.String)
		}
		if !toUser.FieldsTimestamp["email_md5"].Equal(recentDate) {
			t.Errorf("email_md5 should get recent value, got %v, want %v", toUser.FieldsTimestamp["email_md5"], recentDate)
		}

		if toUser.EmailSHA1.String != fromUser.EmailSHA1.String {
			t.Errorf("EmailSHA1 should be equal, got %v, want %v", toUser.EmailSHA1.String, fromUser.EmailSHA1.String)
		}
		if !toUser.FieldsTimestamp["email_sha1"].Equal(recentDate) {
			t.Errorf("email_sha1 should get recent value, got %v, want %v", toUser.FieldsTimestamp["email_sha1"], recentDate)
		}

		if toUser.EmailSHA256.String != fromUser.EmailSHA256.String {
			t.Errorf("EmailSHA256 should be equal, got %v, want %v", toUser.EmailSHA256.String, fromUser.EmailSHA256.String)
		}
		if !toUser.FieldsTimestamp["email_sha256"].Equal(recentDate) {
			t.Errorf("email_sha256 should get recent value, got %v, want %v", toUser.FieldsTimestamp["email_sha256"], recentDate)
		}

		if toUser.Telephone.String != fromUser.Telephone.String {
			t.Errorf("Telephone should be equal, got %v, want %v", toUser.Telephone.String, fromUser.Telephone.String)
		}
		if !toUser.FieldsTimestamp["telephone"].Equal(recentDate) {
			t.Errorf("telephone should get recent value, got %v, want %v", toUser.FieldsTimestamp["telephone"], recentDate)
		}

		if toUser.AddressLine1.String != fromUser.AddressLine1.String {
			t.Errorf("AddressLine1 should be equal, got %v, want %v", toUser.AddressLine1.String, fromUser.AddressLine1.String)
		}
		if !toUser.FieldsTimestamp["address_line_1"].Equal(recentDate) {
			t.Errorf("address_line_1 should get recent value, got %v, want %v", toUser.FieldsTimestamp["address_line_1"], recentDate)
		}

		if toUser.AddressLine2.String != fromUser.AddressLine2.String {
			t.Errorf("AddressLine2 should be equal, got %v, want %v", toUser.AddressLine2.String, fromUser.AddressLine2.String)
		}
		if !toUser.FieldsTimestamp["address_line_2"].Equal(recentDate) {
			t.Errorf("address_line_2 should get recent value, got %v, want %v", toUser.FieldsTimestamp["address_line_2"], recentDate)
		}

		if toUser.City.String != fromUser.City.String {
			t.Errorf("City should be equal, got %v, want %v", toUser.City.String, fromUser.City.String)
		}
		if !toUser.FieldsTimestamp["city"].Equal(recentDate) {
			t.Errorf("city should get recent value, got %v, want %v", toUser.FieldsTimestamp["city"], recentDate)
		}

		if toUser.Region.String != fromUser.Region.String {
			t.Errorf("Region should be equal, got %v, want %v", toUser.Region.String, fromUser.Region.String)
		}
		if !toUser.FieldsTimestamp["region"].Equal(recentDate) {
			t.Errorf("region should get recent value, got %v, want %v", toUser.FieldsTimestamp["region"], recentDate)
		}

		if toUser.PostalCode.String != fromUser.PostalCode.String {
			t.Errorf("PostalCode should be equal, got %v, want %v", toUser.PostalCode.String, fromUser.PostalCode.String)
		}
		if !toUser.FieldsTimestamp["postal_code"].Equal(recentDate) {
			t.Errorf("postal_code should get recent value, got %v, want %v", toUser.FieldsTimestamp["postal_code"], recentDate)
		}

		if toUser.State.String != fromUser.State.String {
			t.Errorf("State should be equal, got %v, want %v", toUser.State.String, fromUser.State.String)
		}
		if !toUser.FieldsTimestamp["state"].Equal(recentDate) {
			t.Errorf("state should get recent value, got %v, want %v", toUser.FieldsTimestamp["state"], recentDate)
		}

		// if !toUser.CartUpdatedAt.Equal(*fromUser.CartUpdatedAt) {
		// 	t.Errorf("CartUpdatedAt should be equal, got %v, want %v", *toUser.CartUpdatedAt, *fromUser.CartUpdatedAt)
		// }
		// if !toUser.MergeableFields["cart_updated_at"].Equal(recentDate) {
		// 	t.Errorf("cart_updated_at should get recent value, got %v, want %v", toUser.MergeableFields["cart_updated_at"], recentDate)
		// }

		// if toUser.CartAbandoned != fromUser.CartAbandoned {
		// 	t.Errorf("CartAbandoned should be equal, got %v, want %v", toUser.CartAbandoned, fromUser.CartAbandoned)
		// }
		// if !toUser.MergeableFields["cart_abandoned"].Equal(recentDate) {
		// 	t.Errorf("cart_abandoned should get recent value, got %v, want %v", toUser.MergeableFields["cart_abandoned"], recentDate)
		// }

		// TODO
		// if toUser.CustomColumns["app_test_num"].(int) != fromUser.CustomColumns["app_test_num"].(int) {
		// 	t.Errorf("app_test_num should be equal, got %v, want %v", toUser.CustomColumns["app_test_num"].(int), fromUser.CustomColumns["app_test_num"].(int))
		// }

		// if !toUser.FieldsTimestamp["app_test_num"].Equal(recentDate) {
		// 	t.Errorf("app_test_num should get recent value, got %v, want %v", toUser.FieldsTimestamp["app_test_num"], recentDate)
		// }

		// if toUser.CustomColumns["app_test_replaced"].(string) != fromUser.CustomColumns["app_test_replaced"].(string) {
		// 	t.Errorf("app_test_replaced should be equal, got %v, want %v", toUser.CustomColumns["app_test_replaced"].(string), fromUser.CustomColumns["app_test_replaced"].(string))
		// }
		// if !toUser.FieldsTimestamp["app_test_replaced"].Equal(recentDate) {
		// 	t.Errorf("app_test_replaced should get recent value, got %v, want %v", toUser.FieldsTimestamp["app_test_replaced"], recentDate)
		// }
	})
}
