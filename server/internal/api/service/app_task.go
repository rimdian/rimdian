package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
	"go.opencensus.io/trace"
)

func TaskExecUpgradeApp(ctx context.Context, pipe *TaskExecPipeline) (result *entity.TaskExecResult) {

	spanCtx, span := trace.StartSpan(ctx, "TaskExecUpgradeApp")
	defer span.End()

	result = &entity.TaskExecResult{
		// keep current state by default
		UpdatedWorkerState: pipe.TaskExec.State.Workers[pipe.TaskExecPayload.WorkerID],
	}

	select {
	case <-spanCtx.Done():
		result.SetError("task execution timeout", false)
		return
	default:
	}

	// log time taken
	startedAt := time.Now()
	defer func() {
		pipe.Logger.Printf("TaskExecUpgradeApp: workspace %s, taskExec %s, worker %d, took %s", pipe.Workspace.ID, pipe.TaskExec.ID, pipe.TaskExecPayload.WorkerID, time.Since(startedAt))
	}()

	// by default, keep current state
	mainState := pipe.TaskExec.State.Workers[0]

	bgCtx := context.Background()

	appID := ""
	if _, ok := mainState["app_id"]; ok {
		appID = mainState["app_id"].(string)
	}

	// get app
	app, err := pipe.Repository.GetApp(bgCtx, pipe.Workspace.ID, appID)

	if err != nil {
		result.SetError(err.Error(), true)
		return
	}

	// stage
	stage := ""
	if _, ok := mainState["stage"]; ok {
		stage = mainState["stage"].(string)
	}

	// migrated table suffix
	migratedTableSuffix := ""
	if _, ok := mainState["migrated_table_suffix"]; ok {
		migratedTableSuffix = mainState["migrated_table_suffix"].(string)
	}
	if migratedTableSuffix == "" {
		// _YYYYMMDD_HHMMSS
		migratedTableSuffix = fmt.Sprintf("_%s_%s", time.Now().Format("20060102"), time.Now().Format("150405"))
		mainState["migrated_table_suffix"] = migratedTableSuffix
	}

	// manifest
	var newManifest *entity.AppManifest

	newManifestString := ""
	if _, ok := mainState["new_manifest"]; ok {
		newManifestString = mainState["new_manifest"].(string)
	}

	if newManifestString != "" {
		if err := json.Unmarshal([]byte(newManifestString), newManifest); err != nil {
			result.SetError(err.Error(), true)
			return
		}
	}

	if newManifest == nil {
		result.SetError("new_nmanifest is required", true)
		return
	}

	if stage == "" {
		stage = "validate"
	}

	var extraColumnsDiff *entity.ExtraColumnsManifestDiff
	var appTablesDiff *entity.AppTablesManifestDiff

	switch stage {
	case "validate":
		// validate new manifest
		if err := newManifest.Validate(nil, false); err != nil {
			result.SetError(err.Error(), true)
			return
		}

		// app name should be the same
		if newManifest.Name != app.Manifest.Name {
			result.SetError("new manifest name should be the same as current manifest name", true)
			return
		}

		// new manifest version should be greater than current version
		if newManifest.Version <= app.Manifest.Version {
			result.SetError("new manifest version should be greater than current version", true)
			return
		}

		// verify extra columns don't get errors
		if _, err := entity.DiffExtraColumns(app.Manifest.ExtraColumns, newManifest.ExtraColumns); err != nil {
			result.SetError(err.Error(), true)
			return
		}

		extraColumnsDiff, err = entity.DiffExtraColumns(app.Manifest.ExtraColumns, newManifest.ExtraColumns)

		if err != nil {
			result.SetError(err.Error(), true)
			return
		}

		appTablesDiff, err = entity.DiffAppTables(app.Manifest.AppTables, newManifest.AppTables)

		if err != nil {
			result.SetError(err.Error(), true)
			return
		}

		// extraColumnsDiff to json
		extraColumnsDiffString, err := json.Marshal(extraColumnsDiff)
		if err != nil {
			result.SetError(err.Error(), true)
			return
		}

		// appTablesDiff to json
		appTablesDiffString, err := json.Marshal(appTablesDiff)
		if err != nil {
			result.SetError(err.Error(), true)
			return
		}

		// persists diffs
		mainState["extra_columns_diff"] = string(extraColumnsDiffString)
		mainState["app_tables_diff"] = string(appTablesDiffString)

		// go to next stage
		mainState["stage"] = "extra_columns"
		result.Message = entity.StringPtr("validation step successful")

	case "extra_columns":

		// unmarshal extra columns diff
		if _, ok := mainState["extra_columns_diff"]; ok {
			if err := json.Unmarshal([]byte(mainState["extra_columns_diff"].(string)), extraColumnsDiff); err != nil {
				result.SetError(err.Error(), true)
				return
			}
		}

		log.Printf("extra_columns diff: %+v\n", extraColumnsDiff)

		// remove columns
		if extraColumnsDiff.ToRemove != nil {
			for _, operation := range extraColumnsDiff.ToRemove {
				if !operation.IsDone {
					if err := pipe.Repository.DeleteColumn(bgCtx, pipe.Workspace, operation.Table, operation.Column.Name); err != nil {
						result.SetError(err.Error(), true)
						return
					}

					// is done
					// update state in case of future failure
					operation.IsDone = true

					extraColumnsDiffString, err := json.Marshal(extraColumnsDiff)
					if err != nil {
						result.SetError(err.Error(), true)
						return
					}
					mainState["extra_columns_diff"] = string(extraColumnsDiffString)
				}
			}
		}

		// add columns
		if extraColumnsDiff.ToAdd != nil {
			for _, operation := range extraColumnsDiff.ToAdd {
				if !operation.IsDone {
					if err := pipe.Repository.AddColumn(bgCtx, pipe.Workspace, operation.Table, operation.Column); err != nil {
						result.SetError(err.Error(), true)
						return
					}

					// is done
					// update state in case of future failure
					operation.IsDone = true

					extraColumnsDiffString, err := json.Marshal(extraColumnsDiff)
					if err != nil {
						result.SetError(err.Error(), true)
						return
					}
					mainState["extra_columns_diff"] = string(extraColumnsDiffString)
				}
			}
		}

		mainState["stage"] = "app_tables"
		result.Message = entity.StringPtr("extra_columns step successful")

	case "app_tables":

		// unmarshal app tables diff
		appTablesDiff := &entity.AppTablesManifestDiff{}
		if _, ok := mainState["app_tables_diff"]; ok {
			if err := json.Unmarshal([]byte(mainState["app_tables_diff"].(string)), appTablesDiff); err != nil {
				result.SetError(err.Error(), true)
				return
			}
		}

		log.Printf("app_tables diff: %+v\n", appTablesDiff)

		// remove tables
		if appTablesDiff.ToRemove != nil {
			for _, operation := range appTablesDiff.ToRemove {
				if !operation.IsDone {
					// rename table with suffix _removed_YYYYMMDD_HHMMSS
					if err := pipe.Repository.DeleteTable(bgCtx, pipe.Workspace.ID, operation.AppTableManifest.Name+"_removed"+migratedTableSuffix); err != nil {
						result.SetError(err.Error(), true)
						return
					}

					// is done
					// update state in case of future failure
					operation.IsDone = true

					appTablesDiffString, err := json.Marshal(appTablesDiff)
					if err != nil {
						result.SetError(err.Error(), true)
						return
					}
					mainState["app_tables_diff"] = string(appTablesDiffString)
				}
			}
		}

		// add tables
		if appTablesDiff.ToAdd != nil {
			for _, operation := range appTablesDiff.ToAdd {
				if !operation.IsDone {
					if err := pipe.Repository.CreateTable(bgCtx, pipe.Workspace, operation.AppTableManifest); err != nil {
						result.SetError(err.Error(), true)
						return
					}
					// is done
					// update state in case of future failure
					operation.IsDone = true

					appTablesDiffString, err := json.Marshal(appTablesDiff)
					if err != nil {
						result.SetError(err.Error(), true)
						return
					}
					mainState["app_tables_diff"] = string(appTablesDiffString)
				}
			}
		}

		// migrate tables
		if appTablesDiff.ToMigrate != nil {
			for _, operation := range appTablesDiff.ToMigrate {
				if !operation.IsDone {
					if err := pipe.Repository.MigrateTable(bgCtx, pipe.Workspace, operation.AppTableManifest, migratedTableSuffix); err != nil {
						result.SetError(err.Error(), true)
						return
					}
					// is done
					// update state in case of future failure
					operation.IsDone = true

					appTablesDiffString, err := json.Marshal(appTablesDiff)
					if err != nil {
						result.SetError(err.Error(), true)
						return
					}
					mainState["app_tables_diff"] = string(appTablesDiffString)
				}
			}
		}

		mainState["app_tables"] = "finalize"
		result.Message = entity.StringPtr("app_tables step successful")

	case "finalize":
		code, err := pipe.Repository.RunInTransactionForSystem(ctx, func(ctx context.Context, tx *sql.Tx) (code int, err error) {

			// replace tasks
			if err := pipe.Repository.DeleteAppTasks(bgCtx, pipe.Workspace.ID, appID, tx); err != nil {
				return 500, err
			}

			if newManifest.DataHooks != nil && len(newManifest.DataHooks) > 0 {
				for _, stDefinition := range newManifest.Tasks {

					task := &entity.Task{
						TaskManifest: *stDefinition,
						WorkspaceID:  pipe.Workspace.ID,
						AppID:        newManifest.ID,
						IsActive:     true,
						IsCron:       stDefinition.IsCron,
					}

					if app.Status == entity.AppStatusInit {
						task.IsActive = false
					}

					task.ComputeNextRun()

					if err := pipe.Repository.InsertTask(ctx, task, tx); err != nil {
						return 500, eris.Wrap(err, "TaskUpgradeApp")
					}
				}
			}

			// replace hooks
			if newManifest.DataHooks != nil && len(newManifest.DataHooks) > 0 {
				for _, dataHook := range newManifest.DataHooks {

					hook := &entity.DataHook{
						ID:      dataHook.ID,
						AppID:   newManifest.ID,
						Name:    dataHook.Name,
						On:      dataHook.On,
						For:     dataHook.For,
						Enabled: false, // will be enabled when the app is activated
					}

					if err := hook.Validate(pipe.Workspace.InstalledApps); err != nil {
						return 400, eris.Wrap(err, "TaskUpgradeApp")
					}

					dataHookExists := false

					// replace eventually existing hook
					for _, existingHook := range pipe.Workspace.DataHooks {
						if existingHook.ID == hook.ID {
							dataHookExists = true
							existingHook = hook
							break
						}
					}

					if !dataHookExists {
						pipe.Workspace.DataHooks = append(pipe.Workspace.DataHooks, hook)
					}
				}
			}

			// update manifest in workspace
			for i, currentApp := range pipe.Workspace.InstalledApps {
				if currentApp.ID == newManifest.ID {
					pipe.Workspace.InstalledApps[i] = newManifest
					break
				}
			}

			if err := pipe.Repository.UpdateWorkspace(ctx, pipe.Workspace, tx); err != nil {
				return 500, eris.Wrap(err, "TaskUpgradeApp")
			}

			return 200, nil
		})

		if err != nil {
			if code == 500 {
				result.SetError(err.Error(), false)
				return
			}
			result.SetError(err.Error(), true)
			return
		}

		// update app manifest
		_, err = pipe.Repository.RunInTransactionForWorkspace(ctx, pipe.Workspace.ID, func(ctx context.Context, tx *sql.Tx) (code int, err error) {
			app.Manifest = *newManifest

			if err := pipe.Repository.UpdateApp(ctx, app, nil); err != nil {
				return 500, eris.Wrap(err, "TaskUpgradeApp")
			}

			return 200, nil
		})

		if err != nil {
			result.SetError(err.Error(), true)
			return
		}

		result.Message = entity.StringPtr("Upgrade successful")
		result.IsDone = true
	default:
		result.SetError("invalid stage", true)
		return
	}

	// update state
	result.UpdatedWorkerState = mainState

	return result
}
