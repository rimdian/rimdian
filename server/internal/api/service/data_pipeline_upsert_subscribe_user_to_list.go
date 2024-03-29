package service

import (
	"context"
	"database/sql"
)

func (pipe *DataLogPipeline) UpsertSubscriptionListUser(ctx context.Context, isChild bool, tx *sql.Tx) (err error) {

	// TODO
	// spanCtx, span := trace.StartSpan(ctx, "UpsertSubscriptionListUser")
	// defer span.End()

	// // find eventual existing subscribe_to_list
	// var existingSubscriptionListUser *entity.SubscriptionListUser
	// updatedFields := []*entity.UpdatedField{}

	// existingSubscriptionListUser, err = pipe.Repository.FindSubscriptionListUserByID(spanCtx, pipe.Workspace, pipe.DataLog.UpsertedSubscriptionListUser.ID, pipe.DataLog.UpsertedSubscriptionListUser.UserID, tx)

	// if err != nil {
	// 	return eris.Wrap(err, "SubscriptionListUserUpsert")
	// }

	// // insert new subscribe_to_list
	// if existingSubscriptionListUser == nil {

	// 	// just for insert: clear fields timestamp if object is new, to avoid storing extra data
	// 	pipe.DataLog.UpsertedSubscriptionListUser.FieldsTimestamp = entity.FieldsTimestamp{}

	// 	if err = pipe.Repository.InsertSubscriptionListUser(spanCtx, pipe.DataLog.UpsertedSubscriptionListUser, tx); err != nil {
	// 		return
	// 	}

	// 	if isChild {
	// 		if err := pipe.InsertChildDataLog(spanCtx, "subscribe_to_list", "create", pipe.DataLog.UpsertedUser.ID, pipe.DataLog.UpsertedSubscriptionListUser.SubscriptionListID, pipe.DataLog.UpsertedSubscriptionListUser.SubscriptionListID, updatedFields, *pipe.DataLog.UpsertedSubscriptionListUser.UpdatedAt, tx); err != nil {
	// 			return err
	// 		}
	// 	} else {
	// 		pipe.DataLog.Action = "create"
	// 	}

	// 	return
	// }

	// // merge fields if subscribe_to_list already exists
	// updatedFields = pipe.DataLog.UpsertedSubscriptionListUser.MergeInto(existingSubscriptionListUser)
	// pipe.DataLog.UpsertedSubscriptionListUser = existingSubscriptionListUser

	// // abort if no fields were updated
	// if len(updatedFields) == 0 {
	// 	if !isChild {
	// 		pipe.DataLog.Action = "noop"
	// 	}
	// 	return nil
	// }

	// if !isChild {
	// 	pipe.DataLog.Action = "update"
	// 	pipe.DataLog.UpdatedFields = updatedFields
	// }

	// // persist changes
	// if err = pipe.Repository.UpdateSubscriptionListUser(spanCtx, pipe.DataLog.UpsertedSubscriptionListUser, tx); err != nil {
	// 	return eris.Wrap(err, "SubscriptionListUserUpsert")
	// }

	// if isChild {
	// 	if err := pipe.InsertChildDataLog(spanCtx, "subscribe_to_list", "update", pipe.DataLog.UpsertedUser.ID, pipe.DataLog.UpsertedSubscriptionListUser.SubscriptionListID, pipe.DataLog.UpsertedSubscriptionListUser.SubscriptionListID, updatedFields, *pipe.DataLog.UpsertedSubscriptionListUser.UpdatedAt, tx); err != nil {
	// 		return err
	// 	}
	// }

	return nil
}
