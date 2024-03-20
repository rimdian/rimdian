package migration

import (
	"context"
	"database/sql"
	"log"

	"github.com/rimdian/rimdian/internal/api/entity"
)

type Migration38 struct {
}

func (m *Migration38) GetMajorVersion() float64 {
	return 38.0
}

func (m *Migration38) HasSystemUpdate() bool {
	log.Printf("running migration %v: HasSystemUpdate()", m.GetMajorVersion())
	return true
}

func (m *Migration38) HasWorkspaceUpdate() bool {
	log.Printf("running migration %v: HasWorkspaceUpdate()", m.GetMajorVersion())
	return false
}

func (m *Migration38) UpdateSystem(ctx context.Context, cfg *entity.Config, systemConnection *sql.Conn) (err error) {
	log.Printf("running migration %v: UpdateSystem()", m.GetMajorVersion())

	// singlestore migration queries here
	queries := []string{
		"ALTER TABLE `workspace` ADD COLUMN files_settings JSON NOT NULL AFTER license_key;",
		"UPDATE `workspace` SET files_settings = '{\"endpoint\": \"\", \"access_key\": \"\", \"encrypted_secret_key\":\"\", \"bucket\": \"\", \"region\": \"\", \"cdn_endpoint\": \"\"}';",
	}

	// mysql migration queries here
	if cfg.DB_TYPE == "mysql" {
	}

	for _, query := range queries {
		log.Printf("executing query: %s", query)

		_, err = systemConnection.ExecContext(ctx, query)

		// ignore specific errors in case of retried migrations
		if err != nil {
			log.Printf("error updating workspace table: %v", err)
		}

	}

	return nil
}

func (m *Migration38) UpdateWorkspace(ctx context.Context, cfg *entity.Config, workspace *entity.Workspace, workspaceConnection *sql.Conn) (err error) {
	log.Printf("running migration %v: UpdateWorkspace()", m.GetMajorVersion())
	return nil
}

func NewMigration38() entity.MajorMigrationInterface {
	return &Migration38{}
}
