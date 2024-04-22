package service

import (
	"context"
	"database/sql"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
	"go.opencensus.io/trace"
)

func (pipe *DataLogPipeline) UpsertMessage(ctx context.Context, isChild bool, tx *sql.Tx) (err error) {

	spanCtx, span := trace.StartSpan(ctx, "UpsertMessage")
	defer span.End()

	// find eventual existing message
	var existingMessage *entity.Message
	updatedFields := []*entity.UpdatedField{}

	existingMessage, err = pipe.Repository.FindMessageByID(spanCtx, pipe.Workspace, pipe.DataLog.UpsertedMessage.ID, pipe.DataLog.UpsertedMessage.UserID, tx)

	if err != nil && !sqlscan.NotFound(err) {
		return eris.Wrap(err, "MessageUpsert")
	}

	// insert new message
	if existingMessage == nil {

		// just for insert: clear fields timestamp if object is new, to avoid storing extra data
		pipe.DataLog.UpsertedMessage.FieldsTimestamp = entity.FieldsTimestamp{}

		pipe.DataLog.UpsertedMessage.BeforeInsert(pipe.Config, pipe.DataLog.UpsertedUser, pipe.Workspace, pipe.DataLog.ID)

		if err = pipe.Repository.InsertMessage(spanCtx, pipe.DataLog.UpsertedMessage, tx); err != nil {
			return
		}

		if isChild {
			if err := pipe.InsertChildDataLog(spanCtx, entity.ChildDataLog{
				Kind:           "message",
				Action:         "create",
				UserID:         pipe.DataLog.UpsertedUser.ID,
				ItemID:         pipe.DataLog.UpsertedMessage.ID,
				ItemExternalID: pipe.DataLog.UpsertedMessage.ExternalID,
				UpdatedFields:  updatedFields,
				EventAt:        *pipe.DataLog.UpsertedMessage.UpdatedAt,
				Tx:             tx,
			}); err != nil {
				return err
			}
		} else {
			pipe.DataLog.Action = "create"
		}

		return
	}

	// merge fields if message already exists
	updatedFields = pipe.DataLog.UpsertedMessage.MergeInto(existingMessage)
	pipe.DataLog.UpsertedMessage = existingMessage

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
	if err = pipe.Repository.UpdateMessage(spanCtx, pipe.DataLog.UpsertedMessage, tx); err != nil {
		return eris.Wrap(err, "MessageUpsert")
	}

	if isChild {
		if err := pipe.InsertChildDataLog(spanCtx, entity.ChildDataLog{
			Kind:           "message",
			Action:         "update",
			UserID:         pipe.DataLog.UpsertedUser.ID,
			ItemID:         pipe.DataLog.UpsertedMessage.ID,
			ItemExternalID: pipe.DataLog.UpsertedMessage.ExternalID,
			UpdatedFields:  updatedFields,
			EventAt:        *pipe.DataLog.UpsertedMessage.UpdatedAt,
			Tx:             tx,
		}); err != nil {
			return err
		}
	}

	return nil
}
