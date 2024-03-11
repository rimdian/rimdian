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
	SQLSelect     = "select"
	SQLFullAccess = "*"

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
		IconURL:          "https://eu.captainmetrics.com/images/apps/test.png",
		ShortDescription: "Test app",
		Description:      "Test app",
		Version:          "1.0.0",
		UIEndpoint:       "https://console.rimdian.com/apps/test",
		WebhookEndpoint:  "API_ENDPOINT/api/webhook.receiver",
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
	SQLQueries       SQLQueriesManifest   `json:"sql_queries,omitempty"`
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

	if x.SQLQueries != nil && len(x.SQLQueries) > 0 {
		for _, query := range x.SQLQueries {
			if err := query.Validate(); err != nil {
				return err
			}
		}
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
	if x.Query != SQLFullAccess {

		if x.Type == SQLSelect && !strings.HasPrefix(x.Query, "SELECT ") && !strings.HasPrefix(x.Query, "select ") {
			return eris.Errorf("query '%v' should start with SELECT statement", x.Query)
		}

		_, err := sqlparser.Parse(x.Query)
		if err != nil {
			return eris.Wrap(err, "query is not valid")
		}
	}

	return nil
}
