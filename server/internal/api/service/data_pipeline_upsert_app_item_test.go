package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/rimdian/rimdian/internal/api/common"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/internal/api/repository"
	commonDTO "github.com/rimdian/rimdian/internal/common/dto"
	"github.com/rimdian/rimdian/internal/common/httpClient"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestServiceImpl_DataPipelineUpsertAppItem(t *testing.T) {

	logger := logrus.New()

	cfgSecretKey := "12345678901234567890123456789012"

	cfg := &entity.Config{
		SECRET_KEY: cfgSecretKey,
		ENV:        entity.ENV_DEV,
	}

	orgID := "testing"
	workspaceID := fmt.Sprintf("%v_%v", orgID, "demoecommerce")

	demoWorkspace, err := entity.GenerateDemoWorkspace(workspaceID, entity.WorkspaceDemoOrder, orgID, cfgSecretKey)

	if err != nil {
		t.Fatalf("generate demo workspace err %v", err)
	}

	demoWorkspace.InstalledApps = append(demoWorkspace.InstalledApps, &entity.AppManifestTest)

	// install hooks
	for _, hook := range entity.AppManifestTest.DataHooks {
		demoWorkspace.DataHooks = append(demoWorkspace.DataHooks, &entity.DataHook{
			ID:      hook.ID,
			AppID:   entity.AppManifestTest.ID,
			Name:    hook.Name,
			On:      hook.On,
			Kind:    hook.Kind,
			Action:  hook.Action,
			Enabled: true,
		})
	}

	var webHost string
	for _, dom := range demoWorkspace.Domains {
		if dom.Type == entity.DomainWeb {
			webHost = dom.Hosts[0].Host
		}
	}

	appEncryptedSecretKey, _ := common.EncryptString(cfgSecretKey, cfgSecretKey)

	t.Run("should insert app_test_table", func(t *testing.T) {

		netClientMock := &httpClient.HTTPClientMock{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				// do ping pong
				body, _ := req.GetBody()
				// return a mock response
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       body,
					// Body:       io.NopCloser(strings.NewReader(`{"browser":{"name":"Chromium","version":"15.0.874.106"},"device":{},"os":{"name":"Ubuntu","version":"11.10"}}`)),
				}, nil

			},
		}

		var updatedDataLog *entity.DataLog

		segments := []*entity.Segment{}
		for _, seg := range entity.GenerateDefaultSegments() {
			copy := seg // copy to avoid pointer overwrite
			segments = append(segments, &copy)
		}

		repoMock := &repository.RepositoryMock{
			GetWorkspaceFunc: func(ctx context.Context, workspaceID string) (*entity.Workspace, error) {
				return demoWorkspace, nil
			},
			ReleaseUsersLockFunc: func(workspaceID string, lock *entity.UsersLock) error {
				return nil
			},
			UpdateWorkspaceFunc: func(ctx context.Context, workspace *entity.Workspace, tx *sql.Tx) error {
				// update fx rates
				return nil
			},
			RunInTransactionForWorkspaceFunc: func(ctx context.Context, workspaceID string, f func(context.Context, *sql.Tx) (int, error)) (int, error) {
				return f(ctx, nil)
			},
			InsertDataLogFunc: func(ctx context.Context, workspaceID string, dataLog *entity.DataLog, tx *sql.Tx) error {
				// logger.Printf("InsertDataLogFunc: %+v\n", dataLog)
				return nil
			},
			EnsureUsersLockFunc: func(ctx context.Context, workspaceID string, lock *entity.UsersLock, withRetry bool) error {
				return nil
			},
			FindUserByIDFunc: func(ctx context.Context, workspace *entity.Workspace, userID string, tx *sql.Tx) (*entity.User, error) {
				return nil, sql.ErrNoRows
			},
			FindUserAliasFunc: func(ctx context.Context, fromUserExternalID string, tx *sql.Tx) (*entity.UserAlias, error) {
				return nil, nil
			},
			InsertUserFunc: func(ctx context.Context, user *entity.User, tx *sql.Tx) error {
				return nil
			},
			ListUsersFunc: func(ctx context.Context, workspace *entity.Workspace, params *dto.UserListParams) ([]*entity.User, string, string, error) {
				return []*entity.User{}, "", "", nil
			},
			ListAppsFunc: func(ctx context.Context, workspaceID string) ([]*entity.App, error) {
				return []*entity.App{
					{
						ID:                 demoWorkspace.InstalledApps[0].ID,
						Name:               demoWorkspace.InstalledApps[0].Name,
						Status:             entity.AppStatusActive,
						State:              entity.MapOfInterfaces{},
						Manifest:           entity.AppManifestTest,
						EncryptedSecretKey: appEncryptedSecretKey,
						IsNative:           true,
						CreatedAt:          time.Now(),
						UpdatedAt:          time.Now(),
					},
				}, nil
			},
			// ListSegmentsFunc: func(ctx context.Context, workspaceID string, withUsersCount bool) ([]*entity.Segment, error) {
			// 	return segments, nil
			// },
			// ListUsersFunc: func(ctx context.Context, workspaceID string, params *dto.UserListParams) ([]*entity.User, string, string, error) {
			// 	return []*entity.User{user}, "", "", nil
			// },
			// ListUserSegmentsFunc: func(ctx context.Context, workspaceID string, userIDs []string, tx *sql.Tx) ([]*entity.UserSegment, error) {
			// 	return []*entity.UserSegment{}, nil
			// },
			// InsertUserSegmentFunc: func(ctx context.Context, userSegment *entity.UserSegment, tx *sql.Tx) error {
			// 	return nil
			// },
			IsDuplicateEntryFunc: func(err error) bool {
				return false
			},
			FindAppItemByIDFunc: func(ctx context.Context, workspace *entity.Workspace, eventID string, userID string, tx *sql.Tx) (*entity.AppItem, error) {
				return nil, sql.ErrNoRows
			},
			UpdateDataLogFunc: func(ctx context.Context, workspaceID string, dataLog *entity.DataLog) error {
				updatedDataLog = dataLog
				if dataLog.Checkpoint != entity.DataLogCheckpointDone {
					return fmt.Errorf("invalid status: %v", dataLog.Checkpoint)
				}
				return nil
			},
			ListSegmentsFunc: func(ctx context.Context, workspaceID string, withUsersCount bool) ([]*entity.Segment, error) {
				return segments, nil
			},
			InsertAppItemFunc: func(ctx context.Context, kind string, upsertedAppItem *entity.AppItem, tx *sql.Tx) error {
				return nil
			},
		}

		kind := entity.AppManifestTest.AppTables[0].Name
		externalID := "1234567890"
		createdAt := time.Now().AddDate(0, 0, -1).Format(time.RFC3339)
		updatedAt := time.Now().Format(time.RFC3339)
		timestamp := time.Now().Unix()

		dataLogInQueue := &commonDTO.DataLogInQueue{
			Origin:   commonDTO.DataLogOriginClient,
			OriginID: webHost,
			Context: commonDTO.DataLogContext{
				WorkspaceID: demoWorkspace.ID,
				HeadersAndParams: commonDTO.MapOfStrings{
					"Origin": "https://www.apple.com",
				},
			},
			Item: fmt.Sprintf(`{
				"kind":"%v",
				"%v": {
					"external_id": "%v",
					"created_at": "%v",
					"updated_at": "%v",
					"required_varchar": "abc",
					"required_longtext": "abc",
					"required_number": 123.456,
					"required_boolean": true,
					"required_date": "2023-01-30",
					"required_timestamp": %v,
					"required_json": {
						"string": "value",
						"number": 123.456,
						"bool": true,
						"is_null": null,
						"array": ["a", "b", "c"]
					}
				},
				"user": {
					"external_id": "1234567890",
					"is_authenticated": false,
					"created_at": "%v"
				}
			}`,
				kind,
				kind,
				externalID,
				createdAt,
				updatedAt,
				timestamp,
				createdAt,
			),
		}

		dataLogInQueue.ComputeID(cfgSecretKey)

		props := &DataPipelineProps{
			Config:         cfg,
			Logger:         logger,
			NetClient:      netClientMock,
			Repository:     repoMock,
			Workspace:      demoWorkspace,
			DataLogInQueue: dataLogInQueue,
		}

		pipeline := NewDataPipeline(props)
		pipeline.Execute(context.Background())

		result := pipeline.GetQueueResult()

		// for _, dl := range pipeline.GetDataLogsGenerated() {
		// 	svc.Logger.Printf("dl %v : %v : %v", dl.Kind, dl.Action, dl.ItemExternalID)
		// }
		// svc.Logger.Printf("GetDataLog: %+v\n", pipeline.GetDataLog())

		assert.NotNil(t, result)
		assert.False(t, result.HasError)
		assert.False(t, result.QueueShouldRetry)
		assert.Equal(t, "", result.Error)
		// assert.Containsf(t, result.Error, "host not allowed", "error should contain: host not allowed")
		assert.Equal(t, entity.DataLogCheckpointDone, pipeline.GetDataLog().Checkpoint)
		assert.Equal(t, kind, pipeline.GetDataLog().Kind)
		assert.Equal(t, "01b307acba4f54f55aafc33bb06bbbf6ca803e9a", pipeline.GetDataLog().UserID)
		assert.Equal(t, 1, len(repoMock.EnsureUsersLockCalls()))
		assert.Greater(t, len(repoMock.ReleaseUsersLockCalls()), 0)
		// upsert user
		assert.Equal(t, 2, len(repoMock.InsertDataLogCalls())) // 1 for app item, 1 for user
		assert.Equal(t, 1, len(repoMock.FindUserByIDCalls()))
		assert.Equal(t, 1, len(repoMock.InsertUserCalls()))
		// segmentation
		assert.Equal(t, 1, len(repoMock.ListSegmentsCalls()))
		assert.Equal(t, 1, len(repoMock.ListUsersCalls()))
		assert.Equal(t, 0, len(repoMock.ListUserSegmentsCalls()))
		assert.Equal(t, 0, len(repoMock.InsertUserSegmentCalls()))

		assert.Equal(t, "01b307acba4f54f55aafc33bb06bbbf6ca803e9a", pipeline.GetDataLogsGenerated()[0].UserID)
		assert.Equal(t, kind, pipeline.GetDataLogsGenerated()[0].Kind)
		assert.Equal(t, "create", pipeline.GetDataLogsGenerated()[0].Action)
		assert.Equal(t, externalID, pipeline.GetDataLogsGenerated()[0].ItemExternalID)

		// data hooks
		assert.Equal(t, 1, len(netClientMock.DoCalls()))
		assert.Equal(t, 1, len(updatedDataLog.Hooks))
		assert.Equal(t, true, updatedDataLog.Hooks[entity.AppManifestTest.DataHooks[0].ID].Done)

		// we did a ping pong with the webhook to test it
		var webhookPayload entity.AppWebhookPayload

		if err := json.Unmarshal([]byte(updatedDataLog.Hooks[entity.AppManifestTest.DataHooks[0].ID].Message), &webhookPayload); err != nil {
			t.Fatalf("unmarshal webhook payload err %v", err)
		}

		assert.Equal(t, "app_test", webhookPayload.AppID)
		assert.Equal(t, entity.AppWebhookKindDataHook, webhookPayload.Kind)
		assert.Equal(t, "app_test", webhookPayload.AppID)
		assert.Equal(t, entity.AppManifestTest.DataHooks[0].Name, webhookPayload.DataHook.DataHookName)
		assert.Equal(t, updatedDataLog.ID, webhookPayload.DataHook.DataLogID)
	})
}
