package entity

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

// Problem with golang json unmarshal : it doesnt differenciate zero-value/null/undefined fields
// These types detect fields existance and "null" values
// They implement DB Scan/Value & JSON marshal unmarshall

// Use-case:
// 1. we parse JSON with eventual "null" values into a struct.
// 2. Then we only insert/update DB fields that have Exists == true.
// 3. When we retrieve DB objects, all structs fields will have Exists=true, and then we can encode them into JSON.

// Warning:
// enconding a non-existing field (Exists=false) into JSON will throw an error "nullable type undefined"
// it is mandatory to retrieve a full object from DB before encoding it into JSON to avoid errors

func GenerateUUID() (str string, err error) {
	id, err := uuid.NewRandom()
	return id.String(), err
}

var nullBytes = []byte("null")

type M = map[string]interface{}

// var ErrNullableTypeUndefined = eris.New("nullable type undefined")

// optimist marshal used to create internal import payloads for app_item
func ToJSON(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		// log but dont panic
		log.Println(err)
	}
	return b
}

func AreEqualJSON(s1, s2 []byte) bool {
	var o1 interface{}
	var o2 interface{}

	err1 := json.Unmarshal(s1, &o1)
	err2 := json.Unmarshal(s2, &o2)

	// if we have errors, we compare error strings
	if err1 != nil || err2 != nil {
		if err1 != nil && err2 != nil {
			return err1.Error() == err2.Error()
		}
		return false
	}
	return reflect.DeepEqual(o1, o2)
}

// STRING

func NewNullableString(value *string) (result *NullableString) {
	result = &NullableString{}
	result.SetValue(value)
	return
}

func NullableStringToInterface(x *NullableString) *string {
	if x == nil || x.IsNull {
		return nil
	}
	return &x.String
}

type NullableString struct {
	IsNull bool   `json:"is_null"`
	String string `json:"string"`
}

func (n *NullableString) SetValue(value *string) {
	if value == nil {
		n.IsNull = true
		n.String = "" // zero-value
	} else {
		n.IsNull = false
		n.String = *value
	}
}

func (ns NullableString) Value() (driver.Value, error) {
	if ns.IsNull {
		return nil, nil
	}
	return ns.String, nil
}

func (ns *NullableString) Scan(value interface{}) error {
	// use native sql scanner here
	dest := sql.NullString{}
	if err := dest.Scan(value); err != nil {
		return err
	}
	ns.IsNull = true
	ns.String = dest.String
	if dest.Valid {
		ns.IsNull = false
	}
	return nil
}

func (s *NullableString) UnmarshalJSON(data []byte) error {

	if bytes.Equal(data, nullBytes) {
		s.IsNull = true
		return nil
	}

	if err := json.Unmarshal(data, &s.String); err != nil {
		return fmt.Errorf("null: couldn't unmarshal JSON: %w", err)
	}

	s.IsNull = false
	return nil
}

func (s NullableString) MarshalJSON() ([]byte, error) {
	if s.IsNull {
		return json.Marshal(nil)
	}
	return json.Marshal(s.String)
}

// INT 64

type NullableInt64 struct {
	IsNull bool  `json:"is_null"`
	Int64  int64 `json:"int64"`
}

func NewNullableInt64(value *int64) (result *NullableInt64) {
	result = &NullableInt64{}
	result.SetValue(value)
	return
}

func NullableInt64ToInterface(x *NullableInt64) *int64 {
	if x == nil || x.IsNull {
		return nil
	}
	return &x.Int64
}

func (n *NullableInt64) SetValue(value *int64) {
	if value == nil {
		n.IsNull = true
		n.Int64 = 0 // zero-value
	} else {
		n.IsNull = false
		n.Int64 = *value
	}
}

func (n NullableInt64) Value() (driver.Value, error) {
	if n.IsNull {
		return nil, nil
	}
	return n.Int64, nil
}

func (n *NullableInt64) Scan(value interface{}) error {
	// use native sql scanner here
	dest := sql.NullInt64{}
	if err := dest.Scan(value); err != nil {
		return err
	}
	n.IsNull = true
	n.Int64 = dest.Int64
	if dest.Valid {
		n.IsNull = false
	}
	return nil
}

func (i *NullableInt64) UnmarshalJSON(data []byte) error {

	if bytes.Equal(data, nullBytes) {
		i.IsNull = true
		return nil
	}

	if err := json.Unmarshal(data, &i.Int64); err != nil {
		var typeError *json.UnmarshalTypeError
		if errors.As(err, &typeError) {
			// special case: accept string input
			if typeError.Value != "string" {
				return fmt.Errorf("null: JSON input is invalid type (need int or string): %w", err)
			}
			var str string
			if err := json.Unmarshal(data, &str); err != nil {
				return fmt.Errorf("null: couldn't unmarshal number string: %w", err)
			}
			n, err := strconv.ParseInt(str, 10, 64)
			if err != nil {
				return fmt.Errorf("null: couldn't convert string to int: %w", err)
			}
			i.Int64 = n
			i.IsNull = false
			return nil
		}
		return fmt.Errorf("null: couldn't unmarshal JSON: %w", err)
	}

	i.IsNull = false
	return nil
}

func (i NullableInt64) MarshalJSON() ([]byte, error) {
	if i.IsNull {
		return json.Marshal(nil)
	}
	return json.Marshal(i.Int64)
}

// FLOAT 64

type NullableFloat64 struct {
	IsNull  bool    `json:"is_null"`
	Float64 float64 `json:"float64"`
}

func NewNullableFloat64(value *float64) (result *NullableFloat64) {
	result = &NullableFloat64{}
	result.SetValue(value)
	return
}

func NullableFloat64ToInterface(x *NullableFloat64) *float64 {
	if x == nil || x.IsNull {
		return nil
	}
	return &x.Float64
}

func (n *NullableFloat64) SetValue(value *float64) {
	if value == nil {
		n.IsNull = true
		n.Float64 = 0
	} else {
		n.IsNull = false
		n.Float64 = *value
	}
}

func (n NullableFloat64) Value() (driver.Value, error) {
	if n.IsNull {
		return nil, nil
	}
	return n.Float64, nil
}

func (n *NullableFloat64) Scan(value interface{}) error {
	// use native sql scanner here
	dest := sql.NullFloat64{}
	if err := dest.Scan(value); err != nil {
		return err
	}
	n.IsNull = true
	n.Float64 = dest.Float64
	if dest.Valid {
		n.IsNull = false
	}
	return nil
}

func (f *NullableFloat64) UnmarshalJSON(data []byte) error {

	if bytes.Equal(data, nullBytes) {
		f.IsNull = true
		return nil
	}

	if err := json.Unmarshal(data, &f.Float64); err != nil {
		var typeError *json.UnmarshalTypeError
		if errors.As(err, &typeError) {
			// special case: accept string input
			if typeError.Value != "string" {
				return fmt.Errorf("null: JSON input is invalid type (need float or string): %w", err)
			}
			var str string
			if err := json.Unmarshal(data, &str); err != nil {
				return fmt.Errorf("null: couldn't unmarshal number string: %w", err)
			}
			n, err := strconv.ParseFloat(str, 64)
			if err != nil {
				return fmt.Errorf("null: couldn't convert string to float: %w", err)
			}
			f.Float64 = n
			f.IsNull = false
			return nil
		}
		return fmt.Errorf("null: couldn't unmarshal JSON: %w", err)
	}

	f.IsNull = false
	return nil
}

func (f NullableFloat64) MarshalJSON() ([]byte, error) {
	if f.IsNull {
		return json.Marshal(nil)
	}
	return json.Marshal(f.Float64)
}

// BOOLEAN

func NewNullableBool(value *bool) (result *NullableBool) {
	result = &NullableBool{}
	result.SetValue(value)
	return
}

func NullableBoolToInterface(x *NullableBool) *bool {
	if x == nil || x.IsNull {
		return nil
	}
	return &x.Bool
}

type NullableBool struct {
	IsNull bool `json:"is_null"`
	Bool   bool `json:"bool"`
}

func (n *NullableBool) SetValue(value *bool) {
	if value == nil {
		n.IsNull = true
		n.Bool = false // zero-value
	} else {
		n.IsNull = false
		n.Bool = *value
	}
}

func (n NullableBool) Value() (driver.Value, error) {
	if n.IsNull {
		return nil, nil
	}
	return n.Bool, nil
}

func (n *NullableBool) Scan(value interface{}) error {
	// use native sql scanner here
	dest := sql.NullBool{}
	if err := dest.Scan(value); err != nil {
		return err
	}
	n.IsNull = true
	n.Bool = dest.Bool
	if dest.Valid {
		n.IsNull = false
	}
	return nil
}

func (b *NullableBool) UnmarshalJSON(data []byte) error {

	if bytes.Equal(data, nullBytes) {
		b.IsNull = true
		return nil
	}

	if err := json.Unmarshal(data, &b.Bool); err != nil {
		return fmt.Errorf("null: couldn't unmarshal JSON: %w", err)
	}

	b.IsNull = false
	return nil
}

func (b NullableBool) MarshalJSON() ([]byte, error) {
	if b.IsNull {
		return json.Marshal(nil)
	}
	return json.Marshal(b.Bool)
}

// TIME

type NullableTime struct {
	IsNull bool      `json:"is_null"`
	Time   time.Time `json:"time"`
}

func NewNullableTime(value *time.Time) (result *NullableTime) {
	result = &NullableTime{}
	result.SetValue(value)
	return
}

func NullableTimeToInterface(x *NullableTime) *string {
	if x == nil || x.IsNull {
		return nil
	}
	date := x.Time.Format(time.RFC3339)
	return &date
}

func (n *NullableTime) SetValue(value *time.Time) {
	if value == nil {
		n.IsNull = true
		n.Time = time.Time{} // zero-value
	} else {
		n.IsNull = false
		n.Time = *value
	}
}

func (n NullableTime) Value() (driver.Value, error) {
	if n.IsNull {
		return nil, nil
	}
	return n.Time, nil
}

func (n *NullableTime) Scan(value interface{}) error {
	// use native sql scanner here
	dest := sql.NullTime{}
	if err := dest.Scan(value); err != nil {
		return err
	}
	n.IsNull = true
	n.Time = dest.Time
	if dest.Valid {
		n.IsNull = false
	}
	return nil
}

func (s *NullableTime) UnmarshalJSON(data []byte) error {

	if bytes.Equal(data, nullBytes) {
		s.IsNull = true
		return nil
	}

	if err := json.Unmarshal(data, &s.Time); err != nil {
		return fmt.Errorf("null: couldn't unmarshal JSON: %w", err)
	}

	s.IsNull = false
	return nil
}

func (s NullableTime) MarshalJSON() ([]byte, error) {
	if s.IsNull {
		return json.Marshal(nil)
	}
	return json.Marshal(s.Time)
}

// JSON

type NullableJSON struct {
	IsNull bool   `json:"is_null"`
	JSON   []byte `json:"json"`
}

func NullableJSONToInterface(x *NullableJSON) *[]byte {
	if x == nil || x.IsNull {
		return nil
	}
	return &x.JSON
}

func NewNullableJSON(value *[]byte) (result *NullableJSON) {
	result = &NullableJSON{}
	result.SetValue(value)
	return
}

func (n *NullableJSON) SetValue(value *[]byte) {
	if value == nil {
		n.IsNull = true
		n.JSON = []byte{} // zero-value
	} else {
		n.IsNull = false
		n.JSON = *value
	}
}

func (n NullableJSON) Value() (driver.Value, error) {
	if n.IsNull {
		return nil, nil
	}
	return n.JSON, nil
}

func (n *NullableJSON) Scan(value interface{}) error {
	n.IsNull = true
	n.JSON = []byte{}
	if value != nil {
		n.IsNull = false
		// VERY IMPORTANT: we need to clone the bytes here
		// The sql driver will reuse the same bytes RAM slots for future queries
		// Thank you St Antoine De Padoue for helping me find this bug
		n.JSON = bytes.Clone(value.([]byte))
	}
	return nil
}

func (s *NullableJSON) UnmarshalJSON(data []byte) error {
	// unmarshal is only called on existing JSON fields
	s.IsNull = false
	s.JSON = data

	if bytes.Equal(data, nullBytes) {
		s.IsNull = true
	}

	return nil
}

func (s NullableJSON) MarshalJSON() ([]byte, error) {
	if s.IsNull || len(s.JSON) == 0 {
		// s.IsNull = true // force isNull if no data
		return json.Marshal(nil)
	}

	return s.JSON, nil
}

type MapOfStrings map[string]string

func (x *MapOfStrings) Scan(val interface{}) error {

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

func (x MapOfStrings) Value() (driver.Value, error) {
	return json.Marshal(x)
}

func (x MapOfStrings) String() string {
	return fmt.Sprintf("%v", map[string]string(x))
}

type StringsArray []string

func (x *StringsArray) Scan(val interface{}) error {

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

func (x StringsArray) Value() (driver.Value, error) {
	return json.Marshal(x)
}

type ArrayOfInterfaces []interface{}

func (x *ArrayOfInterfaces) Scan(val interface{}) error {

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

func (x ArrayOfInterfaces) Value() (driver.Value, error) {
	return json.Marshal(x)
}

type MapOfInterfaces map[string]interface{}

func (x *MapOfInterfaces) Scan(val interface{}) error {

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

func (x MapOfInterfaces) Value() (driver.Value, error) {
	return json.Marshal(x)
}

func IsInterfaceNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}

// keeps track of the most recent timestamp of an updated field
// its used to only mutate fields to their most recent values
// in case of out-of-order incremental fields imports
type FieldsTimestamp map[string]time.Time

func (x *FieldsTimestamp) Scan(val interface{}) error {

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

func (x FieldsTimestamp) Value() (driver.Value, error) {
	return json.Marshal(x)
}

// unmarshal
func (x *FieldsTimestamp) UnmarshalJSON(data []byte) error {
	// use an alias
	type Alias FieldsTimestamp

	// unmarshal
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return eris.Wrap(err, "FieldsTimestamp.UnmarshalJSON")
		// if unmarshal fails, set to empty map
		// alias = map[string]time.Time{}
	}

	// set
	*x = FieldsTimestamp(alias)

	return nil
}

// compares fields updated_at time for provided key
// keeps the most recent one, or set the default time otherwise
// and returns infos about operation
func (x FieldsTimestamp) MergeTimeForRecentKey(key string, compareWithTime time.Time) (didInsert bool, didUpdate bool) {

	didInsert = false
	didUpdate = false

	if x == nil {
		x = FieldsTimestamp{}
	}

	// key doesnt exist yet, insert it
	if _, ok := x[key]; !ok {

		didInsert = true
		x[key] = compareWithTime

	} else if compareWithTime.After(x[key]) {
		// key exist, check if has a more recent key
		x[key] = compareWithTime
		didUpdate = true
	}

	return
}

// will preserve the oldest key
func (x FieldsTimestamp) MergeTimeForOldestKey(key string, compareWithTime time.Time) (didInsert bool, didUpdate bool) {

	didInsert = false
	didUpdate = false

	if x == nil {
		x = FieldsTimestamp{}
	}

	// key doesnt exist yet, insert it
	if _, ok := x[key]; !ok {

		didInsert = true
		x[key] = compareWithTime

	} else if compareWithTime.Before(x[key]) {
		// key exist, check if has a more recent key
		x[key] = compareWithTime
		didUpdate = true
	}

	return
}

func StringPtr(str string) *string {
	return &str
}

func IntPtr(i int) *int {
	return &i
}

func Int64Ptr(i int64) *int64 {
	return &i
}

func BoolPtr(b bool) *bool {
	return &b
}

func TimePtr(t time.Time) *time.Time {
	return &t
}

func StringPointerToInterface(x *string) interface{} {
	if x == nil {
		return nil
	}
	return *x
}
func Int64PointerToInterface(x *int64) interface{} {
	if x == nil {
		return nil
	}
	return *x
}
func Float64PointerToInterface(x *float64) interface{} {
	if x == nil {
		return nil
	}
	return *x
}
func BoolPointerToInterface(x *bool) interface{} {
	if x == nil {
		return nil
	}
	return *x
}
func TimePointerToInterface(x *time.Time) interface{} {
	if x == nil {
		return nil
	}
	// IMPORTANT: convert to UTC timezone to avoid comparing the same time with different timezones
	return x.UTC()
}

func GetInterfaceType(x interface{}) string {
	return reflect.TypeOf(x).String()
}

// reflect.DeepEqual doesn't work with time.Time that are equal but presented in different timezones
func FixedDeepEqual(a, b interface{}) bool {
	// get type
	aType := reflect.TypeOf(a)
	bType := reflect.TypeOf(b)

	if aType != bType {
		return false
	}

	if aType.String() == "time.Time" && bType.String() == "time.Time" {
		return a.(time.Time).Equal(b.(time.Time))
	}

	if aType.String() == "*time.Time" && bType.String() == "*time.Time" {
		aIsNil := reflect.ValueOf(a).IsNil()
		bIsNil := reflect.ValueOf(b).IsNil()
		if aIsNil && bIsNil {
			return true
		}
		if !aIsNil && !bIsNil {
			aTime := a.(*time.Time)
			bTime := b.(*time.Time)
			return aTime.Equal(*bTime)
		}
		return false
	}

	return reflect.DeepEqual(a, b)
}

func StringsEqual(s1, s2 *string) bool {
	if s1 == nil && s2 == nil {
		return true
	}
	if s1 == nil || s2 == nil {
		return false
	}
	return *s1 == *s2
}

func BoolEqual(s1, s2 *bool) bool {
	if s1 == nil && s2 == nil {
		return true
	}
	if s1 == nil || s2 == nil {
		return false
	}
	return *s1 == *s2
}

func Int64Equal(s1, s2 *int64) bool {
	if s1 == nil && s2 == nil {
		return true
	}
	if s1 == nil || s2 == nil {
		return false
	}
	return *s1 == *s2
}

func TimeEqual(x, y *time.Time) bool {
	if x == nil && y == nil {
		return true
	}
	if x == nil || y == nil {
		return false
	}
	return x.Equal(*y)
}

// returns the list of object db row columns, without:
// - computed fields
// - non-persisted fields
func GetNotComputedDBColumnsForObject(object interface{}, computedFields []string, customColumns []*TableColumn) (columns []string) {
	columns = []string{}
	ignoredColumns := []string{"", "-"}
	ignoredColumns = append(ignoredColumns, computedFields...)

	t := reflect.TypeOf(object)

	// Iterate over all available fields and read the tag value
	for i := 0; i < t.NumField(); i++ {
		// Get the field, and tag value
		tag := t.Field(i).Tag.Get("db")

		// ignore non-persisted column, and user_id + merged_from_user_id that will be replaced
		if govalidator.IsIn(tag, ignoredColumns...) {
			continue
		}

		columns = append(columns, tag)
	}

	// add eventual extra columns columns
	if len(customColumns) > 0 {
		for _, col := range customColumns {
			// ignore extra columns that are computed columns
			if col.ExtraDefinition != nil && strings.Contains(*col.ExtraDefinition, "AS ") {
				continue
			}
			columns = append(columns, col.Name)
		}
	}

	return columns
}
