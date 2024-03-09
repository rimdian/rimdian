package repository

import (
	"context"
	"database/sql"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (repo *RepositoryImpl) CleanAfterUserAlias(workspaceID string, fromUserExternalID string) (err error) {

	bgCtx := context.Background()

	conn, err := repo.GetWorkspaceConnection(bgCtx, workspaceID)

	if err != nil {
		return
	}

	defer conn.Close()

	fromUserID := entity.ComputeUserID(fromUserExternalID)

	if _, err = conn.ExecContext(bgCtx, "DELETE FROM `session` WHERE user_id = ?", fromUserID); err != nil {
		return err
	}
	if _, err = conn.ExecContext(bgCtx, "DELETE FROM postview WHERE user_id = ?", fromUserID); err != nil {
		return err
	}
	if _, err = conn.ExecContext(bgCtx, "DELETE FROM pageview WHERE user_id = ?", fromUserID); err != nil {
		return err
	}
	if _, err = conn.ExecContext(bgCtx, "DELETE FROM custom_event WHERE user_id = ?", fromUserID); err != nil {
		return err
	}
	if _, err = conn.ExecContext(bgCtx, "DELETE FROM cart WHERE user_id = ?", fromUserID); err != nil {
		return err
	}
	if _, err = conn.ExecContext(bgCtx, "DELETE FROM cart_item WHERE user_id = ?", fromUserID); err != nil {
		return err
	}
	if _, err = conn.ExecContext(bgCtx, "DELETE FROM `order` WHERE user_id = ?", fromUserID); err != nil {
		return err
	}
	if _, err = conn.ExecContext(bgCtx, "DELETE FROM order_item WHERE user_id = ?", fromUserID); err != nil {
		return err
	}
	// TODO: delete app table items with user_id != none
	return nil
}

func (repo *RepositoryImpl) FindUsersAliased(ctx context.Context, workspaceID string, toUserExternalID string) (aliases []*entity.UserAlias, err error) {

	var conn *sql.Conn

	conn, err = repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	aliases = []*entity.UserAlias{}

	err = sqlscan.Select(ctx, conn, &aliases, `SELECT * FROM user_alias WHERE to_user_external_id = ?`, toUserExternalID)

	if err != nil {
		return nil, eris.Wrap(err, "FindUsersAliased")
	}

	return
}

func (repo *RepositoryImpl) CreateUserAlias(ctx context.Context, fromUserExternalID string, toUserExternalID string, toUserIsAuthenticated bool, tx *sql.Tx) (err error) {

	_, err = tx.ExecContext(ctx, "INSERT INTO user_alias (from_user_external_id, to_user_external_id, to_user_is_authenticated) VALUES (?, ?, ?)", fromUserExternalID, toUserExternalID, toUserIsAuthenticated)

	if err != nil {
		if repo.IsDuplicateEntry(err) {
			return entity.ErrUserAliasAlreadyExists
		}
		return eris.Wrap(err, "CreateUserAlias")
	}

	// update eventual aliases that were pointing to the actual fromUserID. happens when:
	// A -> B
	// B -> C
	// A should now point to C

	_, err = tx.ExecContext(ctx, `UPDATE user_alias SET 
		to_user_external_id = ?, 
		to_user_is_authenticated = ? 
		WHERE to_user_external_id = ?`, toUserExternalID, toUserIsAuthenticated, fromUserExternalID)

	if err != nil {
		return eris.Wrap(err, "CreateUserAlias")
	}

	// update eventual user profiles that have been merged

	// WHERE clause comparing a NULL with string doesnt match the row, so we need to check if row is NULL before
	_, err = tx.ExecContext(ctx, `UPDATE user SET 
		is_merged = true, 
		merged_to = ?, 
		merged_at = NOW() 
		WHERE id = ? AND (merged_to IS NULL OR merged_to != ?)`, entity.ComputeUserID(toUserExternalID), entity.ComputeUserID(fromUserExternalID), entity.ComputeUserID(toUserExternalID))

	if err != nil {
		return eris.Wrap(err, "CreateUserAlias")
	}

	return nil
}

func (repo *RepositoryImpl) FindUserAlias(ctx context.Context, fromUserExternalID string, tx *sql.Tx) (aliasFound *entity.UserAlias, err error) {

	aliasFound = &entity.UserAlias{}

	err = sqlscan.Get(ctx, tx, aliasFound, `SELECT * FROM user_alias WHERE from_user_external_id = ? LIMIT 1`, fromUserExternalID)

	if err != nil {
		if sqlscan.NotFound(err) {
			return nil, nil
		} else {
			return nil, eris.Wrap(err, "FindUserAlias")
		}
	}

	return aliasFound, nil
}
