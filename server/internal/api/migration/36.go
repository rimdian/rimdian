package migration

import (
	"context"
	"database/sql"
	"log"
	"strings"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

type Migration36 struct {
}

func (m *Migration36) GetMajorVersion() float64 {
	return 36.0
}

func (m *Migration36) HasSystemUpdate() bool {
	log.Printf("running migration 36.0: HasSystemUpdate()")
	return true
}

func (m *Migration36) HasWorkspaceUpdate() bool {
	log.Printf("running migration 36.0: HasWorkspaceUpdate()")
	return true
}

func (m *Migration36) UpdateSystem(ctx context.Context, cfg *entity.Config, systemConnection *sql.Conn) (err error) {
	log.Printf("running migration 36.0: UpdateSystem()")

	// system tables queries here
	queries := []string{}

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

func (m *Migration36) UpdateWorkspace(ctx context.Context, cfg *entity.Config, workspace *entity.Workspace, workspaceConnection *sql.Conn) (err error) {
	log.Printf("running migration 36.0: UpdateWorkspace()")

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

func NewMigration36() entity.MajorMigrationInterface {
	return &Migration36{}
}
