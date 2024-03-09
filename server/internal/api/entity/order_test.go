package entity

import (
	"testing"
	"time"
)

func TestEntity_Order(t *testing.T) {

	t.Run("should merge order properly", func(t *testing.T) {

		orderID := "abc"
		userID := "def"
		domainID := "xxx"
		price := NewNullableInt64(Int64Ptr(10000))
		currency := "EUR"
		fromAt := time.Now().Add(-time.Hour * 24)
		updatedAt := time.Now()

		existing := NewOrder(orderID, userID, domainID, fromAt, nil, OrderItems{
			{
				ExternalID:        "item_ext_id",
				ProductExternalID: "ghi",
				Name:              "jkl",
				Price:             10000,
				Quantity:          1,
			},
		})
		existing.SetSubtotalPrice(price, fromAt)
		existing.SetTotalPrice(price, fromAt)
		existing.SetCurrency(&currency, fromAt)

		updated := NewOrder(orderID, userID, domainID, updatedAt, nil, OrderItems{
			{
				ExternalID:        "item_ext_id",
				ProductExternalID: "ghi",
				Name:              "jkl",
				Price:             20000,
				Quantity:          2,
			},
		})

		newPrice := NewNullableInt64(Int64Ptr(20000))
		updated.SetSessionID(StringPtr("session_id"), updatedAt)
		updated.SetSubtotalPrice(newPrice, updatedAt)
		updated.SetTotalPrice(newPrice, updatedAt)
		updated.SetCurrency(&currency, updatedAt)

		updatedFields := updated.MergeInto(existing)

		if len(updatedFields) != 3 {
			t.Errorf("expected 3 updated field, got %d", len(updatedFields))
		}

		if existing.SessionID == nil || *existing.SessionID != "session_id" {
			t.Errorf("expected session_id to be updated")
		}
	})
}
