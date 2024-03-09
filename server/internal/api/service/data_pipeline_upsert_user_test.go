package service

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/rimdian/rimdian/internal/api/common"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/internal/api/repository"
	commonDTO "github.com/rimdian/rimdian/internal/common/dto"
	"github.com/rimdian/rimdian/internal/common/httpClient"
	"github.com/stretchr/testify/assert"
)

func TestServiceImpl_DataPipelineUpsertUser(t *testing.T) {

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

	netClientMock := &httpClient.HTTPClientMock{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// return a mock response
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{}`)),
				// Body:       io.NopCloser(strings.NewReader(`{"browser":{"name":"Chromium","version":"15.0.874.106"},"device":{},"os":{"name":"Ubuntu","version":"11.10"}}`)),
			}, nil

		},
	}

	t.Run("should insert user", func(t *testing.T) {
		createdAt := time.Now().UTC().Add(-time.Hour * 24 * 30)
		signedUpAt := time.Now().UTC().Add(-time.Hour * 24 * 10)

		segments := []*entity.Segment{}
		for _, seg := range entity.GenerateDefaultSegments() {
			copy := seg // copy to avoid pointer overwrite
			segments = append(segments, &copy)
		}

		user := entity.NewUser("1234567890", true, createdAt, createdAt, "Europe/Paris", "fr", "FR", &signedUpAt)

		var segmentDataLog *entity.DataLog

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
				return nil
			},
			EnsureUsersLockFunc: func(ctx context.Context, workspaceID string, lock *entity.UsersLock, withRetry bool) error {
				return nil
			},
			FindUserByIDFunc: func(ctx context.Context, workspace *entity.Workspace, userID string, tx *sql.Tx) (*entity.User, error) {
				return nil, sql.ErrNoRows
			},
			InsertUserFunc: func(ctx context.Context, user *entity.User, tx *sql.Tx) error {
				return nil
			},
			ListSegmentsFunc: func(ctx context.Context, workspaceID string, withUsersCount bool) ([]*entity.Segment, error) {
				return segments, nil
			},
			ListUsersFunc: func(ctx context.Context, workspace *entity.Workspace, params *dto.UserListParams) ([]*entity.User, string, string, error) {
				return []*entity.User{user}, "", "", nil
			},
			ListUserSegmentsFunc: func(ctx context.Context, workspaceID string, userIDs []string, tx *sql.Tx) ([]*entity.UserSegment, error) {
				return []*entity.UserSegment{}, nil
			},
			InsertUserSegmentFunc: func(ctx context.Context, userSegment *entity.UserSegment, tx *sql.Tx) error {
				return nil
			},
			IsDuplicateEntryFunc: func(err error) bool {
				return false
			},
			UpdateDataLogFunc: func(ctx context.Context, workspaceID string, dataLog *entity.DataLog) error {
				log.Printf("UpdateDataLog: %+v\n", dataLog)
				// if dataLog.Checkpoint != entity.DataLogCheckpointDone {
				// 	return fmt.Errorf("invalid status: %v", dataLog.Checkpoint)
				// }
				if dataLog.Kind == "segment" {
					segmentDataLog = dataLog
				}
				return nil
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
		}

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
				"kind":"user",
				"user": {
					"external_id": "%v",
					"is_authenticated": %v,
					"created_at": "%v",
					"signed_up_at": "%v"
				}
			}`, user.ExternalID, user.IsAuthenticated, user.CreatedAt.Format(time.RFC3339), user.SignedUpAt.Format(time.RFC3339)),
		}

		dataLogInQueue.ComputeID(cfgSecretKey)

		props := &DataPipelineProps{
			Config:         cfg,
			NetClient:      netClientMock,
			Repository:     repoMock,
			Workspace:      demoWorkspace,
			DataLogInQueue: dataLogInQueue,
		}

		pipeline := NewDataPipeline(props)
		pipeline.Execute(context.Background())

		result := pipeline.GetQueueResult()

		// for _, dl := range pipeline.GetDataLogsGenerated() {
		// 	log.Printf("dl %v : %v : %v", dl.Kind, dl.Action, dl.ItemExternalID)
		// }
		// log.Printf("GetDataLog: %+v\n", pipeline.GetDataLog())

		assert.NotNil(t, result)
		assert.False(t, result.HasError)
		assert.False(t, result.QueueShouldRetry)
		assert.Equal(t, "", result.Error)
		// assert.Containsf(t, result.Error, "host not allowed", "error should contain: host not allowed")
		assert.Equal(t, entity.DataLogCheckpointDone, pipeline.GetDataLog().Checkpoint)
		assert.Equal(t, "user", pipeline.GetDataLog().Kind)
		assert.Equal(t, user.ID, pipeline.GetDataLog().UserID)
		assert.Equal(t, 1, len(repoMock.EnsureUsersLockCalls()))
		assert.Equal(t, 1, len(repoMock.ReleaseUsersLockCalls()))
		// upsert user
		assert.Equal(t, 2, len(repoMock.InsertDataLogCalls())) // 1 for user:create, 1 for segment:enter
		assert.Equal(t, 2, len(repoMock.UpdateDataLogCalls())) // 1 for status:done, 1 for segment:enter hook
		assert.Equal(t, 1, len(repoMock.FindUserByIDCalls()))
		assert.Equal(t, 1, len(repoMock.InsertUserCalls()))
		// segmentation
		assert.Equal(t, 1, len(repoMock.ListSegmentsCalls()))
		assert.Equal(t, 1, len(repoMock.ListUsersCalls()))
		assert.Equal(t, 1, len(repoMock.ListUserSegmentsCalls()))
		assert.Equal(t, 1, len(repoMock.InsertUserSegmentCalls()))

		assert.Equal(t, user.ID, pipeline.GetDataLogsGenerated()[0].UserID)
		assert.Equal(t, "user", pipeline.GetDataLogsGenerated()[0].Kind)
		assert.Equal(t, "create", pipeline.GetDataLogsGenerated()[0].Action)
		assert.Equal(t, user.ID, pipeline.GetDataLogsGenerated()[0].ItemID)
		assert.Equal(t, user.ExternalID, pipeline.GetDataLogsGenerated()[0].ItemExternalID)

		assert.Equal(t, user.ID, pipeline.GetDataLogsGenerated()[1].UserID)
		assert.Equal(t, "segment", pipeline.GetDataLogsGenerated()[1].Kind)
		assert.Equal(t, "enter", pipeline.GetDataLogsGenerated()[1].Action)
		assert.Equal(t, "authenticated", pipeline.GetDataLogsGenerated()[1].ItemID)
		assert.Equal(t, "authenticated", pipeline.GetDataLogsGenerated()[1].ItemExternalID)

		assert.Equal(t, 1, len(segmentDataLog.Hooks))
		assert.Equal(t, true, segmentDataLog.Hooks["app_test_table_segment_enter"].Done)
	})
}
