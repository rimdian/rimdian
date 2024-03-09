package repository

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

// insert new service account
func (repo *RepositoryImpl) InsertServiceAccount(ctx context.Context, account *entity.Account, organizationID string, fromAccountID string, workspaceScopes entity.WorkspacesScopes) (err error) {

	conn, err := repo.GetSystemConnection(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, nil)

	if err != nil {
		return eris.Wrap(err, "InsertServiceAccount")
	}

	if err := repo.InsertAccount(ctx, account, organizationID, fromAccountID, tx); err != nil {
		tx.Rollback()
		return eris.Wrap(err, "InsertServiceAccount")
	}

	if err := repo.AddAccountToOrganization(ctx, account.ID, organizationID, false, &fromAccountID, workspaceScopes, tx); err != nil {
		tx.Rollback()
		return eris.Wrap(err, "InsertServiceAccount")
	}

	if err := tx.Commit(); err != nil {
		return eris.Wrap(err, "InsertServiceAccount")
	}

	return nil
}

// insert new account
func (repo *RepositoryImpl) InsertAccount(ctx context.Context, account *entity.Account, organizationID string, fromAccountID string, tx *sql.Tx) (err error) {

	// check if an account has same email
	existingCount := 0

	query := "SELECT count(*) FROM `account` WHERE email = ? LIMIT 1"

	row := tx.QueryRowContext(ctx, query, account.Email)

	if err := row.Scan(&existingCount); err != nil {
		return eris.Wrap(err, "InsertAccount")
	}

	if existingCount > 0 {
		return entity.ErrAccountEmailAlreadyUsed
	}

	query, args, err := sq.Insert("account").Columns(
		"id",
		"full_name",
		"timezone",
		"email",
		"hashed_password",
		"is_service_account",
	).Values(
		account.ID,
		account.FullName,
		account.Timezone,
		account.Email,
		account.HashedPassword,
		account.IsServiceAccount,
	).ToSql()

	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(ctx, query, args...)

	if err != nil {
		if repo.IsDuplicateEntry(err) {
			return entity.ErrAccountAlreadyExists
		}
		return eris.Wrapf(err, "InsertAccount exec query %v", query)
	}

	return nil
}

// get account from its email
func (repo *RepositoryImpl) GetAccountFromEmail(ctx context.Context, email string) (account *entity.Account, err error) {

	conn, err := repo.GetSystemConnection(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	account = &entity.Account{}

	query := "SELECT * FROM `account` WHERE email = ? LIMIT 1"

	err = sqlscan.Get(ctx, conn, account, query, email)

	if err != nil {
		if sqlscan.NotFound(err) {
			return nil, eris.Wrap(entity.ErrAccountNotFound, "GetAccountFromEmail")
		} else {

			return nil, eris.Wrapf(err, "GetAccountFromEmail exec query %v, email: %v", query, email)
		}
	}

	return account, nil
}

// get account from its id
func (repo *RepositoryImpl) GetAccountFromID(ctx context.Context, accountID string) (account *entity.Account, err error) {

	conn, err := repo.GetSystemConnection(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	account = &entity.Account{}

	query := "SELECT * FROM `account` WHERE id = ? LIMIT 1"

	err = sqlscan.Get(ctx, conn, account, query, accountID)

	if err != nil {
		if sqlscan.NotFound(err) {
			return nil, eris.Wrap(entity.ErrAccountNotFound, "GetAccountFromID")
		} else {
			return nil, eris.Wrapf(err, "GetAccountFromEmail exec query %v, accountID: %v", query, accountID)
		}
	}

	return account, nil
}

func (repo *RepositoryImpl) UpdateAccountProfile(ctx context.Context, account *entity.Account) error {

	conn, err := repo.GetSystemConnection(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	query, args, err := sq.Update("account").
		Set("full_name", account.FullName).
		Set("timezone", account.Timezone).
		Set("locale", account.Locale).
		Where(sq.Eq{"id": account.ID}).
		ToSql()

	if err != nil {
		return eris.Wrapf(err, "UpdateAccountProfile build query for account %+v\n", *account)
	}

	_, err = conn.ExecContext(ctx, query, args...)

	if err != nil {
		return eris.Wrapf(err, "UpdateAccountProfile exec query %v", query)
	}

	return nil
}

func (repo *RepositoryImpl) ResetAccountPassword(ctx context.Context, accountID string, newPassword string) error {

	conn, err := repo.GetSystemConnection(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, nil)

	if err != nil {
		return eris.Wrap(err, "ChangeAccountPassword")
	}

	// update password
	query, args, err := sq.Update("account").
		Set("hashed_password", newPassword).
		Where(sq.Eq{"id": accountID}).
		ToSql()

	if err != nil {
		tx.Rollback()
		return eris.Wrap(err, "ChangeAccountPassword")
	}

	_, err = tx.ExecContext(ctx, query, args...)

	if err != nil {
		tx.Rollback()
		return eris.Wrapf(err, "ChangeAccountPassword exec query %v", query)
	}

	// block existing sessions

	query, args, err = sq.Update("account_session").
		Set("blocked_at", time.Now().UTC()).
		Where(sq.Eq{"account_id": accountID}).
		ToSql()

	if err != nil {
		tx.Rollback()
		return eris.Wrapf(err, "ChangeAccountPassword build query")
	}

	_, err = tx.ExecContext(ctx, query, args...)

	if err != nil {
		tx.Rollback()
		return eris.Wrapf(err, "ChangeAccountPassword exec query %v", query)
	}

	if err := tx.Commit(); err != nil {
		return eris.Wrap(err, "ChangeAccountPassword")
	}

	return nil
}
