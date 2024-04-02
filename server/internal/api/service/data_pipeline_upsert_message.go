package service

import (
	"context"
	"database/sql"
)

func (pipe *DataLogPipeline) UpsertMessage(ctx context.Context, isChild bool, tx *sql.Tx) (err error) {

	// TODO
	// spanCtx, span := trace.StartSpan(ctx, "UpsertMessage")
	// defer span.End()

	// // find eventual existing message
	// var existingMessage *entity.Message
	// updatedFields := []*entity.UpdatedField{}

	// existingMessage, err = pipe.Repository.FindMessage(spanCtx, pipe.DataLog.UpsertedMessage.SubscriptionListID, pipe.DataLog.UpsertedMessage.UserID, tx)

	// if err != nil {
	// 	return eris.Wrap(err, "MessageUpsert")
	// }

	// // insert new message
	// if existingMessage == nil {

	// 	// just for insert: clear fields timestamp if object is new, to avoid storing extra data
	// 	pipe.DataLog.UpsertedMessage.FieldsTimestamp = entity.FieldsTimestamp{}

	// 	if err = pipe.Repository.InsertMessage(spanCtx, pipe.DataLog.UpsertedMessage, tx); err != nil {
	// 		return
	// 	}

	// 	if isChild {
	// 		if err := pipe.InsertChildDataLog(spanCtx, "message", "create", pipe.DataLog.UpsertedUser.ID, pipe.DataLog.UpsertedMessage.SubscriptionListID, pipe.DataLog.UpsertedMessage.SubscriptionListID, updatedFields, *pipe.DataLog.UpsertedMessage.UpdatedAt, tx); err != nil {
	// 			return err
	// 		}
	// 	} else {
	// 		pipe.DataLog.Action = "create"
	// 	}

	// 	return
	// }

	// // merge fields if message already exists
	// updatedFields = pipe.DataLog.UpsertedMessage.MergeInto(existingMessage)
	// pipe.DataLog.UpsertedMessage = existingMessage

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
	// if err = pipe.Repository.UpdateMessage(spanCtx, pipe.DataLog.UpsertedMessage, tx); err != nil {
	// 	return eris.Wrap(err, "MessageUpsert")
	// }

	// if isChild {
	// 	if err := pipe.InsertChildDataLog(spanCtx, "message", "update", pipe.DataLog.UpsertedUser.ID, pipe.DataLog.UpsertedMessage.SubscriptionListID, pipe.DataLog.UpsertedMessage.SubscriptionListID, updatedFields, *pipe.DataLog.UpsertedMessage.UpdatedAt, tx); err != nil {
	// 		return err
	// 	}
	// }

	return nil
}
