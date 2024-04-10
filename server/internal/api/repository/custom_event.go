package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (repo *RepositoryImpl) FetchCustomEvents(ctx context.Context, workspace *entity.Workspace, query sq.SelectBuilder, tx *sql.Tx) (events []*entity.CustomEvent, err error) {

	events = []*entity.CustomEvent{}
	var rows *sql.Rows

	// query
	sql, args, errSQL := query.ToSql()

	if errSQL != nil {
		err = eris.Wrap(errSQL, "FetchCustomEvents")
		return
	}

	if tx == nil {

		conn, err := repo.GetWorkspaceConnection(ctx, workspace.ID)

		if err != nil {
			return nil, eris.Wrap(err, "FetchCustomEvents")
		}

		defer conn.Close()

		rows, err = conn.QueryContext(ctx, sql, args...)
	} else {
		rows, err = tx.QueryContext(ctx, sql, args...)
	}

	if err != nil {
		return nil, eris.Wrap(err, "FetchCustomEvents")
	}

	defer rows.Close()

	for rows.Next() {

		if rows.Err() != nil {
			return nil, eris.Wrap(rows.Err(), "FetchCustomEvents")
		}

		cols, err := rows.Columns()

		if err != nil {
			return nil, eris.Wrap(err, "FetchCustomEvents")
		}

		event := &entity.CustomEvent{}

		if err = scanCustomEventRow(cols, rows, event, workspace.InstalledApps); err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	return
}

func (repo *RepositoryImpl) ListCustomEventsForUser(ctx context.Context, workspace *entity.Workspace, userID string, orderBy string, tx *sql.Tx) (events []*entity.CustomEvent, err error) {
	query := sq.Select("*").From(entity.ItemKindCustomEvent).Where(sq.Eq{"user_id": userID}).OrderBy(orderBy)
	return repo.FetchCustomEvents(ctx, workspace, query, tx)
}

func (repo *RepositoryImpl) InsertCustomEvent(ctx context.Context, customEvent *entity.CustomEvent, tx *sql.Tx) (err error) {

	if customEvent == nil {
		err = eris.New("customEvent is missing")
		return
	}
	if tx == nil {
		err = eris.New("tx is missing")
		return
	}

	now := time.Now()

	// INSERT
	columns := []string{
		"id",
		"external_id",
		"user_id",
		"created_at",
		"domain_id",
		"session_id",
		// "created_at_trunc", // computed field
		// "db_created_at", // computed field
		// "db_updated_at", // computed field
		// "merged_from_user_id",
		"fields_timestamp",

		"label",
		"string_value",
		"number_value",
		"boolean_value",
		"non_interactive",

		// "extra_columns", // loop over dimensions to add app_custom_fields
	}

	values := []interface{}{
		customEvent.ID,
		customEvent.ExternalID,
		customEvent.UserID,
		customEvent.CreatedAt,
		customEvent.DomainID,
		customEvent.SessionID,
		customEvent.FieldsTimestamp,

		customEvent.Label,
		customEvent.StringValue,
		customEvent.NumberValue,
		customEvent.BooleanValue,
		customEvent.NonInteractive,
	}

	// add extra columns to the query
	// WARNING: values are interfaces, it might not guess well field types...
	// might have to convert them to NullableTypes before adding them
	if customEvent.ExtraColumns != nil && len(customEvent.ExtraColumns) > 0 {
		for field, value := range customEvent.ExtraColumns {
			columns = append(columns, field)

			// if it's a slice, we need to convert it to []byte
			// if reflect.TypeOf(value).Kind() == reflect.Slice {
			// 	value = []byte(fmt.Sprintf("%v", value))
			// }

			// // if its a map
			// if reflect.TypeOf(value).Kind() == reflect.Map {
			// 	value = value.(entity.MapOfInterfaces)
			// }

			values = append(values, value)
		}
	}

	q := sq.Insert(entity.ItemKindCustomEvent).Columns(columns...).Values(values...)

	sql, args, errSQL := q.ToSql()

	if errSQL != nil {
		err = eris.Wrapf(errSQL, "build query insert customEvent: %v\n", customEvent)
		return
	}

	_, errExec := tx.ExecContext(ctx, sql, args...)

	if errExec != nil {
		if repo.IsDuplicateEntry(errExec) {
			return eris.Wrap(ErrRowAlreadyExists, "InsertCustomEvent")
		}

		err = eris.Wrap(errExec, "InsertCustomEvent")
		return
	}

	customEvent.DBCreatedAt = now
	customEvent.DBUpdatedAt = now

	return
}

func (repo *RepositoryImpl) UpdateCustomEvent(ctx context.Context, customEvent *entity.CustomEvent, tx *sql.Tx) (err error) {

	if customEvent == nil {
		err = eris.New("customEvent is missing")
		return
	}
	if tx == nil {
		err = eris.New("tx is missing")
		return
	}

	now := time.Now()

	// UPDATE
	// specify sharding key to avoid deadlocks

	q := sq.Update(entity.ItemKindCustomEvent).Where(sq.Eq{"user_id": customEvent.UserID}).Where(sq.Eq{"id": customEvent.ID}).
		// Set("created_at", customEvent.CreatedAt).
		Set("fields_timestamp", customEvent.FieldsTimestamp).
		Set("label", customEvent.Label).
		Set("string_value", customEvent.StringValue).
		Set("number_value", customEvent.NumberValue).
		Set("boolean_value", customEvent.BooleanValue).
		Set("non_interactive", customEvent.NonInteractive)
		// "extra_columns", // loop over dimensions to add app_custom_fields

	// add extra columns to the query
	// WARNING: values are interfaces, it might not guess well field types...
	// might have to convert them to NullableTypes before adding them
	for field, value := range customEvent.ExtraColumns {
		q = q.Set(field, value)
	}

	sql, args, errSQL := q.ToSql()

	if errSQL != nil {
		err = eris.Wrapf(errSQL, "build query update customEvent: %v\n", customEvent)
		return
	}

	_, errExec := tx.ExecContext(ctx, sql, args...)

	if errExec != nil {
		// if repo.IsDuplicateEntry(errExec) {
		// }

		err = eris.Wrap(errExec, "UpdateCustomEvent")
		return
	}

	customEvent.DBUpdatedAt = now

	return
}

func (repo *RepositoryImpl) FindCustomEventByID(ctx context.Context, workspace *entity.Workspace, eventID string, userID string, tx *sql.Tx) (eventFound *entity.CustomEvent, err error) {

	var rows *sql.Rows
	eventFound = &entity.CustomEvent{}
	eventFound.ExtraColumns = entity.AppItemFields{}

	if tx == nil {

		var conn *sql.Conn

		conn, err = repo.GetWorkspaceConnection(ctx, workspace.ID)

		if err != nil {
			return nil, err
		}

		defer conn.Close()

		rows, err = conn.QueryContext(ctx, "SELECT * FROM custom_event WHERE user_id = ? AND id = ? LIMIT 1", userID, eventID)
	} else {
		rows, err = tx.QueryContext(ctx, "SELECT * FROM custom_event WHERE user_id = ? AND id = ? LIMIT 1", userID, eventID)
	}

	if err != nil {
		return nil, eris.Wrap(err, "FindCustomEventByID")
	}

	defer rows.Close()

	// no rows found
	if !rows.Next() {
		return nil, nil
	}

	// extract columns names
	cols, err := rows.Columns()

	if err != nil {
		return nil, eris.Wrap(err, "FindCustomEventByID")
	}

	// convert raw data fields to app item fields
	err = scanCustomEventRow(cols, rows, eventFound, workspace.InstalledApps)
	if err != nil {
		return nil, eris.Wrap(err, "FindCustomEventByID")
	}

	return eventFound, nil
}

// clones events from a user to another user with
// because the shard key "user_id" is immutable, we can't use an UPDATE
// we have to INSERT FROM SELECT + DELETE
func (repo *RepositoryImpl) MergeUserCustomEvents(ctx context.Context, workspace *entity.Workspace, fromUserID string, toUserID string, tx *sql.Tx) (err error) {

	// find eventual extra columns for the customEvent table
	customEventCustomColumns := workspace.FindExtraColumnsForItemKind(entity.ItemKindCustomEvent)

	customEventStruct := entity.CustomEvent{}
	columns := entity.GetNotComputedDBColumnsForObject(customEventStruct, entity.CustomEventComputedFields, customEventCustomColumns)

	// replace dynamically the user_id+merged_from_user_id on the SELECT statement
	selectedColumns := []string{}
	for _, col := range columns {

		if col == "user_id" {
			selectedColumns = append(selectedColumns, fmt.Sprintf("'%v' as user_id", toUserID))
		} else if col == "merged_from_user_id" {
			selectedColumns = append(selectedColumns, fmt.Sprintf("'%v' as merged_from_user_id", fromUserID))
		} else {
			selectedColumns = append(selectedColumns, col)
		}
	}

	query := fmt.Sprintf(`
		INSERT IGNORE INTO custom_event (%v) 
		SELECT %v FROM custom_event 
		WHERE user_id = ?
	`, strings.Join(columns, ", "), strings.Join(selectedColumns, ", "))

	// log.Println(query)

	if _, err := tx.ExecContext(ctx, query, fromUserID); err != nil {
		return eris.Wrap(err, "MergeUserCustomEvents")
	}

	// BUG deleting these rows might create a deadlock in singlestore
	// // delete "from user" custom_event
	// if _, err := tx.ExecContext(ctx, "DELETE FROM custom_event WHERE user_id = ? OPTION (columnstore_table_lock_threshold = 5000)", fromUserID); err != nil {
	// 	return err
	// }

	return nil
}

func scanCustomEventRow(cols []string, row RowScanner, event *entity.CustomEvent, installedApps entity.InstalledApps) error {

	// scan values
	values := make([]interface{}, len(cols))
	// extraColumns := []*ExtraColumn{}

	if event.ExtraColumns == nil {
		event.ExtraColumns = entity.AppItemFields{}
	}

	for i, col := range cols {
		switch col {
		case "id":
			values[i] = &event.ID
		case "external_id":
			values[i] = &event.ExternalID
		case "user_id":
			values[i] = &event.UserID
		case "domain_id":
			values[i] = &event.DomainID
		case "session_id":
			values[i] = &event.SessionID
		case "created_at":
			values[i] = &event.CreatedAt
		case "created_at_trunc":
			values[i] = &event.CreatedAtTrunc
		case "db_created_at":
			values[i] = &event.DBCreatedAt
		case "db_updated_at":
			values[i] = &event.DBUpdatedAt
		case "is_deleted":
			values[i] = &event.IsDeleted
		case "merged_from_user_id":
			values[i] = &event.MergedFromUserID
		case "fields_timestamp":
			values[i] = &event.FieldsTimestamp
		case "label":
			values[i] = &event.Label
		case "string_value":
			values[i] = &event.StringValue
		case "number_value":
			values[i] = &event.NumberValue
		case "boolean_value":
			values[i] = &event.BooleanValue
		case "non_interactive":
			values[i] = &event.NonInteractive
		default:
			// handle app extra columns
			if !strings.HasPrefix(col, "app_") && !strings.HasPrefix(col, "appx_") {
				return eris.Errorf("event column not mapped %v", col)
			}

			// find app item field in installedApps
			isValueMapped := false

			for _, app := range installedApps {
				for _, extraColumn := range app.ExtraColumns {
					if extraColumn.Kind == entity.ItemKindCustomEvent {

						for _, colDefinition := range extraColumn.Columns {
							if colDefinition.Name == col {
								isValueMapped = true

								// init an extra column in the session
								appItemField := &entity.AppItemField{
									Name: colDefinition.Name,
									Type: colDefinition.Type,
								}

								switch colDefinition.Type {
								case entity.ColumnTypeVarchar, entity.ColumnTypeLongText:
									appItemField.InitForString()
									values[i] = &appItemField.StringValue

								case entity.ColumnTypeNumber:

									appItemField.InitForFloat64()
									values[i] = &appItemField.Float64Value

								case entity.ColumnTypeBoolean:

									appItemField.InitForBool()
									values[i] = &appItemField.BoolValue

								case entity.ColumnTypeDatetime, entity.ColumnTypeDate, entity.ColumnTypeTimestamp:

									appItemField.InitForTime()
									values[i] = &appItemField.TimeValue

								case entity.ColumnTypeJSON:
									appItemField.InitForJSON()
									values[i] = &appItemField.JSONValue
								default:
									return eris.Errorf("unknown column type %v", colDefinition.Type)
								}

								// add the extra column to the session
								event.ExtraColumns[appItemField.Name] = appItemField
							}
						}
					}
				}
			}

			// map value to dumb interface to ignore it
			if !isValueMapped {
				var dumbInterface interface{}
				values[i] = &dumbInterface
			}

		}
	}

	err := row.Scan(values...)

	// scan error
	if err != nil {
		if sqlscan.NotFound(err) {
			return sql.ErrNoRows
		} else {
			return eris.Wrap(err, "scanCustomEventRow")
		}
	}

	return nil
}
