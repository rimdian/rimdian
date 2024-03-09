package repository

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

// insert an account_session on account login
func (repo *RepositoryImpl) InsertAccountSession(ctx context.Context, accountSession *entity.AccountSession) (err error) {

	conn, err := repo.GetSystemConnection(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	query, args, err := sq.Insert("account_session").Columns(
		"id",
		"account_id",
		"encrypted_refresh_token",
		"expires_at",
		"user_agent",
		"client_ip",
		"last_access_token_at",
	).Values(
		accountSession.ID,
		accountSession.AccountID,
		accountSession.EncryptedRefreshToken,
		accountSession.ExpiresAt,
		accountSession.UserAgent,
		accountSession.ClientIP,
		accountSession.LastAccessTokenAt,
	).ToSql()

	if err != nil {
		return eris.Wrapf(err, "InsertAccountSession object %+v\n", accountSession)
	}

	_, err = conn.ExecContext(ctx, query, args...)

	if err != nil {
		return eris.Wrapf(err, "InsertAccountSession exec query %v", query)
	}

	// clean previous expired sessions

	query = "DELETE FROM account_session WHERE account_id = ? AND (expires_at < NOW() OR blocked_at IS NOT NULL)"

	_, err = conn.ExecContext(ctx, query, accountSession.AccountID)

	if err != nil {
		return eris.Wrapf(err, "CleanAccountExpiredSessions exec query %v", query)
	}

	return nil
}

// updates last_access_token_at when account refreshes its token
func (repo *RepositoryImpl) UpdateAccountSessionLastAccess(ctx context.Context, accountID string, accountSessionID string, now time.Time) error {

	conn, err := repo.GetSystemConnection(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	query := "UPDATE account_session SET last_access_token_at = ? WHERE account_id = ? AND id = ?"

	_, err = conn.ExecContext(ctx, query, now, accountID, accountSessionID)

	if err != nil {
		return eris.Wrapf(err, "UpdateAccountSessionLastAccess exec query %v", query)
	}

	return nil
}

func (repo *RepositoryImpl) AccountLogout(ctx context.Context, accountID string, sessionID string) error {

	conn, err := repo.GetSystemConnection(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	query := "UPDATE account_session SET blocked_at = ? WHERE account_id = ? AND id = ?"

	now := time.Now().UTC()

	_, err = conn.ExecContext(ctx, query, now, accountID, sessionID)

	if err != nil {
		return eris.Wrapf(err, "AccountLogout exec query %v", query)
	}

	return nil
}
