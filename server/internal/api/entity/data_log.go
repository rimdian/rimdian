package entity

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rimdian/rimdian/internal/common/dto"
	"github.com/tidwall/gjson"
)

type DataLogStatusType = int

const (
	DataLogHasErrorNone         DataLogStatusType = 0 // has no error
	DataLogHasErrorRetryable    DataLogStatusType = 1 // has retryable error
	DataLogHasErrorNotRetryable DataLogStatusType = 2 // has non retryable error

	// checkpoint from 0 to 100, to have room to add new steps in between
	DataLogCheckpointPending                  DataLogStatusType = 0   // waiting to be processed
	DataLogCheckpointHookOnValidationExecuted DataLogStatusType = 10  // hook on_validation executed
	DataLogCheckpointPersisted                DataLogStatusType = 20  // data_log data extracted and persisted in DB
	DataLogCheckpointItemUpserted             DataLogStatusType = 30  // item upserted
	DataLogCheckpointSpecialActionExecuted    DataLogStatusType = 40  // send eventual message etc... (email, sms...)
	DataLogCheckpointConversionsAttributed    DataLogStatusType = 50  // conversions attributed
	DataLogCheckpointSegmentsRecomputed       DataLogStatusType = 60  // segment processed
	DataLogCheckpointShouldRespawn            DataLogStatusType = 70  // should respawn to process workflows+hooks asynchronously
	DataLogCheckpointWorkflowsTriggered       DataLogStatusType = 80  // workflow triggered
	DataLogCheckpointHooksFinalizeExecuted    DataLogStatusType = 90  // hooks finalize executed
	DataLogCheckpointDone                     DataLogStatusType = 100 // launch eventual child tasks and we are done (success or aborted)

	MicrosecondLayout = "2006-01-02T15:04:05.999999Z" // .99 = trick for microseconds

	ItemKindUser                 = "user"
	ItemKindSession              = "session"
	ItemKindPostview             = "postview"
	ItemKindPageview             = "pageview"
	ItemKindDevice               = "device"
	ItemKindOrder                = "order"
	ItemKindOrderItem            = "order_item"
	ItemKindCart                 = "cart"
	ItemKindCartItem             = "cart_item"
	ItemKindCustomEvent          = "custom_event"
	ItemKindSubscriptionListUser = "subscription_list_user"
	ItemKindMessage              = "message"
)

type ChildDataLog struct {
	Kind           string
	Action         string
	UserID         string
	ItemID         string
	ItemExternalID string
	UpdatedFields  UpdatedFields
	EventAt        time.Time
	Tx             *sql.Tx
}

// The data_log table is used to store all the data received from the collector
// it also acts as the timeline of items/events
type DataLog struct {
	ID         string             `db:"id" json:"id"` // hash of the payload
	Origin     int                `db:"origin" json:"origin"`
	OriginID   string             `db:"origin_id" json:"origin_id"`       // client hostname, token account_id, task_id, workflow_id, data_log_id...
	Context    dto.DataLogContext `db:"context" json:"context,omitempty"` // context provides info regarding the hits
	Item       string             `db:"item" json:"item"`                 // raw data received from the collector
	Checkpoint int                `db:"checkpoint" json:"checkpoint"`
	HasError   int                `db:"has_error" json:"has_error"` // 0: no, 1: retryable, 2: not retryable
	Errors     MapOfStrings       `db:"errors" json:"errors,omitempty"`
	Hooks      DataHooksState     `db:"hooks" json:"hooks,omitempty"` // hooks executed
	// item timeline
	UserID                   string        `db:"user_id" json:"user_id"`
	MergedFromUserExternalID *string       `db:"merged_from_user_external_id" json:"merged_from_user_external_id"`
	Kind                     string        `db:"kind" json:"kind"`     // user, session...
	Action                   string        `db:"action" json:"action"` // signup / alias / create / update / enter / exit ...
	ItemID                   string        `db:"item_id" json:"item_id"`
	ItemExternalID           string        `db:"item_external_id" json:"item_external_id"`
	UpdatedFields            UpdatedFields `db:"updated_fields" json:"updated_fields"`
	EventAt                  time.Time     `db:"event_at" json:"event_at"`
	EventAtTrunc             time.Time     `db:"event_at_trunc" json:"event_at_trunc"` // sort key
	DBCreatedAt              *time.Time    `db:"db_created_at" json:"db_created_at,omitempty"`
	DBUpdatedAt              *time.Time    `db:"db_updated_at" json:"db_updated_at,omitempty"`
	// not persisted
	ReplayID *string `db:"-" json:"replay_id,omitempty"` // used to identify if the data_log is a replay
	DomainID *string `db:"-" json:"domain_id,omitempty"` // domain_id matched with the Host header for client-side origin
	// upserted entities
	UpsertedUser                 *User                 `json:"user,omitempty"`                   // user upserted
	UserAlias                    *UserAlias            `json:"user_alias,omitempty"`             // user_alias upserted
	UpsertedPageview             *Pageview             `json:"pageview,omitempty"`               // pageview upserted
	UpsertedOrder                *Order                `json:"order,omitempty"`                  // order upserted
	UpsertedCart                 *Cart                 `json:"cart,omitempty"`                   // cart upserted
	UpsertedCustomEvent          *CustomEvent          `json:"custom_event,omitempty"`           // custom_event upserted
	UpsertedDevice               *Device               `json:"device,omitempty"`                 // device upserted
	UpsertedSession              *Session              `json:"session,omitempty"`                // session upserted
	UpsertedPostview             *Postview             `json:"postview,omitempty"`               // postview upserted
	UpsertedSubscriptionListUser *SubscriptionListUser `json:"subscription_list_user,omitempty"` // subscription_list_user upserted
	UpsertedMessage              *Message              `json:"message,omitempty"`                // message upserted
	UpsertedAppItem              *AppItem              `json:"-"`                                // app item upserted, not exposed in JSON directly
}

func (dataLog *DataLog) IsPersisted() bool {
	return dataLog.DBCreatedAt != nil
}

func (dataLog *DataLog) IsSpecialAction() bool {
	return dataLog.Kind == ItemKindMessage
}

// some data_logs produce and process children synchronously
// others produce children that are processed asynchronously (ex: user_alias)
// children processed synchronously can't be replayed, as their parent will reproduce them
// children processed asynchronously can be replayed, as their parent will not reprocess them
// a children processed synchronously has no "item" value
func (dataLog *DataLog) IsReplayable() bool {
	if dataLog.HasError == DataLogHasErrorNotRetryable {
		return false
	}
	if dataLog.Origin == dto.DataLogOriginInternalDataLog && dataLog.Item == "{}" {
		return false
	}
	return true
}

type UpdatedFields []*UpdatedField

func (x *UpdatedFields) Scan(val interface{}) error {

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

func (x UpdatedFields) Value() (driver.Value, error) {
	return json.Marshal(x)
}

type UpdatedField struct {
	Field     string      `json:"field"`
	PrevValue interface{} `json:"previous"`
	NewValue  interface{} `json:"new"`
}

func ExtractFieldValueFromGJSON(fieldDefinition *TableColumn, result gjson.Result, clockDifference time.Duration) (fieldValue *AppItemField, err error) {

	switch fieldDefinition.Type {
	case ColumnTypeVarchar, ColumnTypeLongText:

		// init string
		fieldValue = &AppItemField{
			Name:        fieldDefinition.Name,
			Type:        fieldDefinition.Type,
			StringValue: NullableString{IsNull: true},
		}

		if result.Type == gjson.Null {
			return
		}

		if result.Type != gjson.String {
			return nil, fmt.Errorf("ExtractAppItemAndValidate, field %v is not a string", fieldDefinition.Name)
		}

		fieldValue.StringValue.IsNull = false
		fieldValue.StringValue.String = result.String()

	case ColumnTypeNumber:

		// init number
		fieldValue = &AppItemField{
			Name:         fieldDefinition.Name,
			Type:         fieldDefinition.Type,
			Float64Value: NullableFloat64{IsNull: true},
		}

		if result.Type == gjson.Null {
			return
		}

		if result.Type != gjson.Number {
			return nil, fmt.Errorf("ExtractAppItemAndValidate, field %v is not a number", fieldDefinition.Name)
		}

		fieldValue.Float64Value.IsNull = false
		fieldValue.Float64Value.Float64 = result.Float()

	case ColumnTypeBoolean:

		// init boolean
		fieldValue = &AppItemField{
			Name:      fieldDefinition.Name,
			Type:      fieldDefinition.Type,
			BoolValue: NullableBool{IsNull: true},
		}

		if result.Type == gjson.Null {
			return
		}

		if result.Type != gjson.True && result.Type != gjson.False {
			return nil, fmt.Errorf("ExtractAppItemAndValidate, field %v is not a boolean", fieldDefinition.Name)
		}

		fieldValue.BoolValue.IsNull = false
		fieldValue.BoolValue.Bool = result.Bool()

	case ColumnTypeTimestamp:

		// init number
		fieldValue = &AppItemField{
			Name:      fieldDefinition.Name,
			Type:      fieldDefinition.Type,
			TimeValue: NullableTime{IsNull: true},
		}

		if result.Type == gjson.Null {
			return
		}

		if result.Type != gjson.Number {
			return nil, fmt.Errorf("ExtractAppItemAndValidate, field %v is not a number", fieldDefinition.Name)
		}

		fieldValue.TimeValue.IsNull = false
		fieldValue.TimeValue.Time = time.Unix(int64(result.Float()), 0)

	case ColumnTypeDatetime, ColumnTypeDate:

		// init time
		fieldValue = &AppItemField{
			Name:      fieldDefinition.Name,
			Type:      fieldDefinition.Type,
			TimeValue: NullableTime{IsNull: true},
		}

		if result.Type == gjson.Null {
			return
		}

		if result.Type != gjson.String {
			return nil, fmt.Errorf("ExtractAppItemAndValidate, field %v is not a string", fieldDefinition.Name)
		}

		// parse time
		layout := time.RFC3339Nano
		if fieldDefinition.Type == ColumnTypeDate {
			layout = "2006-01-02"
		}

		t, err := time.Parse(layout, result.String())
		if err != nil {
			return nil, fmt.Errorf("ExtractAppItemAndValidate, field %v is not a valid date, got %v, err: %v", fieldDefinition.Name, result.String(), err)
		}

		// apply clock difference on system fields
		if fieldDefinition.Name == "created_at" || fieldDefinition.Name == "updated_at" {
			// apply clockDifference if date in future
			if t.After(time.Now()) {
				t = t.Add(clockDifference)
			}
		}

		fieldValue.TimeValue.IsNull = false
		fieldValue.TimeValue.Time = t

	case ColumnTypeJSON:

		// init json
		fieldValue = &AppItemField{
			Name:      fieldDefinition.Name,
			Type:      fieldDefinition.Type,
			JSONValue: NullableJSON{IsNull: true},
		}

		if result.Type == gjson.Null {
			return
		}

		if result.Type != gjson.JSON {
			return nil, fmt.Errorf("ExtractAppItemAndValidate, field %v is not a json", fieldDefinition.Name)
		}

		fieldValue.JSONValue.IsNull = false
		// fieldValue.JSONValue.JSON = rawMessage
		fieldValue.JSONValue.JSON = []byte(result.Raw)
	default:
		return nil, fmt.Errorf("ExtractAppItemAndValidate, field %v has unknown type", fieldDefinition.Name)
	}

	return
}

func NewInternalDataLogChild(parent *DataLog, data ChildDataLog) *DataLog {

	child := &DataLog{
		Origin:   dto.DataLogOriginInternalDataLog,
		OriginID: parent.ID,
		Context: dto.DataLogContext{
			WorkspaceID:      parent.Context.WorkspaceID,
			IP:               parent.Context.IP,
			HeadersAndParams: map[string]string{},
			ReceivedAt:       parent.Context.ReceivedAt,
		},
		// provide a fake ID to ensure dataLog.ID is unique
		Item:           fmt.Sprintf(`{"_id":"%v"}`, uuid.New().String()),
		Checkpoint:     DataLogCheckpointPersisted,
		Errors:         MapOfStrings{},
		Hooks:          DataHooksState{},
		UserID:         data.UserID,
		Kind:           data.Kind,
		Action:         data.Action,
		ItemID:         data.ItemID,
		ItemExternalID: data.ItemExternalID,
		UpdatedFields:  data.UpdatedFields,
		EventAt:        data.EventAt,
		EventAtTrunc:   data.EventAt.Truncate(time.Hour),
		// propagate upserted user+device+session for data_hooks
		UpsertedUser:    parent.UpsertedUser,
		UpsertedDevice:  parent.UpsertedDevice,
		UpsertedSession: parent.UpsertedSession,
	}

	child.ID = dto.ComputeDataLogID(child.Context.WorkspaceID, child.Origin, child.Item)

	return child
}

func NewTaskDataLog(workspaceID string, taskID string, data ChildDataLog) *DataLog {
	return &DataLog{
		Origin:   dto.DataLogOriginInternalTaskExec,
		OriginID: taskID,
		Context: dto.DataLogContext{
			WorkspaceID:      workspaceID,
			HeadersAndParams: map[string]string{},
			ReceivedAt:       time.Now(),
		},
		Item:           "{}", // empty JSON
		Checkpoint:     DataLogCheckpointDone,
		Errors:         MapOfStrings{},
		Hooks:          DataHooksState{},
		UserID:         data.UserID,
		Kind:           data.Kind,
		Action:         data.Action,
		ItemID:         data.ItemID,
		ItemExternalID: data.ItemExternalID,
		UpdatedFields:  data.UpdatedFields,
		EventAt:        data.EventAt,
		EventAtTrunc:   data.EventAt.Truncate(time.Hour),
	}
}

var DataLogSchema string = `CREATE TABLE IF NOT EXISTS data_log (
	id VARCHAR(128) NOT NULL,
	-- workspace_id VARCHAR(128) NOT NULL,
	origin TINYINT(1) UNSIGNED NOT NULL,
	origin_id VARCHAR(128),
	context JSON NOT NULL,
	item JSON NOT NULL,
	checkpoint TINYINT NOT NULL,
	has_error TINYINT NOT NULL,
	errors JSON NOT NULL,
	hooks JSON NOT NULL,
	user_id VARCHAR(128) NOT NULL,
	merged_from_user_external_id VARCHAR(128),
	kind VARCHAR(128) NOT NULL,
	action VARCHAR(32) NOT NULL,
	item_id VARCHAR(128) NOT NULL,
	item_external_id VARCHAR(128) NOT NULL,
	updated_fields JSON NOT NULL,
	event_at TIMESTAMP(6) NOT NULL,
	event_at_trunc TIMESTAMP(6) NOT NULL SERIES TIMESTAMP, /* micro second date to do pagination before/after */
	db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	
	-- sort DESC by default
	SORT KEY (event_at_trunc DESC, event_at DESC),
	PRIMARY KEY (id),
	KEY (checkpoint),
	KEY (origin, origin_id),
	KEY (has_error),
	KEY (kind),
	SHARD KEY (id)
  ) COLLATE utf8mb4_general_ci;
`

var DataLogSchemaMYSQL string = `CREATE TABLE IF NOT EXISTS data_log (
	id VARCHAR(128) NOT NULL,
	-- workspace_id VARCHAR(128) NOT NULL,
	origin TINYINT(1) UNSIGNED NOT NULL,
	origin_id VARCHAR(128),
	context JSON NOT NULL,
	item JSON NOT NULL,
	checkpoint TINYINT NOT NULL,
	has_error TINYINT NOT NULL,
	errors JSON NOT NULL,
	hooks JSON NOT NULL,
	user_id VARCHAR(128) NOT NULL,
	merged_from_user_external_id VARCHAR(128),
	kind VARCHAR(128) NOT NULL,
	action VARCHAR(32) NOT NULL,
	item_id VARCHAR(128) NOT NULL,
	item_external_id VARCHAR(128) NOT NULL,
	updated_fields JSON NOT NULL,
	event_at TIMESTAMP(6) NOT NULL,
	event_at_trunc TIMESTAMP(6) NOT NULL, /* micro second date to do pagination before/after */
	db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	
	PRIMARY KEY (id),
	KEY (checkpoint),
	KEY (origin, origin_id),
	KEY (has_error),
	KEY (kind),
	KEY (event_at_trunc DESC, event_at DESC)
  ) COLLATE utf8mb4_general_ci;
`

func NewDataLogCube() *CubeJSSchema {
	return &CubeJSSchema{
		Title:       "Data logs",
		Description: "Data logs",
		SQL:         "SELECT * FROM `data_log`",
		Measures: map[string]CubeJSSchemaMeasure{
			"count": {
				Type:        "count",
				Title:       "Count all",
				Description: "Count all",
			},
			"errors_count": {
				Type:        "count",
				Title:       "Errors count",
				Description: "Errors: has_error > 0",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "has_error > 0"},
				},
			},
			"not_done_count": {
				Type:        "count",
				Title:       "Not done",
				Description: "Not done: checkpoint != 100",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "checkpoint != 100"},
				},
			},
			"error_retryable_count": {
				Type:        "count",
				Title:       "Retryable errors count",
				Description: "Retryable errors: has_error = 1",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "has_error = 1"},
				},
			},
			"error_non_retryable_count": {
				Type:        "count",
				Title:       "Non-retryable errors count",
				Description: "Non-retryable errors: has_error = 2",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "has_error = 2"},
				},
			},
			"successful_count": {
				Type:        "count",
				Title:       "Successful count",
				Description: "Successful: checkpoint = 100 and has_error = 0",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "checkpoint = 100 and has_error = 0"},
				},
			},
			"pending_count": {
				Type:        "count",
				Title:       "Pending count",
				Description: "Pending: checkpoint = 0",
				Filters: []CubeJSSchemaMeasureFilter{
					{SQL: "checkpoint = 0"},
				},
			},
		},
		Dimensions: map[string]CubeJSSchemaDimension{
			"id": {
				SQL:         "id",
				Type:        "string",
				PrimaryKey:  true,
				Title:       "Data log ID",
				Description: "field: id",
			},
			"origin": {
				SQL:         "origin",
				Type:        "string",
				Title:       "Data log origin",
				Description: "field: origin",
			},
			"origin_id": {
				SQL:         "origin_id",
				Type:        "string",
				Title:       "Data log origin ID",
				Description: "field: origin_id",
			},
			"checkpoint": {
				SQL:         "checkpoint",
				Type:        "number",
				Title:       "Data log checkpoint",
				Description: "field: checkpoint",
			},
			"has_error": {
				SQL:         "has_error",
				Type:        "number",
				Title:       "Data log has error",
				Description: "field: has_error",
			},
			"kind": {
				SQL:         "kind",
				Type:        "string",
				Title:       "Data log item kind",
				Description: "field: kind",
			},
			"user_id": {
				SQL:         "user_id",
				Type:        "string",
				Title:       "User ID",
				Description: "field: user_id",
			},
			"action": {
				SQL:         "action",
				Type:        "string",
				Title:       "Action",
				Description: "field: action",
			},
			"item_id": {
				SQL:         "item_id",
				Type:        "string",
				Title:       "Item ID",
				Description: "field: item_id",
			},
			"item_external_id": {
				SQL:         "item_external_id",
				Type:        "string",
				Title:       "Item external ID",
				Description: "field: item_external_id",
			},
			"event_at": {
				SQL:         "event_at",
				Type:        "time",
				Title:       "Date of event",
				Description: "field: event_at",
			},
			"event_at_trunc": {
				SQL:         "event_at_trunc",
				Type:        "time",
				Title:       "Date of event (truncated to hour)",
				Description: "field: event_at_trunc",
			},
		},
	}
}
