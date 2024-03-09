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

func (repo *RepositoryImpl) FetchPostviews(ctx context.Context, workspace *entity.Workspace, query sq.SelectBuilder, tx *sql.Tx) (postviews []*entity.Postview, err error) {

	postviews = []*entity.Postview{}
	var rows *sql.Rows

	// query
	sql, args, errSQL := query.ToSql()

	if errSQL != nil {
		return nil, eris.Wrap(errSQL, "FetchPostviews")
	}

	if tx == nil {

		conn, err := repo.GetWorkspaceConnection(ctx, workspace.ID)

		if err != nil {
			return nil, eris.Wrap(err, "FetchPostviews")
		}

		defer conn.Close()

		rows, err = conn.QueryContext(ctx, sql, args...)
	} else {
		rows, err = tx.QueryContext(ctx, sql, args...)
	}

	if err != nil {
		return nil, eris.Wrap(err, "FetchPostviews")
	}

	defer rows.Close()

	for rows.Next() {

		if rows.Err() != nil {
			return nil, eris.Wrap(rows.Err(), "FetchPostviews")
		}

		cols, err := rows.Columns()

		if err != nil {
			return nil, eris.Wrap(err, "FetchPostviews")
		}

		postview := &entity.Postview{}

		if err = scanPostviewRow(cols, rows, postview, workspace.InstalledApps); err != nil {
			return nil, err
		}

		postviews = append(postviews, postview)
	}

	return
}

func (repo *RepositoryImpl) ListPostviewsForUser(ctx context.Context, workspace *entity.Workspace, userID string, orderBy string, tx *sql.Tx) (postviews []*entity.Postview, err error) {
	query := sq.Select("*").From("postview").Where(sq.Eq{"user_id": userID}).OrderBy(orderBy)
	return repo.FetchPostviews(ctx, workspace, query, tx)
}

func (repo *RepositoryImpl) InsertPostview(ctx context.Context, postview *entity.Postview, tx *sql.Tx) (err error) {

	if postview == nil {
		err = eris.New("postview is missing")
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
		"device_id",
		"created_at",
		// "created_at_trunc", // computed field
		// "db_created_at", // computed field
		// "db_updated_at", // computed field
		// "merged_from_user_id",
		"fields_timestamp",

		// channel mapping:
		"channel_origin_id", // computed field
		"channel_id",
		"channel_group_id",

		// utm_ parameters equivalent:
		"utm_source",
		"utm_medium",
		// "utm_id",
		// "utm_id_from",
		"utm_campaign",
		"utm_content",
		"utm_term",

		"country",
		"region",

		// attribution, set when the postview contributed to a conversion
		"conversion_type",
		"conversion_id",
		"conversion_external_id",
		"conversion_at",
		"conversion_amount",
		"linear_amount_attributed",
		"linear_percentage_attributed",
		"time_to_conversion",
		"is_first_conversion",
		"role",
		// "extra_columns", // loop over dimensions to add app_custom_fields
	}

	values := []interface{}{
		postview.ID,
		postview.ExternalID,
		postview.UserID,
		postview.DeviceID,
		postview.CreatedAt,
		postview.FieldsTimestamp,

		// channel mapping:
		postview.ChannelOriginID,
		postview.ChannelID,
		postview.ChannelGroupID,

		// utm_ parameters equivalent:
		postview.UTMSource,
		postview.UTMMedium,
		// postview.UTMID,
		// postview.UTMIDFrom,
		postview.UTMCampaign,
		postview.UTMContent,
		postview.UTMTerm,

		postview.Country,
		postview.Region,

		// attribution, set when the postview contributed to a conversion
		postview.ConversionType,
		postview.ConversionID,
		postview.ConversionExternalID,
		postview.ConversionAt,
		postview.ConversionAmount,
		postview.LinearAmountAttributed,
		postview.LinearPercentageAttributed,
		postview.TimeToConversion,
		postview.IsFirstConversion,
		postview.Role,
	}
	// add extra columns to the query
	// WARNING: values are interfaces, it might not guess well field types...
	// might have to convert them to NullableTypes before adding them
	if postview.ExtraColumns != nil && len(postview.ExtraColumns) > 0 {
		for field, value := range postview.ExtraColumns {
			columns = append(columns, field)
			values = append(values, value)
		}
	}
	q := sq.Insert("postview").Columns(columns...).Values(values...)

	sql, args, errSQL := q.ToSql()

	if errSQL != nil {
		err = eris.Wrapf(errSQL, "build query insert postview: %v\n", postview)
		return
	}

	_, errExec := tx.ExecContext(ctx, sql, args...)

	if errExec != nil {
		if repo.IsDuplicateEntry(errExec) {
			return eris.Wrap(ErrRowAlreadyExists, "InsertPostview")
		}

		err = eris.Wrap(errExec, "InsertPostview")
		return
	}

	postview.DBCreatedAt = now
	postview.DBUpdatedAt = now

	return
}

func (repo *RepositoryImpl) UpdatePostview(ctx context.Context, postview *entity.Postview, tx *sql.Tx) (err error) {

	if postview == nil {
		err = eris.New("postview is missing")
		return
	}
	if tx == nil {
		err = eris.New("tx is missing")
		return
	}

	now := time.Now()

	// specify sharding key user_id to avoid deadlocks
	q := sq.Update("postview").Where(sq.Eq{"user_id": postview.UserID}).Where(sq.Eq{"id": postview.ID}).
		// Set("created_at", postview.CreatedAt).
		Set("fields_timestamp", postview.FieldsTimestamp).
		// channel
		Set("channel_origin_id", postview.ChannelOriginID).
		Set("channel_id", postview.ChannelID).
		Set("channel_group_id", postview.ChannelGroupID).
		// utm
		Set("utm_source", postview.UTMSource).
		Set("utm_medium", postview.UTMMedium).
		// Set("utm_id", postview.UTMID).
		// Set("utm_id_from", postview.UTMIDFrom).
		Set("utm_campaign", postview.UTMCampaign).
		Set("utm_content", postview.UTMContent).
		Set("utm_term", postview.UTMTerm).
		Set("country", postview.Country).
		Set("region", postview.Region).
		// attribution
		Set("conversion_type", postview.ConversionType).
		Set("conversion_id", postview.ConversionID).
		Set("conversion_external_id", postview.ConversionExternalID).
		Set("conversion_at", postview.ConversionAt).
		Set("conversion_amount", postview.ConversionAmount).
		Set("linear_amount_attributed", postview.LinearAmountAttributed).
		Set("linear_percentage_attributed", postview.LinearPercentageAttributed).
		Set("time_to_conversion", postview.TimeToConversion).
		Set("is_first_conversion", postview.IsFirstConversion).
		Set("role", postview.Role)
		// "extra_columns", // loop over dimensions to add app_custom_fields

	// add extra columns to the query
	// WARNING: values are interfaces, it might not guess well field types...
	// might have to convert them to NullableTypes before adding them
	for field, value := range postview.ExtraColumns {
		q = q.Set(field, value)
	}
	sql, args, errSQL := q.ToSql()

	if errSQL != nil {
		err = eris.Wrapf(errSQL, "build query update postview: %v\n", postview)
		return
	}

	_, errExec := tx.ExecContext(ctx, sql, args...)

	if errExec != nil {
		// if repo.IsDuplicateEntry(errExec) {
		// }

		err = eris.Wrap(errExec, "UpdatePostview")
		return
	}

	postview.DBUpdatedAt = now

	return
}

func (repo *RepositoryImpl) FindPostviewByID(ctx context.Context, workspace *entity.Workspace, postviewID string, userID string, tx *sql.Tx) (postviewFound *entity.Postview, err error) {

	var rows *sql.Rows
	postviewFound = &entity.Postview{}
	postviewFound.ExtraColumns = entity.AppItemFields{}

	if tx == nil {

		var conn *sql.Conn

		conn, err = repo.GetWorkspaceConnection(ctx, workspace.ID)

		if err != nil {
			return nil, err
		}

		defer conn.Close()

		rows, err = conn.QueryContext(ctx, "SELECT * FROM postview WHERE user_id = ? AND id = ? LIMIT 1", userID, postviewID)
	} else {
		rows, err = tx.QueryContext(ctx, "SELECT * FROM postview WHERE user_id = ? AND id = ? LIMIT 1", userID, postviewID)
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
		return nil, eris.Wrap(err, "FindPostviewByID")
	}

	// convert raw data fields to app item fields
	err = scanPostviewRow(cols, rows, postviewFound, workspace.InstalledApps)
	if err != nil {
		return nil, eris.Wrap(err, "FindPostviewByID")
	}

	return postviewFound, nil
}

// clones postviews from a user to another user with
// because the shard key "user_id" is immutable, we can't use an UPDATE
// we have to INSERT FROM SELECT + DELETE
func (repo *RepositoryImpl) MergeUserPostviews(ctx context.Context, workspace *entity.Workspace, fromUserID string, toUserID string, tx *sql.Tx) (err error) {

	// find eventual extra columns for the postview table
	postviewCustomColumns := workspace.FindExtraColumnsForItemKind(entity.ItemKindPostview)

	postviewStruct := entity.Postview{}
	columns := entity.GetNotComputedDBColumnsForObject(postviewStruct, entity.PostviewComputedFields, postviewCustomColumns)

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
		INSERT IGNORE INTO postview (%v) 
		SELECT %v FROM postview 
		WHERE user_id = ?
	`, strings.Join(columns, ", "), strings.Join(selectedColumns, ", "))

	// log.Println(query)

	if _, err := tx.ExecContext(ctx, query, fromUserID); err != nil {
		return eris.Wrap(err, "MergeUserPostviews")
	}

	// BUG deleting these rows might create a deadlock in singlestore
	// // delete "from user" postview
	// if _, err := tx.ExecContext(ctx, "DELETE FROM postview WHERE user_id = ? OPTION (columnstore_table_lock_threshold = 5000)", fromUserID); err != nil {
	// 	return err
	// }

	return nil
}

func (repo *RepositoryImpl) ResetPostviewsAttributedForConversion(ctx context.Context, userID string, conversionID string, tx *sql.Tx) (err error) {

	// specify user_id which is the shard key for performance
	q := sq.Update("postview").Where(sq.Eq{"user_id": userID}).Where(sq.Eq{"conversion_id": conversionID}).
		Set("conversion_type", nil).
		Set("conversion_external_id", nil).
		Set("conversion_id", nil).
		Set("conversion_at", nil).
		Set("conversion_amount", 0).
		Set("linear_amount_attributed", 0).
		Set("linear_percentage_attributed", 0).
		Set("time_to_conversion", 0).
		Set("is_first_conversion", false).
		Set("role", nil)

	sql, args, errSQL := q.ToSql()

	if errSQL != nil {
		err = eris.Wrapf(errSQL, "build query reset postviews attributed for conversion: %v\n", conversionID)
		return
	}

	_, err = tx.ExecContext(ctx, sql, args...)

	if err != nil {
		err = eris.Wrap(err, "UpdateOrderAttribution")
	}

	return
}

func scanPostviewRow(cols []string, row RowScanner, postview *entity.Postview, installedApps entity.InstalledApps) error {

	// scan values
	values := make([]interface{}, len(cols))
	// extraColumns := []*ExtraColumn{}

	if postview.ExtraColumns == nil {
		postview.ExtraColumns = entity.AppItemFields{}
	}

	for i, col := range cols {
		switch col {
		case "id":
			values[i] = &postview.ID
		case "external_id":
			values[i] = &postview.ExternalID
		case "user_id":
			values[i] = &postview.UserID
		// case "domain_id":
		// 	values[i] = &postview.DomainID
		case "device_id":
			values[i] = &postview.DeviceID
		case "created_at":
			values[i] = &postview.CreatedAt
		case "created_at_trunc":
			values[i] = &postview.CreatedAtTrunc
		case "db_created_at":
			values[i] = &postview.DBCreatedAt
		case "db_updated_at":
			values[i] = &postview.DBUpdatedAt
		case "is_deleted":
			values[i] = &postview.IsDeleted
		case "merged_from_user_id":
			values[i] = &postview.MergedFromUserID
		case "fields_timestamp":
			values[i] = &postview.FieldsTimestamp
		// case "timezone":
		// 	values[i] = &postview.Timezone
		// case "year":
		// 	values[i] = &postview.Year
		// case "month":
		// 	values[i] = &postview.Month
		// case "month_day":
		// 	values[i] = &postview.MonthDay
		// case "week_day":
		// 	values[i] = &postview.WeekDay
		// case "hour":
		// 	values[i] = &postview.Hour
		case "channel_origin_id":
			values[i] = &postview.ChannelOriginID
		case "channel_id":
			values[i] = &postview.ChannelID
		case "channel_group_id":
			values[i] = &postview.ChannelGroupID
		case "utm_source":
			values[i] = &postview.UTMSource
		case "utm_medium":
			values[i] = &postview.UTMMedium
		case "utm_campaign":
			values[i] = &postview.UTMCampaign
		// case "utm_id":
		// 	values[i] = &postview.UTMID
		// case "utm_id_from":
		// 	values[i] = &postview.UTMIDFrom
		case "utm_content":
			values[i] = &postview.UTMContent
		case "utm_term":
			values[i] = &postview.UTMTerm
		case "country":
			values[i] = &postview.Country
		case "region":
			values[i] = &postview.Region
		case "conversion_type":
			values[i] = &postview.ConversionType
		case "conversion_external_id":
			values[i] = &postview.ConversionExternalID
		case "conversion_id":
			values[i] = &postview.ConversionID
		case "conversion_at":
			values[i] = &postview.ConversionAt
		case "conversion_amount":
			values[i] = &postview.ConversionAmount
		case "linear_amount_attributed":
			values[i] = &postview.LinearAmountAttributed
		case "linear_percentage_attributed":
			values[i] = &postview.LinearPercentageAttributed
		case "time_to_conversion":
			values[i] = &postview.TimeToConversion
		case "is_first_conversion":
			values[i] = &postview.IsFirstConversion
		case "role":
			values[i] = &postview.Role
		default:
			// handle app extra columns
			if !strings.HasPrefix(col, "app_") && !strings.HasPrefix(col, "appx_") {
				return eris.Errorf("postview column not mapped %v", col)
			}

			// find app item field in installedApps
			isValueMapped := false

			for _, app := range installedApps {
				for _, extraColumn := range app.ExtraColumns {
					if extraColumn.Kind == entity.ItemKindPostview {

						for _, colDefinition := range extraColumn.Columns {
							if colDefinition.Name == col {
								isValueMapped = true

								// init an extra column in the postview
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

								// add the extra column to the postview
								postview.ExtraColumns[appItemField.Name] = &appItemField
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
			return eris.Wrap(err, "scanPostviewRow")
		}
	}

	return nil
}
