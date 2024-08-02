package migration

import (
	"context"
	"database/sql"
	"log"

	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/entity"
)

type Migration42 struct {
}

func (m *Migration42) GetMajorVersion() float64 {
	return 42.0
}

func (m *Migration42) HasSystemUpdate() bool {
	log.Printf("running migration %v: HasSystemUpdate()", m.GetMajorVersion())
	return true
}

func (m *Migration42) HasWorkspaceUpdate() bool {
	log.Printf("running migration %v: HasWorkspaceUpdate()", m.GetMajorVersion())
	return false
}

func (m *Migration42) UpdateSystem(ctx context.Context, cfg *entity.Config, systemConnection *sql.Conn) (err error) {
	log.Printf("running migration %v: UpdateSystem()", m.GetMajorVersion())

	// fetch workspace ids
	ids := []string{}

	err = sqlscan.Select(ctx, systemConnection, &ids, `SELECT id FROM workspace`)
	if err != nil {
		log.Printf("error fetching workspace ids: %v", err)
		return err
	}

	for _, workspaceID := range ids {
		log.Printf("inserting task for workspace: %s", workspaceID)

		sql, args, err := squirrel.Insert("task").
			Columns(
				"id",
				"workspace_id",
				"name",
				"on_multiple_exec",
				"app_id",
				"is_active",
				"is_cron",
				"minutes_interval",
			).
			Values(
				entity.TaskKindRecomputeSegment,
				workspaceID,
				entity.TaskNameRecomputeSegment,
				entity.OnMultipleExecAbortExisting,
				"system",
				true,
				false,
				0,
			).
			ToSql()

		if err != nil {
			log.Printf("error building insert query: %v",
				err)
			return err
		}

		_, err = systemConnection.ExecContext(ctx, sql, args...)

		// silently ignore errors in case of retried migrations
		if err != nil {
			log.Printf("error inserting task: %v", err)
		}

		sql, args, err = squirrel.Insert("task").
			Columns(
				"id",
				"workspace_id",
				"name",
				"on_multiple_exec",
				"app_id",
				"is_active",
				"is_cron",
				"minutes_interval",
			).
			Values(
				entity.TaskKindRefreshOutdatedSegments,
				workspaceID,
				"Refresh outdated segments",
				entity.OnMultipleExecDiscardNew,
				"system",
				true,
				true,
				720, // every 12 hours
			).
			ToSql()

		if err != nil {
			log.Printf("error building insert query: %v",
				err)
			return err
		}

		_, err = systemConnection.ExecContext(ctx, sql, args...)

		// silently ignore errors in case of retried migrations
		if err != nil {
			log.Printf("error inserting task: %v", err)
		}
	}
	return nil
}

func (m *Migration42) UpdateWorkspace(ctx context.Context, cfg *entity.Config, workspace *entity.Workspace, workspaceConnection *sql.Conn) (err error) {
	log.Printf("running migration %v: UpdateWorkspace()", m.GetMajorVersion())

	return nil
}

func NewMigration42() entity.MajorMigrationInterface {
	return &Migration42{}
}
