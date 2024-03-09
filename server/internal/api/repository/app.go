package repository

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

// apps are really deleted upon reinstallation, otherwise we keep a record of them
func (repo *RepositoryImpl) DeleteApp(ctx context.Context, appID string, tx *sql.Tx) (err error) {

	query, args, err := sq.Delete("app").Where(sq.Eq{"id": appID}).ToSql()

	if err != nil {
		return eris.Wrapf(err, "DeleteApp build query for app %s\n", appID)
	}

	_, err = tx.ExecContext(ctx, query, args...)

	if err != nil {
		return eris.Wrapf(err, "DeleteApp exec query %v", query)
	}

	return
}

func (repo *RepositoryImpl) UpdateApp(ctx context.Context, app *entity.App, tx *sql.Tx) (err error) {

	q := sq.Update("app").Where(sq.Eq{"id": app.ID}).
		Set("name", app.Name).
		Set("status", app.Status).
		Set("state", app.State).
		Set("manifest", app.Manifest)

	sql, args, errSQL := q.ToSql()

	if errSQL != nil {
		err = eris.Wrapf(errSQL, "build query update app: %v\n", app)
		return
	}

	_, err = tx.ExecContext(ctx, sql, args...)

	if err != nil {
		err = eris.Wrap(err, "UpdateApp")
	}

	return
}

func (repo *RepositoryImpl) ListApps(ctx context.Context, workspaceID string) (apps []*entity.App, err error) {

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return
	}

	defer conn.Close()

	apps = []*entity.App{}

	queryBuilder := sq.Select("*").From("app")

	// fetch apps
	query, args, err := queryBuilder.ToSql()

	if err != nil {
		err = eris.Wrapf(err, "ListApps fetch query: %v, args: %+v", query, args)
		return
	}

	if err = sqlscan.Select(ctx, conn, &apps, query, args...); err != nil {
		err = eris.Wrapf(err, "ListApps query: %v, args: %+v", query, args)
		return
	}

	return
}

func (repo *RepositoryImpl) InsertApp(ctx context.Context, app *entity.App, tx *sql.Tx) (err error) {

	query, args, err := sq.Insert("app").Columns(
		"id",
		"name",
		"status",
		"state",
		"manifest",
		"is_native",
		"encrypted_secret_key",
	).Values(
		app.ID,
		app.Name,
		app.Status,
		app.State,
		app.Manifest,
		app.IsNative,
		app.EncryptedSecretKey,
	).ToSql()

	if err != nil {
		return eris.Wrapf(err, "InsertApp build query for app %+v\n", *app)
	}

	_, err = tx.ExecContext(ctx, query, args...)

	if err != nil {
		if repo.IsDuplicateEntry(err) {
			return entity.ErrAppAlreadyExists
		}
		return eris.Wrapf(err, "InsertApp exec query %v", query)
	}

	return
}

func (repo *RepositoryImpl) GetApp(ctx context.Context, workspaceID string, appID string) (app *entity.App, err error) {

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	app = &entity.App{}

	err = sqlscan.Get(ctx, conn, app, "SELECT * FROM app WHERE id = ? LIMIT 1", appID)

	if err != nil {
		return nil, eris.Wrap(err, "GetApp error")
	}

	return app, nil
}
