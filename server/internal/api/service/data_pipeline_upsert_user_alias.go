package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/internal/api/repository"
	"github.com/rotisserie/eris"
	"go.opencensus.io/trace"
)

func (pipe *DataLogPipeline) UpsertUserAlias(ctx context.Context, isChild bool, tx *sql.Tx) (err error) {

	spanCtx, span := trace.StartSpan(ctx, "UpsertUserAlias")
	defer span.End()

	// check if users are already aliased
	var alias *entity.UserAlias
	alias, err = pipe.Repository.FindUserAlias(spanCtx, pipe.DataLog.UserAlias.FromUserExternalID, tx)

	if err != nil {
		return err
	}

	// abort if alias exists
	if alias != nil {
		return nil
	}

	paramsFromUserID := entity.ComputeUserID(pipe.DataLog.UserAlias.FromUserExternalID)
	paramsToUserID := entity.ComputeUserID(pipe.DataLog.UserAlias.ToUserExternalID)

	// fetch from user
	fromUser, err := pipe.Repository.FindUserByID(spanCtx, pipe.Workspace, paramsFromUserID, tx)

	if err != nil {
		return eris.Wrap(err, "UserAlias")
	}

	// fetch to user
	toUser, err := pipe.Repository.FindUserByID(spanCtx, pipe.Workspace, paramsToUserID, tx)

	if err != nil {
		return eris.Wrap(err, "UserAlias")
	}

	// insert alias
	err = pipe.Repository.CreateUserAlias(spanCtx, pipe.DataLog.UserAlias.FromUserExternalID, pipe.DataLog.UserAlias.ToUserExternalID, pipe.DataLog.UserAlias.ToUserIsAuthenticated, tx)

	if err != nil {

		// abort if already exists
		if eris.Is(err, entity.ErrUserAliasAlreadyExists) {
			return nil
		}

		return eris.Wrap(err, "UserAlias")
	}

	// use data import timestamp as the alias creation timestamp
	aliasAt := pipe.DataLog.UserAlias.ToUserCreatedAt
	aliasUpdatedFields := []*entity.UpdatedField{
		{Field: "user_id", PrevValue: paramsFromUserID, NewValue: paramsToUserID},
		{Field: "user_external_id", PrevValue: pipe.DataLog.UserAlias.FromUserExternalID, NewValue: pipe.DataLog.UserAlias.ToUserExternalID},
	}

	if isChild {
		if err := pipe.InsertChildDataLog(spanCtx, "user_alias", "create", paramsToUserID, paramsToUserID, pipe.DataLog.UserAlias.ToUserExternalID, aliasUpdatedFields, *aliasAt, tx); err != nil {
			return err
		}
	} else {
		pipe.DataLog.Action = "create"
		pipe.DataLog.UpdatedFields = aliasUpdatedFields
	}

	// merge users, and upsert the resulted user profile:
	// at this point it's still possible that both users are not yet in DB because of out-of-order data imports
	// we will still insert the user_alias, create a new 'to' user, but without merging user data as they have no profiles yet

	// if toUser is nil, create a default
	if toUser == nil {
		toUser = entity.NewUser(pipe.DataLog.UserAlias.ToUserExternalID, pipe.DataLog.UserAlias.ToUserIsAuthenticated, *pipe.DataLog.UserAlias.ToUserCreatedAt, *pipe.DataLog.UserAlias.ToUserCreatedAt, pipe.Workspace.DefaultUserTimezone, pipe.Workspace.DefaultUserLanguage, pipe.Workspace.DefaultUserCountry, nil)
	}

	updatedFields := []*entity.UpdatedField{}

	// merge users if both exist
	if fromUser != nil && toUser != nil {
		updatedFields = fromUser.MergeInto(toUser, pipe.Workspace)
	}

	// log upated fields
	// for _, uf := range updatedFields {
	// 	log.Printf("field %+v\n", uf)
	// }

	upsertUser := toUser

	// set updated_at to generate event_time at the right time
	upsertUser.UpdatedAt = pipe.DataLog.UserAlias.ToUserCreatedAt

	// persist in DB
	if upsertUser.IsNew() {

		// clear fields timestamp if object is new, to avoid storing extra data
		upsertUser.FieldsTimestamp = entity.FieldsTimestamp{}

		// insert user
		if err = pipe.Repository.InsertUser(spanCtx, upsertUser, tx); err != nil {
			return eris.Wrap(err, "UserAlias")
		}

		// we need to generate+insert a datalog user create here
		if err := pipe.InsertChildDataLog(spanCtx, "user", "create", upsertUser.ID, upsertUser.ID, upsertUser.ExternalID, nil, upsertUser.CreatedAt, tx); err != nil {
			return err
		}

	} else {

		// update only if fields have changed
		if len(updatedFields) > 0 {
			// persist changes
			if err = pipe.Repository.UpdateUser(spanCtx, upsertUser, tx); err != nil {
				if eris.Is(err, repository.ErrRowNotUpdated) {
					// debug updatedFields
					updatedFieldsJSON, _ := json.Marshal(updatedFields)
					log.Printf("user row not updated with fields: %s", string(updatedFieldsJSON))
					return nil
				}
				return eris.Wrap(err, "UserAlias")
			}

			if err := pipe.InsertChildDataLog(spanCtx, "user", "update", upsertUser.ID, upsertUser.ID, upsertUser.ExternalID, updatedFields, *upsertUser.UpdatedAt, tx); err != nil {
				return err
			}
		}
	}

	// merge data if fromUser already existed

	if fromUser != nil {

		if err = pipe.Repository.MergeUserSessions(spanCtx, pipe.Workspace, paramsFromUserID, paramsToUserID, tx); err != nil {
			return err
		}
		if err = pipe.Repository.MergeUserPostviews(spanCtx, pipe.Workspace, paramsFromUserID, paramsToUserID, tx); err != nil {
			return err
		}
		if err = pipe.Repository.MergeUserPageviews(spanCtx, pipe.Workspace, paramsFromUserID, paramsToUserID, tx); err != nil {
			return err
		}
		if err = pipe.Repository.MergeUserCustomEvents(spanCtx, pipe.Workspace, paramsFromUserID, paramsToUserID, tx); err != nil {
			return err
		}
		if err = pipe.Repository.MergeUserCarts(spanCtx, pipe.Workspace, paramsFromUserID, paramsToUserID, tx); err != nil {
			return err
		}
		if err = pipe.Repository.MergeUserCartItems(spanCtx, pipe.Workspace, paramsFromUserID, paramsToUserID, tx); err != nil {
			return err
		}
		if err = pipe.Repository.MergeUserOrders(spanCtx, pipe.Workspace, paramsFromUserID, paramsToUserID, tx); err != nil {
			return err
		}
		if err = pipe.Repository.MergeUserOrderItems(spanCtx, pipe.Workspace, paramsFromUserID, paramsToUserID, tx); err != nil {
			return err
		}
		if err = pipe.Repository.MergeUserDevices(spanCtx, pipe.Workspace, paramsFromUserID, paramsToUserID, tx); err != nil {
			return err
		}
		if err = pipe.Repository.MergeUserDataLogs(spanCtx, pipe.Workspace, paramsFromUserID, pipe.DataLog.UserAlias.FromUserExternalID, paramsToUserID, tx); err != nil {
			return err
		}

		// TODO: merge app tables with user_id != none
	}

	return nil
}
