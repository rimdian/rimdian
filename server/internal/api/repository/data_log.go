package repository

import (
	"context"
	"database/sql"
	"log"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	common "github.com/rimdian/rimdian/internal/common/dto"
	"github.com/rotisserie/eris"
	"go.opencensus.io/trace"
)

// data_log primary key doesn't use the user_id
func (repo *RepositoryImpl) MergeUserDataLogs(ctx context.Context, workspace *entity.Workspace, fromUserID string, fromUserExternalID string, toUserID string, tx *sql.Tx) (err error) {

	// update all ids with toUserID
	query := "UPDATE data_log SET user_id = ?, merged_from_user_external_id = ? WHERE user_id = ?"

	if _, err := tx.ExecContext(ctx, query, toUserID, fromUserExternalID, fromUserID); err != nil {
		return eris.Wrap(err, "MergeUserDataLogs")
	}

	return nil
}

func (repo *RepositoryImpl) GetDataLogChildren(ctx context.Context, workspaceID string, dataLogID string) (dataLogs []*entity.DataLog, err error) {

	spanCtx, span := trace.StartSpan(ctx, "GetDataLogChildren")
	defer span.End()

	conn, err := repo.GetWorkspaceConnection(spanCtx, workspaceID)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	dataLogs = []*entity.DataLog{}

	err = sqlscan.Select(spanCtx, conn, &dataLogs, "SELECT * FROM data_log WHERE origin = ? AND origin_id = ?", common.DataLogOriginInternalDataLog, dataLogID)

	if err != nil {
		return nil, eris.Wrap(err, "GetDataLogChildren error")
	}

	return dataLogs, nil
}

func (repo *RepositoryImpl) UpdateDataLog(ctx context.Context, workspaceID string, dataLog *entity.DataLog) (err error) {

	spanCtx, span := trace.StartSpan(ctx, "UpdateDataLog")
	defer span.End()

	conn, err := repo.GetWorkspaceConnection(spanCtx, dataLog.Context.WorkspaceID)

	if err != nil {
		return err
	}

	defer conn.Close()

	query, args, err := sq.Update("data_log").
		Set("checkpoint", dataLog.Checkpoint).
		Set("has_error", dataLog.HasError).
		Set("errors", dataLog.Errors).
		Set("hooks", dataLog.Hooks).
		Set("user_id", dataLog.UserID).
		Set("kind", dataLog.Kind).
		Set("action", dataLog.Action).
		Set("item_id", dataLog.ItemID).
		Set("item_external_id", dataLog.ItemExternalID).
		Set("updated_fields", dataLog.UpdatedFields).
		Set("event_at", dataLog.EventAt).
		Set("event_at_trunc", dataLog.EventAtTrunc).
		Where(sq.Eq{"id": dataLog.ID}).
		ToSql()

	if err != nil {
		return eris.Wrapf(err, "UpdateDataLog build query %+v\n", *dataLog)
	}

	_, err = conn.ExecContext(spanCtx, query, args...)

	if err != nil {
		return eris.Wrapf(err, "UpdateDataLog exec query %v", query)
	}

	return nil
}

func (repo *RepositoryImpl) InsertSegmentDataLogs(ctx context.Context, workspaceID string, segmentID string, segmentVersion int, taskID string, isEnter bool, createdAt time.Time, checkpoint int) (err error) {

	spanCtx, span := trace.StartSpan(ctx, "InsertSegmentDataLogs")
	defer span.End()

	conn, err := repo.GetWorkspaceConnection(spanCtx, workspaceID)

	if err != nil {
		return
	}

	defer conn.Close()

	action := "enter"
	enters := 1

	if !isEnter {
		action = "exit"
		enters = 0
	}

	// DataLogOriginInternalTask = 4

	dataLogContext := common.MapOfStrings{
		"workspace_id": workspaceID,
	}
	createdAtTrunc := createdAt.Truncate(1 * time.Hour)

	query := `
		INSERT IGNORE INTO data_log (
			id,
			origin,
			origin_id,
			context,
			item,
			checkpoint,
			has_error,
			errors,
			hooks,
			user_id,
			kind,
			action,
			item_id,
			item_external_id,
			updated_fields,
			event_at,
			event_at_trunc
		)
		SELECT 
			UUID(),
			?,
			?,
			?,
			'{}',
			?,
			?,
			'{}',
			'{}',
			user_id,
			'segment',
			?,
			segment_id,
			segment_id,
			'[]',
			?,
			?
		FROM user_segment_queue
		WHERE segment_id = ? AND segment_version = ? AND enters = ?;
	`

	if _, err = conn.ExecContext(spanCtx, query,
		common.DataLogOriginInternalTaskExec,
		taskID,
		dataLogContext,
		checkpoint, entity.DataLogHasErrorNone, action, createdAt, createdAtTrunc, segmentID, segmentVersion, enters); err != nil {
		return eris.Wrap(err, "InsertSegmentDataLogs")
	}

	return
}

func (repo *RepositoryImpl) InsertDataLog(ctx context.Context, workspaceID string, dataLog *entity.DataLog, tx *sql.Tx) (err error) {

	spanCtx, span := trace.StartSpan(ctx, "InsertDataLog")
	defer span.End()

	query, args, err := sq.Insert("data_log").Columns(
		"id",
		"origin",
		"origin_id",
		"context",
		"item",
		"checkpoint",
		"has_error",
		"errors",
		"hooks",
		"user_id",
		"kind",
		"action",
		"item_id",
		"item_external_id",
		"updated_fields",
		"event_at",
		"event_at_trunc",
	).Values(
		dataLog.ID,
		dataLog.Origin,
		dataLog.OriginID,
		dataLog.Context,
		[]byte(dataLog.Item),
		dataLog.Checkpoint,
		dataLog.HasError,
		dataLog.Errors,
		dataLog.Hooks,
		dataLog.UserID,
		dataLog.Kind,
		dataLog.Action,
		dataLog.ItemID,
		dataLog.ItemExternalID,
		dataLog.UpdatedFields,
		dataLog.EventAt,
		dataLog.EventAtTrunc,
	).ToSql()

	if err != nil {
		return eris.Wrapf(err, "InsertDataLog build query %+v\n", *dataLog)
	}

	_, err = tx.ExecContext(spanCtx, query, args...)

	if err != nil {
		return eris.Wrap(err, "InsertDataLog")
	}

	return nil
}

func (repo *RepositoryImpl) GetDataLog(ctx context.Context, workspaceID string, dataLogID string) (dataLog *entity.DataLog, err error) {

	spanCtx, span := trace.StartSpan(ctx, "GetDataLog")
	defer span.End()

	conn, err := repo.GetWorkspaceConnection(spanCtx, workspaceID)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	dataLog = &entity.DataLog{}

	err = sqlscan.Get(spanCtx, conn, dataLog, "SELECT * FROM data_log WHERE id = ?", dataLogID)

	if err != nil {
		return nil, eris.Wrap(err, "GetDataLog error")
	}

	return dataLog, nil
}

func (repo *RepositoryImpl) ListDataLogsToRespawn(ctx context.Context, workspaceID string, origin int, originID string, checkpoint int, limit int, withNextToken *string) (rows []*dto.DataLogToRespawn, err error) {

	spanCtx, span := trace.StartSpan(ctx, "ListDataLogsToRespawn")
	defer span.End()

	conn, err := repo.GetWorkspaceConnection(spanCtx, workspaceID)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	rows = []*dto.DataLogToRespawn{}

	// select id and event_at
	builder := sq.Select("id", "event_at").
		From("data_log").
		Where(sq.Eq{"origin": origin}).
		Where(sq.Eq{"origin_id": originID}).
		Where(sq.Eq{"checkpoint": checkpoint}).
		OrderBy("event_at_trunc DESC, event_at DESC").
		Limit(uint64(limit))

	if withNextToken != nil {
		// decode pagination token
		lastID, lastDate, err := dto.DecodePaginationToken(*withNextToken)

		if err != nil {
			return nil, eris.Wrap(err, "ListDataLogsToRespawn")
		}

		// add pagination condition
		builder = builder.
			Where(sq.LtOrEq{"event_at_trunc": lastDate.Truncate(1 * time.Hour)}).
			Where(sq.LtOrEq{"event_at": lastDate.Format(entity.MicrosecondLayout)}). // force microsecond format here as driver strips them off
			Where(sq.NotEq{"id": lastID})

	}

	query, args, err := builder.ToSql()

	if err != nil {
		return nil, eris.Wrap(err, "ListDataLogsToRespawn")
	}

	if err = sqlscan.Select(spanCtx, conn, &rows, query, args...); err != nil {
		return nil, eris.Wrapf(err, "ListDataLogsToRespawn query: %v, args: %+v", query, args)
	}

	return rows, nil
}

func (repo *RepositoryImpl) ListDataLogs(ctx context.Context, workspaceID string, params *dto.DataLogListParams) (dataLogs []*entity.DataLog, nextToken string, code int, err error) {

	spanCtx, span := trace.StartSpan(ctx, "ListDataLogs")
	defer span.End()

	dataLogs = []*entity.DataLog{}

	conn, err := repo.GetWorkspaceConnection(spanCtx, workspaceID)

	if err != nil {
		code = 500
		return
	}

	defer conn.Close()

	// fetch an additional row to check if it has more for pagination
	limit := params.Limit + 1

	queryBuilder := sq.Select("*").From("data_log").Limit(uint64(limit))

	// pagination if doesnt filter for a unique row
	if params.DataLogID == nil {

		// before = event_at is an older date = persisted before in the DB
		// after = event_at is a more recent date = persisted after in the DB

		if !params.NextDate.IsZero() {
			// want rows older than x, use DESC sort
			// filter with less-than-equal event_at_trunc to leverage columnstore SORT KEY
			// use less-than-equal to keep eventual rows that had same microseconds
			// refilter with more microsecond precision with event_at
			// exclude id = params.NextID to only select rows before it
			queryBuilder = queryBuilder.
				Where(sq.LtOrEq{"event_at_trunc": params.NextDate.Truncate(1 * time.Hour)}).
				Where(sq.LtOrEq{"event_at": params.NextDate.Format(entity.MicrosecondLayout)}). // force microsecond format here as driver strips them off
				Where(sq.NotEq{"id": params.NextID}).
				OrderBy("event_at_trunc DESC, event_at DESC")
		} else if !params.PreviousDate.IsZero() {
			// want rows newer than x, use ASC sort
			// filter with greater-than-equal event_at_trunc to leverage columnstore SORT KEY
			// use greater-than-equal to keep eventual rows that had same microseconds
			// refilter with more microsecond precision with event_at
			// exclude id = params.PreviousID to only select rows after it
			queryBuilder = queryBuilder.
				Where(sq.GtOrEq{"event_at_trunc": params.PreviousDate.Truncate(1 * time.Hour)}).
				Where(sq.GtOrEq{"event_at": params.PreviousDate.Format(entity.MicrosecondLayout)}). // force microsecond format here as driver strips them off
				Where(sq.NotEq{"id": params.PreviousID}).
				OrderBy("event_at_trunc ASC, event_at ASC")
		} else {
			// default sort
			queryBuilder = queryBuilder.OrderBy("event_at_trunc DESC, event_at DESC")
		}
	} else {
		// look for a specific row, dont paginate
		queryBuilder = queryBuilder.Where(sq.Eq{"id": *params.DataLogID})
	}

	// filters
	if params.Origin != nil {
		queryBuilder = queryBuilder.Where(sq.Eq{"origin": *params.Origin})
	}
	if params.OriginID != nil {
		queryBuilder = queryBuilder.Where(sq.Eq{"origin_id": *params.OriginID})
	}

	if params.HasError != nil {
		queryBuilder = queryBuilder.Where(sq.Eq{"has_error": &params.HasError})
	}

	if params.Checkpoint != nil {
		queryBuilder = queryBuilder.Where(sq.Eq{"checkpoint": &params.Checkpoint})
	}

	if params.Kind != nil {
		queryBuilder = queryBuilder.Where(sq.Eq{"kind": &params.Kind})
	}

	if params.UserID != nil {
		queryBuilder = queryBuilder.Where(sq.Eq{"user_id": &params.UserID})
	}

	if params.ItemID != nil {
		queryBuilder = queryBuilder.Where(sq.Eq{"item_id": &params.ItemID})
	}

	// fetch dataLogs
	query, args, err := queryBuilder.ToSql()

	// log.Println(query)
	// log.Printf("args %+v\n", args)

	if err != nil {
		code = 500
		err = eris.Wrapf(err, "ListDataLogs fetch query: %v, args: %+v", query, args)
		return
	}

	if err = sqlscan.Select(spanCtx, conn, &dataLogs, query, args...); err != nil {
		err = eris.Wrapf(err, "ListDataLogs query: %v, args: %+v", query, args)
		code = 500
		return
	}

	hasMore := false

	if len(dataLogs) == limit {
		hasMore = true
	}

	// if we are going backwards in the list
	if !params.PreviousDate.IsZero() {

		// if we have more
		if hasMore {
			// for _, x := range dataLogs {
			// 	log.Println(x.ID)
			// }
			// remove last item to return the limit
			dataLogs = dataLogs[:len(dataLogs)-1]

		} else if len(dataLogs) < limit-1 {
			// if have less rows than wanted, we reached the begning of the list
			// we should return the list without pagination params then
			params.NextDate = time.Time{}
			params.PreviousDate = time.Time{}
			return repo.ListDataLogs(ctx, workspaceID, params)
		}

		// when querying the most recent items we sort by date ASC
		// but the results are always returned sorted by date DESC
		// so we need to reserve the array order

		reversed := []*entity.DataLog{}
		for i := len(dataLogs) - 1; i >= 0; i-- {
			reversed = append(reversed, dataLogs[i])
		}

		dataLogs = reversed

		// going backwards always has nextToken
		lastRow := dataLogs[len(dataLogs)-1]
		nextToken = dto.EncodePaginationToken(lastRow.ID, lastRow.EventAt)

		// DONT GO BACKWARDS, as there is no "reverse" index, it will be too slow
		// if we had more rows going backwards, we have a previous token
		// if hasMore {
		// firstRow := dataLogs[0]
		// previousToken = dto.EncodePaginationToken(firstRow.ID, firstRow.EventAt)
		// }

	} else if hasMore {
		// if default sort order (=date DESC) has more rows
		// remove last item to return the max limit
		dataLogs = dataLogs[:len(dataLogs)-1]

		lastRow := dataLogs[params.Limit-1]
		nextToken = dto.EncodePaginationToken(lastRow.ID, lastRow.EventAt)

		// DONT GO BACKWARDS, as there is no "reverse" index, it will be too slow
		// was looking for rows before
		// if params.NextID != "" {
		// 	// was looking for rows after
		// 	firstRow := dataLogs[0]
		// 	previousToken = dto.EncodePaginationToken(firstRow.ID, firstRow.EventAt)
		// }
	}

	code = 200
	return
}

func (repo *RepositoryImpl) ListDataLogsToReprocess(ctx context.Context, workspaceID string, lastID string, lastIDEventAt time.Time, limit int) ([]*entity.DataLog, error) {

	spanCtx, span := trace.StartSpan(ctx, "ListDataLogsToReprocess")
	defer span.End()

	startedAt := time.Now()

	defer func() {
		log.Printf("ListDataLogsToReprocess duration %v", time.Since(startedAt))
	}()

	conn, err := repo.GetWorkspaceConnection(spanCtx, workspaceID)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	// select id where is_processed = false and event_at < untilDate order by SORT KEY(event_at_trunc)
	builder := sq.Select("*").
		From("data_log").
		Where(sq.Eq{"checkpoint": 0}).
		Where(sq.LtOrEq{"event_at": lastIDEventAt}).
		OrderBy("event_at_trunc DESC, event_at DESC"). // use default ordering
		Limit(uint64(limit))

	// if lastID is not empty, add a condition to exclude it
	if lastID != "" {
		builder = builder.Where(sq.NotEq{"id": lastID})
	}

	query, args, err := builder.ToSql()

	if err != nil {
		return nil, eris.Wrap(err, "ListDataLogsToReprocess")
	}

	dataLogs := []*entity.DataLog{}

	if err = sqlscan.Select(spanCtx, conn, &dataLogs, query, args...); err != nil {
		return nil, eris.Wrapf(err, "ListDataLogsToReprocess query: %v, args: %+v", query, args)
	}

	return dataLogs, nil
}

func (repo *RepositoryImpl) HasDataLogsToReprocess(ctx context.Context, workspaceID string, untilDate time.Time) (foundOne bool, err error) {

	spanCtx, span := trace.StartSpan(ctx, "HasDataLogsToReprocess")
	defer span.End()

	conn, err := repo.GetWorkspaceConnection(spanCtx, workspaceID)

	if err != nil {
		return false, err
	}

	defer conn.Close()

	// count where checkpoint = 0 and created_at < now - minimumDelay
	query, args, err := sq.Select("COUNT(*)").
		From("data_log").
		Where(sq.Eq{"checkpoint": 0}).
		Where(sq.LtOrEq{"event_at": untilDate}).
		Limit(1).
		ToSql()

	if err != nil {
		return false, eris.Wrap(err, "HasDataLogsToReprocess")
	}

	count := 0

	err = conn.QueryRowContext(spanCtx, query, args...).Scan(&count)

	if err != nil {
		return false, eris.Wrapf(err, "HasDataLogsToReprocess exec query %v", query)
	}

	if count > 0 {
		return true, nil
	}

	return false, nil
}

func (repo *RepositoryImpl) CountSuccessfulDataLogsForDemo(ctx context.Context, workspaceID string) (count int64, err error) {

	spanCtx, span := trace.StartSpan(ctx, "CountSuccessfulDataLogsForDemo")
	defer span.End()

	conn, err := repo.GetWorkspaceConnection(spanCtx, workspaceID)

	if err != nil {
		return 0, err
	}

	defer conn.Close()

	query, args, err := sq.Select("COUNT(*)").From("data_log").Where(sq.Eq{"checkpoint": 100}).ToSql()

	if err != nil {
		return 0, eris.Wrap(err, "CountSuccessfulDataLogsForDemo")
	}

	err = conn.QueryRowContext(spanCtx, query, args...).Scan(&count)

	if err != nil {
		return 0, eris.Wrapf(err, "CountSuccessfulDataLogsForDemo exec query %v", query)
	}

	return count, nil
}
