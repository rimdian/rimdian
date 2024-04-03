package service

import (
	"context"
	"database/sql"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
	"go.opencensus.io/trace"
)

func (pipe *DataLogPipeline) UpsertCustomEvent(ctx context.Context, isChild bool, tx *sql.Tx) (err error) {

	spanCtx, span := trace.StartSpan(ctx, "UpsertCustomEvent")
	defer span.End()

	// find eventual existing custom_event
	var existingCustomEvent *entity.CustomEvent
	updatedFields := []*entity.UpdatedField{}

	existingCustomEvent, err = pipe.Repository.FindCustomEventByID(spanCtx, pipe.Workspace, pipe.DataLog.UpsertedCustomEvent.ID, pipe.DataLog.UpsertedCustomEvent.UserID, tx)

	if err != nil && !sqlscan.NotFound(err) {
		return eris.Wrap(err, "CustomEventUpsert")
	}

	// insert new custom_event
	if existingCustomEvent == nil {

		// just for insert: clear fields timestamp if object is new, to avoid storing extra data
		pipe.DataLog.UpsertedCustomEvent.FieldsTimestamp = entity.FieldsTimestamp{}

		if err = pipe.Repository.InsertCustomEvent(spanCtx, pipe.DataLog.UpsertedCustomEvent, tx); err != nil {
			return
		}

		if isChild {
			if err := pipe.InsertChildDataLog(spanCtx, entity.ChildDataLog{
				Kind:           "custom_event",
				Action:         "create",
				UserID:         pipe.DataLog.UpsertedUser.ID,
				ItemID:         pipe.DataLog.UpsertedCustomEvent.ID,
				ItemExternalID: pipe.DataLog.UpsertedCustomEvent.ExternalID,
				UpdatedFields:  updatedFields,
				EventAt:        *pipe.DataLog.UpsertedCustomEvent.UpdatedAt,
				Tx:             tx,
			}); err != nil {
				return err
			}
		} else {
			pipe.DataLog.Action = "create"
		}

		return
	}

	// merge fields if custom_event already exists
	updatedFields = pipe.DataLog.UpsertedCustomEvent.MergeInto(existingCustomEvent, pipe.Workspace)
	pipe.DataLog.UpsertedCustomEvent = existingCustomEvent

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
	if err = pipe.Repository.UpdateCustomEvent(spanCtx, pipe.DataLog.UpsertedCustomEvent, tx); err != nil {
		return eris.Wrap(err, "CustomEventUpsert")
	}

	if isChild {
		if err := pipe.InsertChildDataLog(spanCtx, entity.ChildDataLog{
			Kind:           "custom_event",
			Action:         "update",
			UserID:         pipe.DataLog.UpsertedUser.ID,
			ItemID:         pipe.DataLog.UpsertedCustomEvent.ID,
			ItemExternalID: pipe.DataLog.UpsertedCustomEvent.ExternalID,
			UpdatedFields:  updatedFields,
			EventAt:        *pipe.DataLog.UpsertedCustomEvent.UpdatedAt,
			Tx:             tx,
		}); err != nil {
			return err
		}
	}

	return nil
}
