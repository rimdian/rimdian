package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

// fetch jobs for a given task_exec_id
func (repo *RepositoryImpl) GetTaskExecJobs(ctx context.Context, workspaceID string, taskExecID string, offset int, limit int) (jobs []*entity.TaskExecJob, total int, err error) {

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return nil, 0, eris.Wrap(err, "GetTaskExecJobs")
	}

	defer conn.Close()

	query := "SELECT * FROM task_exec_job WHERE task_exec_id = ? ORDER BY db_created_at DESC LIMIT ? OFFSET ?"

	jobs = []*entity.TaskExecJob{}

	err = sqlscan.Select(ctx, conn, &jobs, query, taskExecID, limit, offset)

	if err != nil {
		return nil, 0, eris.Wrap(err, "GetTaskExecJobs")
	}

	// count total without limits
	err = sqlscan.Get(ctx, conn, &total, "SELECT COUNT(*) FROM task_exec_job WHERE task_exec_id = ?", taskExecID)

	if err != nil {
		return nil, 0, eris.Wrap(err, "GetTaskExecJobs")
	}

	return jobs, total, nil
}

func (repo *RepositoryImpl) GetTaskExecJob(ctx context.Context, workspaceID string, jobID string) (job *entity.TaskExecJob, err error) {
	job = &entity.TaskExecJob{}

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return nil, eris.Wrap(err, "GetTaskExecJob")
	}

	defer conn.Close()

	err = sqlscan.Get(ctx, conn, job, "SELECT * FROM task_exec_job WHERE id = ? LIMIT 1", jobID)

	if err != nil {
		if sqlscan.NotFound(err) {
			return nil, eris.Wrap(entity.ErrTaskExecJobNotFound, "GetTaskExecJob")
		} else {
			return nil, eris.Wrap(err, "GetTaskExecJob")
		}
	}

	return job, nil
}

func (repo *RepositoryImpl) AddJobToTaskExec(ctxWithTimeout context.Context, taskExecID string, newJobID string, tx *sql.Tx) error {

	query, args, err := sq.Insert("task_exec_job").Columns(
		"id",
		"task_exec_id",
	).Values(
		newJobID,
		taskExecID,
	).ToSql()

	if err != nil {
		return eris.Wrap(err, "AddJobToTaskExec build query")
	}

	_, err = tx.ExecContext(ctxWithTimeout, query, args...)

	if err != nil {
		log.Printf("AddJobToTaskExec query: %v", query)
		return eris.Wrap(err, "AddJobToTaskExec")
	}

	return nil
}

func (repo *RepositoryImpl) GetRunningTaskExecByTaskID(ctx context.Context, taskID string, multipleExecKey *string, tx *sql.Tx) (taskExec *entity.TaskExec, err error) {

	taskExec = &entity.TaskExec{}

	builder := sq.Select("*").From("task_exec").Where(sq.Eq{"task_id": taskID}).Where(sq.Eq{"status": []int{entity.TaskExecStatusProcessing, entity.TaskExecStatusRetryingError}})

	if multipleExecKey != nil {
		builder = builder.Where(sq.Eq{"multiple_exec_key": *multipleExecKey})
	}

	query, args, err := builder.ToSql()

	if err != nil {
		return nil, eris.Wrap(err, "GetRunningTaskExecByTaskID")
	}

	err = sqlscan.Get(ctx, tx, taskExec, query, args...)

	if err != nil {
		if sqlscan.NotFound(err) {
			return nil, nil
		}
		return nil, eris.Wrap(err, "GetRunningTaskExecByTaskID")
	}

	return taskExec, nil
}

func (repo *RepositoryImpl) AbortTaskExec(ctx context.Context, taskExecID string, message string, tx *sql.Tx) error {

	query, args, err := sq.Update("task_exec").
		Set("status", entity.TaskExecStatusAborted).
		Set("message", message).
		Where(sq.Eq{"id": taskExecID}).
		Where(sq.Eq{"status": []int{entity.TaskExecStatusProcessing, entity.TaskExecStatusRetryingError}}).ToSql()

	if err != nil {
		return eris.Wrap(err, "AbortTaskExec")
	}

	_, err = tx.ExecContext(ctx, query, args...)

	if err != nil {
		return eris.Wrap(err, "AbortTaskExec")
	}

	return nil
}

func (repo *RepositoryImpl) StopTaskExecsForApp(ctx context.Context, appID string, tx *sql.Tx) error {

	sql, args, err := sq.Update("task_exec").
		Set("status", entity.TaskExecStatusAborted).
		Set("message", "app stopped").
		Where(sq.Like{"task_id": fmt.Sprint(appID, "%")}).
		Where(sq.Eq{"status": []int{entity.TaskExecStatusProcessing, entity.TaskExecStatusRetryingError}}).ToSql()

	if err != nil {
		return eris.Wrap(err, "StopTasksForApp")
	}
	_, err = tx.ExecContext(ctx, sql, args...)

	if err != nil {
		return eris.Wrap(err, "StopTasksForApp")
	}

	return nil
}

func (repo *RepositoryImpl) ListTaskExecs(ctx context.Context, workspaceID string, params *dto.TaskExecListParams) (taskExecs []*entity.TaskExec, nextToken string, previousToken string, code int, err error) {
	taskExecs = []*entity.TaskExec{}

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		code = 500
		return
	}

	defer conn.Close()

	// fetch an additional row to check if it has more for pagination
	limit := params.Limit + 1

	queryBuilder := sq.Select("*").From("task_exec").Limit(uint64(limit))

	// pagination if doesnt filter for a unique row
	if params.TaskExecID == nil {

		// before = db_created_at is an older date = persisted before in the DB
		// after = db_created_at is a more recent date = persisted after in the DB

		if !params.NextDate.IsZero() {
			// want rows older than x, use DESC sort
			// use less-than-equal to keep eventual rows that had same microseconds
			// exclude id = params.NextID to only select rows before it
			queryBuilder = queryBuilder.
				Where(sq.LtOrEq{"db_created_at": params.NextDate.Format(dto.MicrosecondLayout)}). // force microsecond format here as driver strips them off
				Where(sq.NotEq{"id": params.NextID}).
				OrderBy("db_created_at DESC")
		} else if !params.PreviousDate.IsZero() {
			// want rows newer than x, use ASC sort
			// use greater-than-equal to keep eventual rows that had same microseconds
			// exclude id = params.PreviousID to only select rows after it
			queryBuilder = queryBuilder.
				Where(sq.GtOrEq{"db_created_at": params.PreviousDate.Format(dto.MicrosecondLayout)}). // force microsecond format here as driver strips them off
				Where(sq.NotEq{"id": params.PreviousID}).
				OrderBy("db_created_at ASC")
		} else {
			// default sort
			queryBuilder = queryBuilder.OrderBy("db_created_at DESC")
		}
	} else {
		// look for a specific row, dont paginate
		queryBuilder = queryBuilder.Where(sq.Eq{"id": *params.TaskExecID})
	}

	if params.AppID != nil && (strings.HasPrefix(*params.AppID, "app_") || strings.HasPrefix(*params.AppID, "appx_")) {
		queryBuilder = queryBuilder.Where("task_id LIKE ?", fmt.Sprint(*params.AppID, "%"))
	}

	if params.TaskID != nil {
		queryBuilder = queryBuilder.Where(sq.Eq{"task_id": *params.TaskID})
	}

	if params.MultipleExecKey != nil {
		queryBuilder = queryBuilder.Where(sq.Eq{"multiple_exec_key": *params.MultipleExecKey})
	}

	if params.Status != nil {
		queryBuilder = queryBuilder.Where(sq.Eq{"status": params.Status})
	}

	// fetch tasks
	query, args, err := queryBuilder.ToSql()

	if err != nil {
		code = 500
		err = eris.Wrapf(err, "ListTaskExecs fetch query: %v, args: %+v", query, args)
		return
	}

	// log.Println(query)
	if err = sqlscan.Select(ctx, conn, &taskExecs, query, args...); err != nil {
		err = eris.Wrapf(err, "ListTaskExecs query: %v, args: %+v", query, args)
		code = 500
		return
	}

	hasMore := false

	if len(taskExecs) == limit {
		hasMore = true
	}

	// if we are going backwards in the list
	if !params.PreviousDate.IsZero() {

		// if we have more
		if hasMore {
			// remove last item to return the limit
			taskExecs = taskExecs[:len(taskExecs)-1]

		} else if len(taskExecs) < limit-1 {
			// if have less rows than wanted, we reached the begning of the list
			// we should return the list without pagination params then
			params.NextDate = time.Time{}
			params.PreviousDate = time.Time{}
			return repo.ListTaskExecs(ctx, workspaceID, params)
		}

		// when querying the most recent items we sort by date ASC
		// but the results are always returned sorted by date DESC
		// so we need to reserve the array order

		reversed := []*entity.TaskExec{}
		for i := len(taskExecs) - 1; i >= 0; i-- {
			reversed = append(reversed, taskExecs[i])
		}

		taskExecs = reversed

		// going backwards always has nextToken
		lastRow := taskExecs[len(taskExecs)-1]
		nextToken = dto.EncodePaginationToken(lastRow.ID, *lastRow.DBCreatedAt)

		// if we had more rows going backwards, we have a previous token
		if hasMore {
			firstRow := taskExecs[0]
			previousToken = dto.EncodePaginationToken(firstRow.ID, *firstRow.DBCreatedAt)
		}

	} else if hasMore {
		// if default sort order (=date DESC) has more rows
		// remove last item to return the max limit
		taskExecs = taskExecs[:len(taskExecs)-1]

		lastRow := taskExecs[params.Limit-1]
		nextToken = dto.EncodePaginationToken(lastRow.ID, *lastRow.DBCreatedAt)

		// was looking for rows before
		if params.NextID != "" {
			// was looking for rows after
			firstRow := taskExecs[0]
			previousToken = dto.EncodePaginationToken(firstRow.ID, *firstRow.DBCreatedAt)
		}
	}

	code = 200
	return
}

func (repo *RepositoryImpl) AddTaskExecWorker(ctx context.Context, taskExecID string, newJobID string, newWorker *entity.NewTaskExecWorker, tx *sql.Tx) error {

	jsonState, err := json.Marshal(newWorker.InitialState)

	if err != nil {
		return eris.Wrapf(err, "AddTaskWorker state json err: %v", newWorker.InitialState)
	}

	exprState := fmt.Sprintf("JSON_SET_JSON(state, 'workers', %v, ?)", newWorker.WorkerID)

	if repo.Config.DB_TYPE == "mysql" {
		exprState = fmt.Sprintf("JSON_SET(state, '$.workers.\"%v\"', CAST(? AS JSON))", newWorker.WorkerID)
	}

	query, args, err := sq.Update("task_exec").
		Set("state", sq.Expr(exprState, string(jsonState))).
		Where(sq.Eq{"id": taskExecID}).ToSql()

	if err != nil {
		return eris.Wrap(err, "AddTaskWorker build query")
	}

	if _, err = tx.ExecContext(ctx, query, args...); err != nil {
		return eris.Wrap(err, "AddTaskWorker")
	}

	query, args, err = sq.Insert("task_exec_job").Columns("id", "task_exec_id").Values(newJobID, taskExecID).ToSql()

	if err != nil {
		return eris.Wrap(err, "AddTaskWorker build query")
	}

	if _, err = tx.ExecContext(ctx, query, args...); err != nil {
		return eris.Wrap(err, "AddTaskWorker")
	}

	return nil
}

// set the task worker 'state' and job 'done_at'
func (repo *RepositoryImpl) UpdateTaskExecFromResult(ctx context.Context, taskRequestPayload *dto.TaskExecRequestPayload, taskResult *entity.TaskExecResult, tx *sql.Tx) error {

	if taskRequestPayload.JobID == "" {
		return eris.Errorf("job_id is required")
	}

	status := entity.TaskExecStatusProcessing
	if taskResult.IsError {
		if taskResult.IsDone {
			status = entity.TaskExecStatusAborted
		} else {
			status = entity.TaskExecStatusRetryingError
		}
	} else if taskResult.IsDone {
		status = entity.TaskExecStatusDone
	}

	// truncate message if it's over 500 chars
	if taskResult.Message != nil && len(*taskResult.Message) > 500 {
		taskResult.Message = entity.StringPtr((*taskResult.Message)[:500] + "...")
	}

	// every worker can report an error, but only the main worker can update the status
	update := sq.Update("task_exec").
		Set("message", taskResult.Message).
		Where(sq.Eq{"id": taskRequestPayload.TaskExecID})

	// main worker, only update status if task is not aborted
	// as we can have multiple workers running at the same time
	if taskRequestPayload.WorkerID == 0 {
		update = update.Set("status", status).
			Where(sq.NotEq{"status": entity.TaskExecStatusAborted})
	}

	// update state if provided

	if taskResult.UpdatedWorkerState != nil {

		jsonState, err := json.Marshal(taskResult.UpdatedWorkerState)

		if err != nil {
			return eris.Wrapf(err, "UpdateTaskExecFromResult state json err: %v", taskResult.UpdatedWorkerState)
		}

		// exprJobs := fmt.Sprintf("JSON_SET_STRING(jobs, '%v', 'done_at', ?)", taskRequestPayload.JobID)
		exprState := fmt.Sprintf("JSON_SET_JSON(state, 'workers', %v, ?)", taskRequestPayload.WorkerID)

		if repo.Config.DB_TYPE == "mysql" {
			// workers.index (its a map, not an array)
			// exprJobs = fmt.Sprintf("JSON_SET(jobs, '$.\"%v\".done_at', ?)", taskRequestPayload.JobID)
			exprState = fmt.Sprintf("JSON_SET(state, '$.workers.\"%v\"', CAST(? AS JSON))", taskRequestPayload.WorkerID)
		}

		update = update.Set("state", sq.Expr(exprState, string(jsonState)))
	}

	query, args, err := update.ToSql()

	if err != nil {
		return eris.Wrap(err, "UpdateTaskExecFromResult build query")
	}

	if _, err = tx.ExecContext(ctx, query, args...); err != nil {
		return eris.Wrap(err, "UpdateTaskExecFromResult")
	}

	doneAt := time.Now().UTC()

	// Update job "done_at"
	query = "UPDATE task_exec_job SET done_at = ? WHERE id = ?"
	args = []interface{}{doneAt, taskRequestPayload.JobID}

	// in dev we simulate the Google Task Queue in-memory, all jobs have "dev" id
	if taskRequestPayload.JobID == entity.TaskExecJobIDDev {
		query = "UPDATE task_exec_job SET done_at = ? WHERE task_exec_id = ? AND done_at IS NULL"
		args = []interface{}{doneAt, taskRequestPayload.TaskExecID}
	}

	if _, err = tx.ExecContext(ctx, query, args...); err != nil {
		return eris.Wrap(err, "UpdateTaskExecFromResult")
	}
	return nil
}

func (repo *RepositoryImpl) SetTaskExecError(ctx context.Context, workspaceID string, taskExecID string, workerID int, status int, message string) error {

	// limit message to 500 chars
	if len(message) > 500 {
		message = message[:500] + "..."
	}

	builder := sq.Update("task_exec").
		Set("message", message).
		Set("db_updated_at", time.Now().UTC()). // if error is same, DB won't update and updated_at wont be actualized
		Where(sq.Eq{"id": taskExecID})

	// only main worker can update status
	if workerID == 0 {
		builder = builder.Set("status", status)
	}

	// increment retry count if we retry
	if status == entity.TaskExecStatusRetryingError {
		builder = builder.Set("retry_count", sq.Expr("retry_count + 1"))
	}

	query, args, err := builder.ToSql()

	if err != nil {
		return eris.Wrap(err, "SetTaskExecError build query")
	}

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return err
	}

	defer conn.Close()

	_, err = conn.ExecContext(ctx, query, args...)

	if err != nil {
		return eris.Wrap(err, "SetTaskExecError")
	}

	return nil
}

func (repo *RepositoryImpl) GetTaskExec(ctx context.Context, workspaceID string, taskExecID string) (taskExec *entity.TaskExec, err error) {

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	taskExec = &entity.TaskExec{}

	// in dev only
	if taskExecID == entity.TaskExecIDDev {
		if repo.Config.ENV != entity.ENV_DEV {
			return nil, entity.ErrTaskExecNotFound
		}

		err = sqlscan.Get(ctx, conn, taskExec, "SELECT * FROM task_exec WHERE (status = 0 OR status = -1) LIMIT 1")
	} else {
		err = sqlscan.Get(ctx, conn, taskExec, "SELECT * FROM task_exec WHERE id = ? LIMIT 1", taskExecID)
	}

	if err != nil {
		if sqlscan.NotFound(err) {
			return nil, eris.Wrap(entity.ErrTaskExecNotFound, "GetTaskExec")
		} else {

			return nil, eris.Wrap(err, "GetTaskExec error")
		}
	}

	return taskExec, nil
}

func (repo *RepositoryImpl) InsertTaskExec(ctx context.Context, workspaceID string, taskExec *entity.TaskExec, job *entity.TaskExecJob, tx *sql.Tx) (err error) {

	if taskExec.Name == "" {
		taskExec.Name = taskExec.TaskID
	}

	if job.TaskExecID == "" {
		return eris.Errorf("job is malformatted")
	}

	query, args, err := sq.Insert("task_exec").Columns(
		"id",
		"task_id",
		"name",
		"multiple_exec_key",
		"on_multiple_exec",
		"state",
		"status",
	).Values(
		taskExec.ID,
		taskExec.TaskID,
		taskExec.Name,
		taskExec.MultipleExecKey,
		taskExec.OnMultipleExec,
		taskExec.State,
		taskExec.Status,
	).ToSql()

	if err != nil {
		return eris.Wrapf(err, "InsertTaskExec build query for task %+v\n", *taskExec)
	}

	// in a transaction
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, args...)

		if err != nil {
			if repo.IsDuplicateEntry(err) {
				return entity.ErrTaskExecAlreadyExists
			}
			return eris.Wrapf(err, "InsertTaskExec exec query %v", query)
		}

		if _, err = tx.ExecContext(ctx, "INSERT INTO task_exec_job (id, task_exec_id) VALUES (?, ?)", job.ID, job.TaskExecID); err != nil {
			return eris.Wrapf(err, "InsertTaskExec exec query %v", query)
		}

		return nil
	}

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return err
	}

	defer conn.Close()

	_, err = conn.ExecContext(ctx, query, args...)

	if err != nil {
		if repo.IsDuplicateEntry(err) {
			return entity.ErrWorkspaceAlreadyExists
		}
		return eris.Wrapf(err, "InsertTaskExec exec query %v", query)
	}

	if _, err = conn.ExecContext(ctx, "INSERT INTO task_exec_job (id, task_exec_id) VALUES (?, ?)", job.ID, job.TaskExecID); err != nil {
		return eris.Wrapf(err, "InsertTaskExec exec query %v", query)
	}

	return nil
}
