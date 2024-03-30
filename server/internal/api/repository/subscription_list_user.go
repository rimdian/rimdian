package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (repo *RepositoryImpl) GetUsersNotInSubscriptionList(ctx context.Context, workspaceID string, listID string, offset int64, limit int64) (users []*dto.UserToImportToSubscriptionList, err error) {

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return
	}

	defer conn.Close()

	users = []*dto.UserToImportToSubscriptionList{}

	sql, args, err := squirrel.Select(
		"user.external_id",
		"user.is_authenticated",
		"user.created_at",
	).From("user").
		LeftJoin("subscription_list_user ON user.id = subscription_list_user.user_id AND subscription_list_user.subscription_list_id = ?", listID).
		Where("subscription_list_user.user_id IS NULL").
		Offset(uint64(offset)).
		Limit(uint64(limit)).
		ToSql()

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

	err = sqlscan.Get(ctx, tx, &subscription, "SELECT * FROM subscription_list_user WHERE subscription_list_id = ? AND user_id = ?", listID, userID)

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
