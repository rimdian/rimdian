package service

import (
	"context"
	"database/sql"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
	"go.opencensus.io/trace"
)

func (pipe *DataLogPipeline) UpsertAppItem(ctx context.Context, isChild bool, tx *sql.Tx) (err error) {

	spanCtx, span := trace.StartSpan(ctx, "UpsertAppItem")
	defer span.End()

	// find eventual existing app item
	var existingAppItem *entity.AppItem
	updatedFields := []*entity.UpdatedField{}

	existingAppItem, err = pipe.Repository.FindAppItemByID(spanCtx, pipe.Workspace, pipe.DataLog.Kind, pipe.DataLog.UpsertedAppItem.ID, tx)

	if err != nil && !sqlscan.NotFound(err) {
		return eris.Wrap(err, "AppItemUpsert")
	}

	// insert new app item
	if existingAppItem == nil {

		// just for insert: clear fields timestamp if object is new, to avoid storing extra data
		pipe.DataLog.UpsertedAppItem.FieldsTimestamp = entity.FieldsTimestamp{}

		if err = pipe.Repository.InsertAppItem(spanCtx, pipe.DataLog.Kind, pipe.DataLog.UpsertedAppItem, tx); err != nil {
			return
		}

		if isChild {
			if err := pipe.InsertChildDataLog(spanCtx, entity.ChildDataLog{
				Kind:           pipe.DataLog.Kind,
				Action:         "create",
				UserID:         pipe.DataLog.UpsertedUser.ID,
				ItemID:         pipe.DataLog.UpsertedAppItem.ID,
				ItemExternalID: pipe.DataLog.UpsertedAppItem.ExternalID,
				UpdatedFields:  updatedFields,
				EventAt:        *pipe.DataLog.UpsertedAppItem.UpdatedAt,
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
	updatedFields = pipe.DataLog.UpsertedAppItem.MergeInto(existingAppItem)
	pipe.DataLog.UpsertedAppItem = existingAppItem

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
	if err = pipe.Repository.UpdateAppItem(spanCtx, pipe.DataLog.Kind, pipe.DataLog.UpsertedAppItem, tx); err != nil {
		return eris.Wrap(err, "AppItemUpsert")
	}

	if isChild {
		if err := pipe.InsertChildDataLog(spanCtx, entity.ChildDataLog{
			Kind:           pipe.DataLog.Kind,
			Action:         "update",
			UserID:         pipe.DataLog.UpsertedUser.ID,
			ItemID:         pipe.DataLog.UpsertedAppItem.ID,
			ItemExternalID: pipe.DataLog.UpsertedAppItem.ExternalID,
			UpdatedFields:  updatedFields,
			EventAt:        *pipe.DataLog.UpsertedAppItem.UpdatedAt,
			Tx:             tx,
		}); err != nil {
			return err
		}
	}

	return nil
}
