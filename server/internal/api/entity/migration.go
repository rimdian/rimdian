package entity

import (
	"context"
	"database/sql"
)

// Migration interface
type MajorMigrationInterface interface {
	GetMajorVersion() float64
	HasSystemUpdate() bool
	HasWorkspaceUpdate() bool
	UpdateSystem(ctx context.Context, config *Config, systemConnection *sql.Conn) (err error)
	UpdateWorkspace(ctx context.Context, config *Config, workspace *Workspace, workspaceConnection *sql.Conn) (err error)
}
