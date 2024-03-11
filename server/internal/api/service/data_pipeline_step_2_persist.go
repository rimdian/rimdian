package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/rimdian/rimdian/internal/api/entity"
	"go.opencensus.io/trace"
)

func (pipe *DataLogPipeline) StepPersistDatalog(ctx context.Context) {

	spanCtx, span := trace.StartSpan(ctx, "StepPersistDatalog")
	defer span.End()

	pipe.ExtractAndValidateItem()
	pipe.AddDataLogGenerated(pipe.DataLog)

	// persist dataLog in DB
	_, retryableErr := pipe.Repository.RunInTransactionForWorkspace(spanCtx, pipe.Workspace.ID, func(ctx context.Context, tx *sql.Tx) (txCode int, txErr error) {

		// check if anonymous user has an alias to set the data_svc.Logger.user_id field properly
		if pipe.DataLog.UpsertedUser != nil && !pipe.DataLog.UpsertedUser.IsAuthenticated {
			var alias *entity.UserAlias
			alias, err := pipe.Repository.FindUserAlias(ctx, pipe.DataLog.UpsertedUser.ExternalID, tx)

			if err != nil {
				return 500, err
			}

			if alias != nil {
				pipe.DataLog.UserID = entity.ComputeUserID(alias.ToUserExternalID)
				pipe.UsersLock.AddUser(pipe.DataLog.UserID) // also lock the destination user
			}
		}

		if err := pipe.Repository.InsertDataLog(ctx, pipe.Workspace.ID, pipe.DataLog, tx); err != nil {
			return 500, err
		}

		// set db_created_at and db_updated_at,
		// its used to determine if datalog has been persisted or not
		now := time.Now()
		pipe.DataLog.DBCreatedAt = &now
		pipe.DataLog.DBUpdatedAt = &now

		return 200, nil
	})

	if retryableErr != nil {
		// check if dataLog already exists
		if pipe.Repository.IsDuplicateEntry(retryableErr) {
			pipe.Replay(spanCtx)
			return
		}

		pipe.SetError("server", fmt.Sprintf("will retry: %v", retryableErr), true)
		return
	}

	// move to next step if no error
	if !pipe.HasError() {
		pipe.DataLog.Checkpoint = entity.DataLogCheckpointPersisted
	}
}
