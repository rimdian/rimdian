package repository

import (
	"context"
	"strings"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/common"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

// system db contains tables: account, setting, workspace, organization
// get actual installed settings
func (repo *RepositoryImpl) GetSettings(ctx context.Context) (settings *entity.Settings, err error) {

	conn, err := repo.GetSystemConnection(ctx)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	settings = &entity.Settings{}

	query := `SELECT 
	installed_version
	FROM setting 
	ORDER BY db_created_at DESC 
	LIMIT 1`

	err = conn.QueryRowContext(ctx, query).Scan(&settings.InstalledVersion)

	if err != nil {
		// tables not installed yet
		if sqlscan.NotFound(err) || strings.Contains(err.Error(), "Error 1146: Table") {
			return nil, entity.ErrSettingsTableNotFound
		}
		return nil, eris.Wrapf(err, "GetSettings exec query %v", query)
	}

	return settings, nil
}

// install system schemas and root account
func (repo *RepositoryImpl) Install(ctx context.Context, rootAccount *entity.Account, defaultOrganization *entity.Organization) (err error) {

	conn, err := repo.GetSystemConnection(ctx)

	if err != nil {
		return err
	}

	defer conn.Close()

	schemas := []string{
		entity.AccountSchema,
		entity.AccountSessionsSchema,
		entity.OrganizationInvitationSchema,
		entity.OrganizationSchema,
		entity.OrganizationAccountSchema,
		entity.WorkspaceSchema,
		entity.SettingsSchema,
		entity.TaskSchema,
	}

	if repo.Config.DB_TYPE == "mysql" {
		schemas = []string{
			entity.AccountSchemaMYSQL,
			entity.AccountSessionsSchemaMYSQL,
			entity.OrganizationInvitationSchemaMYSQL,
			entity.OrganizationSchemaMYSQL,
			entity.OrganizationAccountSchemaMYSQL,
			entity.WorkspaceSchemaMYSQL,
			entity.SettingsSchemaMYSQL,
			entity.TaskSchemaMYSQL,
		}
	}

	// Install all system schemas in a transaction

	tx, err := conn.BeginTx(ctx, nil)

	if err != nil {
		return eris.Wrap(err, "Install begin tx")
	}

	for _, schema := range schemas {
		_, err = tx.ExecContext(ctx, schema)

		if err != nil {
			tx.Rollback()
			return eris.Wrapf(err, "Install exec query %v", schema)
		}
	}

	// Insert settings
	query := "INSERT INTO setting (installed_version) VALUES (?)"
	args := []interface{}{
		common.Version,
	}

	_, err = tx.ExecContext(ctx, query, args...)

	if err != nil {
		tx.Rollback()
		return eris.Wrapf(err, "Install insert settings query %v", query)
	}

	// Insert default organization

	query = "INSERT INTO organization (id, name, currency, dpo_id, created_at) VALUES (?, ?, ?, ?, ?)"
	args = []interface{}{
		defaultOrganization.ID,
		defaultOrganization.Name,
		defaultOrganization.Currency,
		defaultOrganization.DataProtectionOfficerID,
		defaultOrganization.CreatedAt,
	}

	_, err = tx.ExecContext(ctx, query, args...)

	if err != nil {
		tx.Rollback()
		return eris.Wrapf(err, "Install insert default organization query %v", query)
	}

	// Insert root account

	query = "INSERT INTO `account` (id, timezone, email, hashed_password, is_root) VALUES (?, ?, ?, ?, true)"
	args = []interface{}{
		rootAccount.ID,
		rootAccount.Timezone,
		rootAccount.Email,
		rootAccount.HashedPassword,
	}

	_, err = tx.ExecContext(ctx, query, args...)

	if err != nil {
		tx.Rollback()
		return eris.Wrapf(err, "Install insert root account query %v", query)
	}

	// Add root account to default organization
	defaultWorkspacesScopes := `[{"workspace_id": "*", "scopes": ["*"]}]`

	query = "INSERT INTO organization_account (account_id, organization_id, is_owner, workspaces_scopes) VALUES (?, ?, true, ?)"
	args = []interface{}{
		rootAccount.ID,
		defaultOrganization.ID,
		defaultWorkspacesScopes,
	}

	_, err = tx.ExecContext(ctx, query, args...)

	if err != nil {
		tx.Rollback()
		return eris.Wrapf(err, "Install add root account to default organization query %v", query)
	}

	if err := tx.Commit(); err != nil {
		return eris.Wrap(err, "Install commit tx")
	}

	return nil
}
