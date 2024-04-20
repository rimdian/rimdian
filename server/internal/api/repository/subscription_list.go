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

	list = &entity.SubscriptionList{}

	if tx != nil {
		err = sqlscan.Get(ctx, tx, &list, "SELECT * FROM subscription_list WHERE id = ?", listID)
	} else {
		conn, errConn := repo.GetWorkspaceConnection(ctx, workspaceID)

		if errConn != nil {
			return
		}

		defer conn.Close()

		err = sqlscan.Get(ctx, conn, list, "SELECT * FROM subscription_list WHERE id = ?", listID)

		if err != nil {
			err = eris.Wrapf(err, "GetSubscriptionList: %v", listID)
			return
		}
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

	query := "SELECT * FROM subscription_list"

	if withUsersCount {
		query = `
			SELECT subscription_list.*,
				COALESCE(active_users, 0) AS active_users,
				COALESCE(paused_users, 0) AS paused_users,
				COALESCE(unsubscribed_users, 0) AS unsubscribed_users
			FROM subscription_list
			LEFT JOIN (
				SELECT subscription_list_id,
				COUNT(CASE WHEN status = 1 THEN 1 END) AS active_users,
				COUNT(CASE WHEN status = 2 THEN 1 END) AS paused_users,
				COUNT(CASE WHEN status = 3 THEN 1 END) AS unsubscribed_users
				FROM subscription_list_user
				GROUP BY subscription_list_id
			) AS subscription_list_user ON subscription_list.id = subscription_list_user.subscription_list_id
		`
		// query = `SELECT
		// 	sl.*,
		// 	COUNT(CASE WHEN ul.status = 1 THEN 1 END) AS active_users,
		// 	COUNT(CASE WHEN ul.status = 2 THEN 1 END) AS paused_users,
		// 	COUNT(CASE WHEN ul.status = 3 THEN 1 END) AS unsubscribed_users
		// FROM
		// 	subscription_list sl
		// JOIN
		// 	subscription_list_user ul ON sl.id = ul.subscription_list_id
		// GROUP BY
		// 	sl.id`
	}

	if err = sqlscan.Select(ctx, conn, &lists, query); err != nil {
		err = eris.Wrapf(err, "ListSubscriptionLists query: %v", query)
		return
	}

	return
}
