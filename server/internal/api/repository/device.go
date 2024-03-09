package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (repo *RepositoryImpl) FetchDevices(ctx context.Context, workspace *entity.Workspace, query sq.SelectBuilder, tx *sql.Tx) (devices []*entity.Device, err error) {

	devices = []*entity.Device{}
	var rows *sql.Rows

	// query
	sql, args, errSQL := query.ToSql()

	if errSQL != nil {
		err = eris.Wrap(errSQL, "FetchDevices")
		return
	}

	if tx == nil {

		conn, err := repo.GetWorkspaceConnection(ctx, workspace.ID)

		if err != nil {
			return nil, eris.Wrap(err, "FetchDevices")
		}

		defer conn.Close()

		if rows, err = conn.QueryContext(ctx, sql, args...); err != nil {
			return nil, eris.Wrap(err, "FetchDevices")
		}
	} else {
		if rows, err = tx.QueryContext(ctx, sql, args...); err != nil {
			return nil, eris.Wrap(err, "FetchDevices")
		}
	}

	defer rows.Close()

	for rows.Next() {

		if rows.Err() != nil {
			return nil, eris.Wrap(rows.Err(), "FetchDevices")
		}

		cols, err := rows.Columns()

		if err != nil {
			return nil, eris.Wrap(err, "FetchDevices")
		}

		device := &entity.Device{}

		if err = scanDeviceRow(cols, rows, device, workspace.InstalledApps); err != nil {
			return nil, err
		}

		devices = append(devices, device)
	}

	return
}

func (repo *RepositoryImpl) ListDevicesForUser(ctx context.Context, workspace *entity.Workspace, userID string, orderBy string, tx *sql.Tx) (devices []*entity.Device, err error) {
	query := sq.Select("*").From("device").Where(sq.Eq{"user_id": userID}).OrderBy(orderBy)
	return repo.FetchDevices(ctx, workspace, query, tx)
}

func (repo *RepositoryImpl) FindDeviceByID(ctx context.Context, workspace *entity.Workspace, deviceID string, userID string, tx *sql.Tx) (deviceFound *entity.Device, err error) {

	var rows *sql.Rows
	deviceFound = &entity.Device{}
	deviceFound.ExtraColumns = entity.AppItemFields{}

	if tx == nil {

		var conn *sql.Conn

		conn, err = repo.GetWorkspaceConnection(ctx, workspace.ID)

		if err != nil {
			return nil, err
		}

		defer conn.Close()

		rows, err = conn.QueryContext(ctx, "SELECT * FROM device WHERE user_id = ? AND id = ? LIMIT 1", userID, deviceID)
	} else {
		rows, err = tx.QueryContext(ctx, "SELECT * FROM device WHERE user_id = ? AND id = ? LIMIT 1", userID, deviceID)
	}

	if err != nil {
		return nil, eris.Wrap(err, "FindDeviceByID")
	}

	defer rows.Close()

	// no rows found
	if !rows.Next() {
		return nil, nil
	}

	// extract columns names
	cols, err := rows.Columns()

	if err != nil {
		return nil, eris.Wrap(err, "FindDeviceByID")
	}

	// convert raw data fields to app item fields
	err = scanDeviceRow(cols, rows, deviceFound, workspace.InstalledApps)
	if err != nil {
		return nil, eris.Wrap(err, "FindDeviceByID")
	}

	return deviceFound, nil
}

func (repo *RepositoryImpl) InsertDevice(ctx context.Context, device *entity.Device, tx *sql.Tx) (err error) {

	if device == nil {
		err = eris.New("device is missing")
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
		// "created_at_trunc", // computed field
		// "db_created_at", // computed field
		// "db_updated_at", // computed field
		// "merged_from_user_id",
		"fields_timestamp",
		"user_agent",
		"user_agent_hash",
		"browser",
		"browser_version",
		"browser_version_major",
		"os",
		"os_version",
		"device_type",
		"resolution",
		"language",
		"ad_blocker",
		"in_webview",
		// "extra_columns", // loop over dimensions to add app_custom_fields
	}

	values := []interface{}{
		device.ID,
		device.ExternalID,
		device.UserID,
		device.CreatedAt,
		device.FieldsTimestamp,
		device.UserAgent,
		device.UserAgentHash,
		device.Browser,
		device.BrowserVersion,
		device.BrowserVersionMajor,
		device.OS,
		device.OSVersion,
		device.DeviceType,
		device.Resolution,
		device.Language,
		device.AdBlocker,
		device.InWebview,
	}

	// add extra columns to the query
	// WARNING: values are interfaces, it might not guess well field types...
	// might have to convert them to NullableTypes before adding them
	if device.ExtraColumns != nil && len(device.ExtraColumns) > 0 {
		for field, value := range device.ExtraColumns {
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

	q := sq.Insert("device").Columns(columns...).Values(values...)

	sql, args, errSQL := q.ToSql()

	if errSQL != nil {
		err = eris.Wrapf(errSQL, "build query insert device: %v\n", device)
		return
	}

	_, errExec := tx.ExecContext(ctx, sql, args...)

	if errExec != nil {
		if repo.IsDuplicateEntry(errExec) {
			return eris.Wrap(ErrRowAlreadyExists, "InsertDevice")
		}

		err = eris.Wrap(errExec, "InsertDevice")
		return
	}

	device.DBCreatedAt = now
	device.DBUpdatedAt = now

	return
}

func (repo *RepositoryImpl) UpdateDevice(ctx context.Context, device *entity.Device, tx *sql.Tx) (err error) {

	if device == nil {
		err = eris.New("device is missing")
		return
	}
	if tx == nil {
		err = eris.New("tx is missing")
		return
	}

	now := time.Now()

	// UPDATE
	// specify sharding key to avoid deadlocks
	q := sq.Update("device").Where(sq.Eq{"user_id": device.UserID}).Where(sq.Eq{"id": device.ID}).
		Set("created_at", device.CreatedAt).
		Set("fields_timestamp", device.FieldsTimestamp).
		Set("user_agent", device.UserAgent).
		Set("user_agent_hash", device.UserAgentHash).
		Set("browser", device.Browser).
		Set("browser_version", device.BrowserVersion).
		Set("browser_version_major", device.BrowserVersionMajor).
		Set("os", device.OS).
		Set("os_version", device.OSVersion).
		Set("device_type", device.DeviceType).
		Set("resolution", device.Resolution).
		Set("language", device.Language).
		Set("ad_blocker", device.AdBlocker).
		Set("in_webview", device.InWebview)

	// add extra columns to the query
	// WARNING: values are interfaces, it might not guess well field types...
	// might have to convert them to NullableTypes before adding them
	for field, value := range device.ExtraColumns {
		q = q.Set(field, value)
	}

	sql, args, errSQL := q.ToSql()

	if errSQL != nil {
		err = eris.Wrapf(errSQL, "build query update device: %v\n", device)
		return
	}

	_, errExec := tx.ExecContext(ctx, sql, args...)

	if errExec != nil {
		// if repo.IsDuplicateEntry(errExec) {
		// }

		err = eris.Wrap(errExec, "UpdateDevice")
		return
	}

	device.DBUpdatedAt = now

	return
}

// clones devices from a user to another user with
// because the shard key "user_id" is immutable, we can't use an UPDATE
// we have to INSERT FROM SELECT + DELETE
func (repo *RepositoryImpl) MergeUserDevices(ctx context.Context, workspace *entity.Workspace, fromUserID string, toUserID string, tx *sql.Tx) (err error) {

	// find eventual extra columns for the device table
	deviceCustomColumns := workspace.FindExtraColumnsForItemKind(entity.ItemKindDevice)

	deviceStruct := entity.Device{}
	columns := entity.GetNotComputedDBColumnsForObject(deviceStruct, entity.DeviceComputedFields, deviceCustomColumns)

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

	// IGNORE
	query := fmt.Sprintf(`
		INSERT IGNORE INTO device (%v) 
		SELECT %v FROM device 
		WHERE user_id = ?
	`, strings.Join(columns, ", "), strings.Join(selectedColumns, ", "))

	// log.Println(query)

	if _, err := tx.ExecContext(ctx, query, fromUserID); err != nil {
		return eris.Wrap(err, "MergeUserDevices")
	}

	// BUG deleting these rows might create a deadlock in singlestore
	// delete "from user" devices
	// if _, err := tx.ExecContext(ctx, "DELETE FROM devices WHERE user_id = ? OPTION (columnstore_table_lock_threshold = 5000)", fromUserID); err != nil {
	// 	return err
	// }

	return nil
}

func scanDeviceRow(cols []string, row RowScanner, device *entity.Device, installedApps entity.InstalledApps) error {

	// scan values
	values := make([]interface{}, len(cols))
	// extraColumns := []*ExtraColumn{}

	if device.ExtraColumns == nil {
		device.ExtraColumns = entity.AppItemFields{}
	}

	for i, col := range cols {
		switch col {
		case "id":
			values[i] = &device.ID
		case "external_id":
			values[i] = &device.ExternalID
		case "user_id":
			values[i] = &device.UserID
		case "created_at":
			values[i] = &device.CreatedAt
		case "created_at_trunc":
			values[i] = &device.CreatedAtTrunc
		case "db_created_at":
			values[i] = &device.DBCreatedAt
		case "db_updated_at":
			values[i] = &device.DBUpdatedAt
		case "merged_from_user_id":
			values[i] = &device.MergedFromUserID
		case "fields_timestamp":
			values[i] = &device.FieldsTimestamp
		case "user_agent":
			values[i] = &device.UserAgent
		case "user_agent_hash":
			values[i] = &device.UserAgentHash
		case "browser":
			values[i] = &device.Browser
		case "browser_version":
			values[i] = &device.BrowserVersion
		case "browser_version_major":
			values[i] = &device.BrowserVersionMajor
		case "os":
			values[i] = &device.OS
		case "os_version":
			values[i] = &device.OSVersion
		case "device_type":
			values[i] = &device.DeviceType
		case "resolution":
			values[i] = &device.Resolution
		case "language":
			values[i] = &device.Language
		case "ad_blocker":
			values[i] = &device.AdBlocker
		case "in_webview":
			values[i] = &device.InWebview
		default:
			// handle app extra columns
			if !strings.HasPrefix(col, "app_") && !strings.HasPrefix(col, "appx_") {
				return eris.Errorf("device column not mapped %v", col)
			}

			// find app item field in installedApps
			isValueMapped := false

			for _, app := range installedApps {
				for _, extraColumn := range app.ExtraColumns {
					if extraColumn.Kind == entity.ItemKindSession {

						for _, colDefinition := range extraColumn.Columns {
							if colDefinition.Name == col {
								isValueMapped = true

								// init an extra column in the device
								appItemField := entity.AppItemField{
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

								// add the extra column to the device
								device.ExtraColumns[appItemField.Name] = &appItemField
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
		if err == sql.ErrNoRows {
			return sql.ErrNoRows
		} else {
			return eris.Wrap(err, "scanDeviceRow")
		}
	}

	return nil
}
