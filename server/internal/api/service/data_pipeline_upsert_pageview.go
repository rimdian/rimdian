package service

import (
	"context"
	"database/sql"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
	"go.opencensus.io/trace"
)

func (pipe *DataLogPipeline) UpsertPageview(ctx context.Context, isChild bool, tx *sql.Tx) (err error) {

	spanCtx, span := trace.StartSpan(ctx, "UpsertPageview")
	defer span.End()

	// find eventual existing pageview
	var existingPageview *entity.Pageview
	updatedFields := []*entity.UpdatedField{}

	existingPageview, err = pipe.Repository.FindPageviewByID(spanCtx, pipe.Workspace, pipe.DataLog.UpsertedPageview.ID, pipe.DataLog.UpsertedPageview.UserID, tx)

	if err != nil {
		return eris.Wrap(err, "PageviewUpsert")
	}

	// insert new pageview
	if existingPageview == nil {

		// just for insert: clear fields timestamp if object is new, to avoid storing extra data
		pipe.DataLog.UpsertedPageview.FieldsTimestamp = entity.FieldsTimestamp{}

		if err = pipe.Repository.InsertPageview(spanCtx, pipe.DataLog.UpsertedPageview, tx); err != nil {
			return
		}

		if isChild {
			if err := pipe.InsertChildDataLog(spanCtx, entity.ChildDataLog{
				Kind:           "pageview",
				Action:         "create",
				UserID:         pipe.DataLog.UpsertedUser.ID,
				ItemID:         pipe.DataLog.UpsertedPageview.ID,
				ItemExternalID: pipe.DataLog.UpsertedPageview.ExternalID,
				UpdatedFields:  updatedFields,
				EventAt:        *pipe.DataLog.UpsertedPageview.UpdatedAt,
				Tx:             tx,
			}); err != nil {
				return err
			}
		} else {
			pipe.DataLog.Action = "create"
		}

		return
	}

	// merge fields if pageview already exists
	updatedFields = pipe.DataLog.UpsertedPageview.MergeInto(existingPageview, pipe.Workspace)
	pipe.DataLog.UpsertedPageview = existingPageview

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
	if err = pipe.Repository.UpdatePageview(spanCtx, pipe.DataLog.UpsertedPageview, tx); err != nil {
		return eris.Wrap(err, "PageviewUpsert")
	}

	if isChild {
		if err := pipe.InsertChildDataLog(spanCtx, entity.ChildDataLog{
			Kind:           "pageview",
			Action:         "update",
			UserID:         pipe.DataLog.UpsertedUser.ID,
			ItemID:         pipe.DataLog.UpsertedPageview.ID,
			ItemExternalID: pipe.DataLog.UpsertedPageview.ExternalID,
			UpdatedFields:  updatedFields,
			EventAt:        *pipe.DataLog.UpsertedPageview.UpdatedAt,
			Tx:             tx,
		}); err != nil {
			return err
		}
	}

	return nil
}
