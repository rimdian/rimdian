package migration

import (
	"context"
	"database/sql"
	"log"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/internal/common/taskorchestrator"
	"google.golang.org/api/option"
)

type Migration39 struct {
}

func (m *Migration39) GetMajorVersion() float64 {
	return 39.0
}

func (m *Migration39) HasSystemUpdate() bool {
	log.Printf("running migration %v: HasSystemUpdate()", m.GetMajorVersion())
	return true
}

func (m *Migration39) HasWorkspaceUpdate() bool {
	log.Printf("running migration %v: HasWorkspaceUpdate()", m.GetMajorVersion())
	return true
}

func (m *Migration39) UpdateSystem(ctx context.Context, cfg *entity.Config, systemConnection *sql.Conn) (err error) {
	log.Printf("running migration %v: UpdateSystem()", m.GetMajorVersion())

	// insert new system task "system_import_users_to_subscription_list" for each workspace

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
				entity.TaskKindImportUsersToSubscriptionList,
				workspaceID,
				"Import users to subscription list",
				entity.OnMultipleExecDiscardNew,
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

		// add column: messaging_settings JSON NOT NULL
		_, err = systemConnection.ExecContext(ctx, `ALTER TABLE workspace ADD COLUMN messaging_settings JSON NOT NULL AFTER files_settings`)
		if err != nil {
			log.Printf("error adding column messaging_settings: %v", err)
			return err
		}

		// set messaging_settings to {}
		_, err = systemConnection.ExecContext(ctx, `UPDATE workspace SET messaging_settings = '{}'`)
		if err != nil {
			log.Printf("error setting messaging_settings: %v", err)
			return err
		}
	}

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
		entity.MessageSchema,
	}

	// mysql migration queries here
	if cfg.DB_TYPE == "mysql" {
		queries = []string{
			entity.SubscriptionListSchemaMYSQL,
			entity.SubscriptionListUserSchemaMYSQL,
			entity.MessageTemplateSchemaMYSQL,
			entity.BroadcastCampaignSchemaMYSQL,
			entity.MessageSchemaMYSQL,
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

	serviceAccount := option.WithCredentialsJSON([]byte(cfg.GCLOUD_JSON_CREDENTIALS))

	cloudTaskClient, err := cloudtasks.NewClient(ctx, serviceAccount)
	if err != nil {
		return err
	}

	taskOrchestrator := taskorchestrator.NewClient(cfg.GCLOUD_PROJECT, cfg.ENV, cloudTaskClient)

	queueName := taskOrchestrator.GetTransactionalMessageQueueNameForWorkspace(workspace.ID)
	if err = taskOrchestrator.EnsureQueue(ctx, cfg.TASK_QUEUE_LOCATION, queueName, 50); err != nil {
		return err
	}

	queueName = taskOrchestrator.GetMarketingMessageQueueNameForWorkspace(workspace.ID)
	return taskOrchestrator.EnsureQueue(ctx, cfg.TASK_QUEUE_LOCATION, queueName, 50)
}

func NewMigration39() entity.MajorMigrationInterface {
	return &Migration39{}
}
