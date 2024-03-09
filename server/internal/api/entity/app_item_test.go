package entity

import (
	"encoding/json"
	"log"
	"testing"
	"time"
)

var (
	// testJSON = M{
	// 	"json_string":  "string",
	// 	"json_number":  123.456,
	// 	"json_boolean": true,
	// 	"json_null":    nil,
	// 	"json_array":   []string{"a", "b", "c"},
	// 	"json_object": M{
	// 		"a": 1,
	// 		"b": 2,
	// 	},
	// }
	// testJSONByte, _ = json.Marshal(testJSON)
	testJSON     = `{"json_null": null, "json_array": ["a", "b", "c"], "json_number": 123.456, "json_object": {"a": 1, "b": 2}, "json_string": "string", "json_boolean": true}`
	testJSONByte = []byte(testJSON)
)

func TestEntity_AppItem(t *testing.T) {

	t.Run("AppItem objects should merge properly", func(t *testing.T) {

		// create 2 app_item objects
		// the first one is the existing one
		// the second one is the one we want to merge into the first one

		createdAt := time.Now().AddDate(-1, 0, 0) // last year
		externalID := "test"
		userID := "1"
		kind := "app_item_test"

		existing := NewAppItem(kind, externalID, userID, createdAt, nil)

		existing.Fields = AppItemFields{
			"string_value": &AppItemField{
				Name: "string_value",
				Type: ColumnTypeVarchar,
				StringValue: NullableString{
					IsNull: false,
					String: "existing",
				},
			},
			"nullable_string": &AppItemField{
				Name: "nullable_string",
				Type: ColumnTypeVarchar,
				StringValue: NullableString{
					IsNull: true,
				},
			},
			"nullable_json": &AppItemField{
				Name: "nullable_json",
				Type: ColumnTypeJSON,
				JSONValue: NullableJSON{
					IsNull: true,
				},
			},
			"identical_string": &AppItemField{
				Name: "identical_string",
				Type: ColumnTypeVarchar,
				StringValue: NullableString{
					IsNull: false,
					String: "identical_string",
				},
			},
			"identical_number": &AppItemField{
				Name: "identical_number",
				Type: ColumnTypeNumber,
				Float64Value: NullableFloat64{
					IsNull:  false,
					Float64: 123.456,
				},
			},
		}

		// the new app_item object
		updatedAt := time.Now() // today
		updated := NewAppItem(kind, externalID, userID, createdAt, &updatedAt)

		updated.Fields = AppItemFields{
			"string_value": &AppItemField{
				Name: "string_value",
				Type: ColumnTypeVarchar,
				StringValue: NullableString{
					IsNull: false,
					String: "updated", // updated value
				},
			},
			"nullable_string": &AppItemField{
				Name: "nullable_string",
				Type: ColumnTypeVarchar,
				StringValue: NullableString{
					IsNull: false,
					String: "updated", // updated value
				},
			},
			"nullable_json": &AppItemField{
				Name: "nullable_json",
				Type: ColumnTypeJSON,
				JSONValue: NullableJSON{
					IsNull: false,
					JSON:   testJSONByte,
				},
			},
			"identical_string": &AppItemField{
				Name: "identical_string",
				Type: ColumnTypeVarchar,
				StringValue: NullableString{
					IsNull: false,
					String: "identical_string",
				},
			},
			"identical_number": &AppItemField{
				Name: "identical_number",
				Type: ColumnTypeNumber,
				Float64Value: NullableFloat64{
					IsNull:  false,
					Float64: 123.456,
				},
			},
		}

		// merge the 2 objects
		updatedFields := updated.MergeInto(existing)

		for k, v := range updatedFields {
			log.Printf("updatedFields[%d] = %+v\n", k, v)
		}

		if len(updatedFields) != 3 {
			t.Errorf("got %v, want %v", len(updatedFields), 3)
		}

		for _, field := range updatedFields {
			if field.Field == "string_value" {

				if field.PrevValue.(string) != "existing" {
					t.Errorf("got %v, want %v", field.PrevValue, "existing")
				}

				if field.NewValue.(string) != "updated" {
					t.Errorf("got %v, want %v", field.NewValue, "updated")
				}
			}

			if field.Field == "nullable_string" {
				if field.PrevValue != nil {
					t.Errorf("got %v, want %v", field.PrevValue, "nil")
				}

				if field.NewValue.(string) != "updated" {
					t.Errorf("got %v, want %v", field.NewValue, "updated")
				}
			}

			if field.Field == "nullable_json" {
				if field.PrevValue != nil {
					t.Errorf("got %v, want %v", field.PrevValue, "nil")
				}

				// marhsal the json value
				jsonValue, _ := json.Marshal(field.NewValue)

				if AreEqualJSON(jsonValue, testJSONByte) == false {
					t.Errorf("got %v, want %v", string(jsonValue), len(testJSON))
				}
			}
		}

		if string(existing.Fields["nullable_json"].JSONValue.JSON) != testJSON {
			t.Errorf("got %v, want %v", string(existing.Fields["nullable_json"].JSONValue.JSON), testJSON)
		}

	})
}
