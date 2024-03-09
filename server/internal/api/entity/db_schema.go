package entity

// type DBTable struct {
// 	Name         string      `json:"name"`
// 	Columns      []*DBColumn `json:"columns"`
// 	StorageType  string      `json:"type"` // COLUMNSTORE | INMEMORY_ROWSTORE
// 	Rows         int64       `json:"rows"`
// 	DataLength   int64       `json:"dataLength"`
// 	MemoryBuffer int64       `json:"memoryBuffer"`
// 	Collation    string      `json:"collation"`
// 	IsView       bool        `json:"isView"`
// 	CreatedAt    string      `json:"createdAt"`
// }

// type DBColumn struct {
// 	Name         string `json:"name"`
// 	Type         string `json:"type"`
// 	Nullable     bool   `json:"nullable"`
// 	DefaultValue string `json:"string"`
// }

type TableStatus struct {
	TableName string `db:"TABLE_NAME"`
	Rows      int64  `db:"TotalRows"`
	MemoryUse int64  `db:"TotalMemoryUse"`
}

type TableInformationSchema struct {
	Name        string `db:"TABLE_NAME" json:"name"`
	Type        string `db:"TABLE_TYPE" json:"type"` // BASE TABLE | VIEW
	CreatedAt   string `db:"CREATE_TIME" json:"created_at"`
	StorageType string `db:"STORAGE_TYPE" json:"storage_type"` // COLUMNSTORE | INMEMORY_ROWSTORE

	// joined server side
	Rows      int64                      `json:"rows"`
	MemoryUse int64                      `json:"memory_use"`
	Columns   []*ColumnInformationSchema `json:"columns"`
}

type ColumnInformationSchema struct {
	TableName          string  `db:"TABLE_NAME" json:"-"`
	Name               string  `db:"COLUMN_NAME" json:"name"`
	Position           int     `db:"ORDINAL_POSITION" json:"position"`
	DefaultValue       *string `db:"COLUMN_DEFAULT" json:"default_value"`
	Nullable           string  `db:"IS_NULLABLE" json:"nullable"`
	ColumnType         string  `db:"COLUMN_TYPE" json:"column_type"`
	DataType           string  `db:"DATA_TYPE" json:"data_type"` // varchar | datetime | timestamp | text | smallint | tinyint
	CharacterMaxLength *int64  `db:"CHARACTER_MAXIMUM_LENGTH" json:"character_max_length"`
	NumericPrecision   *int64  `db:"NUMERIC_PRECISION" json:"numeric_precision"`
	CharacterSet       *string `db:"CHARACTER_SET_NAME" json:"character_set"`
	Collation          *string `db:"COLLATION_NAME" json:"collation"`
	ColumnKey          string  `db:"COLUMN_KEY" json:"column_key"`
	Extra              string  `db:"EXTRA" json:"extra"` // ex: ON UPDATE CURRENT_TIMESTAMP
}
