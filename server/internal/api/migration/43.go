package migration

import (
	"context"
	"database/sql"
	"log"

	"github.com/rimdian/rimdian/internal/api/entity"
)

type Migration43 struct {
}

func (m *Migration43) GetMajorVersion() float64 {
	return 43.0
}

func (m *Migration43) HasSystemUpdate() bool {
	log.Printf("running migration %v: HasSystemUpdate()", m.GetMajorVersion())
	return false
}

func (m *Migration43) HasWorkspaceUpdate() bool {
	log.Printf("running migration %v: HasWorkspaceUpdate()", m.GetMajorVersion())
	return true
}

func (m *Migration43) UpdateSystem(ctx context.Context, cfg *entity.Config, systemConnection *sql.Conn) (err error) {
	log.Printf("running migration %v: UpdateSystem()", m.GetMajorVersion())
	return nil
}

func (m *Migration43) UpdateWorkspace(ctx context.Context, cfg *entity.Config, workspace *entity.Workspace, workspaceConnection *sql.Conn) (err error) {
	log.Printf("running migration %v: UpdateWorkspace()", m.GetMajorVersion())

	query := "ALTER TABLE `app` ADD COLUMN sql_username VARCHAR(32) NOT NULL AFTER encrypted_secret_key;"
	_, err = workspaceConnection.ExecContext(ctx, query)

	// ignore specific errors in case of retried migrations
	if err != nil {
		log.Printf("error updating workspace tables: %v", err)
	}

	// add encrypted_sql_password column to app table
	query = "ALTER TABLE `app` ADD COLUMN encrypted_sql_password VARCHAR(512) NOT NULL AFTER sql_username;"
	_, err = workspaceConnection.ExecContext(ctx, query)

	// ignore specific errors in case of retried migrations
	if err != nil {
		log.Printf("error updating workspace tables: %v", err)
	}

	return nil
}

func NewMigration43() entity.MajorMigrationInterface {
	return &Migration43{}
}
