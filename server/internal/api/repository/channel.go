package repository

import (
	"context"
	"database/sql"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (repo *RepositoryImpl) CreateChannel(ctx context.Context, workspace *entity.Workspace, channel *entity.Channel) error {

	conn, err := repo.GetWorkspaceConnection(ctx, workspace.ID)

	if err != nil {
		return err
	}

	defer conn.Close()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return eris.Wrap(err, "CreateChannel")
	}

	// update sessions+postviews mapping

	if err := SetConversionsAndTouchpointsForChannel(ctx, channel, tx); err != nil {
		tx.Rollback()
		return eris.Wrap(err, "CreateChannel")
	}

	if err := tx.Commit(); err != nil {
		return eris.Wrap(err, "CreateChannel")
	}

	// update workspace
	if err := repo.UpdateWorkspace(context.Background(), workspace, nil); err != nil {
		tx.Rollback()
		return eris.Wrap(err, "CreateChannel")
	}

	return nil
}

func (repo *RepositoryImpl) UpdateChannel(ctx context.Context, workspace *entity.Workspace, channel *entity.Channel) error {

	conn, err := repo.GetWorkspaceConnection(ctx, workspace.ID)

	if err != nil {
		return err
	}

	defer conn.Close()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return eris.Wrap(err, "UpdateChannel")
	}

	// reset + set sessions for this channel origins
	if err := ResetConversionsAndTouchpointsForChannel(ctx, channel.ID, tx); err != nil {
		tx.Rollback()
		return eris.Wrap(err, "UpdateChannel")
	}

	if err := SetConversionsAndTouchpointsForChannel(ctx, channel, tx); err != nil {
		tx.Rollback()
		return eris.Wrap(err, "UpdateChannel")
	}

	if err := tx.Commit(); err != nil {
		return eris.Wrap(err, "UpdateChannel")
	}

	// update workspace
	if err := repo.UpdateWorkspace(context.Background(), workspace, nil); err != nil {
		tx.Rollback()
		return eris.Wrap(err, "UpdateChannel")
	}

	return nil
}

func (repo *RepositoryImpl) DeleteChannel(ctx context.Context, workspace *entity.Workspace, deletedChannelID string) error {

	conn, err := repo.GetWorkspaceConnection(ctx, workspace.ID)

	if err != nil {
		return err
	}

	defer conn.Close()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return eris.Wrap(err, "DeleteChannel")
	}

	if err := ResetConversionsAndTouchpointsForChannel(ctx, deletedChannelID, tx); err != nil {
		tx.Rollback()
		return eris.Wrap(err, "DeleteChannel")
	}

	if err := tx.Commit(); err != nil {
		return eris.Wrap(err, "DeleteChannel")
	}

	// update workspace
	if err := repo.UpdateWorkspace(context.Background(), workspace, nil); err != nil {
		tx.Rollback()
		return eris.Wrap(err, "DeleteChannel")
	}

	return nil
}

func ResetConversionsAndTouchpointsForChannel(ctx context.Context, channelID string, tx *sql.Tx) error {

	// reset order.attribution_updated_at BEFORE update session/postview channel_id
	// use o.user_id = s.user_id to simplify query plan of singlestore sharding
	query := "UPDATE `order` o, `session` s SET o.attribution_updated_at = NULL WHERE o.user_id = s.user_id AND s.conversion_id = o.id AND s.conversion_type = 'order' AND s.channel_id = ?"

	if _, err := tx.ExecContext(ctx, query, channelID); err != nil {
		tx.Rollback()
		return eris.Wrap(err, "ResetConversionsAndTouchpointsForChannel")
	}

	query = "UPDATE `order` o, `postview` p SET o.attribution_updated_at = NULL WHERE o.user_id = p.user_id AND p.conversion_id = o.id AND p.conversion_type = 'order' AND p.channel_id = ?"

	if _, err := tx.ExecContext(ctx, query, channelID); err != nil {
		tx.Rollback()
		return eris.Wrap(err, "ResetConversionsAndTouchpointsForChannel")
	}

	// Reset session+postviews channel_id
	query = "UPDATE `session` SET channel_id = 'not-mapped', channel_group_id = 'not-mapped' WHERE channel_id = ?"

	if _, err := tx.ExecContext(ctx, query, channelID); err != nil {
		tx.Rollback()
		return eris.Wrap(err, "ResetConversionsAndTouchpointsForChannel")
	}

	query = "UPDATE `postview` SET channel_id = 'not-mapped', channel_group_id = 'not-mapped' WHERE channel_id = ?"

	if _, err := tx.ExecContext(ctx, query, channelID); err != nil {
		tx.Rollback()
		return eris.Wrap(err, "ResetConversionsAndTouchpointsForChannel")
	}

	return nil
}

func SetConversionsAndTouchpointsForChannel(ctx context.Context, channel *entity.Channel, tx *sql.Tx) (err error) {
	// Set session+postview with channel_id for all matching origins
	for _, origin := range channel.Origins {

		args := []interface{}{
			channel.ID,
			channel.GroupID,
			origin.ID,
		}

		query := "UPDATE `session` SET channel_id = ?, channel_group_id = ? WHERE channel_origin_id = ?"

		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			tx.Rollback()
			return eris.Wrap(err, "SetConversionsAndTouchpointsForChannel")
		}

		query = "UPDATE `postview` SET channel_id = ?, channel_group_id = ? WHERE channel_origin_id = ?"

		if _, err = tx.ExecContext(ctx, query, args...); err != nil {
			tx.Rollback()
			return eris.Wrap(err, "SetConversionsAndTouchpointsForChannel")
		}

		// reset order.attribution_updated_at, after updating session/postview channel_id
		// use o.user_id = s.user_id to simplify query plan of singlestore sharding
		query = "UPDATE `order` o, `session` s SET o.attribution_updated_at = NULL WHERE o.user_id = s.user_id AND s.conversion_id = o.id AND s.conversion_type = 'order' AND channel_origin_id = ?"

		if _, err = tx.ExecContext(ctx, query, origin.ID); err != nil {
			tx.Rollback()
			return eris.Wrap(err, "SetConversionsAndTouchpointsForChannel")
		}
	}

	return nil
}
