package repository

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

// insert an account_invitation
func (repo *RepositoryImpl) UpsertOrganizationInvitation(ctx context.Context, accountInvitation *entity.OrganizationInvitation) error {

	conn, err := repo.GetSystemConnection(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	query, args, err := sq.Insert("organization_invitation").Columns(
		"email",
		"organization_id",
		"from_account_id",
		"expires_at",
		"workspaces_scopes",
	).Values(
		accountInvitation.Email,
		accountInvitation.OrganizationID,
		accountInvitation.FromAccountID,
		accountInvitation.ExpiresAt,
		accountInvitation.WorkspacesScopes,
	).ToSql()

	if err != nil {
		return eris.Wrapf(err, "UpsertOrganizationInvitation object %+v\n", accountInvitation)
	}

	_, err = conn.ExecContext(ctx, query, args...)

	if err != nil {

		if repo.IsDuplicateEntry(err) {

			// clean previous expired sessions

			query = "UPDATE organization_invitation SET expires_at = ? WHERE email = ? AND organization_id = ? AND consumed_at IS NULL LIMIT 1"

			args := []interface{}{
				accountInvitation.ExpiresAt,
				accountInvitation.Email,
				accountInvitation.OrganizationID,
			}

			result, err := conn.ExecContext(ctx, query, args...)

			if err != nil {
				return eris.Wrapf(err, "UpsertOrganizationInvitation exec query %v, %+v\n", query, args)
			}

			updated, err := result.RowsAffected()

			if err != nil {
				return eris.Wrapf(err, "UpsertOrganizationInvitation")
			}

			// log.Printf("args %+v\n", args)
			// log.Printf("updated %v", updated)
			// if no rows updated, invitation has been consumed
			if updated != 1 {
				return entity.ErrInvitationConsumedOrDeleted
			}

			return nil
		}

		return eris.Wrapf(err, "UpsertOrganizationInvitation exec query %v", query)
	}

	return nil
}

func (repo *RepositoryImpl) CancelOrganizationInvitation(ctx context.Context, organizationID string, email string) error {

	conn, err := repo.GetSystemConnection(ctx)

	if err != nil {
		return err
	}

	defer conn.Close()

	query := "DELETE FROM organization_invitation WHERE organization_id = ? AND email = ? AND consumed_at IS NULL LIMIT 1"

	result, err := conn.ExecContext(ctx, query, organizationID, email)

	if err != nil {
		return eris.Wrapf(err, "CancelOrganizationInvitation exec query %v", query)
	}

	updated, err := result.RowsAffected()

	if err != nil {
		return eris.Wrapf(err, "CancelOrganizationInvitation")
	}

	// if no rows updated, invitation has been consumed
	if updated != 1 {
		return entity.ErrInvitationConsumedOrDeleted
	}

	return nil
}

func (repo *RepositoryImpl) ListInvitationsForOrganization(ctx context.Context, organizationID string) (invitations []*entity.OrganizationInvitation, err error) {

	conn, err := repo.GetSystemConnection(ctx)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	invitations = []*entity.OrganizationInvitation{}

	query := `SELECT * FROM organization_invitation WHERE organization_id = ? AND consumed_at IS NULL`

	if err := sqlscan.Select(ctx, conn, &invitations, query, organizationID); err != nil {
		return nil, eris.Wrapf(err, "ListInvitationsForOrganization query %v", query)
	}

	return invitations, nil
}

func (repo *RepositoryImpl) GetInvitation(ctx context.Context, email string, organizationID string) (invitation *entity.OrganizationInvitation, err error) {

	conn, err := repo.GetSystemConnection(ctx)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	invitation = &entity.OrganizationInvitation{}

	query := `SELECT * FROM organization_invitation WHERE email = ? AND organization_id = ?`

	if err := sqlscan.Get(ctx, conn, invitation, query, email, organizationID); err != nil {
		return nil, eris.Wrapf(err, "GetInvitation query", query)
	}

	return invitation, nil
}

// consume invitation
func (repo *RepositoryImpl) ConsumeInvitation(ctx context.Context, accountID string, insertAccount *entity.Account, invitation *entity.OrganizationInvitation) error {

	conn, err := repo.GetSystemConnection(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return eris.Wrap(err, "ConsumeInvitation")
	}

	// insert account who just created its account while accepting invitation
	if insertAccount != nil {

		if err := repo.InsertAccount(ctx, insertAccount, invitation.OrganizationID, invitation.FromAccountID, tx); err != nil {
			tx.Rollback()
			return eris.Wrap(err, "ConsumeInvitation")
		}
	}

	// consume invitation

	query := "UPDATE organization_invitation SET consumed_at = ? WHERE email = ? AND organization_id = ? AND consumed_at IS NULL LIMIT 1"

	args := []interface{}{
		time.Now().UTC(),
		invitation.Email,
		invitation.OrganizationID,
	}

	result, err := tx.ExecContext(ctx, query, args...)

	if err != nil {
		tx.Rollback()
		return eris.Wrapf(err, "ConsumeInvitation exec query %v, %+v\n", query, args)
	}

	updated, _ := result.RowsAffected()

	if updated != 1 {
		tx.Rollback()
		return eris.New("Failed to update invitation")
	}

	// add to organization
	if err := repo.AddAccountToOrganization(ctx, accountID, invitation.OrganizationID, false, &invitation.FromAccountID, invitation.WorkspacesScopes, tx); err != nil {
		tx.Rollback()
		return eris.Wrap(err, "ConsumeInvitation")
	}

	if err := tx.Commit(); err != nil {
		return eris.Wrap(err, "ConsumeInvitation")
	}

	return nil
}
