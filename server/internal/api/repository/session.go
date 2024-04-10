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

func (repo *RepositoryImpl) FetchSessions(ctx context.Context, workspace *entity.Workspace, query sq.SelectBuilder, tx *sql.Tx) (sessions []*entity.Session, err error) {

	sessions = []*entity.Session{}
	var rows *sql.Rows

	// query
	sql, args, errSQL := query.ToSql()

	if errSQL != nil {
		err = eris.Wrap(errSQL, "FetchSessions")
		return
	}

	if tx == nil {

		conn, err := repo.GetWorkspaceConnection(ctx, workspace.ID)

		if err != nil {
			return nil, eris.Wrap(err, "FetchSessions")
		}

		defer conn.Close()

		if rows, err = conn.QueryContext(ctx, sql, args...); err != nil {
			return nil, eris.Wrap(err, "FetchSessions")
		}
	} else {
		if rows, err = tx.QueryContext(ctx, sql, args...); err != nil {
			return nil, eris.Wrap(err, "FetchSessions")
		}
	}

	defer rows.Close()

	for rows.Next() {

		if rows.Err() != nil {
			return nil, eris.Wrap(rows.Err(), "FetchSessions")
		}

		cols, err := rows.Columns()

		if err != nil {
			return nil, eris.Wrap(err, "FetchSessions")
		}

		session := &entity.Session{}

		if err = scanSessionRow(cols, rows, session, workspace.InstalledApps); err != nil {
			return nil, err
		}

		sessions = append(sessions, session)
	}

	return
}

func (repo *RepositoryImpl) ListSessionsForUser(ctx context.Context, workspace *entity.Workspace, userID string, orderBy string, tx *sql.Tx) (sessions []*entity.Session, err error) {
	query := sq.Select("*").From("session").Where(sq.Eq{"user_id": userID}).OrderBy(orderBy)
	return repo.FetchSessions(ctx, workspace, query, tx)
}

func (repo *RepositoryImpl) InsertSession(ctx context.Context, session *entity.Session, tx *sql.Tx) (err error) {

	if session == nil {
		err = eris.New("session is missing")
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
		"domain_id",
		"device_id",
		"created_at",
		// "created_at_trunc", // computed field
		// "db_created_at", // computed field
		// "db_updated_at", // computed field
		// "merged_from_user_id",
		"fields_timestamp",

		// localized datetime parts for cohorts purpose
		"timezone",
		// "year", // computed field
		// "month", // computed field
		// "month_day", // computed field
		// "week_day", // computed field
		// "hour", // computed field

		// bounce fields:
		"duration",
		// "bounced", // computed
		"pageviews_count",
		"interactions_count",

		// web fields:
		"landing_page",
		"landing_page_path",
		"referrer",
		"referrer_domain",
		"referrer_path",

		// channel mapping:
		"channel_origin_id",
		"channel_id",
		"channel_group_id",

		// utm_ parameters equivalent:
		"utm_source",
		"utm_medium",
		"utm_id",
		"utm_id_from",
		"utm_campaign",
		"utm_content",
		"utm_term",

		// "via utm" fields are used to store a copy of original utm_ parameters in case of overwritte from "data filters"
		"via_utm_source",
		"via_utm_medium",
		"via_utm_id",
		"via_utm_id_from",
		"via_utm_campaign",
		"via_utm_content",
		"via_utm_term",

		// attribution, set when the session contributed to a conversion
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
		session.ID,
		session.ExternalID,
		session.UserID,
		session.DomainID,
		session.DeviceID,
		session.CreatedAt,
		session.FieldsTimestamp,
		// localized datetime parts for cohorts purpose
		session.Timezone,
		// "year", // computed field
		// "month", // computed field
		// "month_day", // computed field
		// "week_day", // computed field
		// "hour", // computed field

		// bounce fields:
		session.Duration,
		// "bounced",
		session.PageviewsCount,
		session.InteractionsCount,

		// web fields:
		session.LandingPage,
		session.LandingPagePath,
		session.Referrer,
		session.ReferrerDomain,
		session.ReferrerPath,

		// channel mapping:
		session.ChannelOriginID,
		session.ChannelID,
		session.ChannelGroupID,

		// utm_ parameters equivalent:
		session.UTMSource,
		session.UTMMedium,
		session.UTMID,
		session.UTMIDFrom,
		session.UTMCampaign,
		session.UTMContent,
		session.UTMTerm,

		// "via utm" fields are used to store a copy of original utm_ parameters in case of overwritte from "data filters"
		session.ViaUTMSource,
		session.ViaUTMMedium,
		session.ViaUTMID,
		session.ViaUTMIDFrom,
		session.ViaUTMCampaign,
		session.ViaUTMContent,
		session.ViaUTMTerm,

		// attribution, set when the session contributed to a conversion
		session.ConversionType,
		session.ConversionID,
		session.ConversionExternalID,
		session.ConversionAt,
		session.ConversionAmount,
		session.LinearAmountAttributed,
		session.LinearPercentageAttributed,
		session.TimeToConversion,
		session.IsFirstConversion,
		session.Role,
	}
	// add extra columns to the query
	// WARNING: values are interfaces, it might not guess well field types...
	// might have to convert them to NullableTypes before adding them
	if session.ExtraColumns != nil && len(session.ExtraColumns) > 0 {
		for field, value := range session.ExtraColumns {
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

	q := sq.Insert("session").Columns(columns...).Values(values...)

	sql, args, errSQL := q.ToSql()

	if errSQL != nil {
		err = eris.Wrapf(errSQL, "build query insert session: %v\n", session)
		return
	}

	_, errExec := tx.ExecContext(ctx, sql, args...)

	if errExec != nil {
		if repo.IsDuplicateEntry(errExec) {
			return eris.Wrap(ErrRowAlreadyExists, "InsertSession")
		}

		err = eris.Wrap(errExec, "InsertSession")
		return
	}

	session.DBCreatedAt = now
	session.DBUpdatedAt = now

	return
}

func (repo *RepositoryImpl) UpdateSession(ctx context.Context, session *entity.Session, tx *sql.Tx) (err error) {

	if session == nil {
		err = eris.New("session is missing")
		return
	}
	if tx == nil {
		err = eris.New("tx is missing")
		return
	}

	now := time.Now()

	// UPDATE
	// specify sharding key to avoid deadlocks

	q := sq.Update("session").Where(sq.Eq{"user_id": session.UserID}).Where(sq.Eq{"id": session.ID}).
		// Set("created_at", session.CreatedAt).
		Set("fields_timestamp", session.FieldsTimestamp).
		Set("timezone", session.Timezone).
		// Set("expires_at", session.ExpiresAt).
		Set("duration", session.Duration).
		Set("pageviews_count", session.PageviewsCount).
		Set("interactions_count", session.InteractionsCount).
		Set("landing_page", session.LandingPage).
		Set("landing_page_path", session.LandingPagePath).
		Set("referrer", session.Referrer).
		Set("referrer_domain", session.ReferrerDomain).
		Set("referrer", session.Referrer).
		// channel
		Set("channel_origin_id", session.ChannelOriginID).
		Set("channel_id", session.ChannelID).
		Set("channel_group_id", session.ChannelGroupID).
		// utm
		Set("utm_source", session.UTMSource).
		Set("utm_medium", session.UTMMedium).
		Set("utm_id", session.UTMID).
		Set("utm_id_from", session.UTMIDFrom).
		Set("utm_campaign", session.UTMCampaign).
		Set("utm_content", session.UTMContent).
		Set("utm_term", session.UTMTerm).
		// via
		Set("via_utm_source", session.ViaUTMSource).
		Set("via_utm_medium", session.ViaUTMMedium).
		Set("via_utm_id", session.ViaUTMID).
		Set("via_utm_id_from", session.ViaUTMIDFrom).
		Set("via_utm_campaign", session.ViaUTMCampaign).
		Set("via_utm_content", session.ViaUTMContent).
		Set("via_utm_term", session.ViaUTMTerm).
		// attribution
		Set("conversion_type", session.ConversionType).
		Set("conversion_id", session.ConversionID).
		Set("conversion_external_id", session.ConversionExternalID).
		Set("conversion_at", session.ConversionAt).
		Set("conversion_amount", session.ConversionAmount).
		Set("linear_amount_attributed", session.LinearAmountAttributed).
		Set("linear_percentage_attributed", session.LinearPercentageAttributed).
		Set("time_to_conversion", session.TimeToConversion).
		Set("is_first_conversion", session.IsFirstConversion).
		Set("role", session.Role)
		// "extra_columns", // loop over dimensions to add app_custom_fields

	// add extra columns to the query
	// WARNING: values are interfaces, it might not guess well field types...
	// might have to convert them to NullableTypes before adding them
	for field, value := range session.ExtraColumns {
		q = q.Set(field, value)
	}

	sql, args, errSQL := q.ToSql()

	if errSQL != nil {
		err = eris.Wrapf(errSQL, "build query update session: %v\n", session)
		return
	}

	_, errExec := tx.ExecContext(ctx, sql, args...)

	if errExec != nil {
		// if repo.IsDuplicateEntry(errExec) {
		// }

		err = eris.Wrap(errExec, "UpdateSession")
		return
	}

	session.DBUpdatedAt = now

	return
}

func (repo *RepositoryImpl) FindSessionByID(ctx context.Context, workspace *entity.Workspace, sessionID string, userID string, tx *sql.Tx) (sessionFound *entity.Session, err error) {

	var rows *sql.Rows
	sessionFound = &entity.Session{}
	sessionFound.ExtraColumns = entity.AppItemFields{}

	if tx == nil {

		var conn *sql.Conn

		conn, err = repo.GetWorkspaceConnection(ctx, workspace.ID)

		if err != nil {
			return nil, err
		}

		defer conn.Close()

		rows, err = conn.QueryContext(ctx, "SELECT * FROM session WHERE user_id = ? AND id = ? LIMIT 1", userID, sessionID)
	} else {
		rows, err = tx.QueryContext(ctx, "SELECT * FROM session WHERE user_id = ? AND id = ? LIMIT 1", userID, sessionID)
	}

	if err != nil {
		return nil, eris.Wrap(err, "FindSessionByID")
	}

	defer rows.Close()

	// no rows found
	if !rows.Next() {
		return nil, nil
	}

	// extract columns names
	cols, err := rows.Columns()

	if err != nil {
		return nil, eris.Wrap(err, "FindSessionByID")
	}

	// convert raw data fields to app item fields
	err = scanSessionRow(cols, rows, sessionFound, workspace.InstalledApps)
	if err != nil {
		return nil, eris.Wrap(err, "FindSessionByID")
	}

	return sessionFound, nil
}

// clones sessions from a user to another user with
// because the shard key "user_id" is immutable, we can't use an UPDATE
// we have to INSERT FROM SELECT + DELETE
func (repo *RepositoryImpl) MergeUserSessions(ctx context.Context, workspace *entity.Workspace, fromUserID string, toUserID string, tx *sql.Tx) (err error) {

	// find eventual extra columns for the session table
	sessionCustomColumns := workspace.FindExtraColumnsForItemKind(entity.ItemKindSession)

	sessionStruct := entity.Session{}
	columns := entity.GetNotComputedDBColumnsForObject(sessionStruct, entity.SessionComputedFields, sessionCustomColumns)

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
		INSERT IGNORE INTO session (%v) 
		SELECT %v FROM session 
		WHERE user_id = ?
	`, strings.Join(columns, ", "), strings.Join(selectedColumns, ", "))

	// log.Println(query)

	if _, err := tx.ExecContext(ctx, query, fromUserID); err != nil {
		return eris.Wrap(err, "MergeUserSessions")
	}

	// BUG deleting these rows might create a deadlock in singlestore
	// // delete "from user" session
	// if _, err := tx.ExecContext(ctx, "DELETE FROM session WHERE user_id = ? OPTION (columnstore_table_lock_threshold = 5000)", fromUserID); err != nil {
	// 	return err
	// }

	return nil
}

func (repo *RepositoryImpl) ResetSessionsAttributedForConversion(ctx context.Context, userID string, conversionID string, tx *sql.Tx) (err error) {

	// specify user_id which is the shard key for performance
	q := sq.Update("session").Where(sq.Eq{"user_id": userID}).Where(sq.Eq{"conversion_id": conversionID}).
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
		err = eris.Wrapf(errSQL, "build query reset sessions attributed for conversion: %v\n", conversionID)
		return
	}

	_, err = tx.ExecContext(ctx, sql, args...)

	if err != nil {
		err = eris.Wrap(err, "UpdateOrderAttribution")
	}

	return
}

func scanSessionRow(cols []string, row RowScanner, session *entity.Session, installedApps entity.InstalledApps) error {

	// scan values
	values := make([]interface{}, len(cols))
	// extraColumns := []*ExtraColumn{}

	if session.ExtraColumns == nil {
		session.ExtraColumns = entity.AppItemFields{}
	}

	for i, col := range cols {
		switch col {
		case "id":
			values[i] = &session.ID
		case "external_id":
			values[i] = &session.ExternalID
		case "user_id":
			values[i] = &session.UserID
		case "domain_id":
			values[i] = &session.DomainID
		case "device_id":
			values[i] = &session.DeviceID
		case "created_at":
			values[i] = &session.CreatedAt
		case "created_at_trunc":
			values[i] = &session.CreatedAtTrunc
		case "db_created_at":
			values[i] = &session.DBCreatedAt
		case "db_updated_at":
			values[i] = &session.DBUpdatedAt
		case "is_deleted":
			values[i] = &session.IsDeleted
		case "merged_from_user_id":
			values[i] = &session.MergedFromUserID
		case "fields_timestamp":
			values[i] = &session.FieldsTimestamp
		case "timezone":
			values[i] = &session.Timezone
		case "year":
			values[i] = &session.Year
		case "month":
			values[i] = &session.Month
		case "month_day":
			values[i] = &session.MonthDay
		case "week_day":
			values[i] = &session.WeekDay
		case "hour":
			values[i] = &session.Hour
		case "duration":
			values[i] = &session.Duration
		case "bounced":
			values[i] = &session.Bounced
		case "pageviews_count":
			values[i] = &session.PageviewsCount
		case "interactions_count":
			values[i] = &session.InteractionsCount
		case "landing_page":
			values[i] = &session.LandingPage
		case "landing_page_path":
			values[i] = &session.LandingPagePath
		case "referrer":
			values[i] = &session.Referrer
		case "referrer_domain":
			values[i] = &session.ReferrerDomain
		case "referrer_path":
			values[i] = &session.ReferrerPath
		case "channel_origin_id":
			values[i] = &session.ChannelOriginID
		case "channel_id":
			values[i] = &session.ChannelID
		case "channel_group_id":
			values[i] = &session.ChannelGroupID
		case "utm_source":
			values[i] = &session.UTMSource
		case "utm_medium":
			values[i] = &session.UTMMedium
		case "utm_campaign":
			values[i] = &session.UTMCampaign
		case "utm_id":
			values[i] = &session.UTMID
		case "utm_id_from":
			values[i] = &session.UTMIDFrom
		case "utm_content":
			values[i] = &session.UTMContent
		case "utm_term":
			values[i] = &session.UTMTerm
		case "via_utm_source":
			values[i] = &session.ViaUTMSource
		case "via_utm_medium":
			values[i] = &session.ViaUTMMedium
		case "via_utm_campaign":
			values[i] = &session.ViaUTMCampaign
		case "via_utm_id":
			values[i] = &session.ViaUTMID
		case "via_utm_id_from":
			values[i] = &session.ViaUTMIDFrom
		case "via_utm_content":
			values[i] = &session.ViaUTMContent
		case "via_utm_term":
			values[i] = &session.ViaUTMTerm
		case "conversion_type":
			values[i] = &session.ConversionType
		case "conversion_external_id":
			values[i] = &session.ConversionExternalID
		case "conversion_id":
			values[i] = &session.ConversionID
		case "conversion_at":
			values[i] = &session.ConversionAt
		case "conversion_amount":
			values[i] = &session.ConversionAmount
		case "linear_amount_attributed":
			values[i] = &session.LinearAmountAttributed
		case "linear_percentage_attributed":
			values[i] = &session.LinearPercentageAttributed
		case "time_to_conversion":
			values[i] = &session.TimeToConversion
		case "is_first_conversion":
			values[i] = &session.IsFirstConversion
		case "role":
			values[i] = &session.Role
		default:
			// handle app extra columns
			if !strings.HasPrefix(col, "app_") && !strings.HasPrefix(col, "appx_") {
				return eris.Errorf("session column not mapped %v", col)
			}

			// find app item field in installedApps
			isValueMapped := false

			for _, app := range installedApps {
				for _, extraColumn := range app.ExtraColumns {
					if extraColumn.Kind == entity.ItemKindSession {

						for _, colDefinition := range extraColumn.Columns {
							if colDefinition.Name == col {
								isValueMapped = true

								// init an extra column in the session
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

								// add the extra column to the session
								session.ExtraColumns[appItemField.Name] = &appItemField
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
			return eris.Wrap(err, "scanSessionRow")
		}
	}

	return nil
}
