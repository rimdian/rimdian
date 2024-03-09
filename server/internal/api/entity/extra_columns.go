package entity

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/rotisserie/eris"
)

var (
	ColumnTypeBoolean   = "boolean"
	ColumnTypeNumber    = "number"
	ColumnTypeDate      = "date"
	ColumnTypeDatetime  = "datetime"
	ColumnTypeTimestamp = "timestamp"
	ColumnTypeVarchar   = "varchar"
	ColumnTypeLongText  = "longtext"
	ColumnTypeJSON      = "json"

	ErrExtraColumnTableInvalid  = eris.New("extra column table name is not valid")
	ErrExtraColumnAlreadyExists = eris.New("extra column key already exists")
	ErrExtraColumnNotFound      = eris.New("extra column not found")
	ErrTableColumnNameInvalid   = eris.New("extra column name should start with: app_appname_field")
)

type ExtraColumnManifest struct {
	// Table   string       `json:"table"` // =
	Kind    string       `json:"kind"` // user, session...
	Columns TableColumns `json:"columns"`
	// enventually in the future: indexes, joins...
}

// map of table and columns (ex: {session: [col1, col2]})
type ExtraColumnsManifest []*ExtraColumnManifest

func (x *ExtraColumnsManifest) Scan(val interface{}) error {

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

func (x ExtraColumnsManifest) Value() (driver.Value, error) {
	return json.Marshal(x)
}

func (x ExtraColumnManifest) GetTable() string {
	return x.Kind
}

type TableColumns []*TableColumn

func (x *TableColumns) HasColumn(name string) bool {
	for _, col := range *x {
		if col.Name == name {
			return true
		}
	}
	return false
}

type TableColumn struct {
	Name string `json:"name"`
	Type string `json:"type"` // boolean | number (=float) | date | datetime | timestamp | varchar | longtext | json. (notNull, default)
	Size *int64 `json:"size"`
	// NotNull          bool                    `json:"not_null"`
	IsRequired       bool                    `json:"is_required"`
	Description      *string                 `json:"description,omitempty"`
	DefaultBoolean   *bool                   `json:"default_boolean,omitempty"`
	DefaultNumber    *float64                `json:"default_number,omitempty"`
	DefaultDate      *string                 `json:"default_date,omitempty"`
	DefaultDateTime  *string                 `json:"default_datetime,omitempty"`
	DefaultTimestamp *string                 `json:"default_timestamp,omitempty"` // CURRENT_TIMESTAMP
	DefaultText      *string                 `json:"default_string,omitempty"`
	DefaultJSON      *map[string]interface{} `json:"default_json,omitempty"`
	ExtraDefinition  *string                 `json:"extra_definition,omitempty"`  // ON UPDATE CURRENT_TIMESTAMP
	HideInAnalytics  bool                    `json:"hide_in_analytics,omitempty"` // hide dimension in exported Cube analytics

	// enriched server-side
	CreatedAt time.Time  `json:"created_at,omitempty"`
	UpdatedAt time.Time  `json:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

func (col *TableColumn) Validate() error {

	col.Name = strings.TrimSpace(col.Name)
	col.Type = strings.TrimSpace(col.Type)

	nameRegex := "^([a-z0-9_])+$"

	re := regexp.MustCompile(nameRegex)

	if !re.MatchString(col.Name) {
		return ErrTableColumnInvalidName
	}

	if len(col.Name) > 48 {
		return ErrTableColumnInvalidColumnLength
	}

	if !govalidator.IsIn(col.Type,
		TableColumnTypeBoolean,
		TableColumnTypeNumber,
		TableColumnTypeDate,
		TableColumnTypeDatetime,
		TableColumnTypeTimestamp,
		TableColumnTypeVarchar,
		TableColumnTypeLongText,
		TableColumnTypeJSON) {
		return ErrTableColumnTypeInvalid
	}

	switch col.Type {
	case TableColumnTypeBoolean:
		// if !col.NotNull && col.DefaultBoolean == nil {
		// 	return ErrTableColumnDefaultValueRequired
		// }
	case TableColumnTypeNumber:
		// if col.DefaultNumber == nil {
		// 	return ErrTableColumnDefaultValueRequired
		// }
	case TableColumnTypeDate:
		if col.DefaultDate != nil {
			if _, err := time.Parse("2006-01-02", *col.DefaultDate); err != nil {
				return eris.Wrapf(ErrTableColumnDefaultValueNotValid, "column: %v", col.Name)
			}
		}
	case TableColumnTypeDatetime:
		if col.DefaultDate != nil {
			if _, err := time.Parse("2006-01-02 15:04:05", *col.DefaultDateTime); err != nil {
				return eris.Wrapf(ErrTableColumnDefaultValueNotValid, "column: %v", col.Name)
			}
		}
	case TableColumnTypeTimestamp:
		if col.DefaultTimestamp != nil && (*col.DefaultTimestamp != "CURRENT_TIMESTAMP" && *col.DefaultTimestamp != "CURRENT_TIMESTAMP(6)") {
			return eris.Wrapf(ErrTableColumnDefaultValueNotValid, "column: %v", col.Name)
			// if !govalidator.IsUnixTime(strconv.Itoa(*col.DefaultTimestamp)) {
			// }
		}
		if col.ExtraDefinition != nil && *col.ExtraDefinition != "ON UPDATE CURRENT_TIMESTAMP" && *col.ExtraDefinition != "ON UPDATE CURRENT_TIMESTAMP(6)" {
			return eris.Wrapf(ErrTableColumnDefaultValueNotValid, "column: %v", col.Name)
		}
	case TableColumnTypeVarchar:
		if col.Size == nil || *col.Size < 1 || *col.Size > 21845 {
			return ErrTableColumnInvalidTextSize
		}
		if col.DefaultText != nil {
			trimed := strings.TrimSpace(*col.DefaultText)
			col.DefaultText = &trimed
		}
	case TableColumnTypeLongText:
		// no default allowed for longtext
	case TableColumnTypeJSON:
	default:
		return eris.Errorf("table column not null type not implemented for %v", col.Type)
	}

	return nil
}
