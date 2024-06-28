package migration

import (
	"context"
	"database/sql"
	"log"

	"github.com/rimdian/rimdian/internal/api/entity"
)

type Migration41 struct {
}

func (m *Migration41) GetMajorVersion() float64 {
	return 41.0
}

func (m *Migration41) HasSystemUpdate() bool {
	log.Printf("running migration %v: HasSystemUpdate()", m.GetMajorVersion())
	return false
}

func (m *Migration41) HasWorkspaceUpdate() bool {
	log.Printf("running migration %v: HasWorkspaceUpdate()", m.GetMajorVersion())
	return true
}

func (m *Migration41) UpdateSystem(ctx context.Context, cfg *entity.Config, systemConnection *sql.Conn) (err error) {
	log.Printf("running migration %v: UpdateSystem()", m.GetMajorVersion())
	return nil
}

func (m *Migration41) UpdateWorkspace(ctx context.Context, cfg *entity.Config, workspace *entity.Workspace, workspaceConnection *sql.Conn) (err error) {
	log.Printf("running migration %v: UpdateWorkspace()", m.GetMajorVersion())

	// add ip column to order table
	query := "ALTER TABLE `order` ADD COLUMN ip VARCHAR(255) AFTER cancel_reason;"
	_, err = workspaceConnection.ExecContext(ctx, query)

	// ignore specific errors in case of retried migrations
	if err != nil {
		log.Printf("error updating workspace tables: %v", err)
	}

	return nil
}

func NewMigration41() entity.MajorMigrationInterface {
	return &Migration41{}
}
