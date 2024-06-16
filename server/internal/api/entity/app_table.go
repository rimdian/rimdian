package entity

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/rotisserie/eris"
)

var (
	StorageTypeRowstore    = "rowstore"
	StorageTypeColumnstore = "columnstore"

	CustomColumnNameRegex = "^app_([a-z0-9])+_([a-z0-9])+$"
	TableIndexRegex       = "^([a-z0-9_])+$"
	// AppExternalIDIndex    = TableIndex{
	// 	Name:    "external_id_hash",
	// 	Columns: []string{"external_id"},
	// }

	ErrAppTableNotFound           = eris.New("app table not found")
	ErrAppTableInvalidName        = eris.New("app table table name is not valid [a-z0-9_]")
	ErrAppTableStorageTypeInvalid = eris.New("app table storage type is not valid")

	ErrAppTableAlreadyExists            = eris.New("app table key already exists")
	ErrAppTableColumnsRequired          = eris.New("app table columns required")
	ErrAppTableShardKeyRequired         = eris.New("app table shard key required")
	ErrAppTableUniqueKeyRequired        = eris.New("app table unique key required")
	ErrAppTableSortKeyRequired          = eris.New("app table sort key required")
	ErrAppTableShardKeyNotFound         = eris.New("app table shard key not found")
	ErrAppTableUniqueKeyNotFound        = eris.New("app table unique key not found")
	ErrAppTableSortKeyNotFound          = eris.New("app table sort key not found")
	ErrAppTableTimeSeriesColumnNotFound = eris.New("app table time series column not found")

	ErrAppTableIndexNameRequired    = eris.New("app table index name required")
	ErrAppTableIndexNameInvalid     = eris.New("app table index name is not valid, should start with appid_")
	ErrAppTableIndexColumnsRequired = eris.New("app table index columns required")
	ErrAppTableIndexColumnNotFound  = eris.New("app table index column not found")

	TableColumnTypeBoolean   = "boolean"
	TableColumnTypeNumber    = "number"
	TableColumnTypeDate      = "date"
	TableColumnTypeDatetime  = "datetime"
	TableColumnTypeTimestamp = "timestamp"
	TableColumnTypeVarchar   = "varchar"
	TableColumnTypeLongText  = "longtext"
	TableColumnTypeJSON      = "json"

	ReservedColumns = []string{
		"id",
		"external_id",
		"user_id",
		"created_at",
		"updated_at",
		"merged_from_user_id",
		"fields_timestamp",
		"db_created_at",
		"db_updated_at",
	}

	ComputedColumns = []string{
		"id",
		"user_id",
		"merged_from_user_id",
		"fields_timestamp",
		"db_created_at",
		"db_updated_at",
	}

	AppReservedTableColumns = []*TableColumn{

		{
			Name:            "id",
			Type:            ColumnTypeVarchar,
			Size:            Int64Ptr(64),
			Description:     StringPtr("ID (sha1 of external_id)"),
			IsRequired:      true,
			HideInAnalytics: true,
		},
		{
			Name:            "external_id",
			Type:            ColumnTypeVarchar,
			Size:            Int64Ptr(256),
			Description:     StringPtr("External ID"),
			IsRequired:      true,
			HideInAnalytics: true,
		},
		{
			Name:        "created_at",
			Type:        ColumnTypeDatetime,
			Description: StringPtr("Created at"),
			IsRequired:  true,
		},
		{
			Name:            "user_id",
			Type:            ColumnTypeVarchar,
			Size:            Int64Ptr(64),
			Description:     StringPtr("User ID"),
			IsRequired:      true,
			HideInAnalytics: true,
		},
		{
			Name:            "merged_from_user_id",
			Type:            ColumnTypeVarchar,
			Size:            Int64Ptr(64),
			IsRequired:      false,
			Description:     StringPtr("Merged from user ID"),
			HideInAnalytics: true,
		},
		{
			Name:            "fields_timestamp",
			Type:            ColumnTypeJSON,
			IsRequired:      true,
			Description:     StringPtr("Fields timestamp"),
			HideInAnalytics: true,
		},
		{
			Name:             "db_created_at",
			Type:             ColumnTypeTimestamp,
			Size:             Int64Ptr(6), // microsecond
			Description:      StringPtr("DB created at"),
			IsRequired:       true,
			DefaultTimestamp: StringPtr("CURRENT_TIMESTAMP(6)"),
			HideInAnalytics:  true,
		},
		{
			Name:             "db_updated_at",
			Type:             ColumnTypeTimestamp,
			Description:      StringPtr("DB updated at"),
			IsRequired:       true,
			DefaultTimestamp: StringPtr("CURRENT_TIMESTAMP"),
			ExtraDefinition:  StringPtr("ON UPDATE CURRENT_TIMESTAMP"),
			HideInAnalytics:  true,
		},
	}

	ErrTableColumnTypeInvalid          = eris.New("table column type is not valid")
	ErrTableColumnInvalidName          = eris.New("table column name is not valid")
	ErrTableColumnAlreadyExists        = eris.New("table column key already exists")
	ErrTableColumnInvalidColumnLength  = eris.New("table column name should contain 32 characters max")
	ErrTableColumnInvalidTextSize      = eris.New("table column invalid text size")
	ErrTableColumnTypeNotImplemented   = eris.New("table column type is not implemented")
	ErrTableColumnDefaultValueNotValid = eris.New("table column default value is not valid")
)

type AppTablesManifest []*AppTableManifest

func (x *AppTablesManifest) Scan(val interface{}) error {

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

func (x AppTablesManifest) Value() (driver.Value, error) {
	return json.Marshal(x)
}

type AppTableManifest struct {
	Name             string       `json:"name"`
	StorageType      string       `json:"storage_type"`
	Description      *string      `json:"description,omitempty"`
	Columns          TableColumns `json:"columns"`
	ShardKey         []string     `json:"shard_key"`
	UniqueKey        []string     `json:"unique_key"`
	SortKey          []string     `json:"sort_key"`
	TimeSeriesColumn *string      `json:"timeseries_column"`
	// Joins            TableJoins   `json:"joins"`
	Indexes TableIndexes `json:"indexes"`
}

func (from *AppTableManifest) HasSameDefinition(to *AppTableManifest) bool {
	if from.Name != to.Name {
		return false
	}

	if from.StorageType != to.StorageType {
		return false
	}

	if len(from.Columns) != len(to.Columns) {
		return false
	}

	for i, col := range from.Columns {
		if !col.HasSameDefinition(*to.Columns[i]) {
			return false
		}
	}

	if len(from.ShardKey) != len(to.ShardKey) {
		return false
	}

	for i, key := range from.ShardKey {
		if key != to.ShardKey[i] {
			return false
		}
	}

	if len(from.UniqueKey) != len(to.UniqueKey) {
		return false
	}

	for i, key := range from.UniqueKey {
		if key != to.UniqueKey[i] {
			return false
		}
	}

	if len(from.SortKey) != len(to.SortKey) {
		return false
	}

	for i, key := range from.SortKey {
		if key != to.SortKey[i] {
			return false
		}
	}

	if from.TimeSeriesColumn != nil && to.TimeSeriesColumn != nil {
		if *from.TimeSeriesColumn != *to.TimeSeriesColumn {
			return false
		}
	}

	if len(from.Indexes) != len(to.Indexes) {
		return false
	}

	for i, index := range from.Indexes {
		if !index.HasSameDefinition(*to.Indexes[i]) {
			return false
		}
	}

	return true
}

func (t *AppTableManifest) HasUserColumn() bool {

	for _, col := range t.Columns {
		if col.Name == "user_id" {
			return true
		}
	}

	return false
}

func (t *AppTableManifest) Validate(appID string, installedApps InstalledApps) error {

	t.Name = strings.TrimSpace(t.Name)

	if !strings.HasPrefix(t.Name, appID+"_") {
		return eris.Wrapf(ErrAppTableInvalidName, "got %v instead of %v", t.Name, appID+"_")
	}

	// check that table name has only 2 underscores
	if strings.Count(t.Name, "_") < 2 {
		return eris.Wrapf(ErrAppTableInvalidName, "got %v", t.Name)
	}

	re := regexp.MustCompile("^([a-z0-9_])+$")

	if !re.MatchString(t.Name) {
		return ErrAppTableInvalidName
	}

	// columnstore only to the moment...
	if t.StorageType != StorageTypeColumnstore {
		t.StorageType = StorageTypeColumnstore
	}

	if t.Columns == nil || len(t.Columns) == 0 {
		return ErrAppTableColumnsRequired
	}

	// remove and add reserved columns
	cleanedColumns := []*TableColumn{}
	hasUserID := false

	for _, col := range t.Columns {
		if col.Name == "user_id" {
			hasUserID = true
		}
		if govalidator.IsIn(col.Name, ReservedColumns...) {
			continue
		}
		cleanedColumns = append(cleanedColumns, col)
	}

	t.Columns = cleanedColumns

	// add reserved columns
	for _, col := range AppReservedTableColumns {
		if !hasUserID && (col.Name == "user_id" || col.Name == "merged_from_user_id") {
			continue
		}
		t.Columns = append(t.Columns, col)
	}

	if t.ShardKey == nil || len(t.ShardKey) == 0 {
		return ErrAppTableShardKeyRequired
	}

	if t.UniqueKey == nil || len(t.UniqueKey) == 0 {
		return ErrAppTableUniqueKeyRequired
	}

	if t.SortKey == nil || len(t.SortKey) == 0 {
		return ErrAppTableSortKeyRequired
	}

	shardKeysFound := []string{}
	uniqueKeysFound := []string{}
	sortKeysFound := []string{}
	timeSeriesColumnFound := false

	if t.Columns != nil && len(t.Columns) > 0 {
		for _, col := range t.Columns {

			// validate basics
			if err := col.Validate(); err != nil {
				return err
			}

			// extract shard keys
			for _, shardKey := range t.ShardKey {
				if shardKey == col.Name {
					shardKeysFound = append(shardKeysFound, shardKey)
				}
			}

			// extract unique keys
			for _, uniqueKey := range t.UniqueKey {
				if uniqueKey == col.Name {
					uniqueKeysFound = append(uniqueKeysFound, uniqueKey)
				}
			}

			// extract sort keys
			for _, sortKey := range t.SortKey {
				if sortKey == col.Name {
					sortKeysFound = append(sortKeysFound, sortKey)
				}
			}

			if t.TimeSeriesColumn != nil && *t.TimeSeriesColumn == col.Name {
				timeSeriesColumnFound = true
			}
		}
	}

	if len(shardKeysFound) != len(t.ShardKey) {
		return ErrAppTableShardKeyNotFound
	}

	if len(uniqueKeysFound) != len(t.UniqueKey) {
		return ErrAppTableUniqueKeyNotFound
	}

	if len(sortKeysFound) != len(t.SortKey) {
		return ErrAppTableSortKeyNotFound
	}

	if t.TimeSeriesColumn != nil && !timeSeriesColumnFound {
		return ErrAppTableTimeSeriesColumnNotFound
	}

	// // TODO: verify joins with external tables/columns exist

	// if t.Joins == nil {
	// 	t.Joins = []*TableJoin{}
	// }

	// // add user_id join
	// if hasUserID {
	// 	t.Joins = append(t.Joins, &TableJoin{
	// 		ExternalTable:  "user",
	// 		Relationship:   "many_to_one",
	// 		LocalColumn:    "user_id",
	// 		ExternalColumn: "id",
	// 	})
	// }

	// validate indexes
	if t.Indexes == nil {
		t.Indexes = []*TableIndex{}
	}

	cleanedIndexes := []*TableIndex{}

	// clean reserved indexes
	// for _, index := range t.Indexes {
	// 	if index.Name == AppExternalIDIndex.Name {
	// 		continue
	// 	}
	// 	cleanedIndexes = append(cleanedIndexes, index)
	// }

	t.Indexes = cleanedIndexes

	// add external_id mandatory index
	// t.Indexes = append(t.Indexes, &AppExternalIDIndex)

	if err := t.Indexes.Validate(t.Columns); err != nil {
		return err
	}

	return nil
}

type TableJoins []*TableJoin

type TableJoin struct {
	ExternalTable string `json:"external_table"`
	// https://cube.dev/docs/schema/reference/joins
	Relationship   string `json:"relationship"` // one_to_one, one_to_many, many_to_one
	LocalColumn    string `json:"local_column"`
	ExternalColumn string `json:"external_column"`
}

type TableIndexes []*TableIndex

func (t TableIndexes) Validate(allColumns TableColumns) error {
	for _, index := range t {
		if err := index.Validate(allColumns); err != nil {
			return err
		}
	}

	return nil
}

func (t TableIndexes) ToDDL() string {
	var indexes = []string{}
	for _, index := range t {
		indexes = append(indexes, index.ToDDL())
	}

	if len(indexes) == 0 {
		return ""
	}

	return "," + strings.Join(indexes, ",")
}

type TableIndex struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
}

func (from *TableIndex) HasSameDefinition(to TableIndex) bool {
	if from.Name != to.Name {
		return false
	}

	if len(from.Columns) != len(to.Columns) {
		return false
	}

	for i, col := range from.Columns {
		if col != to.Columns[i] {
			return false
		}
	}

	return true
}

func (t *TableIndex) ToDDL() string {
	return fmt.Sprintf("KEY %v(%v) USING HASH", t.Name, strings.Join(t.Columns, ","))
}

func (t *TableIndex) Validate(allColumns TableColumns) error {
	t.Name = strings.TrimSpace(t.Name)

	if t.Name == "" {
		return ErrAppTableIndexNameRequired
	}

	// verify if name matches the pattern
	re := regexp.MustCompile(TableIndexRegex)

	if !re.MatchString(t.Name) {
		return eris.Wrapf(ErrAppTableIndexNameInvalid, "got %v instead of %v", t.Name, TableIndexRegex)
	}

	if t.Columns == nil || len(t.Columns) == 0 {
		return ErrAppTableIndexColumnsRequired
	}

	// verify that all columns exist
	for _, col := range t.Columns {
		if !allColumns.HasColumn(col) {
			return eris.Wrapf(ErrAppTableIndexColumnNotFound, "got %v", col)
		}
	}

	return nil
}
