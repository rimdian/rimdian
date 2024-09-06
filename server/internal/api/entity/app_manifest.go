package entity

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/blastrain/vitess-sqlparser/sqlparser"
	"github.com/rotisserie/eris"
)

var (
	SQLSelect = "select"

	AppTestScheduledTask = "app_test_scheduledtask"

	ErrAppManifestIDInvalid                   = eris.New("app manifest id is not valid, it should be like: app_name or appx_name")
	ErrAppManifestNameInvalid                 = eris.New("app manifest name is not valid (25 chars max)")
	ErrAppManifestHomepageInvalid             = eris.New("app manifest homepage is not valid")
	ErrAppManifestAuthorInvalid               = eris.New("app manifest author is not valid")
	ErrAppManifestIconURLInvalid              = eris.New("app manifest icon url is not valid url")
	ErrAppManifestVersionInvalid              = eris.New("app manifest version is not a valid semver")
	ErrAppManifestUIEndpointInvalid           = eris.New("app manifest ui endpoint is not valid url")
	ErrAppManifestWebhookEndpointInvalid      = eris.New("app manifest webhook endpoint is not valid url")
	ErrAppManifestExtraColumnTableNameInvalid = eris.New("app manifest extra column table name is not valid (user, order, session...)")

	AppManifestTest = AppManifest{
		ID:               "app_test",
		Name:             "Test App",
		Homepage:         "https://docs.rimdian.com",
		Author:           "Rimdian",
		IconURL:          "https://cdn-eu.rimdian.com/images/rimdian-email.png",
		ShortDescription: "Test app",
		Description:      "Test app",
		Version:          "1.0.0",
		UIEndpoint:       "https://console.rimdian.com/apps/test",
		WebhookEndpoint:  "API_ENDPOINT/api/webhook.receiver",
		SQLAccess: AppSQLAccessManifest{
			TablesPermissions: []*TablePermission{
				{
					Table: "user",
					Read:  true,
				},
			},
		},
		DataHooks: DataHooksManifest{
			{
				ID:   "app_test_table_oncreate",
				Name: "Test on create",
				On:   DataHookKindOnSuccess,
				For:  []*DataHookFor{{Kind: "app_test_table", Action: "create"}},
			},
			{
				ID:   "app_test_table_onupdate",
				Name: "Test on update",
				On:   DataHookKindOnSuccess,
				For:  []*DataHookFor{{Kind: "app_test_table", Action: "update"}},
			},
			{
				ID:   "app_test_table_segment_enter",
				Name: "Test on segment enter",
				On:   DataHookKindOnSuccess,
				For:  []*DataHookFor{{Kind: "segment", Action: "enter"}},
			},
			{
				ID:   "app_test_table_segment_exit",
				Name: "Test on segment exit",
				On:   DataHookKindOnSuccess,
				For:  []*DataHookFor{{Kind: "segment", Action: "exit"}},
			},
		},
		Tasks: TasksManifest{
			{
				ID:              AppTestScheduledTask,
				Name:            "Test scheduled task",
				IsCron:          true,
				MinutesInterval: 15, // every 15 minutes
				OnMultipleExec:  OnMultipleExecAbortExisting,
			},
		},
		ExtraColumns: ExtraColumnsManifest{
			{
				Kind: ItemKindSession,
				Columns: TableColumns{
					{
						Name:        "app_test_string",
						Type:        ColumnTypeVarchar,
						Size:        Int64Ptr(60),
						IsRequired:  false,
						Description: StringPtr("test string field"),
					},
					{
						Name:        "app_test_bool",
						Type:        ColumnTypeBoolean,
						IsRequired:  false,
						Description: StringPtr("test bool field"),
					},
					{
						Name:        "app_test_number",
						Type:        ColumnTypeNumber,
						IsRequired:  false,
						Description: StringPtr("test number field"),
					},
					{
						Name:        "app_test_date",
						Type:        ColumnTypeDate,
						IsRequired:  false,
						Description: StringPtr("test date field"),
					},
					{
						Name:        "app_test_datetime",
						Type:        ColumnTypeDatetime,
						IsRequired:  false,
						Description: StringPtr("test datetime field"),
					},
					{
						Name:        "app_test_timestamp",
						Type:        ColumnTypeTimestamp,
						IsRequired:  false,
						Description: StringPtr("test timestamp field"),
					},
					{
						Name:        "app_test_longtext",
						Type:        ColumnTypeLongText,
						IsRequired:  false,
						Description: StringPtr("test longtext field"),
					},
					{
						Name:        "app_test_json",
						Type:        ColumnTypeJSON,
						IsRequired:  false,
						Description: StringPtr("test json field"),
					},
				},
			},
		},
		AppTables: AppTablesManifest{
			{
				Name:        "app_test_table",
				StorageType: StorageTypeColumnstore,
				Description: StringPtr("Test table"),
				ShardKey:    []string{"id"},
				UniqueKey:   []string{"id"},
				SortKey:     []string{"created_at"},
				Columns: []*TableColumn{
					AppReservedTableColumns[0], // id
					AppReservedTableColumns[1], // external_id
					AppReservedTableColumns[2], // created_at
					AppReservedTableColumns[3], // user_id
					AppReservedTableColumns[4], // merged_from_user_id
					AppReservedTableColumns[5], // fields_timestamp
					AppReservedTableColumns[6], // db_created_at
					AppReservedTableColumns[7], // db_updated_at
					{
						Name:        "required_varchar",
						Type:        ColumnTypeVarchar,
						Size:        Int64Ptr(256),
						IsRequired:  true,
						Description: StringPtr("varchar required"),
					},
					{
						Name:        "required_number",
						Type:        ColumnTypeNumber,
						Size:        Int64Ptr(128),
						IsRequired:  true,
						Description: StringPtr("number required"),
					},
					{
						Name:        "required_date",
						Type:        ColumnTypeDate,
						Size:        Int64Ptr(128),
						IsRequired:  true,
						Description: StringPtr("date required"),
					},
					{
						Name:        "required_timestamp",
						Type:        ColumnTypeTimestamp,
						Size:        Int64Ptr(128),
						IsRequired:  true,
						Description: StringPtr("timestamp required"),
					},
					{
						Name:        "required_boolean",
						Type:        ColumnTypeBoolean,
						IsRequired:  true,
						Description: StringPtr("boolean required"),
					},
					{
						Name:        "required_longtext",
						Type:        ColumnTypeLongText,
						IsRequired:  true,
						Description: StringPtr("longtext required"),
					},
					{
						Name:        "required_json",
						Type:        ColumnTypeJSON,
						IsRequired:  true,
						Description: StringPtr("json required"),
					},
					{
						Name:        "optional_varchar",
						Type:        ColumnTypeVarchar,
						Size:        Int64Ptr(256),
						IsRequired:  false,
						Description: StringPtr("varchar optional"),
					},
				},
			},
		},
	}
)

type InstalledApps []*AppManifest

func (x *InstalledApps) Scan(val interface{}) error {

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

func (x InstalledApps) Value() (driver.Value, error) {
	return json.Marshal(x)
}

type AppManifest struct {
	ID               string               `json:"id"` // should start with app_
	Name             string               `json:"name"`
	Homepage         string               `json:"homepage"`
	Author           string               `json:"author"`
	IconURL          string               `json:"icon_url"`
	ShortDescription string               `json:"short_description"`
	Description      string               `json:"description"`
	Version          string               `json:"version"`
	UIEndpoint       string               `json:"ui_endpoint"`      // url of the iframe to load the app
	WebhookEndpoint  string               `json:"webhook_endpoint"` // url of the webhook to call when the app receives data
	AppTables        AppTablesManifest    `json:"app_tables,omitempty"`
	ExtraColumns     ExtraColumnsManifest `json:"extra_columns,omitempty"` // i.e: {"user": [customColumn1, customColumn2]}
	Tasks            TasksManifest        `json:"tasks,omitempty"`
	DataHooks        DataHooksManifest    `json:"data_hooks,omitempty"`
	SQLAccess        AppSQLAccessManifest `json:"sql_access,omitempty"`
	CubeSchemas      CubeSchemasManifest  `json:"cube_schemas,omitempty"`
}

func (x *AppManifest) Validate(installedApps InstalledApps, isReinstall bool) error {

	re := regexp.MustCompile("^app_([a-z0-9])+$")
	re2 := regexp.MustCompile("^appx_([a-z0-9])+$")

	if !re.MatchString(x.ID) && !re2.MatchString(x.ID) {
		return ErrAppManifestIDInvalid
	}

	// check that ID has only one underscore
	if strings.Count(x.ID, "_") != 1 {
		return ErrAppManifestIDInvalid
	}

	// check if app id is unique
	for _, app := range installedApps {
		if app.ID == x.ID && !isReinstall {
			return ErrAppAlreadyExists
		}
	}

	if x.Name == "" || len(x.Name) > 50 {
		return ErrAppManifestNameInvalid
	}

	if !govalidator.IsRequestURL(x.Homepage) {
		return ErrAppManifestHomepageInvalid
	}

	if x.Author == "" {
		return ErrAppManifestAuthorInvalid
	}

	if !govalidator.IsRequestURL(x.IconURL) {
		return ErrAppManifestIconURLInvalid
	}

	if !govalidator.IsSemver(x.Version) {
		return ErrAppManifestVersionInvalid
	}

	if !govalidator.IsRequestURL(x.UIEndpoint) {
		return ErrAppManifestUIEndpointInvalid
	}

	if !govalidator.IsRequestURL(x.WebhookEndpoint) {
		return ErrAppManifestWebhookEndpointInvalid
	}

	// validate app tables
	if x.AppTables != nil && len(x.AppTables) > 0 {
		for _, table := range x.AppTables {
			if err := table.Validate(x.ID, installedApps); err != nil {
				return err
			}
		}
	}

	// validate extra columns
	if x.ExtraColumns != nil && len(x.ExtraColumns) > 0 {
		for _, augmentedTable := range x.ExtraColumns {

			if !govalidator.IsIn(
				augmentedTable.Kind,
				ItemKindUser,
				ItemKindOrder,
				ItemKindSession,
				ItemKindPostview,
			) {
				return ErrAppManifestExtraColumnTableNameInvalid
			}

			for _, col := range augmentedTable.Columns {
				if err := col.Validate(); err != nil {
					return err
				}

				// validate that column name is prefixed with app ID
				if !strings.HasPrefix(col.Name, x.ID+"_") {
					return eris.Wrapf(ErrTableColumnNameInvalid, "column: %v", col.Name)
				}

				// check that ID has only 2 underscore
				if strings.Count(col.Name, "_") < 2 {
					return eris.Wrapf(ErrTableColumnNameInvalid, "column: %v", col.Name)
				}
			}
		}
	}

	// TODO: verify that scheduled tasks id are prefixed with app ID

	// verify sql access
	if err := x.SQLAccess.Validate(x.AppTables); err != nil {
		return err
	}

	return nil
}

func (x *AppManifest) Scan(val interface{}) error {

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

func (x AppManifest) Value() (driver.Value, error) {
	return json.Marshal(x)
}

func DiffExtraColumns(from ExtraColumnsManifest, to ExtraColumnsManifest) (diff *ExtraColumnsManifestDiff, err error) {
	diff = &ExtraColumnsManifestDiff{}

	// if a is empty, all columns in b are new
	if len(from) == 0 {
		for _, tableB := range to {
			for _, colB := range tableB.Columns {
				diff.ToAdd = append(diff.ToAdd, &ExtraColumnsManifestOperation{
					Table:  tableB.Kind,
					Column: colB,
				})
			}
		}
		return diff, nil
	}

	// if b is empty, all columns in a are removed
	if len(to) == 0 {
		for _, tableA := range from {
			for _, colA := range tableA.Columns {
				diff.ToRemove = append(diff.ToRemove, &ExtraColumnsManifestOperation{
					Table:  tableA.Kind,
					Column: colA,
				})
			}
		}
		return diff, nil
	}

	// to add
	for _, currentTableB := range to {
		tableFound := false

		// check if extra column already exists in a
		for _, tableA := range from {
			if tableA.Kind == currentTableB.Kind {
				tableFound = true

				// check if column already exists
				for _, colB := range currentTableB.Columns {
					foundCol := false
					for _, colA := range tableA.Columns {
						if colA.Name == colB.Name {
							foundCol = true

							// abort if column type is different
							if !colA.HasSameDefinition(*colB) {
								return diff, eris.Errorf("extra column %v type is different", colB.Name)
							}

							break
						}
					}

					// add new column
					if !foundCol {
						diff.ToAdd = append(diff.ToAdd, &ExtraColumnsManifestOperation{
							Table:  currentTableB.Kind,
							Column: colB,
						})
					}
				}
				break
			}
		}
		if !tableFound {
			// add all columns
			for _, colB := range currentTableB.Columns {
				diff.ToAdd = append(diff.ToAdd, &ExtraColumnsManifestOperation{
					Table:  currentTableB.Kind,
					Column: colB,
				})
			}
		}
	}

	// to remove
	for _, currentTableA := range from {
		tableFound := false

		// check if extra column already exists in b
		for _, tableB := range to {
			if tableB.Kind == currentTableA.Kind {
				tableFound = true

				// check if column already exists
				for _, colA := range currentTableA.Columns {
					foundCol := false
					for _, colB := range tableB.Columns {
						if colA.Name == colB.Name {
							foundCol = true
							break
						}
					}

					// remove column
					if !foundCol {
						diff.ToRemove = append(diff.ToRemove, &ExtraColumnsManifestOperation{
							Table:  currentTableA.Kind,
							Column: colA,
						})
					}
				}
				break
			}
		}
		if !tableFound {
			// remove all columns
			for _, colA := range currentTableA.Columns {
				diff.ToRemove = append(diff.ToRemove, &ExtraColumnsManifestOperation{
					Table:  currentTableA.Kind,
					Column: colA,
				})
			}
		}
	}

	return diff, nil
}

func DiffAppTables(from AppTablesManifest, to AppTablesManifest) (diff *AppTablesManifestDiff, err error) {
	diff = &AppTablesManifestDiff{}

	// if a is empty, all tables in b are new
	if len(from) == 0 {
		for _, tableB := range to {
			diff.ToAdd = append(diff.ToAdd, &AppTableManifestOperation{
				AppTableManifest: tableB,
			})
		}

		return diff, nil
	}

	// if b is empty, all tables in a are removed
	if len(to) == 0 {
		for _, tableA := range from {
			diff.ToRemove = append(diff.ToRemove, &AppTableManifestOperation{
				AppTableManifest: tableA,
			})
		}

		return diff, nil
	}

	// check if table exists in a
	for _, tableB := range to {
		found := false
		for _, tableA := range from {
			if tableA.Name == tableB.Name {
				found = true

				// check if table definition is different
				if !tableA.HasSameDefinition(tableB) {
					diff.ToMigrate = append(diff.ToMigrate, &AppTableManifestOperation{
						AppTableManifest: tableB,
					})
				}

				break
			}
		}
		if !found {
			diff.ToAdd = append(diff.ToAdd, &AppTableManifestOperation{
				AppTableManifest: tableB,
			})
		}
	}

	// check if table exists in b
	for _, tableA := range from {
		found := false
		for _, tableB := range to {
			if tableB.Name == tableA.Name {
				found = true
				break
			}
		}
		if !found {
			diff.ToRemove = append(diff.ToRemove, &AppTableManifestOperation{
				AppTableManifest: tableA,
			})
		}
	}

	return diff, nil
}

type AppSQLAccessManifest struct {
	TablesPermissions []*TablePermission  `json:"tables_permissions,omitempty"`
	PredefinedQueries []*SQLQueryManifest `json:"predefined_queries,omitempty"`
}

type TablePermission struct {
	Table string `json:"table"`
	Read  bool   `json:"read"`
	Write bool   `json:"write"`
}

func (x *AppSQLAccessManifest) Validate(appTables AppTablesManifest) error {

	// tables allowed to read
	readAllowed := []string{
		"cart",
		"cart_item",
		"custom_event",
		"device",
		"order",
		"order_item",
		"pageview",
		"postview",
		"session",
		"user",
		"user_alias",
	}

	// add app tables
	for _, table := range appTables {
		readAllowed = append(readAllowed, table.Name)
	}

	// check that read tables are allowed
	for _, perm := range x.TablesPermissions {

		// read
		if perm.Read && !govalidator.IsIn(perm.Table, readAllowed...) {
			return eris.Errorf("table '%v' is not allowed to read", perm.Table)
		}

		// write
		// tables allowed to write should belong to the app
		if perm.Write {
			found := false
			for _, appTable := range appTables {
				if appTable.Name == perm.Table {
					found = true
					break
				}
			}
			if !found {
				return eris.Errorf("table '%v' is not allowed to write", perm.Table)
			}
		}

	}

	// check that predefined queries are valid
	for _, query := range x.PredefinedQueries {
		if err := query.Validate(); err != nil {
			return err
		}
	}

	return nil
}

type DataHooksManifest []*DataHookManifest

type DataHookManifest struct {
	ID   string         `json:"id"`
	Name string         `json:"name"`
	On   string         `json:"on"` // on_validation, on_success
	For  []*DataHookFor `json:"for"`
}

type SQLQueriesManifest []*SQLQueryManifest

type SQLQueryManifest struct {
	ID          string        `json:"id"`
	Type        string        `json:"type"` // query type: select, insert, update, delete
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Query       string        `json:"query"`
	TestArgs    []interface{} `json:"test_args"`
}

func (x *SQLQueryManifest) Validate() error {
	if x.ID == "" {
		return eris.New("query id is required")
	}

	if x.Name == "" {
		return eris.New("query name is required")
	}

	if x.Description == "" {
		return eris.New("query description is required")
	}

	if x.Query == "" {
		return eris.New("query is required")
	}

	// only SELECT statement is allowed for now
	if x.Type != SQLSelect {
		return eris.New("query type should be a 'select'")
	}

	// parse query to check if it's valid

	if x.Type == SQLSelect && !strings.HasPrefix(x.Query, "SELECT ") && !strings.HasPrefix(x.Query, "select ") {
		return eris.Errorf("query '%v' should start with SELECT statement", x.Query)
	}

	_, err := sqlparser.Parse(x.Query)
	if err != nil {
		return eris.Wrap(err, "query is not valid")
	}

	return nil
}
