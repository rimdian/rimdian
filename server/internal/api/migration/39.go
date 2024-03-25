package migration

import (
	"context"
	"database/sql"
	"log"

	"github.com/rimdian/rimdian/internal/api/entity"
)

type Migration39 struct {
}

func (m *Migration39) GetMajorVersion() float64 {
	return 39.0
}

func (m *Migration39) HasSystemUpdate() bool {
	log.Printf("running migration %v: HasSystemUpdate()", m.GetMajorVersion())
	return false
}

func (m *Migration39) HasWorkspaceUpdate() bool {
	log.Printf("running migration %v: HasWorkspaceUpdate()", m.GetMajorVersion())
	return true
}

func (m *Migration39) UpdateSystem(ctx context.Context, cfg *entity.Config, systemConnection *sql.Conn) (err error) {
	log.Printf("running migration %v: UpdateSystem()", m.GetMajorVersion())
	return nil
}

func (m *Migration39) UpdateWorkspace(ctx context.Context, cfg *entity.Config, workspace *entity.Workspace, workspaceConnection *sql.Conn) (err error) {
	log.Printf("running migration %v: UpdateWorkspace()", m.GetMajorVersion())

	// singlestore migration queries here
	queries := []string{
		entity.SubscriptionListSchema,
		entity.SubscriptionListUserSchema,
		entity.MessageTemplateSchema,
		entity.BroadcastCampaignSchema,
	}

	// mysql migration queries here
	if cfg.DB_TYPE == "mysql" {
		queries = []string{
			entity.SubscriptionListSchemaMYSQL,
			entity.SubscriptionListUserSchemaMYSQL,
			entity.MessageTemplateSchemaMYSQL,
			entity.BroadcastCampaignSchemaMYSQL,
		}
	}

	for _, query := range queries {
		log.Printf("executing query: %s", query)

		_, err = workspaceConnection.ExecContext(ctx, query)

		// ignore specific errors in case of retried migrations
		if err != nil {
			log.Printf("error updating workspace tables: %v", err)
		}

	}

	return nil
}

func NewMigration39() entity.MajorMigrationInterface {
	return &Migration39{}
}
