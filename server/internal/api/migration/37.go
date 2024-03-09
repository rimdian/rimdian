package migration

import (
	"context"
	"database/sql"
	"log"
	"strings"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

type Migration37 struct {
}

func (m *Migration37) GetMajorVersion() float64 {
	return 37.0
}

func (m *Migration37) HasSystemUpdate() bool {
	log.Printf("running migration %v: HasSystemUpdate()", m.GetMajorVersion())
	return true
}

func (m *Migration37) HasWorkspaceUpdate() bool {
	log.Printf("running migration %v: HasWorkspaceUpdate()", m.GetMajorVersion())
	return true
}

func (m *Migration37) UpdateSystem(ctx context.Context, cfg *entity.Config, systemConnection *sql.Conn) (err error) {
	log.Printf("running migration %v: UpdateSystem()", m.GetMajorVersion())

	// system tables queries here
	queries := []string{
		"ALTER TABLE `workspace` ADD COLUMN license_key VARCHAR(1024) AFTER data_hooks;",
	}

	for _, query := range queries {
		log.Printf("Executing query: %s", query)
		if _, err = systemConnection.ExecContext(ctx, query); err != nil {
			if !strings.Contains(err.Error(), "Duplicate column name") && !strings.Contains(err.Error(), "check that column/key exists") && !strings.Contains(err.Error(), "Error 1146") {
				log.Printf("error updating workspace table: %v", err)
				return err
			}
		}
	}

	return nil
}

func (m *Migration37) UpdateWorkspace(ctx context.Context, cfg *entity.Config, workspace *entity.Workspace, workspaceConnection *sql.Conn) (err error) {
	log.Printf("running migration %v: UpdateWorkspace()", m.GetMajorVersion())

	// singlestore migration queries here
	queries := []string{}

	// mysql migration queries here
	if cfg.DB_TYPE == "mysql" {
		queries = []string{}
	}

	for _, query := range queries {
		log.Printf("executing query: %s", query)

		_, err = workspaceConnection.ExecContext(ctx, query)

		if err != nil {
			log.Printf("error updating workspace %s: %v", workspace.ID, err)
		}

		// ignore specific errors in case of retried migrations
		if err != nil &&
			!strings.Contains(err.Error(), "Error 1146") &&
			!strings.Contains(err.Error(), "Error 1050") &&
			!strings.Contains(err.Error(), "Error 1054") &&
			!strings.Contains(err.Error(), "Error 1060") &&
			!strings.Contains(err.Error(), "Error 1091") &&
			!strings.Contains(err.Error(), "Error 1061") &&
			!strings.Contains(err.Error(), "Error 1062") &&
			!strings.Contains(err.Error(), "Error 1146") {
			return eris.Wrap(err, "ExecuteMigration")
		}
	}

	return nil
}

func NewMigration37() entity.MajorMigrationInterface {
	return &Migration37{}
}
