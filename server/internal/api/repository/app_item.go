package repository

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/asaskevich/govalidator"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (repo *RepositoryImpl) FetchAppItems(ctx context.Context, workspace *entity.Workspace, kind string, query sq.SelectBuilder, tx *sql.Tx) (items []*entity.AppItem, err error) {

	// find app table in workspace
	var tableDefinition = workspace.FindAppTableDefinitionForItem(kind)

	if tableDefinition == nil {
		err = eris.Wrapf(entity.ErrAppTableNotFound, "FetchAppItems: table definition %v not found", kind)
		return
	}

	items = []*entity.AppItem{}
	var rows *sql.Rows

	query = query.From(kind)

	q, args, errSQL := query.ToSql()

	if errSQL != nil {
		err = eris.Wrapf(errSQL, "build query FetchAppItems: %v\n", kind)
		return
	}

	if tx == nil {

		var conn *sql.Conn
		conn, err = repo.GetWorkspaceConnection(ctx, workspace.ID)

		if err != nil {
			return nil, err
		}

		defer conn.Close()

		rows, err = conn.QueryContext(ctx, q, args...)

	} else {
		rows, err = tx.QueryContext(ctx, q, args...)
	}

	if err != nil {
		log.Printf("with query: %v", q)
		return nil, eris.Wrap(err, "FetchAppItems")
	}

	defer rows.Close()

	now := time.Now()
	cols, err := rows.Columns()

	if err != nil {
		return nil, eris.Wrap(err, "FetchAppItems")
	}

	for rows.Next() {

		if rows.Err() != nil {
			return nil, eris.Wrap(rows.Err(), "FetchAppItems")
		}

		item := entity.NewAppItem(kind, "", "", now, nil)

		err = rawDataToAppItem(cols, rows, tableDefinition, item)

		if err != nil {
			if err == sql.ErrNoRows {
				return items, nil
			}

			return nil, eris.Wrap(err, "FetchAppItems")
		}

		items = append(items, item)
	}

	return items, nil
}

func (repo *RepositoryImpl) DeleteAppItemByExternalID(ctx context.Context, workspace *entity.Workspace, kind string, externalID string, tx *sql.Tx) (err error) {
	// compute id from external id
	id := fmt.Sprintf("%x", sha1.Sum([]byte(externalID)))
	return repo.DeleteAppItemByID(ctx, workspace, kind, id, tx)
}

func (repo *RepositoryImpl) DeleteAppItemByID(ctx context.Context, workspace *entity.Workspace, kind string, ID string, tx *sql.Tx) (err error) {

	// find app table in workspace
	var tableDefinition = workspace.FindAppTableDefinitionForItem(kind)

	if tableDefinition == nil {
		return eris.Wrapf(entity.ErrAppTableNotFound, "DeleteAppItemByID: table definition %v not found", kind)
	}

	if tx == nil {

		var conn *sql.Conn

		conn, err = repo.GetWorkspaceConnection(ctx, workspace.ID)

		if err != nil {
			return err
		}

		defer conn.Close()

		_, err = conn.ExecContext(ctx, fmt.Sprintf("DELETE FROM %v WHERE id = ? LIMIT 1", kind), ID)

		if err != nil {
			return eris.Wrap(err, "DeleteAppItemByID")
		}

		_, err = conn.ExecContext(ctx, "DELETE FROM data_log WHERE kind = ? AND id = ? LIMIT 1", kind, ID)

	} else {
		_, err = tx.ExecContext(ctx, fmt.Sprintf("DELETE FROM %v WHERE item_id = ? LIMIT 1", kind), ID)

		if err != nil {
			return eris.Wrap(err, "DeleteAppItemByID")
		}

		_, err = tx.ExecContext(ctx, "DELETE FROM data_log WHERE kind = ? AND id = ? LIMIT 1", kind, ID)
	}

	if err != nil {
		return eris.Wrap(err, "DeleteAppItemByID")
	}

	return nil
}

func (repo *RepositoryImpl) FindAppItemByExternalID(ctx context.Context, workspace *entity.Workspace, kind string, externalID string, tx *sql.Tx) (item *entity.AppItem, err error) {

	// find app table in workspace
	var tableDefinition = workspace.FindAppTableDefinitionForItem(kind)

	if tableDefinition == nil {
		err = eris.Wrapf(entity.ErrAppTableNotFound, "FindAppItemByExternalID: table definition %v not found", kind)
		return
	}

	var rows *sql.Rows

	if tx == nil {

		var conn *sql.Conn

		conn, err = repo.GetWorkspaceConnection(ctx, workspace.ID)

		if err != nil {
			return nil, err
		}

		defer conn.Close()

		rows, err = conn.QueryContext(ctx, fmt.Sprintf("SELECT * FROM %v WHERE external_id = ? LIMIT 1", kind), externalID)

	} else {
		rows, err = tx.QueryContext(ctx, fmt.Sprintf("SELECT * FROM %v WHERE external_id = ? LIMIT 1", kind), externalID)
	}

	defer rows.Close()

	// no rows found
	if !rows.Next() {
		return nil, nil
	}

	// extract columns names
	cols, err := rows.Columns()

	if err != nil {
		return nil, eris.Wrap(err, "FindAppItemByExternalID")
	}

	item = entity.NewAppItem(kind, "", "", time.Now(), nil)

	// convert raw data fields to app item fields
	err = rawDataToAppItem(cols, rows, tableDefinition, item)
	if err != nil {
		return nil, eris.Wrap(err, "FindAppItemByExternalID")
	}

	return item, nil
}
func (repo *RepositoryImpl) FindAppItemByID(ctx context.Context, workspace *entity.Workspace, kind string, id string, tx *sql.Tx) (item *entity.AppItem, err error) {

	// find app table in workspace
	var tableDefinition = workspace.FindAppTableDefinitionForItem(kind)

	if tableDefinition == nil {
		err = eris.Wrapf(entity.ErrAppTableNotFound, "FindAppItemByID: table definition %v not found", kind)
		return
	}

	var rows *sql.Rows

	if tx == nil {

		var conn *sql.Conn

		conn, err = repo.GetWorkspaceConnection(ctx, workspace.ID)

		if err != nil {
			return nil, err
		}

		defer conn.Close()

		rows, err = conn.QueryContext(ctx, fmt.Sprintf("SELECT * FROM %v WHERE id = ? LIMIT 1", kind), id)

	} else {
		rows, err = tx.QueryContext(ctx, fmt.Sprintf("SELECT * FROM %v WHERE id = ? LIMIT 1", kind), id)
	}

	if err != nil {
		return nil, eris.Wrap(err, "FindAppItemByID")
	}

	defer rows.Close()

	// no rows found
	if !rows.Next() {
		return nil, nil
	}

	// extract columns names
	cols, err := rows.Columns()

	if err != nil {
		return nil, eris.Wrap(err, "FindAppItemByID")
	}

	item = entity.NewAppItem(kind, "", "", time.Now(), nil)

	// convert raw data fields to app item fields
	err = rawDataToAppItem(cols, rows, tableDefinition, item)
	if err != nil {
		return nil, eris.Wrap(err, "FindAppItemByID")
	}

	return item, nil
}

func (repo *RepositoryImpl) InsertAppItem(ctx context.Context, kind string, upsertedAppItem *entity.AppItem, tx *sql.Tx) (err error) {

	if upsertedAppItem == nil {
		err = eris.New("app item is missing")
		return
	}
	if tx == nil {
		err = eris.New("tx is missing")
		return
	}

	columns := []string{
		"id",
		"external_id",
		"created_at",
		"fields_timestamp",
	}

	values := []interface{}{
		upsertedAppItem.ID,
		upsertedAppItem.ExternalID,
		upsertedAppItem.CreatedAt,
		upsertedAppItem.FieldsTimestamp,
	}

	if upsertedAppItem.UserID != entity.None {
		columns = append(columns, "user_id")
		values = append(values, upsertedAppItem.UserID)
	}

	now := time.Now()

	builder := sq.Insert(kind)

	for _, field := range upsertedAppItem.Fields {

		// ignore reserved columns
		if govalidator.IsIn(field.Name, entity.ReservedColumns...) {
			continue
		}

		if field.Type == entity.TableColumnTypeBoolean {
			columns = append(columns, field.Name)
			if field.BoolValue.IsNull {
				values = append(values, nil)
			} else {
				values = append(values, field.BoolValue.Bool)
			}
			continue
		}

		if field.Type == entity.TableColumnTypeNumber {
			columns = append(columns, field.Name)
			if field.Float64Value.IsNull {
				values = append(values, nil)
			} else {
				values = append(values, field.Float64Value.Float64)
			}
			continue
		}

		if field.Type == entity.TableColumnTypeDate || field.Type == entity.TableColumnTypeDatetime || field.Type == entity.TableColumnTypeTimestamp {
			columns = append(columns, field.Name)
			if field.TimeValue.IsNull {
				values = append(values, nil)
			} else {
				values = append(values, field.TimeValue.Time)
			}
			continue
		}

		if field.Type == entity.TableColumnTypeVarchar || field.Type == entity.TableColumnTypeLongText {
			columns = append(columns, field.Name)
			if field.StringValue.IsNull {
				values = append(values, nil)
			} else {
				values = append(values, field.StringValue.String)
			}
			continue
		}

		if field.Type == entity.TableColumnTypeJSON {
			columns = append(columns, field.Name)
			if field.JSONValue.IsNull {
				values = append(values, nil)
			} else {
				values = append(values, field.JSONValue.JSON)
			}
			continue
		}
	}

	query, args, err := builder.Columns(columns...).Values(values...).ToSql()

	if err != nil {
		err = eris.Wrap(err, "InsertAppItem")
		return
	}

	_, err = tx.ExecContext(ctx, query, args...)

	if err != nil {

		if strings.Contains(err.Error(), "Error 1146: Table") {
			return eris.Wrap(entity.ErrAppTableNotFound, "InsertAppItem")
		}

		err = eris.Wrapf(err, "InsertAppItem: %v, %+v", query, args)
		return
	}

	upsertedAppItem.DBCreatedAt = now
	upsertedAppItem.DBUpdatedAt = now
	return
}

func (repo *RepositoryImpl) UpdateAppItem(ctx context.Context, kind string, upsertedAppItem *entity.AppItem, tx *sql.Tx) (err error) {

	if upsertedAppItem == nil {
		err = eris.New("app item is missing")
		return
	}
	if tx == nil {
		err = eris.New("tx is missing")
		return
	}

	builder := sq.Update(kind).
		Set("user_id", upsertedAppItem.UserID).
		Set("created_at", upsertedAppItem.CreatedAt).
		Set("fields_timestamp", upsertedAppItem.FieldsTimestamp).
		Where(sq.Eq{"id": upsertedAppItem.ID})

	for _, field := range upsertedAppItem.Fields {

		// skip reserved fields
		if govalidator.IsIn(field.Name, entity.ReservedColumns...) {
			continue
		}

		if field.Type == entity.TableColumnTypeBoolean {
			if field.BoolValue.IsNull {
				builder = builder.Set(field.Name, nil)
			} else {
				builder = builder.Set(field.Name, field.BoolValue.Bool)
			}
			continue
		}

		if field.Type == entity.TableColumnTypeNumber {
			if field.Float64Value.IsNull {
				builder = builder.Set(field.Name, nil)
			} else {
				builder = builder.Set(field.Name, field.Float64Value.Float64)
			}
			continue
		}

		if field.Type == entity.TableColumnTypeDate || field.Type == entity.TableColumnTypeDatetime || field.Type == entity.TableColumnTypeTimestamp {
			if field.TimeValue.IsNull {
				builder = builder.Set(field.Name, nil)
			} else {
				builder = builder.Set(field.Name, field.TimeValue.Time)
			}
			continue
		}

		if field.Type == entity.TableColumnTypeVarchar || field.Type == entity.TableColumnTypeLongText {
			if field.StringValue.IsNull {
				builder = builder.Set(field.Name, nil)
			} else {
				builder = builder.Set(field.Name, field.StringValue.String)
			}
			continue
		}

		if field.Type == entity.TableColumnTypeJSON {
			if field.JSONValue.IsNull {
				builder = builder.Set(field.Name, nil)
			} else {
				builder = builder.Set(field.Name, field.JSONValue.JSON)
			}
			continue
		}
	}

	query, args, err := builder.ToSql()

	if err != nil {
		err = eris.Wrap(err, "UpdateAppItem")
		return
	}

	_, err = tx.ExecContext(ctx, query, args...)

	if err != nil {
		if strings.Contains(err.Error(), "Error 1146: Table") {
			return eris.Wrap(entity.ErrAppTableNotFound, "UpdateAppItem")
		}
		err = eris.Wrap(err, "UpdateAppItem")
		return
	}

	upsertedAppItem.DBUpdatedAt = time.Now()

	return
}

// populate app item with raw DB data
// use Scan instead of MapScan, as MapScan considers "int strings" as ints...
func rawDataToAppItem(cols []string, rows RowScanner, tableDefinition *entity.AppTableManifest, item *entity.AppItem) (err error) {

	scanValues := make([]interface{}, len(cols))
	// columnsDefinitionsByName := map[string]*entity.TableColumn{}

	for i := range cols {

		// default
		var value interface{}
		scanValues[i] = &value

		// find column definition and generate a custom item field that will be populated by the scanner
		for _, field := range tableDefinition.Columns {
			if field.Name == cols[i] {
				// columnsDefinitionsByName[cols[i]] = field

				var appItemField *entity.AppItemField

				switch field.Type {
				case entity.TableColumnTypeBoolean:
					appItemField = &entity.AppItemField{
						Name:      field.Name,
						Type:      field.Type,
						BoolValue: entity.NullableBool{},
					}
					scanValues[i] = &appItemField.BoolValue

				case entity.TableColumnTypeNumber:
					appItemField = &entity.AppItemField{
						Name:         field.Name,
						Type:         field.Type,
						Float64Value: entity.NullableFloat64{},
					}
					scanValues[i] = &appItemField.Float64Value

				case entity.TableColumnTypeDate, entity.TableColumnTypeDatetime, entity.TableColumnTypeTimestamp:
					appItemField = &entity.AppItemField{
						Name:      field.Name,
						Type:      field.Type,
						TimeValue: entity.NullableTime{},
					}
					scanValues[i] = &appItemField.TimeValue

				case entity.TableColumnTypeVarchar, entity.TableColumnTypeLongText:
					appItemField = &entity.AppItemField{
						Name:        field.Name,
						Type:        field.Type,
						StringValue: entity.NullableString{},
					}
					scanValues[i] = &appItemField.StringValue

				case entity.TableColumnTypeJSON:
					appItemField = &entity.AppItemField{
						Name:      field.Name,
						Type:      field.Type,
						JSONValue: entity.NullableJSON{},
					}
					scanValues[i] = &appItemField.JSONValue

				default:
					err = eris.Errorf("rawDataToAppItem unknown column type %s", field.Type)
					return
				}

				// append to item fields
				item.Fields[appItemField.Name] = appItemField
			}
		}
	}

	err = rows.Scan(scanValues...)

	if err == sql.ErrNoRows {
		return err
	}

	if err != nil {
		return eris.Wrap(err, "rawDataToAppItem")
	}

	// populate reserved/required fields at the item level
	for _, field := range item.Fields {
		if field.Name == "id" {
			if !field.StringValue.IsNull {
				item.ID = field.StringValue.String
			}
		}

		if field.Name == "external_id" {
			if !field.StringValue.IsNull {
				item.ExternalID = field.StringValue.String
			}
		}

		if field.Name == "created_at" {
			if !field.TimeValue.IsNull {
				item.CreatedAt = field.TimeValue.Time
			}
		}

		if field.Name == "user_id" {
			if !field.StringValue.IsNull {
				item.UserID = field.StringValue.String
			}
		}

		if field.Name == "merged_from_user_id" {
			if !field.StringValue.IsNull {
				item.MergedFromUserID = &field.StringValue.String
			}
		}

		if field.Name == "fields_timestamp" {
			// init
			item.FieldsTimestamp = entity.FieldsTimestamp{}

			if field.Type == entity.TableColumnTypeJSON {
				if !field.JSONValue.IsNull {
					if err = json.Unmarshal(field.JSONValue.JSON, &item.FieldsTimestamp); err != nil {
						return eris.Wrap(err, "rawDataToAppItem")
					}
				}
			}
		}

		if field.Name == "db_created_at" {
			if !field.TimeValue.IsNull {
				item.DBCreatedAt = field.TimeValue.Time
			}
		}

		if field.Name == "db_updated_at" {
			if !field.TimeValue.IsNull {
				item.DBUpdatedAt = field.TimeValue.Time
			}
		}
	}

	return nil
}
