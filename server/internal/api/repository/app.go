package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (repo *RepositoryImpl) RemoveSQLUser(ctx context.Context, username string) (err error) {

	conn, err := repo.GetSystemConnection(ctx)

	if err != nil {
		return
	}

	defer conn.Close()

	// drop user if exists
	log.Printf("Dropping SQL user %s\n", username)
	_, err = conn.ExecContext(ctx, fmt.Sprintf("DROP USER IF EXISTS %s;", username))

	if err != nil {
		return eris.Wrapf(err, "RemoveSQLUser drop user %s", username)
	}

	return
}

func (repo *RepositoryImpl) CreateAppSQLAccess(ctx context.Context, workspaceID string, app *entity.App) (err error) {

	conn, err := repo.GetWorkspaceConnection(ctx, workspaceID)

	if err != nil {
		return
	}

	defer conn.Close()

	password, err := app.DecryptSQLPassword(repo.Config.SECRET_KEY)

	if err != nil {
		return eris.Wrap(err, "CreateAppSQLAccess")
	}

	dbName := repo.Config.DB_PREFIX + workspaceID

	// remove user if exists
	_, err = conn.ExecContext(ctx, fmt.Sprintf("DROP USER IF EXISTS %s;", app.SQLUsername))

	if err != nil {
		return eris.Wrapf(err, "CreateAppSQLAccess drop user if exists %s", app.SQLUsername)
	}

	// create user
	sqlUser := fmt.Sprintf("CREATE USER %s IDENTIFIED BY '%s' WITH FAILED_LOGIN_ATTEMPTS = 10 PASSWORD_LOCK_TIME = 20 REQUIRE SSL;", app.SQLUsername, password)

	// mysql
	if repo.Config.DB_TYPE == "mysql" {
		sqlUser = fmt.Sprintf("CREATE USER %s IDENTIFIED BY '%s';", app.SQLUsername, password)
	}

	_, err = conn.ExecContext(ctx, sqlUser)

	if err != nil {
		return eris.Wrapf(err, "CreateAppSQLAccess create user %s", app.SQLUsername)
	}

	// grant access to database
	for _, perm := range app.Manifest.SQLAccess.TablesPermissions {

		if perm.Read {
			_, err = conn.ExecContext(ctx, fmt.Sprintf("GRANT SELECT ON %s.%s TO %s;", dbName, perm.Table, app.SQLUsername))

			if err != nil {
				return eris.Wrapf(err, "CreateAppSQLAccess grant select %s.%s to %s", dbName, perm.Table, app.SQLUsername)
			}
		}

		if perm.Write {
			_, err = conn.ExecContext(ctx, fmt.Sprintf("GRANT INSERT, UPDATE, DELETE ON %s.%s TO %s;", dbName, perm.Table, app.SQLUsername))

			if err != nil {
				return eris.Wrapf(err, "CreateAppSQLAccess grant insert, update, delete %s.%s to %s", dbName, perm.Table, app.SQLUsername)
			}
		}
	}

	// automatically allow read/writes on app tables
	for _, table := range app.Manifest.AppTables {
		_, err = conn.ExecContext(ctx, fmt.Sprintf("GRANT SELECT, INSERT, UPDATE, DELETE ON %s.%s TO %s;", dbName, table.Name, app.SQLUsername))

		if err != nil {
			return eris.Wrapf(err, "CreateAppSQLAccess grant select, insert, update, delete %s.%s to %s", dbName, table.Name, app.SQLUsername)
		}
	}

	return
}

// apps are really deleted upon reinstallation, otherwise we keep a record of them
func (repo *RepositoryImpl) DeleteApp(ctx context.Context, app *entity.App, tx *sql.Tx) (err error) {

	query, args, err := sq.Delete("app").Where(sq.Eq{"id": app.ID}).ToSql()

	if err != nil {
		return eris.Wrapf(err, "DeleteApp build query for app %s\n", app.ID)
	}

	_, err = tx.ExecContext(ctx, query, args...)

	if err != nil {
		return eris.Wrapf(err, "DeleteApp exec query %v", query)
	}

	// remove eventual app sql access
	if app.SQLUsername != "" {
		if err := repo.RemoveSQLUser(ctx, app.SQLUsername); err != nil {
			return eris.Wrap(err, "AppDelete")
		}
	}

	return
}

func (repo *RepositoryImpl) UpdateApp(ctx context.Context, workspaceID string, app *entity.App, tx *sql.Tx) (err error) {

	// handle sql access
	switch app.Status {
	// create access if missing and app is not stopped
	case entity.AppStatusInit, entity.AppStatusActive:
		// if should create access
		if app.SQLUsername == "" {

			_, err = app.GenerateSQLCredentials(workspaceID, repo.Config.SECRET_KEY)

			if err != nil {
				return eris.Wrap(err, "UpdateApp")
			}

			// create sql user
			if err = repo.CreateAppSQLAccess(ctx, workspaceID, app); err != nil {
				return eris.Wrap(err, "UpdateApp")
			}
		}

	// remove access if exists and app is stopped
	case entity.AppStatusStopped, entity.AppStatusDeleted:
		// remove old sql access
		if app.SQLUsername != "" {
			log.Printf("Removing SQL access for app %s\n", app.ID)
			if err = repo.RemoveSQLUser(ctx, app.SQLUsername); err != nil {
				return eris.Wrap(err, "UpdateApp")
			}
			app.SQLUsername = ""
			app.EncryptedSQLPassword = ""
		}

	default:
	}

	q := sq.Update("app").Where(sq.Eq{"id": app.ID}).
		Set("name", app.Name).
		Set("status", app.Status).
		Set("state", app.State).
		Set("sql_username", app.SQLUsername).
		Set("encrypted_sql_password", app.EncryptedSQLPassword).
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

func (repo *RepositoryImpl) InsertApp(ctx context.Context, workspaceID string, app *entity.App, tx *sql.Tx) (err error) {

	_, err = app.GenerateSQLCredentials(workspaceID, repo.Config.SECRET_KEY)

	if err != nil {
		return eris.Wrap(err, "InsertApp")
	}

	// create sql user
	if err = repo.CreateAppSQLAccess(ctx, workspaceID, app); err != nil {
		return eris.Wrap(err, "InsertApp")
	}

	query, args, err := sq.Insert("app").Columns(
		"id",
		"name",
		"status",
		"state",
		"manifest",
		"is_native",
		"encrypted_secret_key",
		"sql_username",
		"encrypted_sql_password",
	).Values(
		app.ID,
		app.Name,
		app.Status,
		app.State,
		app.Manifest,
		app.IsNative,
		app.EncryptedSecretKey,
		app.SQLUsername,
		app.EncryptedSQLPassword,
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
