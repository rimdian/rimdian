package repository

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

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
		queryBuilder = queryBuilder.LeftJoin("subscribe_to_list ON subscription_list.id = subscribe_to_list.subscription_list_id")
		queryBuilder = queryBuilder.GroupBy("subscription_list.id")
		queryBuilder = queryBuilder.Column("COALESCE(COUNT(subscribe_to_list.user_id), 0) AS users_count")
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
