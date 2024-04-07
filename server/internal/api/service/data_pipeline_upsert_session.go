package service

import (
	"context"
	"database/sql"
	"strings"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
	"go.opencensus.io/trace"
)

func (pipe *DataLogPipeline) UpsertSession(ctx context.Context, isChild bool, tx *sql.Tx) (err error) {

	spanCtx, span := trace.StartSpan(ctx, "UpsertSession")
	defer span.End()

	// find eventual existing session
	var existingSession *entity.Session
	updatedFields := []*entity.UpdatedField{}

	existingSession, err = pipe.Repository.FindSessionByID(spanCtx, pipe.Workspace, pipe.DataLog.UpsertedSession.ID, pipe.DataLog.UpsertedSession.UserID, tx)

	if err != nil {
		return eris.Wrap(err, "SessionUpsert")
	}

	// insert new session
	if existingSession == nil {

		// just for insert: clear fields timestamp if object is new, to avoid storing extra data
		pipe.DataLog.UpsertedSession.FieldsTimestamp = entity.FieldsTimestamp{}

		if err = pipe.Repository.InsertSession(spanCtx, pipe.DataLog.UpsertedSession, tx); err != nil {
			return
		}

		if isChild {
			if err := pipe.InsertChildDataLog(spanCtx, entity.ChildDataLog{
				Kind:           "session",
				Action:         "create",
				UserID:         pipe.DataLog.UpsertedUser.ID,
				ItemID:         pipe.DataLog.UpsertedSession.ID,
				ItemExternalID: pipe.DataLog.UpsertedSession.ExternalID,
				UpdatedFields:  updatedFields,
				EventAt:        *pipe.DataLog.UpsertedSession.UpdatedAt,
				Tx:             tx,
			}); err != nil {
				return err
			}
		} else {
			pipe.DataLog.Action = "create"
		}

		// process eventual rimdian utm_id
		if pipe.DataLog.UpsertedSession.UTMID != nil && strings.HasPrefix(pipe.DataLog.UpsertedSession.UTMID.String, "rmd~") {
			// format: rmd~CHANNEL~ENTITY_EXTERNAL_ID
			bits := strings.Split(pipe.DataLog.UpsertedSession.UTMID.String, "~")
			if len(bits) > 2 {
				channel := bits[1] // msg = entity type is message
				entityExternalID := bits[2]
				if channel == "email" {
					// update message "first click at"
					pipe.DataLog.UpsertedMessage = entity.NewMessageFromUTMID(channel, entityExternalID, pipe.DataLog.UpsertedSession.UserID, *pipe.DataLog.UpsertedSession.UpdatedAt)
					pipe.DataLog.UpsertedMessage.FirstClickAt = &pipe.DataLog.UpsertedSession.CreatedAt
				}
			}
		}

		return
	}

	// merge fields if session already exists
	updatedFields = pipe.DataLog.UpsertedSession.MergeInto(existingSession, pipe.Workspace)
	pipe.DataLog.UpsertedSession = existingSession

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
	if err = pipe.Repository.UpdateSession(spanCtx, pipe.DataLog.UpsertedSession, tx); err != nil {
		return eris.Wrap(err, "SessionUpsert")
	}

	if isChild {
		if err := pipe.InsertChildDataLog(spanCtx, entity.ChildDataLog{
			Kind:           "session",
			Action:         "update",
			UserID:         pipe.DataLog.UpsertedUser.ID,
			ItemID:         pipe.DataLog.UpsertedSession.ID,
			ItemExternalID: pipe.DataLog.UpsertedSession.ExternalID,
			UpdatedFields:  updatedFields,
			EventAt:        *pipe.DataLog.UpsertedSession.UpdatedAt,
			Tx:             tx,
		}); err != nil {
			return err
		}
	}

	return nil
}
