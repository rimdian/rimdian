package repository

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (repo *RepositoryImpl) FetchUsers(ctx context.Context, workspace *entity.Workspace, query sq.SelectBuilder, tx *sql.Tx) (users []*entity.User, err error) {

	users = []*entity.User{}
	var rows *sql.Rows

	// query
	sql, args, errSQL := query.ToSql()

	if errSQL != nil {
		err = eris.Wrap(errSQL, "FetchUsers")
		return
	}

	if tx == nil {

		conn, err := repo.GetWorkspaceConnection(ctx, workspace.ID)

		if err != nil {
			return nil, eris.Wrap(err, "FetchUsers")
		}

		defer conn.Close()

		if rows, err = conn.QueryContext(ctx, sql, args...); err != nil {
			return nil, eris.Wrap(err, "FetchUsers")
		}
	} else {
		if rows, err = tx.QueryContext(ctx, sql, args...); err != nil {
			return nil, eris.Wrap(err, "FetchUsers")
		}
	}

	defer rows.Close()

	for rows.Next() {

		if rows.Err() != nil {
			return nil, eris.Wrap(rows.Err(), "FetchUsers")
		}

		cols, err := rows.Columns()

		if err != nil {
			return nil, eris.Wrap(err, "FetchUsers")
		}

		user := &entity.User{}

		if err = scanUserRow(cols, rows, user, workspace.InstalledApps); err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return
}

func (repo *RepositoryImpl) ListUsers(ctx context.Context, workspace *entity.Workspace, params *dto.UserListParams) (users []*entity.User, nextToken string, previousToken string, err error) {

	users = []*entity.User{}

	conn, err := repo.GetWorkspaceConnection(ctx, workspace.ID)

	if err != nil {
		return
	}

	defer conn.Close()

	// fetch an additional row to check if it has more for pagination
	limit := params.Limit + 1

	queryBuilder := sq.Select("u.*").From("`user` as u").Where(sq.Eq{"u.is_merged": false}).Limit(uint64(limit))

	// pagination if doesnt filter for a unique row
	if params.UserIDs == nil {

		// before = created_at is an older date = persisted before in the DB
		// after = created_at is a more recent date = persisted after in the DB

		if !params.NextDate.IsZero() {
			// want rows older than x, use DESC sort
			// filter with less-than-equal created_at_trunc to leverage columnstore SORT KEY
			// use less-than-equal to keep eventual rows that had same microseconds
			// refilter with more microsecond precision with created_at
			// refilter with external_id in case created_at is the same
			// exclude id = params.NextID to only select rows before it
			queryBuilder = queryBuilder.
				Where(sq.LtOrEq{"u.created_at_trunc": params.NextDate.Truncate(1 * time.Hour)}).
				Where(sq.LtOrEq{"u.created_at": params.NextDate.Format(dto.MicrosecondLayout)}). // force microsecond format here as driver strips them off
				Where(sq.NotEq{"u.id": params.NextID}).
				OrderBy("u.created_at_trunc DESC, u.created_at DESC, u.external_id DESC")
		} else if !params.PreviousDate.IsZero() {
			// want rows newer than x, use ASC sort
			// filter with greater-than-equal created_at_trunc to leverage columnstore SORT KEY
			// use greater-than-equal to keep eventual rows that had same microseconds
			// refilter with more microsecond precision with created_at
			// refilter with external_id in case created_at is the same
			// exclude id = params.PreviousID to only select rows after it
			queryBuilder = queryBuilder.
				Where(sq.GtOrEq{"u.created_at_trunc": params.PreviousDate.Truncate(1 * time.Hour)}).
				Where(sq.GtOrEq{"u.created_at": params.PreviousDate.Format(dto.MicrosecondLayout)}). // force microsecond format here as driver strips them off
				Where(sq.NotEq{"u.id": params.PreviousID}).
				OrderBy("u.created_at_trunc ASC, u.created_at ASC, u.external_id ASC")
		} else {
			// default sort
			queryBuilder = queryBuilder.OrderBy("u.created_at_trunc DESC, u.created_at DESC, u.external_id ASC")
		}
	} else {
		// split user_ids into array
		userIDs := strings.Split(*params.UserIDs, ",")
		// look for a specific rows, dont paginate
		queryBuilder = queryBuilder.Where(sq.Eq{"u.id": userIDs}).OrderBy("u.created_at_trunc DESC, u.created_at DESC, u.external_id ASC")
	}

	// filters

	// check if param.IsAuthenticated is set and filter accordingly
	if params.SegmentID != nil && *params.SegmentID != "" {
		// if segmentID is anonymous or authenticated
		if *params.SegmentID == "anonymous" {
			queryBuilder = queryBuilder.Where(sq.Eq{"u.is_authenticated": false})
		} else if *params.SegmentID == "authenticated" {
			queryBuilder = queryBuilder.Where(sq.Eq{"u.is_authenticated": true})
		} else {
			// if segmentID is a real segment
			queryBuilder = queryBuilder.InnerJoin("user_segment ON u.id = user_segment.user_id").Where(sq.Eq{"user_segment.segment_id": *params.SegmentID})
		}
	}

	if params.ListID != nil && *params.ListID != "" {
		queryBuilder = queryBuilder.InnerJoin("subscribe_to_list ON u.id = subscribe_to_list.user_id").Where(sq.Eq{"subscribe_to_list.subscription_list_id": *params.ListID})
	}

	users, err = repo.FetchUsers(ctx, workspace, queryBuilder, nil)

	if err != nil {
		return
	}

	hasMore := false

	if len(users) == limit {
		hasMore = true
	}

	// if we are going backwards in the list
	if !params.PreviousDate.IsZero() {

		// if we have more
		if hasMore {
			// remove last item to return the limit
			users = users[:len(users)-1]

		} else if len(users) < limit-1 {
			// if have less rows than wanted, we reached the begning of the list
			// we should return the list without pagination params then
			params.NextDate = time.Time{}
			params.PreviousDate = time.Time{}
			return repo.ListUsers(ctx, workspace, params)
		}

		// when querying the most recent items we sort by date ASC
		// but the results are always returned sorted by date DESC
		// so we need to reserve the array order

		reversed := []*entity.User{}
		for i := len(users) - 1; i >= 0; i-- {
			reversed = append(reversed, users[i])
		}

		users = reversed

		// going backwards always has nextToken
		lastRow := users[len(users)-1]
		nextToken = dto.EncodePaginationToken(lastRow.ID, lastRow.CreatedAt)

		// if we had more rows going backwards, we have a previous token
		if hasMore {
			firstRow := users[0]
			previousToken = dto.EncodePaginationToken(firstRow.ID, firstRow.CreatedAt)
		}

	} else if hasMore {
		// if default sort order (=date DESC) has more rows
		// remove last item to return the max limit
		users = users[:len(users)-1]

		lastRow := users[params.Limit-1]
		nextToken = dto.EncodePaginationToken(lastRow.ID, lastRow.CreatedAt)

		// was looking for rows before
		if params.NextID != "" {
			// was looking for rows after
			firstRow := users[0]
			previousToken = dto.EncodePaginationToken(firstRow.ID, firstRow.CreatedAt)
		}
	}

	return
}

// receives a user to update
func (repo *RepositoryImpl) UpdateUser(ctx context.Context, user *entity.User, tx *sql.Tx) (err error) {

	if user == nil {
		err = eris.New("user is missing")
		return
	}
	if tx == nil {
		err = eris.New("tx is missing")
		return
	}

	now := time.Now()

	q := sq.Update("user").Where(sq.Eq{"id": user.ID}).Limit(1).
		// user merge fields are controlled by user_alias
		// Set("is_merged", user.IsMerged).
		// Set("merged_to", user.MergedTo).
		// Set("merged_at", user.MergedAt).
		Set("signed_up_at", user.SignedUpAt).
		Set("created_at", user.CreatedAt).
		Set("last_interaction_at", user.LastInteractionAt).
		Set("timezone", user.Timezone).
		Set("language", user.Language).
		Set("country", user.Country).
		// "db_created_at", // auto set by DB
		// "db_updated_at", // auto set by DB
		Set("fields_timestamp", user.FieldsTimestamp).
		// optional fields:
		Set("consent_all", user.ConsentAll).
		Set("consent_personalization", user.ConsentPersonalization).
		Set("consent_marketing", user.ConsentMarketing).
		Set("last_ip", user.LastIP).
		Set("longitude", user.Longitude).
		Set("latitude", user.Latitude).
		Set("first_name", user.FirstName).
		Set("last_name", user.LastName).
		Set("gender", user.Gender).
		Set("birthday", user.Birthday).
		Set("photo_url", user.PhotoURL).
		Set("email", user.Email).
		Set("email_md5", user.EmailMD5).
		Set("email_sha1", user.EmailSHA1).
		Set("email_sha256", user.EmailSHA256).
		Set("telephone", user.Telephone).
		Set("address_line_1", user.AddressLine1).
		Set("address_line_2", user.AddressLine2).
		Set("city", user.City).
		Set("region", user.Region).
		Set("postal_code", user.PostalCode).
		Set("state", user.State).
		// order KPIs
		Set("orders_count", user.OrdersCount).
		Set("orders_ltv", user.OrdersLTV).
		Set("orders_avg_cart", user.OrdersAvgCart).
		Set("first_order_at", user.FirstOrderAt).
		Set("first_order_subtotal", user.FirstOrderSubtotal).
		Set("first_order_ttc", user.FirstOrderTTC).
		Set("first_order_domain_id", user.FirstOrderDomainID).
		Set("first_order_domain_type", user.FirstOrderDomainType).
		Set("last_order_at", user.LastOrderAt).
		Set("avg_repeat_cart", user.AvgRepeatCart).
		Set("avg_repeat_order_ttc", user.AvgRepeatOrderTTC)

	// add extra columns to the query
	// WARNING: values are interfaces, it might not guess well field types...
	// might have to convert them to NullableTypes before adding them
	for field, value := range user.ExtraColumns {
		q = q.Set(field, value)
	}

	sql, args, errSQL := q.ToSql()

	if errSQL != nil {
		err = eris.Wrapf(errSQL, "build query update user: %v\n", user)
		return
	}

	result, errExec := tx.ExecContext(ctx, sql, args...)

	if errExec != nil {
		err = errExec
		return
	}

	user.DBUpdatedAt = now

	// check result
	rowsAffected, err := result.RowsAffected()

	if err != nil {
		err = eris.Wrapf(err, "get rows affected update user: %v\n", user)
		return
	}

	if rowsAffected == 0 {
		err = ErrRowNotUpdated
		return
	}

	return
}

// receives a user to insert
func (repo *RepositoryImpl) InsertUser(ctx context.Context, user *entity.User, tx *sql.Tx) (err error) {

	if user == nil {
		err = eris.New("user is missing")
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
		// user merge fields are controlled by user_alias
		// "is_merged",
		// "merged_to",
		// "merged_at",
		"is_authenticated",
		"signed_up_at",
		"created_at",
		"last_interaction_at",
		"timezone",
		"language",
		"country",
		// "db_created_at", // auto set by DB
		// "db_updated_at", // auto set by DB
		"fields_timestamp",
		// optional fields:
		"consent_all",
		"consent_personalization",
		"consent_marketing",
		"last_ip",
		// "longitude", // set with latitude or nothing
		// "latitude", // set with longitude or nothing
		"first_name",
		"last_name",
		"gender",
		"birthday",
		"photo_url",
		"email",
		"email_md5",
		"email_sha1",
		"email_sha256",
		"telephone",
		"address_line_1",
		"address_line_2",
		"city",
		"region",
		"postal_code",
		"state",
		// "cart",
		// "cart_abandoned",
		// "wishlist",
	}

	values := []interface{}{
		user.ID,
		user.ExternalID,
		// user.IsMerged,
		// user.MergedTo,
		// user.MergedAt,
		user.IsAuthenticated,
		user.SignedUpAt,
		user.CreatedAt,
		user.LastInteractionAt,
		user.Timezone,
		user.Language,
		user.Country,
		user.FieldsTimestamp,
		// optional fields:
		user.ConsentAll,
		user.ConsentPersonalization,
		user.ConsentMarketing,
		user.LastIP,
		user.FirstName,
		user.LastName,
		user.Gender,
		user.Birthday,
		user.PhotoURL,
		user.Email,
		user.EmailMD5,
		user.EmailSHA1,
		user.EmailSHA256,
		user.Telephone,
		user.AddressLine1,
		user.AddressLine2,
		user.City,
		user.Region,
		user.PostalCode,
		user.State,
		// user.Cart,
		// user.CartAbandoned,
		// user.WishList,
	}

	// set geo if has all coords
	if user.Latitude != nil && !user.Latitude.IsNull && user.Longitude != nil && !user.Longitude.IsNull {
		columns = append(columns, "latitude", "longitude")
		values = append(values, user.Latitude, user.Longitude)
	}

	// add extra columns to the query
	// WARNING: values are interfaces, it might not guess well field types...
	// might have to convert them to NullableTypes before adding them
	if user.ExtraColumns != nil && len(user.ExtraColumns) > 0 {
		for field, value := range user.ExtraColumns {
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
	q := sq.Insert("user").Columns(columns...).Values(values...)

	sql, args, errSQL := q.ToSql()

	if errSQL != nil {
		err = eris.Wrapf(errSQL, "build query insert user: %v\n", user)
		return
	}

	_, errExec := tx.ExecContext(ctx, sql, args...)

	if errExec != nil {
		if repo.IsDuplicateEntry(errExec) {
			log.Printf("InsertUser dup: %v\n", sql)
			return eris.Wrap(ErrRowAlreadyExists, "InsertUser")
		}

		err = eris.Wrap(errExec, "InsertUser")
		return
	}

	user.DBCreatedAt = now
	user.DBUpdatedAt = now

	return
}

// takes a list of reconciliation keys to compare with existing users
// in order to eventually merge users
func (repo *RepositoryImpl) FindEventualUsersToMergeWith(ctx context.Context, workspace *entity.Workspace, withUser *entity.User, withReconciliationKeys entity.MapOfInterfaces, tx *sql.Tx) (usersFound []*entity.User, err error) {

	// kill if no withReconciliationKeys are provided, we can't match a random guy
	if len(withReconciliationKeys) == 0 {
		return nil, nil
	}

	// retrieve 1 user max, that is not the user we compare
	queryBuilder := sq.Select("*").From("user").Where(sq.NotEq{"id": withUser.ID}).OrderBy("is_authenticated DESC").Limit(1)

	// only match with anonymous users if destination user is authenticated
	if withUser.IsAuthenticated {
		queryBuilder = queryBuilder.Where(sq.Eq{"is_authenticated": false})
	}

	orConditions := sq.Or{}

	for column, value := range withReconciliationKeys {
		orConditions = append(orConditions, sq.Eq{column: value})
	}

	queryBuilder = queryBuilder.Where(orConditions)

	return repo.FetchUsers(ctx, workspace, queryBuilder, tx)
}

func (repo *RepositoryImpl) FindUserByID(ctx context.Context, workspace *entity.Workspace, userID string, tx *sql.Tx) (userFound *entity.User, err error) {

	var rows *sql.Rows
	userFound = &entity.User{}
	userFound.ExtraColumns = entity.AppItemFields{}

	if tx == nil {

		var conn *sql.Conn

		conn, err = repo.GetWorkspaceConnection(ctx, workspace.ID)

		if err != nil {
			return nil, err
		}

		defer conn.Close()

		rows, err = conn.QueryContext(ctx, "SELECT * FROM `user` WHERE id = ? LIMIT 1", userID)
	} else {
		rows, err = tx.QueryContext(ctx, "SELECT * FROM `user` WHERE id = ? LIMIT 1", userID)
	}

	if err != nil {
		return nil, eris.Wrap(err, "FindUserByID")
	}

	defer rows.Close()

	// no rows found
	if !rows.Next() {
		return nil, nil
	}

	// extract columns names
	cols, err := rows.Columns()

	if err != nil {
		return nil, eris.Wrap(err, "FindUserByID")
	}

	// convert raw data fields to app item fields
	err = scanUserRow(cols, rows, userFound, workspace.InstalledApps)
	if err != nil {
		return nil, eris.Wrap(err, "FindUserByID")
	}

	return userFound, nil
}

func scanUserRow(cols []string, row RowScanner, user *entity.User, installedApps entity.InstalledApps) error {

	// scan values
	values := make([]interface{}, len(cols))
	// extraColumns := []*ExtraColumn{}

	if user.ExtraColumns == nil {
		user.ExtraColumns = entity.AppItemFields{}
	}

	for i, col := range cols {
		switch col {
		case "id":
			values[i] = &user.ID
		case "external_id":
			values[i] = &user.ExternalID
		case "is_merged":
			values[i] = &user.IsMerged
		case "merged_to":
			values[i] = &user.MergedTo
		case "merged_at":
			values[i] = &user.MergedAt
		case "is_authenticated":
			values[i] = &user.IsAuthenticated
		case "signed_up_at":
			values[i] = &user.SignedUpAt
		case "created_at":
			values[i] = &user.CreatedAt
		case "created_at_trunc":
			values[i] = &user.CreatedAtTrunc
		case "last_interaction_at":
			values[i] = &user.LastInteractionAt
		case "timezone":
			values[i] = &user.Timezone
		case "language":
			values[i] = &user.Language
		case "country":
			values[i] = &user.Country
		case "db_created_at":
			values[i] = &user.DBCreatedAt
		case "db_updated_at":
			values[i] = &user.DBUpdatedAt
		case "fields_timestamp":
			values[i] = &user.FieldsTimestamp
		case "consent_all":
			values[i] = &user.ConsentAll
		case "consent_personalization":
			values[i] = &user.ConsentPersonalization
		case "consent_marketing":
			values[i] = &user.ConsentMarketing
		case "last_ip":
			values[i] = &user.LastIP
		case "latitude":
			values[i] = &user.Latitude
		case "longitude":
			values[i] = &user.Longitude
		case "geo":
			values[i] = &user.Geo
		case "first_name":
			values[i] = &user.FirstName
		case "last_name":
			values[i] = &user.LastName
		case "gender":
			values[i] = &user.Gender
		case "birthday":
			values[i] = &user.Birthday
		case "photo_url":
			values[i] = &user.PhotoURL
		case "email":
			values[i] = &user.Email
		case "email_md5":
			values[i] = &user.EmailMD5
		case "email_sha1":
			values[i] = &user.EmailSHA1
		case "email_sha256":
			values[i] = &user.EmailSHA256
		case "telephone":
			values[i] = &user.Telephone
		case "address_line_1":
			values[i] = &user.AddressLine1
		case "address_line_2":
			values[i] = &user.AddressLine2
		case "city":
			values[i] = &user.City
		case "region":
			values[i] = &user.Region
		case "postal_code":
			values[i] = &user.PostalCode
		case "state":
			values[i] = &user.State
		case "orders_count":
			values[i] = &user.OrdersCount
		case "orders_ltv":
			values[i] = &user.OrdersLTV
		case "orders_avg_cart":
			values[i] = &user.OrdersAvgCart
		case "first_order_at":
			values[i] = &user.FirstOrderAt
		case "first_order_domain_id":
			values[i] = &user.FirstOrderDomainID
		case "first_order_domain_type":
			values[i] = &user.FirstOrderDomainType
		case "first_order_subtotal":
			values[i] = &user.FirstOrderSubtotal
		case "first_order_ttc":
			values[i] = &user.FirstOrderTTC
		case "last_order_at":
			values[i] = &user.LastOrderAt
		case "avg_repeat_cart":
			values[i] = &user.AvgRepeatCart
		case "avg_repeat_order_ttc":
			values[i] = &user.AvgRepeatOrderTTC

		default:
			// handle app extra columns
			if !strings.HasPrefix(col, "app_") && !strings.HasPrefix(col, "appx_") {
				return eris.Errorf("user column not mapped %v", col)
			}

			// find app item field in installedApps
			isValueMapped := false

			for _, app := range installedApps {
				for _, extraColumn := range app.ExtraColumns {
					if extraColumn.Kind == entity.ItemKindUser {

						for _, colDefinition := range extraColumn.Columns {
							if colDefinition.Name == col {
								isValueMapped = true

								// init an extra column in the user
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

								// add the extra column to the user
								user.ExtraColumns[appItemField.Name] = &appItemField
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
			return eris.Wrap(err, "scanUserRow")
		}
	}

	return nil
}
