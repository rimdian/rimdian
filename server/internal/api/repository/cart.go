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

func (repo *RepositoryImpl) FindCartByID(ctx context.Context, workspaceID string, cartID string, userID string, tx *sql.Tx) (cartFound *entity.Cart, err error) {

	cartFound = &entity.Cart{}

	if tx == nil {

		var conn *sql.Conn

		conn, err = repo.GetWorkspaceConnection(ctx, workspaceID)

		if err != nil {
			return nil, err
		}

		defer conn.Close()

		err = sqlscan.Get(ctx, conn, cartFound, "SELECT * FROM cart WHERE user_id = ? AND id = ? LIMIT 1", userID, cartID)
	} else {
		err = sqlscan.Get(ctx, tx, cartFound, "SELECT * FROM cart WHERE user_id = ? AND id = ? LIMIT 1", userID, cartID)
	}

	if err != nil {
		if sqlscan.NotFound(err) {
			return nil, nil
		} else {
			return nil, eris.Wrapf(err, "FindCartByID %v", cartID)
		}
	}

	return cartFound, nil
}

func (repo *RepositoryImpl) InsertCart(ctx context.Context, cart *entity.Cart, tx *sql.Tx) (err error) {

	if cart == nil {
		err = eris.New("cart is missing")
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
		"session_id",
		"created_at",
		// "created_at_trunc", // computed field
		// "db_created_at", // computed field
		// "db_updated_at", // computed field
		// "merged_from_user_id",
		"fields_timestamp",
		"currency",
		"fx_rate",
		"public_url",
		// "items",
		"status",
		// "extra_columns", // loop over dimensions to add app_custom_fields
	}

	values := []interface{}{
		cart.ID,
		cart.ExternalID,
		cart.UserID,
		cart.DomainID,
		cart.SessionID,
		cart.CreatedAt,
		cart.FieldsTimestamp,
		cart.Currency,
		cart.FxRate,
		cart.PublicURL,
		// cart.Items,
		cart.Status,
	}

	// add extra columns to the query
	// WARNING: values are interfaces, it might not guess well field types...
	// might have to convert them to NullableTypes before adding them
	if cart.ExtraColumns != nil && len(cart.ExtraColumns) > 0 {
		for field, value := range cart.ExtraColumns {
			columns = append(columns, field)
			values = append(values, value)
		}
	}

	q := sq.Insert("cart").Columns(columns...).Values(values...)

	sql, args, errSQL := q.ToSql()

	if errSQL != nil {
		err = eris.Wrapf(errSQL, "build query insert cart: %v\n", cart)
		return
	}

	_, errExec := tx.ExecContext(ctx, sql, args...)

	if errExec != nil {
		if repo.IsDuplicateEntry(errExec) {
			return eris.Wrap(ErrRowAlreadyExists, "InsertCart")
		}

		err = eris.Wrap(errExec, "InsertCart")
		return
	}

	cart.DBCreatedAt = now
	cart.DBUpdatedAt = now

	return
}

func (repo *RepositoryImpl) UpdateCart(ctx context.Context, cart *entity.Cart, tx *sql.Tx) (err error) {

	if cart == nil {
		err = eris.New("cart is missing")
		return
	}
	if tx == nil {
		err = eris.New("tx is missing")
		return
	}

	now := time.Now()

	// UPDATE
	// specify sharding key to avoid deadlocks

	q := sq.Update("cart").Where(sq.Eq{"user_id": cart.UserID}).Where(sq.Eq{"id": cart.ID}).
		Set("created_at", cart.CreatedAt).
		Set("fields_timestamp", cart.FieldsTimestamp).
		Set("currency", cart.Currency).
		Set("fx_rate", cart.FxRate).
		Set("public_url", cart.PublicURL)
		// Set("items", cart.Items)
		// "extra_columns", // loop over dimensions to add app_custom_fields

	// add extra columns to the query
	// WARNING: values are interfaces, it might not guess well field types...
	// might have to convert them to NullableTypes before adding them
	for field, value := range cart.ExtraColumns {
		q = q.Set(field, value)
	}
	sql, args, errSQL := q.ToSql()

	if errSQL != nil {
		err = eris.Wrapf(errSQL, "build query update cart: %v\n", cart)
		return
	}

	if repo.Config.DB_TYPE == "singlestore" {
		sql = sql + " OPTION (columnstore_table_lock_threshold = 5000)"
	}

	_, errExec := tx.ExecContext(ctx, sql, args...)

	if errExec != nil {
		// if repo.IsDuplicateEntry(errExec) {
		// }

		err = eris.Wrap(errExec, "UpdateCart")
		return
	}

	cart.DBUpdatedAt = now

	return
}

// clones cart from a user to another user with
// because the shard key "user_id" is immutable, we can't use an UPDATE
// we have to INSERT FROM SELECT + DELETE
func (repo *RepositoryImpl) MergeUserCarts(ctx context.Context, workspace *entity.Workspace, fromUserID string, toUserID string, tx *sql.Tx) (err error) {

	// find eventual extra columns for the cart table
	cartCustomColumns := workspace.FindExtraColumnsForItemKind(entity.ItemKindCart)

	cartStruct := entity.Cart{}
	columns := entity.GetNotComputedDBColumnsForObject(cartStruct, entity.CartComputedFields, cartCustomColumns)

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
		INSERT IGNORE INTO cart (%v) 
		SELECT %v FROM cart 
		WHERE user_id = ?
	`, strings.Join(columns, ", "), strings.Join(selectedColumns, ", "))

	// log.Println(query)

	if _, err := tx.ExecContext(ctx, query, fromUserID); err != nil {
		return eris.Wrap(err, "MergeUserCarts")
	}

	// BUG deleting these rows might create a deadlock in singlestore
	// delete "from user" cart
	// if _, err := tx.ExecContext(ctx, "DELETE FROM cart WHERE user_id = ? "+" OPTION (columnstore_table_lock_threshold = 5000)", fromUserID); err != nil {
	// 	return err
	// }

	return nil
}
