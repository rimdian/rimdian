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

func (repo *RepositoryImpl) FindPageviewByID(ctx context.Context, workspace *entity.Workspace, pageviewID string, userID string, tx *sql.Tx) (pageviewFound *entity.Pageview, err error) {

	var rows *sql.Rows
	pageviewFound = &entity.Pageview{}
	pageviewFound.ExtraColumns = entity.AppItemFields{}

	if tx == nil {

		var conn *sql.Conn

		conn, err = repo.GetWorkspaceConnection(ctx, workspace.ID)

		if err != nil {
			return nil, err
		}

		defer conn.Close()

		rows, err = conn.QueryContext(ctx, "SELECT * FROM pageview WHERE user_id = ? AND id = ? LIMIT 1", userID, pageviewID)
	} else {
		rows, err = tx.QueryContext(ctx, "SELECT * FROM pageview WHERE user_id = ? AND id = ? LIMIT 1", userID, pageviewID)
	}

	if err != nil {
		return nil, eris.Wrap(err, "FindPageviewByID")
	}

	defer rows.Close()

	// no rows found
	if !rows.Next() {
		return nil, nil
	}

	// extract columns names
	cols, err := rows.Columns()

	if err != nil {
		return nil, eris.Wrap(err, "FindPageviewByID")
	}

	// convert raw data fields to app item fields
	err = scanPageviewRow(cols, rows, pageviewFound, workspace.InstalledApps)
	if err != nil {
		return nil, eris.Wrap(err, "FindPageviewByID")
	}

	return pageviewFound, nil
}

func (repo *RepositoryImpl) InsertPageview(ctx context.Context, pageview *entity.Pageview, tx *sql.Tx) (err error) {

	if pageview == nil {
		err = eris.New("pageview is missing")
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
		"page_id",
		"title",
		"referrer",
		"referrer_domain",
		"referrer_path",
		"duration",
		"image_url",
		"product_external_id",
		"product_sku",
		"product_name",
		"product_brand",
		"product_category",
		"product_variant_external_id",
		"product_variant_title",
		"product_price",
		"product_currency",
		"product_fx_rate",
		// "product_conversion_rate_error",

		// "extra_columns", // loop over dimensions to add app_custom_fields
	}

	values := []interface{}{
		pageview.ID,
		pageview.ExternalID,
		pageview.UserID,
		pageview.DomainID,
		pageview.SessionID,
		pageview.CreatedAt,
		pageview.FieldsTimestamp,
		pageview.PageID,
		pageview.Title,
		pageview.Referrer,
		pageview.ReferrerDomain,
		pageview.ReferrerPath,
		pageview.Duration,
		pageview.ImageURL,
		pageview.ProductExternalID,
		pageview.ProductSKU,
		pageview.ProductName,
		pageview.ProductBrand,
		pageview.ProductCategory,
		pageview.ProductVariantExternalID,
		pageview.ProductVariantTitle,
		pageview.ProductPrice,
		pageview.ProductCurrency,
		pageview.ProductFxRate,
		// pageview.ProductConversionRateError,
	}

	// add extra columns to the query
	// WARNING: values are interfaces, it might not guess well field types...
	// might have to convert them to NullableTypes before adding them
	if pageview.ExtraColumns != nil && len(pageview.ExtraColumns) > 0 {
		for field, value := range pageview.ExtraColumns {
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

	q := sq.Insert("pageview").Columns(columns...).Values(values...)

	sql, args, errSQL := q.ToSql()

	if errSQL != nil {
		err = eris.Wrapf(errSQL, "build query insert pageview: %v\n", pageview)
		return
	}

	_, errExec := tx.ExecContext(ctx, sql, args...)

	if errExec != nil {
		if repo.IsDuplicateEntry(errExec) {
			return eris.Wrap(ErrRowAlreadyExists, "InsertPageview")
		}

		err = eris.Wrap(errExec, "InsertPageview")
		return
	}

	pageview.DBCreatedAt = now
	pageview.DBUpdatedAt = now

	return
}

func (repo *RepositoryImpl) UpdatePageview(ctx context.Context, pageview *entity.Pageview, tx *sql.Tx) (err error) {

	if pageview == nil {
		err = eris.New("pageview is missing")
		return
	}
	if tx == nil {
		err = eris.New("tx is missing")
		return
	}

	now := time.Now()

	// UPDATE
	// specify sharding key to avoid deadlocks
	q := sq.Update("pageview").Where(sq.Eq{"user_id": pageview.UserID}).Where(sq.Eq{"id": pageview.ID}).
		Set("created_at", pageview.CreatedAt).
		Set("fields_timestamp", pageview.FieldsTimestamp).
		Set("referrer", pageview.Referrer).
		Set("referrer_domain", pageview.ReferrerDomain).
		Set("referrer_path", pageview.ReferrerPath).
		Set("duration", pageview.Duration).
		Set("image_url", pageview.ImageURL).
		Set("product_external_id", pageview.ProductExternalID).
		Set("product_sku", pageview.ProductSKU).
		Set("product_name", pageview.ProductName).
		Set("product_brand", pageview.ProductBrand).
		Set("product_category", pageview.ProductCategory).
		Set("product_variant_external_id", pageview.ProductVariantExternalID).
		Set("product_variant_title", pageview.ProductVariantTitle).
		Set("product_price", pageview.ProductPrice).
		Set("product_currency", pageview.ProductCurrency).
		Set("product_fx_rate", pageview.ProductFxRate)
		// Set("product_conversion_rate_error", pageview.ProductConversionRateError)
		// "extra_columns", // loop over dimensions to add app_custom_fields

	// add extra columns to the query
	// WARNING: values are interfaces, it might not guess well field types...
	// might have to convert them to NullableTypes before adding them
	for field, value := range pageview.ExtraColumns {
		q = q.Set(field, value)
	}

	sql, args, errSQL := q.ToSql()

	if errSQL != nil {
		err = eris.Wrapf(errSQL, "build query update pageview: %v\n", pageview)
		return
	}

	if repo.Config.DB_TYPE == "singlestore" {
		sql = sql + " OPTION (columnstore_table_lock_threshold = 5000)"
	}

	_, errExec := tx.ExecContext(ctx, sql, args...)

	if errExec != nil {
		// if repo.IsDuplicateEntry(errExec) {
		// }

		err = eris.Wrap(errExec, "UpdatePageview")
		return
	}

	pageview.DBUpdatedAt = now

	return
}

// clones pageviews from a user to another user with
// because the shard key "user_id" is immutable, we can't use an UPDATE
// we have to INSERT FROM SELECT + DELETE
func (repo *RepositoryImpl) MergeUserPageviews(ctx context.Context, workspace *entity.Workspace, fromUserID string, toUserID string, tx *sql.Tx) (err error) {

	// find eventual extra columns for the pageview table
	pageviewCustomColumns := workspace.FindExtraColumnsForItemKind(entity.ItemKindPageview)

	pageviewStruct := entity.Pageview{}
	columns := entity.GetNotComputedDBColumnsForObject(pageviewStruct, entity.PageviewComputedFields, pageviewCustomColumns)

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
		INSERT IGNORE INTO pageview (%v) 
		SELECT %v FROM pageview 
		WHERE user_id = ?
	`, strings.Join(columns, ", "), strings.Join(selectedColumns, ", "))

	// log.Println(query)

	if _, err := tx.ExecContext(ctx, query, fromUserID); err != nil {
		return eris.Wrap(err, "MergeUserPageviews")
	}

	// BUG deleting these rows might create a deadlock in singlestore
	// delete "from user" pageviews
	// if _, err := tx.ExecContext(ctx, "DELETE FROM pageviews WHERE user_id = ? OPTION (columnstore_table_lock_threshold = 5000)", fromUserID); err != nil {
	// 	return err
	// }

	return nil
}

func scanPageviewRow(cols []string, row RowScanner, pageview *entity.Pageview, installedApps entity.InstalledApps) error {

	// scan values
	values := make([]interface{}, len(cols))
	// extraColumns := []*ExtraColumn{}

	if pageview.ExtraColumns == nil {
		pageview.ExtraColumns = entity.AppItemFields{}
	}

	for i, col := range cols {
		switch col {
		case "id":
			values[i] = &pageview.ID
		case "external_id":
			values[i] = &pageview.ExternalID
		case "user_id":
			values[i] = &pageview.UserID
		case "domain_id":
			values[i] = &pageview.DomainID
		case "session_id":
			values[i] = &pageview.SessionID
		case "created_at":
			values[i] = &pageview.CreatedAt
		case "created_at_trunc":
			values[i] = &pageview.CreatedAtTrunc
		case "db_created_at":
			values[i] = &pageview.DBCreatedAt
		case "db_updated_at":
			values[i] = &pageview.DBUpdatedAt
		case "is_deleted":
			values[i] = &pageview.IsDeleted
		case "merged_from_user_id":
			values[i] = &pageview.MergedFromUserID
		case "fields_timestamp":
			values[i] = &pageview.FieldsTimestamp
		case "page_id":
			values[i] = &pageview.PageID
		case "title":
			values[i] = &pageview.Title
		case "referrer":
			values[i] = &pageview.Referrer
		case "referrer_domain":
			values[i] = &pageview.ReferrerDomain
		case "referrer_path":
			values[i] = &pageview.ReferrerPath
		case "duration":
			values[i] = &pageview.Duration
		case "image_url":
			values[i] = &pageview.ImageURL
		case "product_external_id":
			values[i] = &pageview.ProductExternalID
		case "product_sku":
			values[i] = &pageview.ProductSKU
		case "product_name":
			values[i] = &pageview.ProductName
		case "product_brand":
			values[i] = &pageview.ProductBrand
		case "product_category":
			values[i] = &pageview.ProductCategory
		case "product_variant_external_id":
			values[i] = &pageview.ProductVariantExternalID
		case "product_variant_title":
			values[i] = &pageview.ProductVariantTitle
		case "product_price":
			values[i] = &pageview.ProductPrice
		case "product_currency":
			values[i] = &pageview.ProductCurrency
		case "product_fx_rate":
			values[i] = &pageview.ProductFxRate
		case "product_converted_price":
			values[i] = &pageview.ProductConvertedPrice

		default:
			// handle app extra columns
			if !strings.HasPrefix(col, "app_") && !strings.HasPrefix(col, "appx_") {
				return eris.Errorf("pageview column not mapped %v", col)
			}

			// find app item field in installedApps
			isValueMapped := false

			for _, app := range installedApps {
				for _, extraColumn := range app.ExtraColumns {
					if extraColumn.Kind == entity.ItemKindPageview {

						for _, colDefinition := range extraColumn.Columns {
							if colDefinition.Name == col {
								isValueMapped = true

								// init an extra column in the pageview
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

								// add the extra column to the pageview
								pageview.ExtraColumns[appItemField.Name] = appItemField
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
			return eris.Wrap(err, "scanPageviewRow")
		}
	}

	return nil
}
