package repository

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (repo *RepositoryImpl) ActivateAppTasks(ctx context.Context, workspaceID string, appID string, tx *sql.Tx) (err error) {

	query, args, err := sq.Update("task").Set("is_active", true).Where(sq.Eq{"workspace_id": workspaceID}).Where(sq.Eq{"app_id": appID}).ToSql()

	if err != nil {
		return eris.Wrap(err, "ActivateTasks build query")
	}

	_, err = tx.ExecContext(ctx, query, args...)

	if err != nil {
		return eris.Wrapf(err, "ActivateTasks exec query %v", query)
	}

	return nil
}

func (repo *RepositoryImpl) GetTask(ctx context.Context, workspaceID string, taskID string, tx *sql.Tx) (task *entity.Task, err error) {

	task = &entity.Task{}

	query := "SELECT * FROM task WHERE workspace_id = ? AND id = ?"

	if tx == nil {
		conn, err := repo.GetSystemConnection(ctx)

		if err != nil {
			return nil, err
		}

		defer conn.Close()

		if err = sqlscan.Get(ctx, conn, task, query, workspaceID, taskID); err != nil {
			if sqlscan.NotFound(err) {
				return nil, eris.Errorf("task not found: %v", taskID)
			}
			return nil, eris.Wrapf(err, "GetTask exec query %v", query)
		}

		return task, nil
	}

	if err = sqlscan.Get(ctx, tx, task, query, workspaceID, taskID); err != nil {
		if sqlscan.NotFound(err) {
			return nil, eris.Errorf("task not found: %v", taskID)
		}
		return nil, eris.Wrapf(err, "GetTask exec query %v", query)
	}

	return task, nil
}

func (repo *RepositoryImpl) ListTasksToWakeUp(ctx context.Context) (crons []*entity.Task, err error) {

	conn, err := repo.GetSystemConnection(ctx)

	if err != nil {
		return
	}

	defer conn.Close()

	crons = []*entity.Task{}

	query := "SELECT * FROM task WHERE is_cron IS TRUE AND is_active IS TRUE AND next_run <= NOW()"

	if err = sqlscan.Select(ctx, conn, &crons, query); err != nil {
		return nil, eris.Wrapf(err, "ListTasksToWakeUp exec query %v", query)
	}

	return crons, nil
}

func (repo *RepositoryImpl) ListTasks(ctx context.Context, workspaceID string) (tasks []*entity.Task, err error) {

	conn, err := repo.GetSystemConnection(ctx)

	if err != nil {
		return
	}

	defer conn.Close()

	tasks = []*entity.Task{}

	queryBuilder := sq.Select("*").From("task").Where(sq.Eq{"workspace_id": workspaceID})

	// fetch tasks
	query, args, err := queryBuilder.ToSql()

	if err != nil {
		err = eris.Wrapf(err, "ListTasks fetch query: %v, args: %+v", query, args)
		return
	}

	if err = sqlscan.Select(ctx, conn, &tasks, query, args...); err != nil {
		err = eris.Wrapf(err, "ListTasks query: %v, args: %+v", query, args)
		return
	}

	return
}

func (repo *RepositoryImpl) InsertTask(ctx context.Context, task *entity.Task, tx *sql.Tx) (err error) {

	// insert the task

	query, args, err := sq.Insert("task").Columns(
		"id",
		"workspace_id",
		"name",
		"is_cron",
		"on_multiple_exec",
		"app_id",
		"is_active",
		"minutes_interval",
		"next_run",
	).Values(
		task.ID,
		task.WorkspaceID,
		task.Name,
		task.IsCron,
		task.OnMultipleExec,
		task.AppID,
		task.IsActive,
		task.MinutesInterval,
		task.NextRun,
	).ToSql()

	if err != nil {
		return eris.Wrapf(err, "InsertTask build query for task %+v\n", *task)
	}

	_, err = tx.ExecContext(ctx, query, args...)

	if err != nil {
		return eris.Wrapf(err, "InsertTask exec query for task %+v\n", *task)
	}

	return nil
}

func (repo *RepositoryImpl) UpdateTask(ctx context.Context, task *entity.Task, tx *sql.Tx) error {

	query, args, err := sq.Update("task").
		Set("name", task.Name).
		Set("is_active", task.IsActive).
		Set("is_cron", task.IsCron).
		Set("minutes_interval", task.MinutesInterval).
		Set("last_run", task.LastRun).
		Set("next_run", task.NextRun).
		Where(sq.Eq{"workspace_id": task.WorkspaceID}).
		Where(sq.Eq{"id": task.ID}).
		ToSql()

	if err != nil {
		return eris.Wrapf(err, "UpdateTask build query for task %+v\n", *task)
	}

	// log.Printf("UpdateTask query: %v, args: %+v", query, args)
	_, err = tx.ExecContext(ctx, query, args...)

	if err != nil {
		return eris.Wrapf(err, "UpdateTask exec query %v", query)
	}

	return nil
}

func (repo *RepositoryImpl) StopAppTasks(ctx context.Context, workspaceID string, appID string, tx *sql.Tx) (err error) {

	query, args, err := sq.Update("task").
		Set("is_active", false).
		Where(sq.Eq{"workspace_id": workspaceID}).
		Where(sq.Eq{"app_id": appID}).
		ToSql()

	if err != nil {
		return eris.Wrap(err, "StopAppTasks build query")
	}

	_, err = tx.ExecContext(ctx, query, args...)

	if err != nil {
		return eris.Wrapf(err, "StopAppTasks exec query %v", query)
	}

	return nil
}
func (repo *RepositoryImpl) DeleteAppTasks(ctx context.Context, workspaceID string, appID string, tx *sql.Tx) (err error) {

	query, args, err := sq.Delete("task").
		Where(sq.Eq{"workspace_id": workspaceID}).
		Where(sq.Eq{"app_id": appID}).
		ToSql()

	if err != nil {
		return eris.Wrap(err, "DeleteAppTasks build query")
	}

	_, err = tx.ExecContext(ctx, query, args...)

	if err != nil {
		return eris.Wrapf(err, "DeleteAppTasks exec query %v", query)
	}

	return nil
}
