package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/asaskevich/govalidator"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/common"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/internal/api/repository"
	"github.com/rimdian/rimdian/internal/common/dto"
	"github.com/rotisserie/eris"
	"go.opencensus.io/trace"
)

func (pipe *DataLogPipeline) UpsertUser(ctx context.Context, isChild bool, tx *sql.Tx) (err error) {

	spanCtx, span := trace.StartSpan(ctx, "UpsertUser")
	defer span.End()

	upsertedUser := pipe.DataLog.UpsertedUser

	// check if anonymous user has an alias
	if !upsertedUser.IsAuthenticated {
		var alias *entity.UserAlias
		alias, err = pipe.Repository.FindUserAlias(spanCtx, upsertedUser.ExternalID, tx)

		if err != nil {
			return eris.Wrap(err, "UserUpsert")
		}

		// mutate the user_id + is_authenticated if has an alias to point to the right user
		if alias != nil {
			upsertedUser.ID = entity.ComputeUserID(alias.ToUserExternalID)
			upsertedUser.ExternalID = alias.ToUserExternalID
			upsertedUser.IsAuthenticated = alias.ToUserIsAuthenticated
		}
	}

	// find eventual existing user
	var existingUser *entity.User

	existingUser, err = pipe.Repository.FindUserByID(spanCtx, pipe.Workspace, upsertedUser.ID, tx)

	if err != nil && !sqlscan.NotFound(err) {
		return eris.Wrap(err, "UserUpsert")
	}

	// compute email hashes if possible, before matching reconciliation keys
	// as an anonymous user can just have an email hash in its profile without the full email address
	upsertedUser.ComputeEmailHashes()

	// merge eventually matching users asynchronously if conditions are met
	// extract reconciliation keys and their values to match them

	reconciliationKeysFound := entity.MapOfInterfaces{}

	// the existing user might have a reconciliation key that will be updated by this upsert payload
	// in this case, we need to refind matching users for this updated reconciliation key
	hasDifferentKeysThanExisting := false

	for _, key := range pipe.Workspace.UserReconciliationKeys {

		// is the reconciliation key an extra column?
		// TODO
		// if strings.HasPrefix(key, "app_") || strings.HasPrefix(pipe.DataLog.Kind, "appx_") {

		// 	// abort if no extra columns are sent in the user payload
		// 	if upsertedUser.CustomColumns == nil || len(upsertedUser.CustomColumns) == 0 {
		// 		continue
		// 	}

		// 	for customKey, value := range upsertedUser.CustomColumns {
		// 		// find a extra columns that matches a reconciliation key
		// 		// value should not be null
		// 		if customKey == key && !entity.IsInterfaceNil(value) {
		// 			reconciliationKeysFound[key] = value

		// 			// if a user already exists, check if reconciliation key is different (new/updated)
		// 			if existingUser != nil {
		// 				_, keyExists := existingUser.CustomColumns[customKey]

		// 				// if new reconciliation key is provided
		// 				if !keyExists {
		// 					hasDifferentKeysThanExisting = true
		// 				}

		// 				// if key already exists and is different
		// 				if keyExists && existingUser.CustomColumns[customKey] != value {
		// 					hasDifferentKeysThanExisting = true
		// 				}
		// 			}
		// 		}
		// 	}
		// 	continue
		// }

		// native reconciliation key
		switch key {
		case "email":
			// if upserted user has an email provided, we will find matching users by "email"
			// if the upserted user already exists in DB, but has no email yet
			if upsertedUser.Email != nil && !upsertedUser.Email.IsNull {
				reconciliationKeysFound[key] = upsertedUser.Email.String

				// existing user has new or update reconciliation key
				if existingUser != nil && ((existingUser.Email == nil || existingUser.Email.IsNull) || existingUser.Email.String != upsertedUser.Email.String) {
					hasDifferentKeysThanExisting = true
				}
			}
		case "email_md5":
			if upsertedUser.EmailMD5 != nil && !upsertedUser.EmailMD5.IsNull {
				reconciliationKeysFound[key] = upsertedUser.EmailMD5.String

				// existing user has new or update reconciliation key
				if existingUser != nil && ((existingUser.EmailMD5 == nil || existingUser.EmailMD5.IsNull) || existingUser.EmailMD5.String != upsertedUser.EmailMD5.String) {
					hasDifferentKeysThanExisting = true
				}
			}
		case "email_sha1":
			if upsertedUser.EmailSHA1 != nil && !upsertedUser.EmailSHA1.IsNull {
				reconciliationKeysFound[key] = upsertedUser.EmailSHA1.String

				// existing user has new or update reconciliation key
				if existingUser != nil && ((existingUser.EmailSHA1 == nil || existingUser.EmailSHA1.IsNull) || existingUser.EmailSHA1.String != upsertedUser.EmailSHA1.String) {
					hasDifferentKeysThanExisting = true
				}
			}
		case "email_sha256":
			if upsertedUser.EmailSHA256 != nil && !upsertedUser.EmailSHA256.IsNull {
				reconciliationKeysFound[key] = upsertedUser.EmailSHA256.String

				// existing user has new or update reconciliation key
				if existingUser != nil && ((existingUser.EmailSHA256 == nil || existingUser.EmailSHA256.IsNull) || existingUser.EmailSHA256.String != upsertedUser.EmailSHA256.String) {
					hasDifferentKeysThanExisting = true
				}
			}
		case "telephone":
			if upsertedUser.Telephone != nil && !upsertedUser.Telephone.IsNull {
				reconciliationKeysFound[key] = upsertedUser.Telephone.String

				// existing user has new or update reconciliation key
				if existingUser != nil && ((existingUser.Telephone == nil || existingUser.Telephone.IsNull) || existingUser.Telephone.String != upsertedUser.Telephone.String) {
					hasDifferentKeysThanExisting = true
				}
			}
		default:
			return eris.Errorf("native user reconciliation key %v not implemented", key)
		}
	}

	// merge eventual users if:
	// - has reconciliation key in payload
	// - (user already exists & reconciliation is different) OR (user is new)

	if len(reconciliationKeysFound) > 0 && (existingUser == nil || hasDifferentKeysThanExisting) {

		// we have to check if it matches existing users
		// it could match as much users as there are reconciliation keys configured in the workspace
		var matchingUsers []*entity.User

		matchingUsers, err = pipe.Repository.FindEventualUsersToMergeWith(spanCtx, pipe.Workspace, upsertedUser, reconciliationKeysFound, tx)

		if err != nil {
			return eris.Wrap(err, "UserUpsert")
		}

		// FOUND users to merge! hard work here :)
		if len(matchingUsers) > 0 {

			// if many users matched, check if one is authenticated
			// otherwise merge all into the oldest user
			var authenticatedUserMatched *entity.User
			var oldestUserMatched *entity.User

			if len(matchingUsers) > 0 {
				for _, u := range matchingUsers {
					if u.IsAuthenticated {
						authenticatedUserMatched = u
						continue
					}

					if oldestUserMatched == nil {
						oldestUserMatched = u
					}

					if u.CreatedAt.Before(oldestUserMatched.CreatedAt) {
						oldestUserMatched = u
					}
				}
			}

			// merging matched users:
			// - merge upsertedUser now
			// - create an async "data import" to merge other users

			type mergingOperation struct {
				FromUser entity.User
				ToUser   entity.User
			}

			toMergeAsync := []mergingOperation{}

			for _, matchedUser := range matchingUsers {

				var fromUser *entity.User // the merged user
				var toUser *entity.User   // the survivor of the merge

				if upsertedUser.IsAuthenticated {
					fromUser = matchedUser
					toUser = upsertedUser
				} else {

					// if we have one match, keep the existing user
					if len(matchingUsers) == 1 {
						fromUser = upsertedUser
						toUser = matchedUser
					}

					// if many users matched, check if one is authenticated
					// otherwise merge all into the oldest user
					if len(matchingUsers) > 1 {
						if authenticatedUserMatched != nil {

							// if the matched user is the authenticated one, keep it
							if matchedUser.ID == authenticatedUserMatched.ID {
								fromUser = upsertedUser
								toUser = matchedUser
							} else {
								// merge matched user into authenticated matched user
								fromUser = matchedUser
								toUser = authenticatedUserMatched
							}
						}

						// if all matched users are anonymous, merge into the oldest one
						if authenticatedUserMatched == nil {

							// if the matched user is the oldest, keep it
							if matchedUser.ID == oldestUserMatched.ID {
								fromUser = upsertedUser
								toUser = matchedUser
							} else {
								// merge this matched user into the oldest one
								fromUser = matchedUser
								toUser = oldestUserMatched
							}
						}
					}
				}

				toMergeAsync = append(toMergeAsync, mergingOperation{
					FromUser: *fromUser,
					ToUser:   *toUser,
				})
			}

			// create an async data import batch to process matched users
			if len(toMergeAsync) > 0 {
				items := []string{}

				for _, mergingOpe := range toMergeAsync {
					isAuthenticatedBool := "false"
					if mergingOpe.ToUser.IsAuthenticated {
						isAuthenticatedBool = "true"
					}
					items = append(items, fmt.Sprintf(`{
						"workspace_id":"%s",
						"kind":"user_alias",
						"user_alias": {
							"from_user_external_id":"%s",
							"to_user_external_id":"%s",
							"to_user_is_authenticated":%v,
							"to_user_created_at":"%s"
						}
					}`,
						pipe.Workspace.ID,
						mergingOpe.FromUser.ID,
						mergingOpe.ToUser.ID,
						isAuthenticatedBool,
						mergingOpe.ToUser.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
					))
				}

				pipe.DataLogEnqueue(spanCtx, nil, dto.DataLogOriginInternalDataLog, pipe.DataLog.ID, pipe.Workspace.ID, items, false)

				if pipe.HasError() {
					return eris.New(pipe.QueueResult.Error)
				}
			}
		}
	}

	// merge fields if user already exists

	updatedFields := []*entity.UpdatedField{}

	if existingUser != nil {
		// svc.Logger.Printf("USER ALREADY EXISTS %+v\n", existingUser)
		updatedFields = upsertedUser.MergeInto(existingUser, pipe.Workspace)
		upsertedUser = existingUser
		// svc.Logger.Printf("AFTER MERGE %+v\n", upsertedUser)
	}

	// handle default mandatory values
	if upsertedUser.IsNew() {

		// enrich country+latitude+longitude if context is provided only for new users
		if country, ok := pipe.DataLog.Context.HeadersAndParams["country"]; ok {
			if govalidator.IsIn(country, common.CountriesCodes...) {
				upsertedUser.SetCountry(&country, pipe.DataLog.EventAt)
			}
		}

		if latitude, ok := pipe.DataLog.Context.HeadersAndParams["latitude"]; ok {
			if longitude, ok := pipe.DataLog.Context.HeadersAndParams["longitude"]; ok {
				// convert to float64
				if latitudeFloat64, err := strconv.ParseFloat(latitude, 32); err == nil {
					if longitudeFloat64, err := strconv.ParseFloat(longitude, 32); err == nil {
						upsertedUser.SetLatitude(&entity.NullableFloat64{IsNull: false, Float64: latitudeFloat64}, pipe.DataLog.EventAt)
						upsertedUser.SetLongitude(&entity.NullableFloat64{IsNull: false, Float64: longitudeFloat64}, pipe.DataLog.EventAt)
					}
				}
			}
		}

		// clear fields timestamp if object is new, to avoid storing extra data
		upsertedUser.FieldsTimestamp = entity.FieldsTimestamp{}

		// insert user
		if err = pipe.Repository.InsertUser(spanCtx, upsertedUser, tx); err != nil {
			return eris.Wrap(err, "UserUpsert")
		}

		if isChild {
			if err := pipe.InsertChildDataLog(spanCtx, "user", "create", upsertedUser.ID, upsertedUser.ID, upsertedUser.ExternalID, updatedFields, *upsertedUser.UpdatedAt, tx); err != nil {
				return err
			}
		} else {
			pipe.DataLog.Action = "create"
		}

	} else {

		if !isChild {
			pipe.DataLog.Action = "update"
			pipe.DataLog.UpdatedFields = updatedFields
		}

		// abort if no fields were updated
		if len(updatedFields) == 0 {
			if !isChild {
				pipe.DataLog.Action = "noop"
			}
			return nil
		}

		// persist changes
		if err = pipe.Repository.UpdateUser(spanCtx, upsertedUser, tx); err != nil {
			if eris.Is(err, repository.ErrRowNotUpdated) {
				// debug updatedFields
				updatedFieldsJSON, _ := json.Marshal(updatedFields)
				pipe.Logger.Printf("user row not updated with fields: %s", string(updatedFieldsJSON))
				return nil
			}
			return eris.Wrap(err, "UserUpsert")
		}

		if isChild {
			if err := pipe.InsertChildDataLog(spanCtx, "user", "update", upsertedUser.ID, upsertedUser.ID, upsertedUser.ExternalID, updatedFields, *upsertedUser.UpdatedAt, tx); err != nil {
				return err
			}
		}
	}

	return nil
}
