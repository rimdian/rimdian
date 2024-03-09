package entity

import (
	"testing"
	"time"
)

func TestEntity_Device(t *testing.T) {

	t.Run("should merge item properly", func(t *testing.T) {

		externalID := "device_1"
		userID := "aaa"
		createdAt := time.Now().Add(-time.Hour * 24)
		updatedAt := time.Now()

		existing := NewDevice(externalID, userID, createdAt, createdAt)

		updated := NewDevice(externalID, userID, createdAt, updatedAt)
		updated.SetUserAgent(NewNullableString(StringPtr(`"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:109.0) Gecko/20100101 Firefox/119.0"`)), updatedAt)

		updatedFields := updated.MergeInto(existing)

		if len(updatedFields) != 1 {
			t.Errorf("expected 1 updated field, got %d", len(updatedFields))
		}

		if updatedFields[0].Field != "user_agent" {
			t.Errorf("expected updated field to be user_agent, got %s", updatedFields[0].Field)
		}
	})
}
