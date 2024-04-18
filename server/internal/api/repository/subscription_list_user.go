package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Masterminds/squirrel"
	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (repo *RepositoryImpl) GetSubscriptionListUsersToMessage(ctx context.Context, workspaceID string, listIDs []string, offset int64, limit int64) (subscribers []*entity.SubscriptionListUser, err error) {

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return
	}

	defer conn.Close()

	subscribers = []*entity.SubscriptionListUser{}

	query, args, err := squirrel.Select("subscription_list_user.*, user.external_id as user_external_id, user.is_authenticated as user_is_authenticated").
		From("subscription_list_user").
		Where(sq.Eq{"subscription_list_id": listIDs}).
		Where(sq.Eq{"subscription_list_user.status": 1}).
		Join("`user` ON subscription_list_user.user_id = user.id").
		OrderBy("subscription_list_user.updated_at ASC"). // allow recent unsubscribed users to be skipped
		Offset(uint64(offset)).
		Limit(uint64(limit)).
		ToSql()

	if err != nil {
		err = eris.Wrap(err, "GetSubscriptionListUsersToMessage")
		return
	}

	err = sqlscan.Select(ctx, conn, &subscribers, query, args...)

	return
}

func (repo *RepositoryImpl) ListSubscriptionListUsers(ctx context.Context, workspaceID string, userIDs []string) (subscriptions []*entity.SubscriptionListUser, err error) {

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return
	}

	defer conn.Close()

	subscriptions = []*entity.SubscriptionListUser{}

	query, args, err := squirrel.Select("*").From("subscription_list_user").Where(sq.Eq{"user_id": userIDs}).ToSql()

	if err != nil {
		err = eris.Wrap(err, "ListUserSegments")
		return
	}

	err = sqlscan.Select(ctx, conn, &subscriptions, query, args...)

	return
}

func (repo *RepositoryImpl) GetUsersNotInSubscriptionList(ctx context.Context, workspaceID string, listID string, offset int64, limit int64, segmentID *string) (users []*dto.UserToImportToSubscriptionList, err error) {

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return
	}

	defer conn.Close()

	users = []*dto.UserToImportToSubscriptionList{}

	builder := squirrel.Select(
		"user.external_id",
		"user.is_authenticated",
		"user.created_at",
	).From("user").
		LeftJoin("subscription_list_user ON user.id = subscription_list_user.user_id AND subscription_list_user.subscription_list_id = ?", listID).
		Where("subscription_list_user.user_id IS NULL").
		Offset(uint64(offset)).
		Limit(uint64(limit))

	if segmentID != nil {
		builder = builder.LeftJoin("user_segment ON user.id = user_segment.user_id AND user_segment.segment_id = ?", *segmentID).
			Where("user_segment.user_id IS NOT NULL")
	}

	sql, args, err := builder.ToSql()

	if err != nil {
		return nil, eris.Wrap(err, "GetUsersNotInSubscriptionList")
	}

	err = sqlscan.Select(ctx, conn, &users, sql, args...)

	if err != nil {
		return nil, eris.Wrap(err, "GetUsersNotInSubscriptionList")
	}

	return
}

func (repo *RepositoryImpl) FindSubscriptionListUser(ctx context.Context, listID string, userID string, tx *sql.Tx) (subscription *entity.SubscriptionListUser, err error) {

	subscription = &entity.SubscriptionListUser{}

	err = sqlscan.Get(ctx, tx, subscription, "SELECT * FROM subscription_list_user WHERE subscription_list_id = ? AND user_id = ?", listID, userID)

	if err != nil && sqlscan.NotFound(err) {
		return nil, err
	}

	return
}

func (repo *RepositoryImpl) InsertSubscriptionListUser(ctx context.Context, subscription *entity.SubscriptionListUser, tx *sql.Tx) (err error) {

	query, args, err := squirrel.Insert("subscription_list_user").
		Columns(
			"subscription_list_id",
			"user_id",
			"status",
			"comment",
			"created_at",
			"fields_timestamp").
		Values(
			subscription.SubscriptionListID,
			subscription.UserID,
			subscription.Status,
			subscription.Comment,
			subscription.CreatedAt,
			subscription.FieldsTimestamp).
		ToSql()

	if err != nil {
		return eris.Wrap(err, "InsertSubscriptionListUser")
	}

	_, err = tx.ExecContext(ctx, query, args...)

	if err != nil {
		return eris.Wrap(err, "InsertSubscriptionListUser")
	}

	now := time.Now()
	subscription.DBCreatedAt = now
	subscription.DBUpdatedAt = now

	return
}

func (repo *RepositoryImpl) UpdateSubscriptionListUser(ctx context.Context, subscription *entity.SubscriptionListUser, tx *sql.Tx) (err error) {

	sql, args, err := squirrel.Update("subscription_list_user").
		Set("status", subscription.Status).
		Set("comment", subscription.Comment).
		Set("fields_timestamp", subscription.FieldsTimestamp).
		Where("subscription_list_id = ? AND user_id = ?", subscription.SubscriptionListID, subscription.UserID).
		ToSql()

	if err != nil {
		return eris.Wrap(err, "UpdateSubscriptionListUser")
	}

	_, err = tx.ExecContext(ctx, sql, args...)

	if err != nil {
		return eris.Wrap(err, "UpdateSubscriptionListUser")
	}

	subscription.DBUpdatedAt = time.Now()

	return
}
