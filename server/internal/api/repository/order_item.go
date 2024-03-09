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

func (repo *RepositoryImpl) DeleteOrderItem(ctx context.Context, orderItemID string, userID string, tx *sql.Tx) (err error) {

	if tx == nil {
		err = eris.New("tx is missing")
		return
	}

	var result sql.Result
	result, err = tx.ExecContext(ctx, "UPDATE order_item SET is_deleted = 1 WHERE id = ? AND user_id = ?", orderItemID, userID)

	if err != nil {
		return eris.Wrapf(err, "DeleteOrderItem %v", orderItemID)
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return eris.Wrapf(err, "DeleteOrderItem %v", orderItemID)
	}

	if rowsAffected == 0 {
		// return eris.Errorf("DeleteOrderItem %v not found", orderItemID)
	}

	return
}

func (repo *RepositoryImpl) FindOrderItemsByOrderID(ctx context.Context, workspaceID string, orderID string, userID string, tx *sql.Tx) (orderItems []*entity.OrderItem, err error) {

	orderItems = []*entity.OrderItem{}

	if tx == nil {

		var conn *sql.Conn

		conn, err = repo.GetWorkspaceConnection(ctx, workspaceID)

		if err != nil {
			return nil, err
		}

		defer conn.Close()

		err = sqlscan.Select(ctx, conn, &orderItems, "SELECT * FROM order_item WHERE order_id = ? AND user_id = ?", orderID, userID)
	} else {
		err = sqlscan.Select(ctx, tx, &orderItems, "SELECT * FROM order_item WHERE order_id = ? AND user_id = ?", orderID, userID)
	}

	if err != nil {
		if sqlscan.NotFound(err) {
			return nil, nil
		} else {
			return nil, eris.Wrapf(err, "FindOrderItemsByOrderID %v", orderID)
		}
	}

	return
}

func (repo *RepositoryImpl) InsertOrderItem(ctx context.Context, item *entity.OrderItem, tx *sql.Tx) (err error) {

	if item == nil {
		err = eris.New("item is missing")
		return
	}
	if tx == nil {
		err = eris.New("tx is missing")
		return
	}

	now := time.Now()

	q := sq.Insert("order_item").Columns(
		"id",
		"external_id",
		"order_id",
		"user_id",
		"product_external_id",
		"name",
		"sku",
		"brand",
		"category",
		"variant_external_id",
		"variant_title",
		"image_url",
		"price",
		"currency",
		"fx_rate",
		"quantity",
		"created_at",
		// "created_at_trunc",
		// "db_created_at",
		// "db_updated_at",
		// "is_deleted",
		"merged_from_user_id",
		"fields_timestamp",
	).Values(
		item.ID,
		item.ExternalID,
		item.OrderID,
		item.UserID,
		item.ProductExternalID,
		item.Name,
		item.SKU,
		item.Brand,
		item.Category,
		item.VariantExternalID,
		item.VariantTitle,
		item.ImageURL,
		item.Price,
		item.Currency,
		item.FxRate,
		item.Quantity,
		item.CreatedAt,
		// item.CreatedAtTrunc,
		// item.DBCreatedAt,
		// item.DBUpdatedAt,
		// item.IsDeleted,
		item.MergedFromUserID,
		item.FieldsTimestamp,
	)

	sql, args, errSQL := q.ToSql()

	if errSQL != nil {
		err = eris.Wrapf(errSQL, "build query insert order item: %v\n", item)
		return
	}

	_, errExec := tx.ExecContext(ctx, sql, args...)

	if errExec != nil {
		if repo.IsDuplicateEntry(errExec) {
			return eris.Wrap(ErrRowAlreadyExists, "InsertOrderItem")
		}

		err = eris.Wrap(errExec, "InsertOrderItem")
		return
	}

	item.DBCreatedAt = now
	item.DBUpdatedAt = now

	return
}

func (repo *RepositoryImpl) UpdateOrderItem(ctx context.Context, item *entity.OrderItem, tx *sql.Tx) (err error) {

	if item == nil {
		err = eris.New("item is missing")
		return
	}
	if tx == nil {
		err = eris.New("tx is missing")
		return
	}

	now := time.Now()

	// UPDATE
	// specify sharding/primary keys to avoid deadlocks

	q := sq.Update("order_item").Where(sq.Eq{"user_id": item.UserID}).Where(sq.Eq{"id": item.ID}).
		Set("created_at", item.CreatedAt).
		Set("fields_timestamp", item.FieldsTimestamp).
		Set("order_id", item.OrderID).
		Set("name", item.Name).
		Set("sku", item.SKU).
		Set("brand", item.Brand).
		Set("category", item.Category).
		Set("variant_external_id", item.VariantExternalID).
		Set("variant_title", item.VariantTitle).
		Set("image_url", item.ImageURL).
		Set("price", item.Price).
		Set("quantity", item.Quantity).
		Set("currency", item.Currency).
		Set("fx_rate", item.FxRate)

		// "extra_columns", // loop over dimensions to add app_custom_fields

	// add extra columns to the query
	// WARNING: values are interfaces, it might not guess well field types...
	// might have to convert them to NullableTypes before adding them
	for field, value := range item.ExtraColumns {
		q = q.Set(field, value)
	}

	sql, args, errSQL := q.ToSql()

	if errSQL != nil {
		err = eris.Wrapf(errSQL, "build query update order item: %v\n", item)
		return
	}

	_, errExec := tx.ExecContext(ctx, sql, args...)

	if errExec != nil {
		// if repo.IsDuplicateEntry(errExec) {
		// }

		err = eris.Wrap(errExec, "UpdateOrderItem")
		return
	}

	item.DBUpdatedAt = now

	return
}

// clones orders from a user to another user with
// because the shard key "user_id" is immutable, we can't use an UPDATE
// we have to INSERT FROM SELECT + DELETE
func (repo *RepositoryImpl) MergeUserOrderItems(ctx context.Context, workspace *entity.Workspace, fromUserID string, toUserID string, tx *sql.Tx) (err error) {

	// find eventual extra columns for the order table
	itemCustomColumns := workspace.FindExtraColumnsForItemKind(entity.ItemKindOrderItem)

	itemStruct := entity.OrderItem{}
	columns := entity.GetNotComputedDBColumnsForObject(itemStruct, entity.OrderItemComputedFields, itemCustomColumns)

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
		INSERT IGNORE INTO order_item (%v) 
		SELECT %v FROM order_item 
		WHERE user_id = ?
	`, strings.Join(columns, ", "), strings.Join(selectedColumns, ", "))

	// log.Println(query)

	if _, err := tx.ExecContext(ctx, query, fromUserID); err != nil {
		return eris.Wrap(err, "MergeUserOrderItems")
	}

	return nil
}
