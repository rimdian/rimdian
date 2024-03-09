package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (repo *RepositoryImpl) CreateOrganization(ctx context.Context, organization *entity.Organization, tx *sql.Tx) error {

	query, args, err := sq.Insert("organization").
		Columns("id", "name", "currency", "dpo_id").
		Values(organization.ID, organization.Name, organization.Currency, organization.DataProtectionOfficerID).
		ToSql()

	if err != nil {
		return eris.Wrapf(err, "CreateOrganization build query for organization %+v\n", *organization)
	}

	_, err = tx.ExecContext(ctx, query, args...)

	if err != nil {
		return eris.Wrapf(err, "CreateOrganization exec query %v", query)
	}

	return nil
}

// get organization from id
func (repo *RepositoryImpl) GetOrganization(ctx context.Context, organizationID string) (organization *entity.Organization, err error) {

	conn, err := repo.GetSystemConnection(ctx)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	organization = &entity.Organization{}

	err = sqlscan.Get(ctx, conn, organization, "SELECT * FROM organization WHERE id = ? LIMIT 1", organizationID)

	if err != nil {
		if sqlscan.NotFound(err) {
			return nil, eris.Wrap(entity.ErrOrganizationNotFound, "GetOrganization")
		} else {

			return nil, eris.Wrap(err, "GetOrganization error")
		}
	}

	return organization, nil
}

// check if account belongs to organization and not deleted
func (repo *RepositoryImpl) IsAccountOfOrganization(ctx context.Context, accountId string, organizationId string, shouldBeOwner bool) (isAccount bool, err error) {

	conn, err := repo.GetSystemConnection(ctx)

	if err != nil {
		return false, err
	}

	defer conn.Close()

	isOwnerCondition := ""

	if shouldBeOwner {
		isOwnerCondition = "AND oa.is_owner IS TRUE"
	}

	query := fmt.Sprintf(`SELECT a.id FROM account a 
		INNER JOIN organization_account oa ON a.id = oa.account_id 
		WHERE a.id = ? 
		AND oa.organization_id = ? 
		%v
		AND oa.deactivated_at IS NULL 
		LIMIT 1`, isOwnerCondition)

	row := conn.QueryRowContext(ctx, query, accountId, organizationId)

	var id string

	if err := row.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}

		return false, eris.Wrapf(err, "IsAccountOfOrganization query %v", query)
	}

	return true, nil
}

func (repo *RepositoryImpl) UpdateOrganizationProfile(ctx context.Context, organization *entity.Organization) error {

	conn, err := repo.GetSystemConnection(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	query, args, err := sq.Update("organization").
		Set("name", organization.Name).
		Set("currency", organization.Currency).
		Set("dpo_id", organization.DataProtectionOfficerID).
		Where(sq.Eq{"id": organization.ID}).
		ToSql()

	if err != nil {
		return eris.Wrapf(err, "UpdateOrganizationProfile build query for organization %+v\n", *organization)
	}

	_, err = conn.ExecContext(ctx, query, args...)

	if err != nil {
		return eris.Wrapf(err, "UpdateOrganizationProfile exec query %v", query)
	}

	return nil
}

func (repo *RepositoryImpl) ListAccountsForOrganization(ctx context.Context, organizationID string) (accounts []*entity.AccountWithOrganizationRole, err error) {

	conn, err := repo.GetSystemConnection(ctx)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	accounts = []*entity.AccountWithOrganizationRole{}

	query := `SELECT a.*, oa.is_owner, oa.workspaces_scopes, oa.deactivated_at FROM account a 
		INNER JOIN organization_account oa ON a.id = oa.account_id 
		WHERE oa.organization_id = ?`

	if err := sqlscan.Select(ctx, conn, &accounts, query, organizationID); err != nil {
		return nil, eris.Wrapf(err, "ListAccountsForOrganization query %v, %v", query, organizationID)
	}

	return accounts, nil
}

func (repo *RepositoryImpl) ListOrganizationsForAccount(ctx context.Context, accountID string) (organizations []*entity.Organization, err error) {

	conn, err := repo.GetSystemConnection(ctx)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	query := `SELECT o.*, oa.is_owner as account_is_owner, oa.workspaces_scopes as account_workspaces_scopes FROM organization AS o 
	INNER JOIN organization_account AS oa ON o.id = oa.organization_id 
	WHERE o.deleted_at IS NULL AND oa.account_id = ? AND oa.deactivated_at IS NULL`

	if err := sqlscan.Select(ctx, conn, &organizations, query, accountID); err != nil {
		return nil, eris.Wrapf(err, "ListOrganizationsForAccount query %v", query)
	}

	return organizations, nil
}

func (repo *RepositoryImpl) DeactivateOrganizationAccount(ctx context.Context, accountID string, deactivateAccountID string, organizationID string) error {

	conn, err := repo.GetSystemConnection(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return eris.Wrap(err, "DeactivateOrganizationAccount")
	}

	result, err := tx.ExecContext(ctx, "UPDATE organization_account SET deactivated_at = ? WHERE account_id = ? AND organization_id = ? AND deactivated_at IS NULL LIMIT 1", time.Now().UTC(), deactivateAccountID, organizationID)

	if err != nil {
		tx.Rollback()
		return eris.Wrap(err, "DeactivateOrganizationAccount")
	}

	updated, _ := result.RowsAffected()

	if updated != 1 {
		tx.Rollback()
		return eris.Wrap(eris.New("Failed to deactivate organization account"), "DeactivateOrganizationAccount")
	}

	if err := tx.Commit(); err != nil {
		return eris.Wrap(err, "DeactivateOrganizationAccount")
	}

	return nil
}

func (repo *RepositoryImpl) TransferOrganizationOwnsership(ctx context.Context, accountID string, toAccountID string, organizationID string) error {

	conn, err := repo.GetSystemConnection(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return eris.Wrap(err, "TransferOrganizationOwnsership")
	}

	result, err := tx.ExecContext(ctx, "UPDATE organization_account SET is_owner = TRUE WHERE account_id = ? AND organization_id = ? AND deactivated_at IS NULL LIMIT 1", toAccountID, organizationID)

	if err != nil {
		tx.Rollback()
		return eris.Wrap(err, "TransferOrganizationOwnsership")
	}

	updated, _ := result.RowsAffected()

	if updated != 1 {
		tx.Rollback()
		return eris.Wrap(eris.New("Failed to update new organization owner"), "TransferOrganizationOwnsership")
	}

	// remove previous ownership
	result, err = tx.ExecContext(ctx, "UPDATE organization_account SET is_owner = FALSE WHERE account_id = ? AND organization_id = ? AND deactivated_at IS NULL LIMIT 1", accountID, organizationID)

	if err != nil {
		tx.Rollback()
		return eris.Wrap(err, "TransferOrganizationOwnsership")
	}

	updated, _ = result.RowsAffected()

	if updated != 1 {
		tx.Rollback()
		return eris.Wrap(eris.New("Failed to update previous organization owner"), "TransferOrganizationOwnsership")
	}

	if err := tx.Commit(); err != nil {
		return eris.Wrap(err, "TransferOrganizationOwnsership")
	}

	return nil
}
