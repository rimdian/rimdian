package entity

import (
	"encoding/json"
	"testing"
)

type testStruct struct {
	String    *NullableString `json:"string_value,omitempty"`
	Null      *NullableString `json:"null_value,omitempty"`
	Undefined *NullableString `json:"undefined_value,omitempty"`
}

// The NullableString type is used to compensate the lack
// of golang handling for JSON null values
// the golang json:omitempty will remove data with nil pointers
// but when a value is defined as null, we want to keep it in the resulting json.Marshal

func EntityTest_NullableString(t *testing.T) {

	t.Run("nullable string should convert to JSON properly", func(t *testing.T) {

		user := testStruct{
			String: &NullableString{String: "ok"},
			Null:   &NullableString{IsNull: true},
			// dont set the Undefined field here
		}

		b, err := json.Marshal(user)

		if err != nil {
			t.Fatalf("marshal err, got: %v", err)
		}

		// the Undefined struct key should not appear in the JSON
		want := "{\"string_value\":\"ok\",\"null_value\":null}"

		if string(b) != want {
			t.Errorf("got %v, want %v", string(b), want)
		}
	})
}
