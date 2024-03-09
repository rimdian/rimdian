package entity

import (
	"testing"
	"time"
)

func TestEntity_CartItem(t *testing.T) {

	t.Run("should merge item properly", func(t *testing.T) {

		itemID := "abc"
		userID := "def"
		cartID := "xxx"
		productExternalID := "ghi"
		updatedName := "jkl"
		fromAt := time.Now().Add(-time.Hour * 24)
		updatedAt := time.Now()

		existing := NewCartItem(itemID, "a", userID, cartID, productExternalID, fromAt, nil)
		updated := NewCartItem(itemID, updatedName, userID, cartID, productExternalID, fromAt, &updatedAt)

		updated.SetSKU(NewNullableString(StringPtr("sku")), updatedAt)
		updated.SetBrand(NewNullableString(StringPtr("brand")), updatedAt)
		updated.SetCategory(NewNullableString(StringPtr("category")), updatedAt)
		updated.SetVariantExternalID(NewNullableString(StringPtr("variant ext id")), updatedAt)
		updated.SetVariantTitle(NewNullableString(StringPtr("variant title")), updatedAt)
		updated.SetImageURL(NewNullableString(StringPtr("image url")), updatedAt)
		updated.SetPrice(10050, updatedAt)
		updated.SetQuantity(2, updatedAt)

		updatedFields := updated.MergeInto(existing)

		if len(updatedFields) != 9 {
			t.Errorf("expected 9 updated field, got %d", len(updatedFields))
		}

		if updatedFields[0].Field != "name" {
			t.Errorf("expected updated field to be name, got %s", updatedFields[0].Field)
		}

		if updatedFields[1].Field != "sku" {
			t.Errorf("expected updated field to be sku, got %s", updatedFields[0].Field)
		}

		if updatedFields[2].Field != "brand" {
			t.Errorf("expected updated field to be brand, got %s", updatedFields[0].Field)
		}

		if updatedFields[3].Field != "category" {
			t.Errorf("expected updated field to be category, got %s", updatedFields[0].Field)
		}

		if updatedFields[4].Field != "variant_external_id" {
			t.Errorf("expected updated field to be variant_external_id, got %s", updatedFields[0].Field)
		}

		if updatedFields[5].Field != "variant_title" {
			t.Errorf("expected updated field to be variant_title, got %s", updatedFields[0].Field)
		}

		if updatedFields[6].Field != "image_url" {
			t.Errorf("expected updated field to be image_url, got %s", updatedFields[0].Field)
		}

		if updatedFields[7].Field != "price" {
			t.Errorf("expected updated field to be price, got %s", updatedFields[0].Field)
		}

		if updatedFields[8].Field != "quantity" {
			t.Errorf("expected updated field to be quantity, got %s", updatedFields[0].Field)
		}
	})
}
