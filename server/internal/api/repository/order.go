package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (repo *RepositoryImpl) FindUserIDsWithOrdersToReattribute(ctx context.Context, workspaceID string, limit int) (userIDs []string, err error) {

	var conn *sql.Conn

	conn, err = repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	userIDs = []string{}

	sql := "SELECT DISTINCT user_id FROM `order` WHERE attribution_updated_at IS NULL LIMIT ?"

	err = sqlscan.Select(ctx, conn, &userIDs, sql, limit)

	if err != nil {
		return nil, eris.Wrap(err, "FindUserIDsWithOrdersToReattribute")
	}

	return
}

func (repo *RepositoryImpl) FindOrderByID(ctx context.Context, workspace *entity.Workspace, orderID string, userID string, tx *sql.Tx) (orderFound *entity.Order, err error) {

	var rows *sql.Rows
	orderFound = &entity.Order{}
	orderFound.ExtraColumns = entity.AppItemFields{}

	if tx == nil {

		var conn *sql.Conn

		conn, err = repo.GetWorkspaceConnection(ctx, workspace.ID)

		if err != nil {
			return nil, err
		}

		defer conn.Close()

		rows, err = conn.QueryContext(ctx, "SELECT * FROM `order` WHERE user_id = ? AND id = ? LIMIT 1", userID, orderID)
	} else {
		rows, err = tx.QueryContext(ctx, "SELECT * FROM `order` WHERE user_id = ? AND id = ? LIMIT 1", userID, orderID)
	}

	if err != nil {
		return nil, eris.Wrap(err, "FindOrderByID")
	}

	defer rows.Close()

	// no rows found
	if !rows.Next() {
		return nil, nil
	}

	// extract columns names
	cols, err := rows.Columns()

	if err != nil {
		return nil, eris.Wrap(err, "FindOrderByID")
	}

	// convert raw data fields to app item fields
	err = scanOrderRow(cols, rows, orderFound, workspace.InstalledApps)
	if err != nil {
		return nil, eris.Wrap(err, "FindOrderByID")
	}

	return orderFound, nil
}

func (repo *RepositoryImpl) InsertOrder(ctx context.Context, order *entity.Order, tx *sql.Tx) (err error) {

	if order == nil {
		err = eris.New("order is missing")
		return
	}
	if tx == nil {
		err = eris.New("tx is missing")
		return
	}

	now := time.Now()

	columns := []string{
		"id",
		"external_id",
		"user_id",
		"domain_id",
		"session_id",
		"created_at",
		// "created_at_trunc", // computed field
		// "db_created_at", // computed field
		// "db_updated_at", // computed field
		// "merged_from_user_id",
		"fields_timestamp",
		"discount_codes",
		"subtotal_price",
		"total_price",
		"currency",
		"fx_rate",
		"cancelled_at",
		"cancel_reason",
		// "items",
		// attribution fields are only set by attribution algorithm
		// "is_first_conversion",
		// "time_to_conversion",
		// "devices_funnel",
		// "devices_type_count",
		// "domains_funnel",
		// "domains_type_funnel",
		// "domains_count",
		"funnel", // default JSON array is required
		// "funnel_hash",
		// "attribution_updated_at",
	}

	values := []interface{}{
		order.ID,
		order.ExternalID,
		order.UserID,
		order.DomainID,
		order.SessionID,
		order.CreatedAt,
		order.FieldsTimestamp,
		order.DiscountCodes,
		order.SubtotalPrice,
		order.TotalPrice,
		order.Currency,
		order.FxRate,
		order.CancelledAt,
		order.CancelReason,
		// order.Items,
		// attribution fields
		// order.IsFirstConversion,
		// order.TimeToConversion,
		// order.DevicesFunnel,
		// order.DevicesTypeCount,
		// order.DomainsFunnel,
		// order.DomainsTypeFunnel,
		// order.DomainsCount,
		order.Funnel,
		// order.FunnelHash,
		// order.AttributionUpdatedAt,
	}

	// add extra columns to the query
	// WARNING: values are interfaces, it might not guess well field types...
	// might have to convert them to NullableTypes before adding them
	if order.ExtraColumns != nil && len(order.ExtraColumns) > 0 {
		for field, value := range order.ExtraColumns {
			columns = append(columns, field)
			values = append(values, value)
		}
	}

	q := sq.Insert("`order`").Columns(columns...).Values(values...)

	sql, args, errSQL := q.ToSql()

	if errSQL != nil {
		err = eris.Wrapf(errSQL, "build query insert order: %v\n", order)
		return
	}

	_, errExec := tx.ExecContext(ctx, sql, args...)

	if errExec != nil {
		if repo.IsDuplicateEntry(errExec) {
			return eris.Wrap(ErrRowAlreadyExists, "InsertOrder")
		}

		err = eris.Wrap(errExec, "InsertOrder")
		log.Printf("InsertOrder: %v\n", err)
		return
	}

	order.DBCreatedAt = now
	order.DBUpdatedAt = now

	return
}

func (repo *RepositoryImpl) UpdateOrder(ctx context.Context, order *entity.Order, tx *sql.Tx) (err error) {

	if order == nil {
		err = eris.New("order is missing")
		return
	}
	if tx == nil {
		err = eris.New("tx is missing")
		return
	}

	now := time.Now()

	// specify sharding key (user_id) to avoid deadlocks

	q := sq.Update("`order`").Where(sq.Eq{"user_id": order.UserID}).Where(sq.Eq{"id": order.ID}).
		Set("created_at", order.CreatedAt).
		Set("fields_timestamp", order.FieldsTimestamp).
		Set("session_id", order.SessionID).
		Set("discount_codes", order.DiscountCodes).
		Set("subtotal_price", order.SubtotalPrice).
		Set("total_price", order.TotalPrice).
		Set("currency", order.Currency).
		Set("fx_rate", order.FxRate).
		Set("cancelled_at", order.CancelledAt).
		Set("cancel_reason", order.CancelReason)
		// Set("items", order.Items)
		// attribution fields are only set by attribution algorithm
		// Set("is_first_conversion", order.IsFirstConversion).
		// Set("time_to_conversion", order.TimeToConversion).
		// Set("devices_funnel", order.DevicesFunnel).
		// Set("devices_type_count", order.DevicesTypeCount).
		// Set("domains_funnel", order.DomainsFunnel).
		// Set("domains_type_funnel", order.DomainsTypeFunnel).
		// Set("domains_count", order.DomainsCount).
		// Set("funnel", order.Funnel).
		// Set("funnel_hash", order.FunnelHash).
		// Set("attribution_updated_at", order.AttributionUpdatedAt)
		// "extra_columns", // loop over dimensions to add app_custom_fields

	// add extra columns to the query
	// WARNING: values are interfaces, it might not guess well field types...
	// might have to convert them to NullableTypes before adding them
	for field, value := range order.ExtraColumns {
		q = q.Set(field, value)
	}

	sql, args, errSQL := q.ToSql()

	if errSQL != nil {
		err = eris.Wrapf(errSQL, "build query update order: %v\n", order)
		return
	}

	_, errExec := tx.ExecContext(ctx, sql, args...)

	if errExec != nil {
		// if repo.IsDuplicateEntry(errExec) {
		// }

		err = eris.Wrap(errExec, "UpdateOrder")
		return
	}

	order.DBUpdatedAt = now

	return
}

// clones orders from a user to another user with
// because the shard key "user_id" is immutable, we can't use an UPDATE
// we have to INSERT FROM SELECT + DELETE
func (repo *RepositoryImpl) MergeUserOrders(ctx context.Context, workspace *entity.Workspace, fromUserID string, toUserID string, tx *sql.Tx) (err error) {

	// find eventual extra columns for the order table
	orderCustomColumns := workspace.FindExtraColumnsForItemKind(entity.ItemKindOrder)

	orderStruct := entity.Order{}
	columns := entity.GetNotComputedDBColumnsForObject(orderStruct, entity.OrderComputedFields, orderCustomColumns)

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

	query := fmt.Sprintf("INSERT IGNORE INTO `order` (%v) SELECT %v FROM `order` WHERE user_id = ?", strings.Join(columns, ", "), strings.Join(selectedColumns, ", "))

	// log.Println(query)

	if _, err := tx.ExecContext(ctx, query, fromUserID); err != nil {
		return eris.Wrap(err, "MergeUserOrders")
	}

	// BUG deleting these rows might create a deadlock in singlestore
	// delete "from user" orders
	// if _, err := tx.ExecContext(ctx, "DELETE FROM orders WHERE user_id = ?"+" OPTION (columnstore_table_lock_threshold = 5000)", fromUserID); err != nil {
	// 	return err
	// }

	return nil
}

func (repo *RepositoryImpl) ListOrdersForUser(ctx context.Context, workspace *entity.Workspace, userID string, orderBy string, tx *sql.Tx) (orders []*entity.Order, err error) {

	orders = []*entity.Order{}
	var rows *sql.Rows

	// query
	sql, args, errSQL := sq.Select("*").From("`order`").Where(sq.Eq{"user_id": userID}).OrderBy(orderBy).ToSql()

	if errSQL != nil {
		err = eris.Wrapf(errSQL, "build query list orders for user: %v\n", userID)
		return
	}

	if tx == nil {

		conn, err := repo.GetWorkspaceConnection(ctx, workspace.ID)

		if err != nil {
			return nil, eris.Wrap(err, "ListOrdersForUser")
		}

		defer conn.Close()

		rows, err = conn.QueryContext(ctx, sql, args...)
	} else {
		rows, err = tx.QueryContext(ctx, sql, args...)
	}

	if err != nil {
		return nil, eris.Wrap(err, "ListOrdersForUser")
	}

	defer rows.Close()

	for rows.Next() {

		if rows.Err() != nil {
			return nil, eris.Wrap(rows.Err(), "ListOrdersForUser")
		}

		cols, err := rows.Columns()

		if err != nil {
			return nil, eris.Wrap(err, "ListOrdersForUser")
		}

		order := &entity.Order{}

		if err = scanOrderRow(cols, rows, order, workspace.InstalledApps); err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	return
}

func (repo *RepositoryImpl) UpdateOrderAttribution(ctx context.Context, order *entity.Order, tx *sql.Tx) (err error) {

	q := sq.Update("`order`").Where(sq.Eq{"user_id": order.UserID}).Where(sq.Eq{"id": order.ID}).
		Set("is_first_conversion", order.IsFirstConversion).
		Set("time_to_conversion", order.TimeToConversion).
		Set("devices_funnel", order.DevicesFunnel).
		Set("devices_type_count", order.DevicesTypeCount).
		Set("domains_funnel", order.DomainsFunnel).
		Set("domains_type_funnel", order.DomainsTypeFunnel).
		Set("domains_count", order.DomainsCount).
		Set("funnel", order.Funnel).
		Set("funnel_hash", order.FunnelHash).
		Set("attribution_updated_at", order.AttributionUpdatedAt).
		Set("session_id", order.SessionID)

	sql, args, errSQL := q.ToSql()

	if errSQL != nil {
		err = eris.Wrapf(errSQL, "build query update order: %v\n", order)
		return
	}

	_, err = tx.ExecContext(ctx, sql, args...)

	if err != nil {
		err = eris.Wrap(err, "UpdateOrderAttribution")
	}

	return
}

func scanOrderRow(cols []string, row RowScanner, order *entity.Order, installedApps entity.InstalledApps) error {

	// scan values
	values := make([]interface{}, len(cols))
	// extraColumns := []*ExtraColumn{}

	if order.ExtraColumns == nil {
		order.ExtraColumns = entity.AppItemFields{}
	}

	for i, col := range cols {
		switch col {
		case "id":
			values[i] = &order.ID
		case "external_id":
			values[i] = &order.ExternalID
		case "user_id":
			values[i] = &order.UserID
		case "domain_id":
			values[i] = &order.DomainID
		case "session_id":
			values[i] = &order.SessionID
		case "created_at":
			values[i] = &order.CreatedAt
		case "created_at_trunc":
			values[i] = &order.CreatedAtTrunc
		case "db_created_at":
			values[i] = &order.DBCreatedAt
		case "db_updated_at":
			values[i] = &order.DBUpdatedAt
		case "is_deleted":
			values[i] = &order.IsDeleted
		case "merged_from_user_id":
			values[i] = &order.MergedFromUserID
		case "fields_timestamp":
			values[i] = &order.FieldsTimestamp
		case "discount_codes":
			values[i] = &order.DiscountCodes
		case "subtotal_price":
			values[i] = &order.SubtotalPrice
		case "total_price":
			values[i] = &order.TotalPrice
		case "currency":
			values[i] = &order.Currency
		case "fx_rate":
			values[i] = &order.FxRate
		case "converted_subtotal_price":
			values[i] = &order.ConvertedSubtotalPrice
		case "converted_total_price":
			values[i] = &order.ConvertedTotalPrice
		case "cancelled_at":
			values[i] = &order.CancelledAt
		case "cancel_reason":
			values[i] = &order.CancelReason
		// case "items":
		// 	values[i] = &order.Items
		case "is_first_conversion":
			values[i] = &order.IsFirstConversion
		case "time_to_conversion":
			values[i] = &order.TimeToConversion
		case "devices_funnel":
			values[i] = &order.DevicesFunnel
		case "devices_type_count":
			values[i] = &order.DevicesTypeCount
		case "domains_funnel":
			values[i] = &order.DomainsFunnel
		case "domains_type_funnel":
			values[i] = &order.DomainsTypeFunnel
		case "domains_count":
			values[i] = &order.DomainsCount
		case "funnel":
			values[i] = &order.Funnel
		case "funnel_hash":
			values[i] = &order.FunnelHash
		case "attribution_updated_at":
			values[i] = &order.AttributionUpdatedAt
		default:
			// handle app extra columns
			if !strings.HasPrefix(col, "app_") && !strings.HasPrefix(col, "appx_") {
				return eris.Errorf("order column not mapped %v", col)
			}

			// find app item field in installedApps
			isValueMapped := false

			for _, app := range installedApps {
				for _, extraColumn := range app.ExtraColumns {
					if extraColumn.Kind == entity.ItemKindOrder {

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
								order.ExtraColumns[appItemField.Name] = &appItemField
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
			return eris.Wrap(err, "scanOrderRow")
		}
	}

	return nil
}
