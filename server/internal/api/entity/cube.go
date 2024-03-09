package entity

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/asaskevich/govalidator"
)

// type CubeJSQueryResponse struct {
// 	Query           interface{}       `json:"query"`
// 	Data            []MapOfInterfaces `json:"data"`
// 	LastRefreshTime string            `json:"lastRefreshTime"`
// }

// type CubeJSQuery struct {
// 	// https://cube.dev/docs/product/apis-integrations/rest-api/query-format
// 	// measures?: string[];
// 	// dimensions?: string[];
// 	// filters?: Filter[];
// 	// timeDimensions?: TimeDimension[];
// 	// segments?: string[];
// 	// limit?: null | number;
// 	// offset?: number;
// 	// order?: TQueryOrderObject | TQueryOrderArray;
// 	// timezone?: string;
// 	// renewQuery?: boolean;
// 	// ungrouped?: boolean;
// 	// responseFormat?: 'compact' | 'default';
// 	// total?: boolean;
// 	Measures       []string                   `json:"measures"`
// 	Dimensions     []string                   `json:"dimensions,omitempty"`
// 	Filters        []CubeJSQueryFilter        `json:"filters,omitempty"`
// 	TimeDimensions []CubeJSQueryTimeDimension `json:"timeDimensions"`
// 	Segments       []string                   `json:"segments,omitempty"`
// 	Limit          int64                      `json:"limit,omitempty"`
// 	Offset         int64                      `json:"offset,omitempty"`
// 	Order          []MapOfInterfaces          `json:"order,omitempty"`
// 	Timezone       string                     `json:"timezone"`
// 	RenewQuery     bool                       `json:"renewQuery,omitempty"`
// 	// If ungrouped is set to true no GROUP BY statement will be added to the query.
// 	// Instead, the raw results after filtering and joining will be returned without grouping.
// 	// By default ungrouped queries require a primary key as a dimension of every cube involved
// 	// in the query for security purposes. In case of ungrouped query measures will be rendered
// 	// as underlying sql of measures without aggregation and time dimensions will be truncated
// 	// as usual however not grouped by.
// 	Ungrouped bool `json:"ungrouped,omitempty"`
// }

type CubeJSQueryFilter struct {
	Dimension string   `json:"dimension,omitempty"`
	Member    string   `json:"member,omitempty"`
	Operator  string   `json:"operator"`
	Values    []string `json:"values"`
}

type CubeJSQueryTimeDimension struct {
	Dimension   string   `json:"dimension"`
	DateRange   []string `json:"dateRange"`
	Granularity string   `json:"granularity,omitempty"`
}

type CubeJSSchema struct {
	SQL         string `json:"sql"`
	Title       string `json:"title"`
	Description string `json:"description"`
	// Set this flag to true if you want Cube to rewrite your queries after final SQL has been generated. This may be helpful to apply filter pushdown optimizations or reduce unnecessary query nesting.
	RewriteQueries bool `json:"rewriteQueries,omitempty"`
	// The shown property is used to manage the visibility of a cube. Valid values for shown are true and false.
	Shown      bool                             `json:"shown,omitempty"`
	Joins      map[string]CubeJSSchemaJoin      `json:"joins,omitempty"`
	Segments   map[string]CubeJSSchemaSegment   `json:"segments,omitempty"`
	Measures   map[string]CubeJSSchemaMeasure   `json:"measures"`
	Dimensions map[string]CubeJSSchemaDimension `json:"dimensions"`
}

func (schema *CubeJSSchema) AddDimension(column *TableColumn) {
	if column.HideInAnalytics || column.Type == TableColumnTypeJSON {
		return
	}

	dimension := CubeJSSchemaDimension{
		Title:       column.Name,
		Description: *column.Description,
		Type:        "string",
		SQL:         column.Name,
	}

	switch column.Type {
	case TableColumnTypeBoolean:
		dimension.Type = "boolean"
	case TableColumnTypeNumber:
		dimension.Type = "number"
	case TableColumnTypeDate, TableColumnTypeDatetime, TableColumnTypeTimestamp:
		dimension.Type = "time"
	default:
		// default is string
	}

	schema.Dimensions[column.Name] = dimension
}

func (schema *CubeJSSchema) BuildContent(cubeName string) (content string) {

	b, err := json.Marshal(schema)
	if err != nil {
		log.Fatalf("build schema %v err: %v", cubeName, err)
	}

	return string(b)
	// return "cube('" + cubeName + "', " + string(b) + ");"
}

// create a ne CubeJS schema for a table
func NewTableCube(table *AppTableManifest) (schema *CubeJSSchema) {
	schema = &CubeJSSchema{
		Title: table.Name,
		SQL:   "SELECT * FROM " + table.Name,
		Joins: map[string]CubeJSSchemaJoin{},
		Measures: map[string]CubeJSSchemaMeasure{
			// count *
			"count": {
				Title:       "Count",
				Description: "Count",
				Type:        "count",
			},
		},
		Dimensions: map[string]CubeJSSchemaDimension{
			// reserved dimensions
			"id": {
				Title:       AppReservedTableColumns[0].Name,
				Description: *AppReservedTableColumns[0].Description,
				Type:        "string",
				PrimaryKey:  true,
				Shown:       false,
				SQL:         "id",
			},
			"external_id": {
				Title:       AppReservedTableColumns[1].Name,
				Description: *AppReservedTableColumns[1].Description,
				Type:        "string",
				SQL:         "external_id",
			},
			"created_at": {
				Title:       AppReservedTableColumns[2].Name,
				Description: *AppReservedTableColumns[2].Description,
				Type:        "time",
				SQL:         "created_at",
			},
			// Don't add user_id yet, should be added when we implement JOINs between apps and users
			// "user_id": {
			// 	Title:       AppReservedTableColumns[4].Name,
			// 	Description: *AppReservedTableColumns[4].Description,
			// 	Type:        "string",
			// 	SQL:         "user_id",
			// },
		},
	}

	if table.Description != nil {
		schema.Description = *table.Description
	}

	// joins
	if table.Joins != nil && len(table.Joins) > 0 {
		for _, join := range table.Joins {
			cubeName := strings.ToUpper(string(join.ExternalTable[0])) + join.ExternalTable[1:]
			// relationship := "belongsTo"
			// if join.Relationship == "has_one" {
			// 	relationship = "hasOne"
			// }
			// if join.Relationship == "has_many" {
			// 	relationship = "hasMany"
			// }
			schema.Joins[cubeName] = CubeJSSchemaJoin{
				Relationship: join.Relationship,
				SQL:          fmt.Sprintf("${CUBE}.%v = ${%v}.%v", join.LocalColumn, cubeName, join.ExternalColumn),
			}
		}
	}

	// dimensions + measures
	if table.Columns != nil && len(table.Columns) > 0 {
		for _, column := range table.Columns {

			// ignore reserved columns and JSON columns
			if govalidator.IsIn(column.Name, ReservedColumns...) || column.Type == ColumnTypeJSON {
				continue
			}

			dimensionType := "string"

			// https://cube.dev/docs/schema/reference/types-and-formats#dimensions-types
			switch column.Type {
			case ColumnTypeVarchar, ColumnTypeLongText:
				dimensionType = "string"
			case ColumnTypeNumber:
				dimensionType = "number"
			case ColumnTypeDate, ColumnTypeTimestamp, ColumnTypeDatetime:
				dimensionType = "time"
			case ColumnTypeBoolean:
				// boolean should be considered as number
				dimensionType = "number"
			default:
			}

			schema.Dimensions[column.Name] = CubeJSSchemaDimension{
				Title:       column.Name,
				Description: *column.Description,
				Type:        dimensionType,
				SQL:         column.Name,
			}

			// Number columns can be used as measures, we add the SUM and AVG measures
			if column.Type == ColumnTypeNumber {
				schema.Measures[column.Name+"_sum"] = CubeJSSchemaMeasure{
					Title:       "SUM of " + column.Name,
					Description: "(SUM of) " + *column.Description,
					Type:        "sum",
					SQL:         column.Name,
				}

				schema.Measures[column.Name+"_avg"] = CubeJSSchemaMeasure{
					Title:       "AVG of " + column.Name,
					Description: "(Average of) " + *column.Description,
					Type:        "avg",
					SQL:         column.Name,
				}
			}
		}
	}

	return
}

type CubeJSSchemaJoin struct {
	Relationship string `json:"relationship"` // one_to_one` || `one_to_many` || `many_to_one`
	SQL          string `json:"sql"`
}

type CubeJSSchemaSegment struct {
	SQL string `json:"sql"`
}

type CubeJSSchemaMeasure struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	// types: number, count, countDistinct, countDistrinctApprox, sum, min, max, avg, runningTotal
	Type          string                      `json:"type"`
	SQL           string                      `json:"sql"`
	DrillMembers  []string                    `json:"drillMembers"`
	Filters       []CubeJSSchemaMeasureFilter `json:"filters,omitempty"`
	Format        string                      `json:"format,omitempty"` // percent, currency
	RollingWindow *CubeJSSchemaRollingWindow  `json:"rollingWindow,omitempty"`
	Meta          M                           `json:"meta,omitempty"`  // used to pass data to the frontend
	Shown         bool                        `json:"shown,omitempty"` // default: true
}

type CubeJSSchemaMeasureFilter struct {
	SQL string `json:"sql"`
}

// https://cube.dev/docs/schema/reference/measures#parameters-rolling-window
type CubeJSSchemaRollingWindow struct {
	Trailing string `json:"trailing,omitempty"`
	Leading  string `json:"leading,omitempty"`
	Offset   string `json:"offset,omitempty"`
}

type CubeJSSchemaCase struct {
	When []CubeJSSchemaCaseWhen `json:"when"`
	Else CubeJSSchemaCaseElse   `json:"else"`
}

type CubeJSSchemaCaseWhen struct {
	SQL   string `json:"sql"`
	Label string `json:"label"`
}

type CubeJSSchemaCaseElse struct {
	Label string `json:"label"`
}

type CubeJSSchemaDimension struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	// types: string, number, time, boolean, geo
	Type       string            `json:"type"`
	SQL        string            `json:"sql"`
	PrimaryKey bool              `json:"primaryKey,omitempty"`
	Shown      bool              `json:"shown,omitempty"` // default: true
	Case       *CubeJSSchemaCase `json:"case,omitempty"`
	// The subQuery statement allows you to reference a measure in a dimension. It's an advanced concept and you can learn more about it here.
	// https://cube.dev/docs/schema/fundamentals/additional-concepts#subquery
	Subquery bool `json:"subquery,omitempty"`
	// When this statement is set to true, the filters applied to the query will be passed to the subquery.
	PropagateFiltersToSubQuery bool `json:"propagateFiltersToSubQuery,omitempty"`
	// https://cube.dev/docs/schema/reference/types-and-formats#dimension-formats-image-url
	Format string `json:"format,omitempty"` // percent, currency, link, id, imageUrl
	Meta   M      `json:"meta,omitempty"`   // used to pass data to the frontend
}

func GenerateSchemas(installedApps InstalledApps) (schemas map[string]*CubeJSSchema) {

	schemas = map[string]*CubeJSSchema{
		"User":         NewUserCube(),
		"Order":        NewOrderCube(),
		"Order_item":   NewOrderItemCube(),
		"Session":      NewSessionCube(),
		"Postview":     NewPostviewCube(),
		"Pageview":     NewPageviewCube(),
		"Device":       NewDeviceCube(),
		"Cart":         NewCartCube(),
		"Custom_event": NewCustomEventCube(),
		"Data_log":     NewDataLogCube(),
	}

	// loop over installed apps in workspace and enrich schemas
	for _, app := range installedApps {
		if app.AppTables != nil {
			// loop over tables
			for _, table := range app.AppTables {
				// generate schema for each table
				// convert first letter to uppercase
				cubeName := strings.ToUpper(string(table.Name[0])) + table.Name[1:]
				schemas[cubeName] = NewTableCube(table)
			}
		}

		if app.ExtraColumns != nil {
			// loop over extra columns
			for _, column := range app.ExtraColumns {
				if column.Kind == ItemKindUser {
					for _, col := range column.Columns {
						schemas["User"].AddDimension(col)
					}
				}
				if column.Kind == ItemKindSession {
					for _, col := range column.Columns {
						schemas["Session"].AddDimension(col)
					}
				}
				if column.Kind == ItemKindPostview {
					for _, col := range column.Columns {
						schemas["Postview"].AddDimension(col)
					}
				}
				if column.Kind == ItemKindPageview {
					for _, col := range column.Columns {
						schemas["Pageview"].AddDimension(col)
					}
				}
				if column.Kind == ItemKindDevice {
					for _, col := range column.Columns {
						schemas["Device"].AddDimension(col)
					}
				}
				if column.Kind == ItemKindOrder {
					for _, col := range column.Columns {
						schemas["Order"].AddDimension(col)
					}
				}
				if column.Kind == ItemKindOrderItem {
					for _, col := range column.Columns {
						schemas["Order_item"].AddDimension(col)
					}
				}
				if column.Kind == ItemKindCart {
					for _, col := range column.Columns {
						schemas["Cart"].AddDimension(col)
					}
				}
				if column.Kind == ItemKindCustomEvent {
					for _, col := range column.Columns {
						schemas["Custom_event"].AddDimension(col)
					}
				}
			}
		}
	}

	return schemas
}

// {
//   sql: 'SELECT * FROM users WHERE is_merged = FALSE',

//   joins: {
//     // UserSegments: {
//     //     relationship: 'hasMany',
//     //     sql: CUBE+'.id = '+UserSegments+'.user_id'
//     // },
//   },

//   segments: {
//     authenticated: {
//       sql: CUBE + '.is_authenticated = true'
//     }
//   },

//   measures: {
//     count: {
//       type: 'count'
//     },

//     sign_ups: {
//       type: 'count',
//       sql: CUBE + '.signed_up_at IS NOT NULL'
//     }
//   },

//   dimensions: {
//     id: { sql: 'id', type: 'string', primaryKey: true },

//     is_authenticated: { sql: 'is_authenticated', type: 'number' },
//     signed_up_at: { sql: 'signed_up_at', type: 'time' },
//     created_at: { sql: 'created_at', type: 'time' },
//     created_at_trunc: { sql: 'created_at_trunc', type: 'time' },
//     db_created_at: { sql: 'db_created_at', type: 'time' },
//     db_updated_at: { sql: 'db_updated_at', type: 'time' },
//     last_interaction_at: { sql: 'last_interaction_at', type: 'time' },
//     timezone: { sql: 'timezone', type: 'string' },
//     language: { sql: 'language', type: 'string' },
//     country: { sql: 'country', type: 'string' },

//     orders_count: { sql: 'orders_count', type: 'number' },
//     orders_ltv: { sql: 'orders_ltv', type: 'number' },
//     orders_avg_cart: { sql: 'orders_avg_cart', type: 'number' },
//     first_order_at: { sql: 'first_order_at', type: 'time' },
//     first_order_subtotal: { sql: 'first_order_subtotal', type: 'number' },
//     first_order_ttc: { sql: 'first_order_ttc', type: 'number' },
//     last_order_at: { sql: 'last_order_at', type: 'time' },
//     avg_repeat_cart: { sql: 'avg_repeat_cart', type: 'number' },
//     avg_repeat_order_ttc: { sql: 'avg_repeat_order_ttc', type: 'number' }
//   }
// }
