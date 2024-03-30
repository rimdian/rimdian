package repository

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (repo *RepositoryImpl) GetSubscriptionList(ctx context.Context, workspaceID string, listID string, tx *sql.Tx) (list *entity.SubscriptionList, err error) {

	if tx != nil {
		err = sqlscan.Get(ctx, tx, &list, "SELECT * FROM subscription_list WHERE id = ?", listID)
	} else {
		conn, errConn := repo.GetWorkspaceConnection(ctx, workspaceID)

		if errConn != nil {
			return
		}

		defer conn.Close()

		err = sqlscan.Get(ctx, conn, &list, "SELECT * FROM subscription_list WHERE id = ?", listID)
	}

	return
}

func (repo *RepositoryImpl) CreateSubscriptionList(ctx context.Context, workspaceID string, list *entity.SubscriptionList) (err error) {

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return
	}

	defer conn.Close()

	// insert list
	query, args, err := sq.Insert("subscription_list").
		Columns("id", "name", "color", "channel", "double_opt_in", "message_template_id").
		Values(list.ID, list.Name, list.Color, list.Channel, list.DoubleOptIn, list.MessageTemplateID).
		ToSql()

	if err != nil {
		err = eris.Wrapf(err, "CreateSubscriptionList insert query: %v, args: %+v", query, args)
		return
	}

	if _, err = conn.ExecContext(ctx, query, args...); err != nil {
		err = eris.Wrapf(err, "CreateSubscriptionList query: %v, args: %+v", query, args)
		return
	}

	return
}

func (repo *RepositoryImpl) ListSubscriptionLists(ctx context.Context, workspaceID string, withUsersCount bool) (lists []*entity.SubscriptionList, err error) {

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return
	}

	defer conn.Close()

	lists = []*entity.SubscriptionList{}

	queryBuilder := sq.Select("subscription_list.*").From("subscription_list")

	if withUsersCount {
		queryBuilder = queryBuilder.LeftJoin("subscription_list_user ON subscription_list.id = subscription_list_user.subscription_list_id")
		queryBuilder = queryBuilder.GroupBy("subscription_list.id")
		queryBuilder = queryBuilder.Column("COALESCE(COUNT(subscription_list_user.user_id), 0) AS users_count")
	}

	// fetch lists
	query, args, err := queryBuilder.ToSql()

	// log.Printf("query: %v, args: %+v", query, args)

	if err != nil {
		err = eris.Wrapf(err, "ListSubscriptionLists fetch query: %v, args: %+v", query, args)
		return
	}

	if err = sqlscan.Select(ctx, conn, &lists, query, args...); err != nil {
		err = eris.Wrapf(err, "ListSubscriptionLists query: %v, args: %+v", query, args)
		return
	}

	return
}
