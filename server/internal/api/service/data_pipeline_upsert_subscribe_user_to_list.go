package service

import (
	"context"
	"database/sql"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
	"go.opencensus.io/trace"
)

func (pipe *DataLogPipeline) UpsertSubscriptionListUser(ctx context.Context, isChild bool, tx *sql.Tx) (err error) {

	spanCtx, span := trace.StartSpan(ctx, "UpsertSubscriptionListUser")
	defer span.End()

	// find eventual existing subscription_list_user
	var existingSubscriptionListUser *entity.SubscriptionListUser
	updatedFields := []*entity.UpdatedField{}

	existingSubscriptionListUser, err = pipe.Repository.FindSubscriptionListUser(spanCtx, pipe.DataLog.UpsertedSubscriptionListUser.SubscriptionListID, pipe.DataLog.UpsertedSubscriptionListUser.UserID, tx)

	if err != nil {
		return eris.Wrap(err, "SubscriptionListUserUpsert")
	}

	// insert new subscription_list_user
	if existingSubscriptionListUser == nil {

		// default status to active if not set
		if pipe.DataLog.UpsertedSubscriptionListUser.Status == nil {
			pipe.DataLog.UpsertedSubscriptionListUser.Status = entity.Int64Ptr(entity.SubscriptionListUserStatusActive)
		}

		// just for insert: clear fields timestamp if object is new, to avoid storing extra data
		pipe.DataLog.UpsertedSubscriptionListUser.FieldsTimestamp = entity.FieldsTimestamp{}

		if err = pipe.Repository.InsertSubscriptionListUser(spanCtx, pipe.DataLog.UpsertedSubscriptionListUser, tx); err != nil {
			return
		}

		if isChild {
			if err := pipe.InsertChildDataLog(spanCtx, "subscription_list_user", "create", pipe.DataLog.UpsertedUser.ID, pipe.DataLog.UpsertedSubscriptionListUser.SubscriptionListID, pipe.DataLog.UpsertedSubscriptionListUser.SubscriptionListID, updatedFields, *pipe.DataLog.UpsertedSubscriptionListUser.UpdatedAt, tx); err != nil {
				return err
			}
		} else {
			pipe.DataLog.Action = "create"
		}

		return
	}

	// merge fields if subscription_list_user already exists
	updatedFields = pipe.DataLog.UpsertedSubscriptionListUser.MergeInto(existingSubscriptionListUser)
	pipe.DataLog.UpsertedSubscriptionListUser = existingSubscriptionListUser

	// abort if no fields were updated
	if len(updatedFields) == 0 {
		if !isChild {
			pipe.DataLog.Action = "noop"
		}
		return nil
	}

	if !isChild {
		pipe.DataLog.Action = "update"
		pipe.DataLog.UpdatedFields = updatedFields
	}

	// persist changes
	if err = pipe.Repository.UpdateSubscriptionListUser(spanCtx, pipe.DataLog.UpsertedSubscriptionListUser, tx); err != nil {
		return eris.Wrap(err, "SubscriptionListUserUpsert")
	}

	if isChild {
		if err := pipe.InsertChildDataLog(spanCtx, "subscription_list_user", "update", pipe.DataLog.UpsertedUser.ID, pipe.DataLog.UpsertedSubscriptionListUser.SubscriptionListID, pipe.DataLog.UpsertedSubscriptionListUser.SubscriptionListID, updatedFields, *pipe.DataLog.UpsertedSubscriptionListUser.UpdatedAt, tx); err != nil {
			return err
		}
	}

	return nil
}
