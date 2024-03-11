package service

import (
	"context"
	"database/sql"
	"fmt"
	"io"
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
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestServiceImpl_DataPipelineUpsertCart(t *testing.T) {

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

	var webDomain *entity.Domain
	var webHost string

	for _, dom := range demoWorkspace.Domains {
		if dom.Type == entity.DomainWeb {
			webDomain = dom
			webHost = dom.Hosts[0].Host
		}
	}

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

	appEncryptedSecretKey, _ := common.EncryptString(cfgSecretKey, cfgSecretKey)

	t.Run("should insert cart", func(t *testing.T) {
		createdAt := time.Now().UTC().Add(-time.Hour * 24 * 30)
		signedUpAt := time.Now().UTC().Add(-time.Hour * 24 * 10)

		segments := []*entity.Segment{}
		for _, seg := range entity.GenerateDefaultSegments() {
			copy := seg // copy to avoid pointer overwrite
			segments = append(segments, &copy)
		}

		user := entity.NewUser("user_1234567890", true, createdAt, createdAt, "Europe/Paris", "fr", "FR", &signedUpAt)

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
				// logger.Printf("InsertDataLogFunc: %v, %v, %v", dataLog.Kind, dataLog.Action, dataLog.ItemExternalID)
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
			FindCartByIDFunc: func(ctx context.Context, workspaceID, cartID, userID string, tx *sql.Tx) (*entity.Cart, error) {
				return nil, nil
			},
			InsertCartFunc: func(ctx context.Context, cart *entity.Cart, tx *sql.Tx) error {
				return nil
			},
			InsertCartItemFunc: func(ctx context.Context, cartItem *entity.CartItem, tx *sql.Tx) error {
				return nil
			},
			UpdateDataLogFunc: func(ctx context.Context, workspaceID string, dataLog *entity.DataLog) error {
				if dataLog.Checkpoint != entity.DataLogCheckpointDone {
					return fmt.Errorf("invalid status: %v", dataLog.Checkpoint)
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
				"kind":"cart",
				"user": {
					"external_id": "%v",
					"is_authenticated": %v,
					"created_at": "%v",
					"signed_up_at": "%v"
				},
				"cart": {
					"external_id": "cart_1234567890",
					"domain_id": "%v",
					"items": [
						{
							"external_id": "item_a",
							"name": "Product A",
							"product_external_id": "product_a",
							"quantity": 1,
							"price": 10,
							"currency": "EUR"
						},
						{
							"external_id": "item_b",
							"name": "Product B",
							"product_external_id": "product_b",
							"quantity": 2,
							"price": 20,
							"currency": "EUR"
						}
					],
					"created_at": "%v"
				}
			}`,
				user.ExternalID,
				user.IsAuthenticated,
				user.CreatedAt.Format(time.RFC3339),
				user.SignedUpAt.Format(time.RFC3339),
				webDomain.ID,
				createdAt.Format(time.RFC3339),
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
		assert.Equal(t, "cart", pipeline.GetDataLog().Kind)
		assert.Equal(t, user.ID, pipeline.GetDataLog().UserID)
		assert.Equal(t, 1, len(repoMock.EnsureUsersLockCalls()))
		assert.Greater(t, len(repoMock.ReleaseUsersLockCalls()), 0)
		// upsert user
		assert.Equal(t, 5, len(repoMock.InsertDataLogCalls())) // 1 for user:create, 1 for segment:enter, 1 for cart:create, 2 for cart_item:create
		assert.Equal(t, 1, len(repoMock.FindUserByIDCalls()))
		assert.Equal(t, 1, len(repoMock.InsertUserCalls()))
		// upsert cart
		assert.Equal(t, 1, len(repoMock.FindCartByIDCalls()))
		assert.Equal(t, 1, len(repoMock.InsertCartCalls()))
		// upsert cart_items
		assert.Equal(t, 2, len(repoMock.InsertCartItemCalls()))
		// segmentation
		assert.Equal(t, 1, len(repoMock.ListSegmentsCalls()))
		assert.Equal(t, 1, len(repoMock.ListUsersCalls()))
		assert.Equal(t, 1, len(repoMock.ListUserSegmentsCalls()))
		assert.Equal(t, 1, len(repoMock.InsertUserSegmentCalls()))

		assert.Equal(t, user.ID, pipeline.GetDataLogsGenerated()[0].UserID)
		assert.Equal(t, "cart", pipeline.GetDataLogsGenerated()[0].Kind)
		assert.Equal(t, "create", pipeline.GetDataLogsGenerated()[0].Action)
		assert.Equal(t, "cart_1234567890", pipeline.GetDataLogsGenerated()[0].ItemExternalID)

		assert.Equal(t, user.ID, pipeline.GetDataLogsGenerated()[1].UserID)
		assert.Equal(t, "user", pipeline.GetDataLogsGenerated()[1].Kind)
		assert.Equal(t, "create", pipeline.GetDataLogsGenerated()[1].Action)
		assert.Equal(t, user.ID, pipeline.GetDataLogsGenerated()[1].ItemID)
		assert.Equal(t, user.ExternalID, pipeline.GetDataLogsGenerated()[1].ItemExternalID)

		assert.Equal(t, user.ID, pipeline.GetDataLogsGenerated()[2].UserID)
		assert.Equal(t, "cart_item", pipeline.GetDataLogsGenerated()[2].Kind)
		assert.Equal(t, "create", pipeline.GetDataLogsGenerated()[2].Action)
		assert.Equal(t, "item_a", pipeline.GetDataLogsGenerated()[2].ItemExternalID)

		assert.Equal(t, user.ID, pipeline.GetDataLogsGenerated()[3].UserID)
		assert.Equal(t, "cart_item", pipeline.GetDataLogsGenerated()[3].Kind)
		assert.Equal(t, "create", pipeline.GetDataLogsGenerated()[3].Action)
		assert.Equal(t, "item_b", pipeline.GetDataLogsGenerated()[3].ItemExternalID)

		assert.Equal(t, user.ID, pipeline.GetDataLogsGenerated()[4].UserID)
		assert.Equal(t, "segment", pipeline.GetDataLogsGenerated()[4].Kind)
		assert.Equal(t, "enter", pipeline.GetDataLogsGenerated()[4].Action)
		assert.Equal(t, "authenticated", pipeline.GetDataLogsGenerated()[4].ItemID)
		assert.Equal(t, "authenticated", pipeline.GetDataLogsGenerated()[4].ItemExternalID)
	})

	t.Run("should update cart and items", func(t *testing.T) {
		createdAt := time.Now().UTC().Add(-time.Hour * 24 * 30)
		signedUpAt := time.Now().UTC().Add(-time.Hour * 24 * 10)

		segments := []*entity.Segment{}
		for _, seg := range entity.GenerateDefaultSegments() {
			copy := seg // copy to avoid pointer overwrite
			segments = append(segments, &copy)
		}

		user := entity.NewUser("user_1234567890", true, createdAt, createdAt, "Europe/Paris", "fr", "FR", &signedUpAt)

		cartID := "cart_1234567890"

		existingItems := entity.CartItems{
			entity.NewCartItem("item_a", "Product A", user.ID, cartID, "product_a", createdAt, nil),
			entity.NewCartItem("item_b", "Product B", user.ID, cartID, "product_b", createdAt, nil),
		}

		existingCart := entity.NewCart(cartID, user.ID, webDomain.ID, createdAt, nil, existingItems)
		existingCart.Currency = &demoWorkspace.Currency

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
				// logger.Printf("InsertDataLogFunc: %v, %v, %v", dataLog.Kind, dataLog.Action, dataLog.ItemExternalID)
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
			FindCartByIDFunc: func(ctx context.Context, workspaceID, cartID, userID string, tx *sql.Tx) (*entity.Cart, error) {
				return existingCart, nil
			},
			FindCartItemsByCartIDFunc: func(ctx context.Context, workspaceID, cartID, userID string, tx *sql.Tx) ([]*entity.CartItem, error) {
				return existingItems, nil
			},
			UpdateCartFunc: func(ctx context.Context, cart *entity.Cart, tx *sql.Tx) error {
				return nil
			},
			UpdateCartItemFunc: func(ctx context.Context, cartItem *entity.CartItem, tx *sql.Tx) error {
				return nil
			},
			UpdateDataLogFunc: func(ctx context.Context, workspaceID string, dataLog *entity.DataLog) error {
				if dataLog.Checkpoint != entity.DataLogCheckpointDone {
					return fmt.Errorf("invalid status: %v", dataLog.Checkpoint)
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
				"kind":"cart",
				"user": {
					"external_id": "%v",
					"is_authenticated": %v,
					"created_at": "%v",
					"signed_up_at": "%v"
				},
				"cart": {
					"external_id": "%v",
					"domain_id": "%v",
					"items": [
						{
							"external_id": "item_a",
							"name": "Product A",
							"product_external_id": "product_a",
							"quantity": 1,
							"price": 10,
							"currency": "EUR"
						},
						{
							"external_id": "item_b",
							"name": "Product B",
							"product_external_id": "product_b",
							"quantity": 2,
							"price": 20,
							"currency": "EUR"
						}
					],
					"public_url": "https://www.apple.com/cart=xxxx",
					"created_at": "%v",
					"updated_at": "%v"
				}
			}`,
				user.ExternalID,
				user.IsAuthenticated,
				user.CreatedAt.Format(time.RFC3339),
				user.SignedUpAt.Format(time.RFC3339),
				cartID,
				webDomain.ID,
				createdAt.Format(time.RFC3339),
				time.Now().Format(time.RFC3339),
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
		assert.Equal(t, "cart", pipeline.GetDataLog().Kind)
		assert.Equal(t, user.ID, pipeline.GetDataLog().UserID)
		assert.Equal(t, 1, len(repoMock.EnsureUsersLockCalls()))
		assert.Greater(t, len(repoMock.ReleaseUsersLockCalls()), 0)
		// upsert user
		assert.Equal(t, 5, len(repoMock.InsertDataLogCalls())) // 1 for user:create, 1 for segment:enter, 1 for cart:create, 2 for cart_item:update
		assert.Equal(t, 1, len(repoMock.FindUserByIDCalls()))
		assert.Equal(t, 1, len(repoMock.InsertUserCalls()))
		// upsert cart
		assert.Equal(t, 1, len(repoMock.FindCartByIDCalls()))
		assert.Equal(t, 1, len(repoMock.UpdateCartCalls()))
		// upsert cart_items
		assert.Equal(t, 0, len(repoMock.InsertCartItemCalls()))
		assert.Equal(t, 2, len(repoMock.UpdateCartItemCalls()))
		// segmentation
		assert.Equal(t, 1, len(repoMock.ListSegmentsCalls()))
		assert.Equal(t, 1, len(repoMock.ListUsersCalls()))
		assert.Equal(t, 1, len(repoMock.ListUserSegmentsCalls()))
		assert.Equal(t, 1, len(repoMock.InsertUserSegmentCalls()))

		assert.Equal(t, user.ID, pipeline.GetDataLogsGenerated()[0].UserID)
		assert.Equal(t, "cart", pipeline.GetDataLogsGenerated()[0].Kind)
		assert.Equal(t, "update", pipeline.GetDataLogsGenerated()[0].Action)
		assert.Equal(t, "cart_1234567890", pipeline.GetDataLogsGenerated()[0].ItemExternalID)

		assert.Equal(t, user.ID, pipeline.GetDataLogsGenerated()[1].UserID)
		assert.Equal(t, "user", pipeline.GetDataLogsGenerated()[1].Kind)
		assert.Equal(t, "create", pipeline.GetDataLogsGenerated()[1].Action)
		assert.Equal(t, user.ID, pipeline.GetDataLogsGenerated()[1].ItemID)
		assert.Equal(t, user.ExternalID, pipeline.GetDataLogsGenerated()[1].ItemExternalID)

		assert.Equal(t, user.ID, pipeline.GetDataLogsGenerated()[2].UserID)
		assert.Equal(t, "cart_item", pipeline.GetDataLogsGenerated()[2].Kind)
		assert.Equal(t, "update", pipeline.GetDataLogsGenerated()[2].Action)
		assert.Equal(t, "item_a", pipeline.GetDataLogsGenerated()[2].ItemExternalID)

		assert.Equal(t, user.ID, pipeline.GetDataLogsGenerated()[3].UserID)
		assert.Equal(t, "cart_item", pipeline.GetDataLogsGenerated()[3].Kind)
		assert.Equal(t, "update", pipeline.GetDataLogsGenerated()[3].Action)
		assert.Equal(t, "item_b", pipeline.GetDataLogsGenerated()[3].ItemExternalID)

		assert.Equal(t, user.ID, pipeline.GetDataLogsGenerated()[4].UserID)
		assert.Equal(t, "segment", pipeline.GetDataLogsGenerated()[4].Kind)
		assert.Equal(t, "enter", pipeline.GetDataLogsGenerated()[4].Action)
		assert.Equal(t, "authenticated", pipeline.GetDataLogsGenerated()[4].ItemID)
		assert.Equal(t, "authenticated", pipeline.GetDataLogsGenerated()[4].ItemExternalID)
	})

	t.Run("should update cart_item only", func(t *testing.T) {
		createdAt := time.Now().UTC().Add(-time.Hour * 24 * 30)
		signedUpAt := time.Now().UTC().Add(-time.Hour * 24 * 10)

		segments := []*entity.Segment{}
		for _, seg := range entity.GenerateDefaultSegments() {
			copy := seg // copy to avoid pointer overwrite
			segments = append(segments, &copy)
		}

		user := entity.NewUser("user_1234567890", true, createdAt, createdAt, "Europe/Paris", "fr", "FR", &signedUpAt)

		cartID := "cart_1234567890"

		existingItems := entity.CartItems{
			entity.NewCartItem("item_a", "Product A", user.ID, cartID, "product_a", createdAt, nil),
			entity.NewCartItem("item_b", "Product B", user.ID, cartID, "product_b", createdAt, nil),
		}

		existingCart := entity.NewCart(cartID, user.ID, webDomain.ID, createdAt, nil, existingItems)
		existingCart.Currency = &demoWorkspace.Currency

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
				// logger.Printf("InsertDataLogFunc: %v, %v, %v", dataLog.Kind, dataLog.Action, dataLog.ItemExternalID)
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
			FindCartByIDFunc: func(ctx context.Context, workspaceID, cartID, userID string, tx *sql.Tx) (*entity.Cart, error) {
				return existingCart, nil
			},
			FindCartItemsByCartIDFunc: func(ctx context.Context, workspaceID, cartID, userID string, tx *sql.Tx) ([]*entity.CartItem, error) {
				return existingItems, nil
			},
			UpdateCartItemFunc: func(ctx context.Context, cartItem *entity.CartItem, tx *sql.Tx) error {
				return nil
			},
			UpdateDataLogFunc: func(ctx context.Context, workspaceID string, dataLog *entity.DataLog) error {
				// for _, field := range dataLog.UpdatedFields {
				// 	svc.Logger.Printf("data log updated field %+v\n", field)
				// }
				if dataLog.Checkpoint != entity.DataLogCheckpointDone {
					return fmt.Errorf("invalid status: %v", dataLog.Checkpoint)
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
				"kind":"cart",
				"user": {
					"external_id": "%v",
					"is_authenticated": %v,
					"created_at": "%v",
					"signed_up_at": "%v"
				},
				"cart": {
					"external_id": "%v",
					"domain_id": "%v",
					"items": [
						{
							"external_id": "item_a",
							"name": "Product A",
							"product_external_id": "product_a",
							"quantity": 2,
							"price": 20,
							"currency": "EUR"
						},
						{
							"external_id": "item_b",
							"name": "Product B",
							"product_external_id": "product_b"
						}
					],
					"created_at": "%v",
					"updated_at": "%v"
				}
			}`,
				user.ExternalID,
				user.IsAuthenticated,
				user.CreatedAt.Format(time.RFC3339),
				user.SignedUpAt.Format(time.RFC3339),
				cartID,
				webDomain.ID,
				createdAt.Format(time.RFC3339),
				time.Now().Format(time.RFC3339),
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
		assert.Equal(t, "cart", pipeline.GetDataLog().Kind)
		assert.Equal(t, user.ID, pipeline.GetDataLog().UserID)
		assert.Equal(t, 1, len(repoMock.EnsureUsersLockCalls()))
		assert.Greater(t, len(repoMock.ReleaseUsersLockCalls()), 0)
		// upsert user
		assert.Equal(t, 4, len(repoMock.InsertDataLogCalls())) // 1 for user:create, 1 for segment:enter, 1 for cart:noop, 1 for cart_item:update
		assert.Equal(t, 1, len(repoMock.FindUserByIDCalls()))
		assert.Equal(t, 1, len(repoMock.InsertUserCalls()))
		// upsert cart
		assert.Equal(t, 1, len(repoMock.FindCartByIDCalls()))
		assert.Equal(t, 0, len(repoMock.InsertCartCalls()))
		assert.Equal(t, 0, len(repoMock.UpdateCartCalls()))
		// upsert cart_items
		assert.Equal(t, 0, len(repoMock.InsertCartItemCalls()))
		assert.Equal(t, 1, len(repoMock.UpdateCartItemCalls()))
		// segmentation
		assert.Equal(t, 1, len(repoMock.ListSegmentsCalls()))
		assert.Equal(t, 1, len(repoMock.ListUsersCalls()))
		assert.Equal(t, 1, len(repoMock.ListUserSegmentsCalls()))
		assert.Equal(t, 1, len(repoMock.InsertUserSegmentCalls()))

		assert.Equal(t, user.ID, pipeline.GetDataLogsGenerated()[0].UserID)
		assert.Equal(t, "cart", pipeline.GetDataLogsGenerated()[0].Kind)
		assert.Equal(t, "noop", pipeline.GetDataLogsGenerated()[0].Action)
		assert.Equal(t, "cart_1234567890", pipeline.GetDataLogsGenerated()[0].ItemExternalID)

		assert.Equal(t, user.ID, pipeline.GetDataLogsGenerated()[1].UserID)
		assert.Equal(t, "user", pipeline.GetDataLogsGenerated()[1].Kind)
		assert.Equal(t, "create", pipeline.GetDataLogsGenerated()[1].Action)
		assert.Equal(t, user.ID, pipeline.GetDataLogsGenerated()[1].ItemID)
		assert.Equal(t, user.ExternalID, pipeline.GetDataLogsGenerated()[1].ItemExternalID)

		assert.Equal(t, user.ID, pipeline.GetDataLogsGenerated()[2].UserID)
		assert.Equal(t, "cart_item", pipeline.GetDataLogsGenerated()[2].Kind)
		assert.Equal(t, "update", pipeline.GetDataLogsGenerated()[2].Action)
		assert.Equal(t, "item_a", pipeline.GetDataLogsGenerated()[2].ItemExternalID)

		assert.Equal(t, user.ID, pipeline.GetDataLogsGenerated()[3].UserID)
		assert.Equal(t, "segment", pipeline.GetDataLogsGenerated()[3].Kind)
		assert.Equal(t, "enter", pipeline.GetDataLogsGenerated()[3].Action)
		assert.Equal(t, "authenticated", pipeline.GetDataLogsGenerated()[3].ItemID)
		assert.Equal(t, "authenticated", pipeline.GetDataLogsGenerated()[3].ItemExternalID)
	})

	t.Run("should delete 1 cart_item, update 1 and add 1", func(t *testing.T) {
		createdAt := time.Now().UTC().Add(-time.Hour * 24 * 30)
		signedUpAt := time.Now().UTC().Add(-time.Hour * 24 * 10)

		segments := []*entity.Segment{}
		for _, seg := range entity.GenerateDefaultSegments() {
			copy := seg // copy to avoid pointer overwrite
			segments = append(segments, &copy)
		}

		user := entity.NewUser("user_1234567890", true, createdAt, createdAt, "Europe/Paris", "fr", "FR", &signedUpAt)

		cartID := "cart_1234567890"

		existingItems := entity.CartItems{
			entity.NewCartItem("item_a", "Product A", user.ID, cartID, "product_a", createdAt, nil),
			entity.NewCartItem("item_b", "Product B", user.ID, cartID, "product_b", createdAt, nil),
		}

		existingCart := entity.NewCart(cartID, user.ID, webDomain.ID, createdAt, nil, existingItems)
		existingCart.Currency = &demoWorkspace.Currency

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
				// logger.Printf("InsertDataLogFunc: %v, %v, %v", dataLog.Kind, dataLog.Action, dataLog.ItemExternalID)
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
			FindCartByIDFunc: func(ctx context.Context, workspaceID, cartID, userID string, tx *sql.Tx) (*entity.Cart, error) {
				return existingCart, nil
			},
			FindCartItemsByCartIDFunc: func(ctx context.Context, workspaceID, cartID, userID string, tx *sql.Tx) ([]*entity.CartItem, error) {
				return existingItems, nil
			},
			DeleteCartItemFunc: func(ctx context.Context, cartItemID, userID string, tx *sql.Tx) error {
				return nil
			},
			InsertCartItemFunc: func(ctx context.Context, cartItem *entity.CartItem, tx *sql.Tx) error {
				return nil
			},
			UpdateCartItemFunc: func(ctx context.Context, cartItem *entity.CartItem, tx *sql.Tx) error {
				return nil

			},
			UpdateDataLogFunc: func(ctx context.Context, workspaceID string, dataLog *entity.DataLog) error {
				// for _, field := range dataLog.UpdatedFields {
				// 	svc.Logger.Printf("data log updated field %+v\n", field)
				// }
				if dataLog.Checkpoint != entity.DataLogCheckpointDone {
					return fmt.Errorf("invalid status: %v", dataLog.Checkpoint)
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
				"kind":"cart",
				"user": {
					"external_id": "%v",
					"is_authenticated": %v,
					"created_at": "%v",
					"signed_up_at": "%v"
				},
				"cart": {
					"external_id": "%v",
					"domain_id": "%v",
					"items": [
						{
							"external_id": "item_b",
							"name": "Product B updated",
							"product_external_id": "product_b",
							"quantity": 1,
							"price": 10,
							"currency": "EUR"
						},
						{
							"external_id": "item_c",
							"name": "Product C",
							"product_external_id": "product_c",
							"quantity": 2,
							"price": 20,
							"currency": "EUR"
						},
						{
							"external_id": "item_d",
							"name": "Product D",
							"product_external_id": "product_d"
						}
					],
					"created_at": "%v",
					"updated_at": "%v"
				}
			}`,
				user.ExternalID,
				user.IsAuthenticated,
				user.CreatedAt.Format(time.RFC3339),
				user.SignedUpAt.Format(time.RFC3339),
				cartID,
				webDomain.ID,
				createdAt.Format(time.RFC3339),
				time.Now().Format(time.RFC3339),
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
		assert.Equal(t, "cart", pipeline.GetDataLog().Kind)
		assert.Equal(t, user.ID, pipeline.GetDataLog().UserID)
		assert.Equal(t, 1, len(repoMock.EnsureUsersLockCalls()))
		assert.Greater(t, len(repoMock.ReleaseUsersLockCalls()), 0)
		// upsert user
		assert.Equal(t, 7, len(repoMock.InsertDataLogCalls())) // 1 for user:create, 1 for segment:enter, 1 for cart:noop, 2 for cart_item:delete, 3 for cart_item:create
		assert.Equal(t, 1, len(repoMock.FindUserByIDCalls()))
		assert.Equal(t, 1, len(repoMock.InsertUserCalls()))
		// upsert cart
		assert.Equal(t, 1, len(repoMock.FindCartByIDCalls()))
		assert.Equal(t, 0, len(repoMock.InsertCartCalls()))
		assert.Equal(t, 0, len(repoMock.UpdateCartCalls()))
		// upsert cart_items
		assert.Equal(t, 1, len(repoMock.DeleteCartItemCalls()))
		assert.Equal(t, 2, len(repoMock.InsertCartItemCalls()))
		assert.Equal(t, 1, len(repoMock.UpdateCartItemCalls()))
		// segmentation
		assert.Equal(t, 1, len(repoMock.ListSegmentsCalls()))
		assert.Equal(t, 1, len(repoMock.ListUsersCalls()))
		assert.Equal(t, 1, len(repoMock.ListUserSegmentsCalls()))
		assert.Equal(t, 1, len(repoMock.InsertUserSegmentCalls()))

		assert.Equal(t, user.ID, pipeline.GetDataLogsGenerated()[0].UserID)
		assert.Equal(t, "cart", pipeline.GetDataLogsGenerated()[0].Kind)
		assert.Equal(t, "noop", pipeline.GetDataLogsGenerated()[0].Action)
		assert.Equal(t, "cart_1234567890", pipeline.GetDataLogsGenerated()[0].ItemExternalID)

		assert.Equal(t, user.ID, pipeline.GetDataLogsGenerated()[1].UserID)
		assert.Equal(t, "user", pipeline.GetDataLogsGenerated()[1].Kind)
		assert.Equal(t, "create", pipeline.GetDataLogsGenerated()[1].Action)
		assert.Equal(t, user.ID, pipeline.GetDataLogsGenerated()[1].ItemID)
		assert.Equal(t, user.ExternalID, pipeline.GetDataLogsGenerated()[1].ItemExternalID)

		assert.Equal(t, user.ID, pipeline.GetDataLogsGenerated()[2].UserID)
		assert.Equal(t, "cart_item", pipeline.GetDataLogsGenerated()[2].Kind)
		assert.Equal(t, "delete", pipeline.GetDataLogsGenerated()[2].Action)
		assert.Equal(t, "item_a", pipeline.GetDataLogsGenerated()[2].ItemExternalID)

		assert.Equal(t, user.ID, pipeline.GetDataLogsGenerated()[3].UserID)
		assert.Equal(t, "cart_item", pipeline.GetDataLogsGenerated()[3].Kind)
		assert.Equal(t, "update", pipeline.GetDataLogsGenerated()[3].Action)
		assert.Equal(t, "item_b", pipeline.GetDataLogsGenerated()[3].ItemExternalID)

		assert.Equal(t, user.ID, pipeline.GetDataLogsGenerated()[4].UserID)
		assert.Equal(t, "cart_item", pipeline.GetDataLogsGenerated()[4].Kind)
		assert.Equal(t, "create", pipeline.GetDataLogsGenerated()[4].Action)
		assert.Equal(t, "item_c", pipeline.GetDataLogsGenerated()[4].ItemExternalID)

		assert.Equal(t, user.ID, pipeline.GetDataLogsGenerated()[5].UserID)
		assert.Equal(t, "cart_item", pipeline.GetDataLogsGenerated()[5].Kind)
		assert.Equal(t, "create", pipeline.GetDataLogsGenerated()[5].Action)
		assert.Equal(t, "item_d", pipeline.GetDataLogsGenerated()[5].ItemExternalID)

		assert.Equal(t, user.ID, pipeline.GetDataLogsGenerated()[6].UserID)
		assert.Equal(t, "segment", pipeline.GetDataLogsGenerated()[6].Kind)
		assert.Equal(t, "enter", pipeline.GetDataLogsGenerated()[6].Action)
		assert.Equal(t, "authenticated", pipeline.GetDataLogsGenerated()[6].ItemID)
		assert.Equal(t, "authenticated", pipeline.GetDataLogsGenerated()[6].ItemExternalID)
	})
}
