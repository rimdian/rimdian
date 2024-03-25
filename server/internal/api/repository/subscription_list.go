package repository

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

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
