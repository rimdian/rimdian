package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rimdian/rimdian/internal/api/common"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (svc *ServiceImpl) AppGet(ctx context.Context, accountID string, params *dto.AppGetParams) (app *entity.App, code int, err error) {

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, params.WorkspaceID, accountID)

	if err != nil {
		return nil, code, eris.Wrap(err, "AppGet")
	}

	// get app
	app, err = svc.Repo.GetApp(ctx, params.WorkspaceID, params.ID)

	if err != nil {
		if sqlscan.NotFound(err) {
			return nil, 400, eris.New("AppGet: app not found")
		}
		return nil, 500, eris.Wrap(err, "AppGet")
	}

	// get account for timezone
	account, err := svc.Repo.GetAccountFromID(ctx, accountID)

	if err != nil {
		return nil, 500, eris.Wrap(err, "AppGet")
	}

	if err = app.EnrichMetadatas(svc.Config, workspace.Currency, params.WorkspaceID, account.ID, account.Timezone, true); err != nil {
		return nil, 500, eris.Wrap(err, "AppList")
	}

	return app, 200, nil
}

func (svc *ServiceImpl) AppFromToken(ctx context.Context, params *dto.AppFromTokenParams) (result *dto.AppFromToken, code int, err error) {

	result = &dto.AppFromToken{}

	token, err := jwt.Parse(params.Token, func(token *jwt.Token) (interface{}, error) {

		// extract claims from token
		claims, ok := token.Claims.(jwt.MapClaims)

		if !ok {
			return nil, eris.New("AppFromToken: invalid claims")
		}

		workspaceID, ok := claims["workspace_id"].(string)
		if !ok {
			return nil, eris.New("AppFromToken: invalid workspace_id")
		}

		appID, ok := claims["app_id"].(string)
		if !ok {
			return nil, eris.New("AppFromToken: invalid app_id")
		}

		timezone, ok := claims["account_timezone"].(string)
		if !ok {
			return nil, eris.New("AppFromToken: invalid account_timezone")
		}

		accountID, ok := claims["account_id"].(string)
		if !ok {
			return nil, eris.New("AppFromToken: invalid account°id")
		}

		// fetch app from DB
		result.App, err = svc.Repo.GetApp(ctx, workspaceID, appID)

		if err != nil {
			if sqlscan.NotFound(err) {
				return nil, err
			}
			return nil, eris.Wrap(err, "AppFromToken")
		}

		// fetch workspace
		workspace, err := svc.Repo.GetWorkspace(ctx, workspaceID)

		if err != nil {
			return nil, eris.Wrap(err, "AppFromToken")
		}

		if err = result.App.EnrichMetadatas(svc.Config, workspace.Currency, workspaceID, accountID, timezone, false); err != nil {
			return nil, eris.Wrap(err, "AppFromToken")
		}

		return []byte(result.App.SecretKey), nil
	})

	if err != nil {
		return nil, 400, eris.Wrap(err, "AppFromToken parse jwt")
	}

	if !token.Valid {
		return nil, 400, eris.New("AppFromToken: invalid token")
	}

	return result, 200, nil
}

func (svc *ServiceImpl) AppMutateState(ctx context.Context, accountID string, params *dto.AppMutateState) (code int, err error) {

	if err := params.Validate(); err != nil {
		return 400, eris.Wrap(err, "AppMutateState")
	}

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, params.WorkspaceID, accountID)

	if err != nil {
		return code, eris.Wrap(err, "AppMutateState")
	}

	// save app state
	code, err = svc.Repo.RunInTransactionForWorkspace(ctx, workspace.ID, func(ctx context.Context, tx *sql.Tx) (code int, err error) {

		// get app
		app, err := svc.Repo.GetApp(ctx, params.WorkspaceID, params.ID)

		if err != nil {
			if sqlscan.NotFound(err) {
				return 400, err
			}
			return 500, eris.Wrap(err, "AppMutateState")
		}

		// apply mutations
		app.ApplyMutations(params.Mutations)
		app.UpdatedAt = time.Now()

		if err := svc.Repo.UpdateApp(ctx, app, tx); err != nil {
			if sqlscan.NotFound(err) {
				return 400, err
			}
			return 500, eris.Wrap(err, "AppMutateState")
		}

		return 200, nil
	})

	if err != nil {
		return code, eris.Wrap(err, "AppMutateState")
	}

	return 200, nil
}

func (svc *ServiceImpl) AppExecQuery(ctx context.Context, accountID string, params *dto.AppExecQuery) (result *dto.AppExecQueryResult, code int, err error) {
	result = &dto.AppExecQueryResult{
		Rows: []map[string]interface{}{},
	}

	if err := params.Validate(); err != nil {
		return result, 400, eris.Wrap(err, "AppExecQuery")
	}

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, params.WorkspaceID, accountID)

	if err != nil {
		return result, code, eris.Wrap(err, "AppExecQuery")
	}

	// get app
	app, err := svc.Repo.GetApp(ctx, params.WorkspaceID, params.AppID)

	if err != nil {
		if sqlscan.NotFound(err) {
			return result, 400, err
		}
		return result, 500, eris.Wrap(err, "AppExecQuery")
	}

	query := ""

	if app.Manifest.SQLQueries == nil {
		return result, 400, eris.New("AppExecQuery: app has no SQL queries")
	}

	// find query
	for _, q := range app.Manifest.SQLQueries {
		if q.ID == params.QueryID {
			query = q.Query // use predefined query
			if q.Query == entity.SQLFullAccess {
				query = params.Query // use custom query
			}
			break
		}
	}

	if query == "" {
		return result, 400, eris.New("AppExecQuery: query not found")
	}

	// start timer
	start := time.Now()

	// svc.Logger.Printf("AppExecQuery: %v", query)

	output, err := svc.DoDBSelect(workspace.ID, query, params.Args)

	if err != nil {
		svc.Logger.Printf("error AppExecQuery output: %v, %v", string(output), err)
		return nil, 500, eris.Errorf("AppExecQuery %v", output)
	}

	if output == nil {
		return result, 200, nil
	}

	// check if  output is an error
	if strings.HasPrefix(string(output), "error") {
		return nil, 400, eris.Errorf("AppExecQuery %v", string(output))
	}

	// decode output
	if err = json.Unmarshal(output, &result.Rows); err != nil {
		svc.Logger.Printf("error AppExecQuery output: %v", string(output))
		return nil, 400, eris.Wrap(err, "AppExecQuery")
	}

	// stop timer and get ms
	result.TookMs = time.Since(start).Milliseconds()

	return result, 200, nil
}

func (svc *ServiceImpl) AppActivate(ctx context.Context, accountID string, params *dto.AppActivate) (code int, err error) {

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, params.WorkspaceID, accountID)

	if err != nil {
		return code, eris.Wrap(err, "AppActivate")
	}

	// get app
	app, err := svc.Repo.GetApp(ctx, params.WorkspaceID, params.ID)

	if err != nil {
		if sqlscan.NotFound(err) {
			return 400, err
		}
		return 500, eris.Wrap(err, "AppActivate")
	}

	// check that app is stopped
	if app.Status != entity.AppStatusInit {
		return 400, eris.Errorf("AppActivate: app status is already: %v", app.Status)
	}

	app.Status = entity.AppStatusActive
	app.UpdatedAt = time.Now()

	// save app state
	code, err = svc.Repo.RunInTransactionForWorkspace(ctx, workspace.ID, func(ctx context.Context, tx *sql.Tx) (code int, err error) {

		if err := svc.Repo.UpdateApp(ctx, app, tx); err != nil {
			if sqlscan.NotFound(err) {
				return 400, err
			}
			return 500, eris.Wrap(err, "AppActivate")
		}

		// activate tasks and hooks
		code, err = svc.Repo.RunInTransactionForSystem(ctx, func(ctx context.Context, tx *sql.Tx) (code int, err error) {

			if err = svc.Repo.ActivateAppTasks(ctx, workspace.ID, app.ID, tx); err != nil {
				return 500, eris.Wrap(err, "AppActivate")
			}

			// activate hooks
			for _, hook := range workspace.DataHooks {
				if hook.AppID == app.ID {
					hook.Enabled = true
				}
			}

			if err := svc.Repo.UpdateWorkspace(ctx, workspace, tx); err != nil {
				return 500, eris.Wrap(err, "AppActivate")
			}

			return 200, nil
		})

		if err != nil {
			return code, eris.Wrap(err, "AppActivate")
		}

		return 200, nil
	})

	if err != nil {
		return code, eris.Wrap(err, "AppActivate")
	}

	return 200, nil
}

func (svc *ServiceImpl) AppDelete(ctx context.Context, accountID string, params *dto.AppDelete) (code int, err error) {

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, params.WorkspaceID, accountID)

	if err != nil {
		return code, eris.Wrap(err, "AppDelete")
	}

	// get app
	app, err := svc.Repo.GetApp(ctx, params.WorkspaceID, params.ID)

	if err != nil {
		if sqlscan.NotFound(err) {
			return 400, err
		}
		return 500, eris.Wrap(err, "AppDelete")
	}

	// check that app is stopped or initializing
	if app.Status != entity.AppStatusStopped && app.Status != entity.AppStatusInit {
		return 400, eris.New("AppDelete: app is not stopped or initializing")
	}

	app.Status = entity.AppStatusDeleted
	app.DeletedAt = entity.TimePtr(time.Now())

	code, err = svc.Repo.RunInTransactionForWorkspace(ctx, workspace.ID, func(ctx context.Context, tx *sql.Tx) (code int, err error) {
		if err := svc.Repo.DeleteApp(ctx, app.ID, tx); err != nil {
			return 500, eris.Wrap(err, "AppDelete")
		}
		return 200, nil
	})

	if err != nil {
		return code, eris.Wrap(err, "AppDelete")
	}

	// delete tasks, update workspace
	code, err = svc.Repo.RunInTransactionForSystem(ctx, func(ctx context.Context, tx *sql.Tx) (code int, err error) {

		// remove app from installed_app
		remainingApps := entity.InstalledApps{}
		for _, a := range workspace.InstalledApps {
			if a.ID != params.ID {
				remainingApps = append(remainingApps, a)
			}
		}
		workspace.InstalledApps = remainingApps

		// remove hooks
		remainingHooks := entity.DataHooks{}
		for _, hook := range workspace.DataHooks {
			if hook.AppID != params.ID {
				remainingHooks = append(remainingHooks, hook)
			}
		}
		workspace.DataHooks = remainingHooks

		if err := svc.Repo.UpdateWorkspace(ctx, workspace, tx); err != nil {
			return 500, eris.Wrap(err, "AppInstall")
		}

		if err := svc.Repo.DeleteAppTasks(ctx, workspace.ID, app.ID, tx); err != nil {
			return 500, eris.Wrap(err, "AppDelete")
		}
		return 200, nil
	})

	if err != nil {
		return code, eris.Wrap(err, "AppDelete")
	}

	// delete custom tables if erase data
	if app.Manifest.AppTables != nil && len(app.Manifest.AppTables) > 0 {

		// extract custom tables from app manifest
		for _, table := range app.Manifest.AppTables {
			if err := svc.Repo.DeleteTable(ctx, workspace.ID, table.Name); err != nil {
				return 500, eris.Wrap(err, "AppDelete")
			}

		}

		if err != nil {
			return 500, eris.Wrap(err, "AppDelete")
		}
	}

	// delete augmented columns
	if app.Manifest.ExtraColumns != nil && len(app.Manifest.ExtraColumns) > 0 {

		// delete augmented tables columns from app manifest
		for _, augTable := range app.Manifest.ExtraColumns {
			for _, col := range augTable.Columns {
				if err := svc.Repo.DeleteColumn(ctx, workspace, augTable.Kind, col.Name); err != nil {
					return 500, eris.Wrap(err, "AppDelete")
				}
			}
		}

		if err != nil {
			return 500, eris.Wrap(err, "AppDelete")
		}
	}

	return 200, nil
}

func (svc *ServiceImpl) AppStop(ctx context.Context, accountID string, params *dto.AppDelete) (app *entity.App, code int, err error) {

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, params.WorkspaceID, accountID)

	if err != nil {
		return nil, code, eris.Wrap(err, "AppStop")
	}

	// get app
	app, err = svc.Repo.GetApp(ctx, params.WorkspaceID, params.ID)

	if err != nil {
		if sqlscan.NotFound(err) {
			return nil, 400, err
		}
		return nil, 500, eris.Wrap(err, "AppStop")
	}

	// stop app
	app.Status = entity.AppStatusStopped

	code, err = svc.Repo.RunInTransactionForWorkspace(ctx, workspace.ID, func(ctx context.Context, tx *sql.Tx) (code int, err error) {

		if err := svc.Repo.StopTaskExecsForApp(ctx, app.ID, tx); err != nil {
			return 500, eris.Wrap(err, "AppStop")
		}

		if err := svc.Repo.UpdateApp(ctx, app, tx); err != nil {
			return 500, eris.Wrap(err, "AppStop")
		}

		return 200, nil
	})

	if err != nil {
		return nil, code, eris.Wrap(err, "AppStop")
	}

	// stop tasks and hooks
	code, err = svc.Repo.RunInTransactionForSystem(ctx, func(ctx context.Context, tx *sql.Tx) (code int, err error) {

		if err := svc.Repo.StopAppTasks(ctx, workspace.ID, app.ID, tx); err != nil {
			return 500, eris.Wrap(err, "AppStop")
		}

		// stop hooks
		for _, hook := range workspace.DataHooks {
			if hook.AppID == app.ID {
				hook.Enabled = false
			}
		}

		if err := svc.Repo.UpdateWorkspace(ctx, workspace, tx); err != nil {
			return 500, eris.Wrap(err, "AppStop")
		}

		return 200, nil
	})

	if err != nil {
		return nil, code, eris.Wrap(err, "AppStop")
	}

	return app, 200, nil
}

func (svc *ServiceImpl) AppInstall(ctx context.Context, accountID string, params *dto.AppInstall) (installedApp *entity.App, code int, err error) {

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, params.WorkspaceID, accountID)

	if err != nil {
		return nil, code, eris.Wrap(err, "AppInstall")
	}

	if params.Manifest == nil {
		return nil, 400, eris.New("AppInstall: manifest is required")
	}

	// rewrite the api endpoint with the env var
	if strings.Contains(params.Manifest.WebhookEndpoint, "API_ENDPOINT") {
		params.Manifest.WebhookEndpoint = strings.ReplaceAll(params.Manifest.WebhookEndpoint, "API_ENDPOINT", svc.Config.API_ENDPOINT)
	}

	// validate manifest
	if err := params.Manifest.Validate(workspace.InstalledApps, params.Reinstall); err != nil {
		return nil, 400, eris.Wrap(err, "AppInstall")
	}

	now := time.Now().UTC()

	installedApp = &entity.App{
		ID:       params.Manifest.ID,
		Name:     params.Manifest.Name,
		Status:   entity.AppStatusInit,
		State:    entity.MapOfInterfaces{},
		Manifest: *params.Manifest,
		// EncryptedSecretKey: encryptedSecretKey,
		IsNative:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// native apps use the secret key from the config
	switch params.Manifest.ID {

	case "appx_metacapi", "appx_meta", "appx_googleads", "appx_woocommerce", "appx_shopify", "appx_googlecm360", "appx_admo", "appx_affilae":
		params.SecretKey = &svc.Config.SECRET_KEY

	// test
	case entity.AppManifestTest.ID:
		params.SecretKey = &svc.Config.SECRET_KEY
		installedApp.IsNative = true
		// init state with somes values for testing
		installedApp.State["test_string"] = "abc"
		installedApp.State["test_float64"] = 123.456
		installedApp.State["test_boolean"] = true
		installedApp.State["test_date"] = time.Now()

	default:
	}

	if params.SecretKey == nil {
		return nil, 400, eris.New("AppInstall: secret_key is required")
	}

	encrypted, err := common.EncryptString(strings.TrimSpace(*params.SecretKey), svc.Config.SECRET_KEY)

	if err != nil {
		return nil, 500, eris.Wrap(err, "AppInstall")
	}

	installedApp.EncryptedSecretKey = encrypted

	code, err = svc.Repo.RunInTransactionForWorkspace(ctx, workspace.ID, func(ctx context.Context, tx *sql.Tx) (code int, err error) {

		// delete previous app if we reinstall it
		if params.Reinstall {
			if err := svc.Repo.DeleteApp(ctx, installedApp.Manifest.ID, tx); err != nil {
				return 500, eris.Wrap(err, "AppInstall")
			}
		}

		if err := svc.Repo.InsertApp(ctx, installedApp, tx); err != nil {
			return 500, eris.Wrap(err, "AppInstall")
		}

		// install extra columns
		if installedApp.Manifest.ExtraColumns != nil && len(installedApp.Manifest.ExtraColumns) > 0 {
			for _, extraColumnsManifest := range installedApp.Manifest.ExtraColumns {

				if extraColumnsManifest.Columns != nil && len(extraColumnsManifest.Columns) > 0 {
					for _, col := range extraColumnsManifest.Columns {

						if err := svc.Repo.AddColumn(ctx, workspace, extraColumnsManifest.Kind, col); err != nil {
							return 500, eris.Wrap(err, "AppInstall")
						}
					}
				}
			}
		}

		// install custom tables
		if installedApp.Manifest.AppTables != nil && len(installedApp.Manifest.AppTables) > 0 {
			for _, tableManifest := range installedApp.Manifest.AppTables {

				if err := svc.Repo.CreateTable(ctx, workspace, tableManifest); err != nil {
					return 500, eris.Wrap(err, "AppInstall")
				}
			}
		}

		// update system DB
		// update workspace and install tasks
		code, err = svc.Repo.RunInTransactionForSystem(ctx, func(ctx context.Context, tx *sql.Tx) (code int, err error) {

			if params.Reinstall {
				// delete eventually existing tasks
				if err := svc.Repo.DeleteAppTasks(ctx, workspace.ID, installedApp.ID, tx); err != nil {
					return 500, eris.Wrap(err, "AppInstall")
				}
			}

			if installedApp.Manifest.Tasks != nil && len(installedApp.Manifest.Tasks) > 0 {
				for _, stDefinition := range installedApp.Manifest.Tasks {

					task := &entity.Task{
						TaskManifest: *stDefinition,
						WorkspaceID:  workspace.ID,
						AppID:        installedApp.Manifest.ID,
						IsActive:     true,
						IsCron:       stDefinition.IsCron,
					}

					if installedApp.Status == entity.AppStatusInit {
						task.IsActive = false
					}

					task.ComputeNextRun()

					if err := svc.Repo.InsertTask(ctx, task, tx); err != nil {
						return 500, eris.Wrap(err, "AppInstall")
					}
				}
			}

			// append app manifest to installed apps
			workspace.InstalledApps = append(workspace.InstalledApps, &installedApp.Manifest)

			// add data_hooks
			if installedApp.Manifest.DataHooks != nil && len(installedApp.Manifest.DataHooks) > 0 {
				for _, dataHook := range installedApp.Manifest.DataHooks {

					hook := &entity.DataHook{
						ID:      dataHook.ID,
						AppID:   installedApp.Manifest.ID,
						Name:    dataHook.Name,
						On:      dataHook.On,
						For:     dataHook.For,
						Enabled: false, // will be enabled when the app is activated
					}

					if err := hook.Validate(workspace.InstalledApps); err != nil {
						return 400, eris.Wrap(err, "AppInstall")
					}

					dataHookExists := false

					// replace eventually existing hook
					for _, existingHook := range workspace.DataHooks {
						if existingHook.ID == hook.ID {
							dataHookExists = true
							existingHook = hook
							break
						}
					}

					if !dataHookExists {
						workspace.DataHooks = append(workspace.DataHooks, hook)
					}
				}
			}

			// update workspace
			if err := svc.Repo.UpdateWorkspace(ctx, workspace, tx); err != nil {
				return 500, eris.Wrap(err, "AppInstall")
			}

			return 200, nil
		})

		if err != nil {
			return code, eris.Wrap(err, "AppInstall")
		}

		return 200, nil
	})

	if err != nil {
		return nil, code, eris.Wrap(err, "AppInstall")
	}

	return installedApp, 200, nil
}

func (svc *ServiceImpl) AppList(ctx context.Context, accountID string, params *dto.AppListParams) (result *dto.AppListResult, code int, err error) {

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, params.WorkspaceID, accountID)

	if err != nil {
		return nil, code, eris.Wrap(err, "AppList")
	}

	account, err := svc.Repo.GetAccountFromID(ctx, accountID)

	if err != nil {
		return nil, 500, eris.Wrap(err, "AppList")
	}

	// fetch apps
	result = &dto.AppListResult{}

	result.Apps, err = svc.Repo.ListApps(ctx, workspace.ID)

	if err != nil {
		return nil, 500, err
	}

	// create UI JWT for each private app
	for _, app := range result.Apps {

		if app.IsNative {
			continue
		}

		if err = app.EnrichMetadatas(svc.Config, workspace.Currency, params.WorkspaceID, account.ID, account.Timezone, true); err != nil {
			return nil, 500, eris.Wrap(err, "AppList")
		}
	}

	return result, 200, nil
}
