package repository

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

// add account to organization
func (repo *RepositoryImpl) AddAccountToOrganization(ctx context.Context, accountID string, organizationID string, isOwner bool, fromAccountID *string, workspaceScopes entity.WorkspacesScopes, tx *sql.Tx) error {

	query, args, err := sq.Insert("organization_account").Columns(
		"account_id",
		"organization_id",
		"is_owner",
		"from_account_id",
		"workspaces_scopes",
	).Values(
		accountID,
		organizationID,
		isOwner,
		fromAccountID,
		workspaceScopes,
	).ToSql()

	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(ctx, query, args...)

	if err != nil {
		if repo.IsDuplicateEntry(err) {
			return entity.ErrAccountAlreadyInOrganization
		}
		return eris.Wrapf(err, "AddAccountToOrganization exec query %v", query)
	}

	return nil
}
