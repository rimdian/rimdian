package repository

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (repo *RepositoryImpl) CreateWorkspaceTables(ctx context.Context, workspaceID string, tx *sql.Tx) (err error) {

	schemas := []string{
		entity.SegmentSchema,
		entity.UserSchema,
		entity.UserAliasSchema,
		entity.UserSegmentSchema,
		entity.UserSegmentQueueSchema,
		entity.DeviceSchema,
		entity.SessionSchema,
		entity.PostviewSchema,
		entity.PageviewSchema,
		entity.CustomEventSchema,
		entity.CartSchema,
		entity.CartItemSchema,
		entity.OrderSchema,
		entity.OrderItemSchema,
		entity.TaskExecSchema,
		entity.TaskExecJobSchema,
		entity.DataLogSchema,
		entity.UserLockSchema,
		entity.AppSchema,
		entity.MessageTemplateSchema,
		entity.BroadcastCampaignSchema,
		entity.SubscriptionListSchema,
		entity.SubscriptionListUserSchema,
		entity.MessageSchema,
	}

	if repo.Config.DB_TYPE == "mysql" {
		schemas = []string{
			entity.SegmentSchemaMYSQL,
			entity.UserSchemaMYSQL,
			entity.UserAliasSchemaMYSQL,
			entity.UserSegmentSchemaMYSQL,
			entity.UserSegmentQueueSchemaMYSQL,
			entity.DeviceSchemaMYSQL,
			entity.SessionSchemaMYSQL,
			entity.PostviewSchemaMYSQL,
			entity.PageviewSchemaMYSQL,
			entity.CustomEventSchemaMYSQL,
			entity.CartSchemaMYSQL,
			entity.CartItemSchemaMYSQL,
			entity.OrderSchemaMYSQL,
			entity.OrderItemSchemaMYSQL,
			entity.TaskExecSchemaMYSQL,
			entity.TaskExecJobSchemaMYSQL,
			entity.DataLogSchemaMYSQL,
			entity.UserLockSchemaMYSQL,
			entity.AppSchemaMYSQL,
			entity.MessageTemplateSchemaMYSQL,
			entity.BroadcastCampaignSchemaMYSQL,
			entity.SubscriptionListSchemaMYSQL,
			entity.SubscriptionListUserSchemaMYSQL,
			entity.MessageSchemaMYSQL,
		}
	}

	for _, schema := range schemas {
		if _, err = tx.ExecContext(ctx, schema); err != nil {
			return eris.Wrap(err, "CreateWorkspaceTables")
		}
	}

	return nil
}

// get workspace from id
func (repo *RepositoryImpl) GetWorkspace(ctx context.Context, workspaceID string) (workspace *entity.Workspace, err error) {

	conn, err := repo.GetSystemConnection(ctx)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	workspace = &entity.Workspace{}

	err = sqlscan.Get(ctx, conn, workspace, `SELECT * FROM workspace WHERE id = ? LIMIT 1`, workspaceID)

	// err = sqlscan.Get(ctx, conn, workspace, "SELECT * FROM workspace WHERE id = ? LIMIT 1", workspaceID)

	if err != nil {
		return nil, eris.Wrap(err, "GetWorkspace error")
	}

	return workspace, nil
}

// get workspace from id
func (repo *RepositoryImpl) InsertWorkspace(ctx context.Context, workspace *entity.Workspace, tx *sql.Tx) (err error) {

	query, args, err := sq.Insert("workspace").Columns(
		"id",
		"name",
		"is_demo",
		"demo_kind",
		"website_url",
		"privacy_policy_url",
		"industry",
		"currency",
		"organization_id",
		"secret_keys",
		"default_user_timezone",
		"default_user_country",
		"default_user_language",
		"user_reconciliation_keys",
		"user_id_signing",
		"session_timeout",
		"abandoned_carts_processed_until",
		"channel_groups",
		"channels",
		"domains",
		"has_orders",
		"has_leads",
		"lead_stages",
		"installed_apps",
		"fx_rates",
		"data_hooks",
		"license_key",
		"files_settings",
	).Values(
		workspace.ID,
		workspace.Name,
		workspace.IsDemo,
		workspace.DemoKind,
		workspace.WebsiteURL,
		workspace.PrivacyPolicyURL,
		workspace.Industry,
		workspace.Currency,
		workspace.OrganizationID,
		workspace.SecretKeys,
		workspace.DefaultUserTimezone,
		workspace.DefaultUserCountry,
		workspace.DefaultUserLanguage,
		workspace.UserReconciliationKeys,
		workspace.UserIDSigning,
		workspace.SessionTimeout,
		workspace.AbandonedCartsProcessedUntil,
		workspace.ChannelGroups,
		workspace.Channels,
		workspace.Domains,
		workspace.HasOrders,
		workspace.HasLeads,
		workspace.LeadStages,
		workspace.InstalledApps,
		// workspace.ObservabilityGroups,
		workspace.FxRates,
		workspace.DataHooks,
		workspace.LicenseKey,
		workspace.FilesSettings,
	).ToSql()

	if err != nil {
		return eris.Wrapf(err, "InsertWorkspace build query for workspace %+v\n", *workspace)
	}

	// in a transaction
	if tx != nil {

		// create workspace DB
		realName := repo.Config.DB_PREFIX + workspace.ID
		if _, err := tx.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %v;", realName)); err != nil {
			return eris.Wrapf(err, "Create workspace DB %v error", realName)
		}

		_, err = tx.ExecContext(ctx, query, args...)

		if err != nil {
			if repo.IsDuplicateEntry(err) {
				return entity.ErrWorkspaceAlreadyExists
			}
			return eris.Wrapf(err, "InsertWorkspace exec query %v", query)
		}

		// insert system tasks
		for _, task := range entity.SystemTasks {
			if _, err := tx.ExecContext(ctx, "INSERT INTO `task` (id, workspace_id, name, on_multiple_exec, app_id, is_active, is_cron, minutes_interval) VALUES (?, ?, ?, ?, 'system', 1, 0, 0)", task.ID, workspace.ID, task.Name, task.OnMultipleExec); err != nil {
				return eris.Wrapf(err, "InsertWorkspace exec query %v", task)
			}
		}

		return nil
	}

	conn, err := repo.GetSystemConnection(ctx)

	if err != nil {
		return err
	}

	defer conn.Close()

	// create workspace DB
	realName := repo.Config.DB_PREFIX + workspace.ID
	if _, err := conn.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %v;", realName)); err != nil {
		return eris.Wrapf(err, "Create workspace DB %v error", realName)
	}

	_, err = conn.ExecContext(ctx, query, args...)

	if err != nil {
		if repo.IsDuplicateEntry(err) {
			return entity.ErrWorkspaceAlreadyExists
		}
		return eris.Wrapf(err, "InsertWorkspace exec query %v", query)
	}

	// insert system tasks
	for _, task := range entity.SystemTasks {
		if _, err := conn.ExecContext(ctx, "INSERT INTO `task` (id, workspace_id, name, on_multiple_exec, app_id, is_active, is_cron, minutes_interval) VALUES (?, ?, ?, ?, 'system', 1, 0, 0)", task.ID, workspace.ID, task.Name, task.OnMultipleExec); err != nil {
			return eris.Wrapf(err, "InsertWorkspace exec query %v", task)
		}
	}

	return nil
}

// update workspace
func (repo *RepositoryImpl) UpdateWorkspace(ctx context.Context, workspace *entity.Workspace, tx *sql.Tx) error {

	query, args, err := sq.Update("workspace").
		Set("name", workspace.Name).
		Set("website_url", workspace.WebsiteURL).
		Set("privacy_policy_url", workspace.PrivacyPolicyURL).
		Set("industry", workspace.Industry).
		Set("currency", workspace.Currency).
		Set("secret_keys", workspace.SecretKeys).
		Set("default_user_timezone", workspace.DefaultUserTimezone).
		Set("default_user_country", workspace.DefaultUserCountry).
		Set("default_user_language", workspace.DefaultUserLanguage).
		Set("user_id_signing", workspace.UserIDSigning).
		Set("user_reconciliation_keys", workspace.UserReconciliationKeys).
		Set("session_timeout", workspace.SessionTimeout).
		Set("channel_groups", workspace.ChannelGroups).
		Set("domains", workspace.Domains).
		Set("installed_apps", workspace.InstalledApps).
		Set("channels", workspace.Channels).
		Set("has_orders", workspace.HasOrders).
		Set("has_leads", workspace.HasLeads).
		Set("lead_stages", workspace.LeadStages).
		Set("outdated_conversions_attribution", workspace.OutdatedConversionsAttribution).
		Set("fx_rates", workspace.FxRates).
		Set("data_hooks", workspace.DataHooks).
		Set("license_key", workspace.LicenseKey).
		Set("files_settings", workspace.FilesSettings).
		Where(sq.Eq{"id": workspace.ID}).
		ToSql()

	if err != nil {
		return eris.Wrapf(err, "UpdateWorkspace build query for workspace %+v\n", *workspace)
	}

	if tx != nil {

		// tx can be initiated on the workspace DB, make sure we USE the system DB here
		_, err = tx.ExecContext(ctx, "USE "+repo.Config.DB_PREFIX+repo.GetSystemDB())

		if err != nil {
			return eris.Wrapf(err, "UpdateWorkspace exec query %v", "USE "+repo.Config.DB_PREFIX+repo.GetSystemDB())
		}

		_, err = tx.ExecContext(ctx, query, args...)

		if err != nil {
			return eris.Wrapf(err, "UpdateWorkspace exec query %v", query)
		}

		return nil
	}

	conn, err := repo.GetSystemConnection(ctx)

	if err != nil {
		return err
	}

	defer conn.Close()

	_, err = conn.ExecContext(ctx, query, args...)

	if err != nil {
		return eris.Wrapf(err, "UpdateWorkspace exec query %v", query)
	}

	return nil
}

func (repo *RepositoryImpl) ShowTables(ctx context.Context, workspaceID string) (tables []*entity.TableInformationSchema, err error) {

	conn, err := repo.DB.Conn(ctx)

	if err != nil {
		return nil, eris.Wrap(err, "GetConnection error")
	}

	if _, err := conn.ExecContext(ctx, "USE information_schema"); err != nil {
		return nil, eris.Wrap(err, "GetConnection use DB information_schema error")
	}

	defer conn.Close()

	// fetch tables
	tables = []*entity.TableInformationSchema{}

	query := fmt.Sprintf(`SELECT 
		TABLE_NAME,
		TABLE_TYPE,
		CREATE_TIME,
		STORAGE_TYPE
	FROM TABLES WHERE TABLE_SCHEMA = '%v' ORDER BY TABLE_NAME ASC`, repo.Config.DB_PREFIX+workspaceID)

	if repo.Config.DB_TYPE == "mysql" {
		query = fmt.Sprintf(`SELECT 
			TABLE_NAME,
			TABLE_TYPE,
			CREATE_TIME
		FROM TABLES WHERE TABLE_SCHEMA = '%v' ORDER BY TABLE_NAME ASC`, repo.Config.DB_PREFIX+workspaceID)
	}
	if err := sqlscan.Select(ctx, conn, &tables, query); err != nil {
		return nil, err
	}

	// fetch table extra status
	status := []*entity.TableStatus{}

	query = fmt.Sprintf(`SELECT 
		TABLE_NAME, 
		SUM(ROWS) as TotalRows, 
		SUM(MEMORY_USE) as TotalMemoryUse
	FROM TABLE_STATISTICS 
	WHERE DATABASE_NAME = '%v' AND PARTITION_TYPE = 'Master'
	GROUP BY TABLE_NAME`, repo.Config.DB_PREFIX+workspaceID)

	// not available in mysql
	if repo.Config.DB_TYPE != "mysql" {
		if err := sqlscan.Select(ctx, conn, &status, query); err != nil {
			return nil, err
		}
	}

	for _, s := range status {
		// log.Printf("s %+v\n", s)
		for _, t := range tables {
			if t.Name == s.TableName {
				t.Rows = s.Rows
				t.MemoryUse = s.MemoryUse
			}
		}
	}

	// fetch columns
	columns := []*entity.ColumnInformationSchema{}

	query = fmt.Sprintf(`SELECT 
		TABLE_NAME,
		COLUMN_NAME,
		ORDINAL_POSITION,
		COLUMN_DEFAULT,
		IS_NULLABLE,
		COLUMN_TYPE,
		DATA_TYPE,
		CHARACTER_MAXIMUM_LENGTH,
		NUMERIC_PRECISION,
		CHARACTER_SET_NAME,
		COLLATION_NAME,
		COLUMN_KEY,
		EXTRA
	FROM COLUMNS WHERE TABLE_SCHEMA = '%v' ORDER BY ORDINAL_POSITION ASC`, repo.Config.DB_PREFIX+workspaceID)

	if err := sqlscan.Select(ctx, conn, &columns, query); err != nil {
		return nil, err
	}

	for _, col := range columns {

		for _, t := range tables {
			if t.Name == col.TableName {
				t.Columns = append(t.Columns, col)
			}
		}
	}

	return tables, nil
}

func (repo *RepositoryImpl) ListWorkspaces(ctx context.Context, organizationID *string) (workspaces []*entity.Workspace, err error) {

	conn, err := repo.GetSystemConnection(ctx)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	builder := sq.Select("*").
		From("workspace").
		Where(sq.Eq{"deleted_at": nil})

	if organizationID != nil && *organizationID != "" {
		builder = builder.Where(sq.Eq{"organization_id": organizationID})
	}

	query, args, err := builder.ToSql()

	if err != nil {
		return nil, eris.Wrap(err, "ListWorkspaces build query")
	}

	workspaces = []*entity.Workspace{}

	if err := sqlscan.Select(ctx, conn, &workspaces, query, args...); err != nil {
		return nil, eris.Wrap(err, "ListWorkspaces exec query")
	}

	return workspaces, nil
}

func (repo *RepositoryImpl) DeleteWorkspace(ctx context.Context, workspaceID string) error {

	conn, err := repo.GetSystemConnection(ctx)

	if err != nil {
		return err
	}

	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "DELETE FROM workspace WHERE id = ?", workspaceID); err != nil {
		return eris.Wrap(err, "DeleteWorkspace")
	}

	if _, err := conn.ExecContext(ctx, "DELETE FROM `task` WHERE workspace_id = ?", workspaceID); err != nil {
		return eris.Wrap(err, "DeleteWorkspace")
	}

	if _, err := conn.ExecContext(ctx, "DROP DATABASE IF EXISTS "+repo.Config.DB_PREFIX+workspaceID); err != nil {
		return eris.Wrapf(err, "DeleteWorkspace db %v", repo.Config.DB_PREFIX+workspaceID)
	}

	return nil
}
