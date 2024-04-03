package service

import (
	"context"
	"database/sql"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
	"go.opencensus.io/trace"
)

func (pipe *DataLogPipeline) UpsertPostview(ctx context.Context, isChild bool, tx *sql.Tx) (err error) {

	spanCtx, span := trace.StartSpan(ctx, "UpsertPostview")
	defer span.End()

	// find eventual existing postview
	var existingPostview *entity.Postview
	updatedFields := []*entity.UpdatedField{}

	existingPostview, err = pipe.Repository.FindPostviewByID(spanCtx, pipe.Workspace, pipe.DataLog.UpsertedPostview.ID, pipe.DataLog.UpsertedPostview.UserID, tx)

	if err != nil {
		return eris.Wrap(err, "PostviewUpsert")
	}

	// insert new postview
	if existingPostview == nil {

		// just for insert: clear fields timestamp if object is new, to avoid storing extra data
		pipe.DataLog.UpsertedPostview.FieldsTimestamp = entity.FieldsTimestamp{}

		if err = pipe.Repository.InsertPostview(spanCtx, pipe.DataLog.UpsertedPostview, tx); err != nil {
			return
		}

		if isChild {
			if err := pipe.InsertChildDataLog(spanCtx, entity.ChildDataLog{
				Kind:           "postview",
				Action:         "create",
				UserID:         pipe.DataLog.UpsertedUser.ID,
				ItemID:         pipe.DataLog.UpsertedPostview.ID,
				ItemExternalID: pipe.DataLog.UpsertedPostview.ExternalID,
				UpdatedFields:  updatedFields,
				EventAt:        *pipe.DataLog.UpsertedPostview.UpdatedAt,
				Tx:             tx,
			}); err != nil {
				return err
			}
		} else {
			pipe.DataLog.Action = "create"
		}

		return
	}

	// merge fields if postview already exists
	updatedFields = pipe.DataLog.UpsertedPostview.MergeInto(existingPostview, pipe.Workspace)
	pipe.DataLog.UpsertedPostview = existingPostview

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
	if err = pipe.Repository.UpdatePostview(spanCtx, pipe.DataLog.UpsertedPostview, tx); err != nil {
		return eris.Wrap(err, "PostviewUpsert")
	}

	if isChild {
		if err := pipe.InsertChildDataLog(spanCtx, entity.ChildDataLog{
			Kind:           "postview",
			Action:         "update",
			UserID:         pipe.DataLog.UpsertedUser.ID,
			ItemID:         pipe.DataLog.UpsertedPostview.ID,
			ItemExternalID: pipe.DataLog.UpsertedPostview.ExternalID,
			UpdatedFields:  updatedFields,
			EventAt:        *pipe.DataLog.UpsertedPostview.UpdatedAt,
			Tx:             tx,
		}); err != nil {
			return err
		}
	}

	return nil
}
