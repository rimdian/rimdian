package service

// import (
// 	"context"
// 	"database/sql"
// 	"fmt"
// 	"testing"
// 	"time"

// 	"github.com/rimdian/rimdian/internal/api/dto"
// 	"github.com/rimdian/rimdian/internal/api/entity"
// 	"github.com/rimdian/rimdian/internal/api/repository"
// )

// func TestServiceImpl_UserAlias(t *testing.T) {

// 	cfgSecretKey := "12345678901234567890123456789012"
// 	orgID := "testing"
// 	workspaceID := fmt.Sprintf("%v_%v", orgID, "demoecommerce")

// 	demoWorkspace, err := entity.GenerateDemoWorkspace(workspaceID, entity.WorkspaceDemoOrder, orgID, cfgSecretKey)

// 	if err != nil {
// 		t.Fatalf("generate demo workspace err %v", err)
// 	}

// 	t.Run("should fail on creating existing alias", func(t *testing.T) {

// 		fromUserExternalID := "from-user"
// 		toUserExternalID := "to-user"
// 		createdAt := time.Now().AddDate(0, -1, 0)
// 		called := 0

// 		repoMock := &repository.RepositoryMock{
// 			FindUserByIDFunc: func(ctx context.Context, workspaceID, userID string, tx *sql.Tx) (*entity.User, error) {
// 				if called == 0 {
// 					called = 1
// 					return entity.NewUser(fromUserExternalID, true, createdAt, createdAt, "Europe/Paris", "fr", "FR", nil), nil
// 				}
// 				return entity.NewUser(toUserExternalID, true, createdAt, createdAt, "Europe/Paris", "fr", "FR", nil), nil
// 			},
// 			FindUserAliasFunc: func(ctx context.Context, fromUserExternalID string, tx *sql.Tx) (*entity.UserAlias, error) {
// 				return &entity.UserAlias{
// 					DBCreatedAt:           createdAt,
// 					FromUserExternalID:    fromUserExternalID,
// 					ToUserExternalID:      toUserExternalID,
// 					ToUserIsAuthenticated: true,
// 				}, nil
// 			},
// 			CreateUserAliasFunc: func(ctx context.Context, fromUserID, toUserID string, toUserIsAuthenticated bool, tx *sql.Tx) error {
// 				return entity.ErrUserAliasAlreadyExists
// 			},
// 		}

// 		svc := &ServiceImpl{
// 			Config: &entity.Config{SECRET_KEY: "12345678901234567890123456789012"},
// 			Repo:   repoMock,
// 		}

// 		params := &dto.UserAliasParams{
// 			Workspace:              demoWorkspace,
// 			DataImportID:           "data-import-id",
// 			FromUserExternalID:     fromUserExternalID,
// 			ToUserExternalID:       toUserExternalID,
// 			ToUserIsAuthenticated:  true,
// 			ToUserDefaultCreatedAt: time.Now(),
// 		}

// 		rts, code, err := svc.UserAlias(context.Background(), params)

// 		if err != nil {
// 			t.Fatalf("should not fail on existing user_alias, got: %v, %v", code, err)
// 		}
// 		if len(repoMock.CreateUserAliasCalls()) != 0 {
// 			t.Errorf("should not CreateUserAlias, did: %+v\n", len(repoMock.CreateUserAliasCalls()))
// 		}
// 		if len(rts) > 0 {
// 			t.Errorf("should have urts on existing user_alias, got: %+v\n", rts)
// 		}
// 	})

// 	t.Run("should create & alias upserted user when 'from' & 'to' users are missing", func(t *testing.T) {

// 		fromUserID := "from-user"
// 		toUserID := "to-user"

// 		// createdAt := time.Now().AddDate(0, -1, 0)
// 		// called := 0

// 		repoMock := &repository.RepositoryMock{
// 			FindUserByIDFunc: func(ctx context.Context, workspaceID, userID string, tx *sql.Tx) (*entity.User, error) {
// 				return nil, nil
// 			},
// 			FindUserAliasFunc: func(ctx context.Context, fromUserExternalID string, tx *sql.Tx) (*entity.UserAlias, error) {
// 				return nil, nil
// 			},
// 			CreateUserAliasFunc: func(ctx context.Context, fromUserID, toUserID string, toUserIsAuthenticated bool, tx *sql.Tx) error {
// 				return nil
// 			},
// 			InsertUserFunc: func(ctx context.Context, user *entity.User, tx *sql.Tx) (err error) {
// 				return
// 			},
// 			MergeUserSessionsFunc: func(ctx context.Context, workspace *entity.Workspace, fromUserID, toUserID string, tx *sql.Tx) error {
// 				return nil
// 			},
// 			InsertItemTimelineFunc: func(ctx context.Context, userResourceTimeline *entity.ItemTimeline, tx *sql.Tx) error {
// 				return nil
// 			},
// 		}

// 		svc := &ServiceImpl{
// 			Config: &entity.Config{SECRET_KEY: "12345678901234567890123456789012"},
// 			Repo:   repoMock,
// 		}

// 		params := &dto.UserAliasParams{
// 			Workspace:              demoWorkspace,
// 			DataImportID:           "data-import-id",
// 			FromUserExternalID:     fromUserID,
// 			ToUserExternalID:       toUserID,
// 			ToUserIsAuthenticated:  true,
// 			ToUserDefaultCreatedAt: time.Now().AddDate(0, 0, -1),
// 		}

// 		rts, code, err := svc.UserAlias(context.Background(), params)

// 		if err != nil || code != 200 {
// 			t.Fatalf("should not fail, got: %v, %v", code, err)
// 		}
// 		if len(repoMock.InsertUserCalls()) != 1 {
// 			t.Fatalf("should call InsertUser once, did %+v\n", len(repoMock.InsertUserCalls()))
// 		}
// 		if repoMock.InsertUserCalls()[0].User.ExternalID != toUserID {
// 			t.Fatalf("upserted user ext id should be %v, got %v", toUserID, repoMock.InsertUserCalls()[0].User.ExternalID)
// 		}
// 		if !repoMock.InsertUserCalls()[0].User.CreatedAt.Equal(params.ToUserDefaultCreatedAt) {
// 			t.Fatalf("upserted user default date should be %v, got %v", params.ToUserDefaultCreatedAt, repoMock.InsertUserCalls()[0].User.CreatedAt)
// 		}
// 		if len(repoMock.InsertItemTimelineCalls()) != 2 {
// 			t.Fatalf("should call InsertItemTimeline twice, did %+v\n", len(repoMock.InsertItemTimelineCalls()))
// 		}
// 		if len(repoMock.MergeUserSessionsCalls()) > 0 {
// 			t.Fatal("should not merge user timeline for missing users")
// 		}
// 		if len(rts) != 2 {
// 			t.Fatalf("should have 2 urts only (user.create & user.alias), got %+v\n", rts)
// 		}
// 		if rts[0].Kind != "user" {
// 			t.Fatalf("want first kind to be user, got %+v\n", rts[0].Kind)
// 		}
// 		if rts[0].ItemExternalID != toUserID {
// 			t.Fatalf("want urts user id %v, got %+v\n", toUserID, rts[0].ItemExternalID)
// 		}
// 		if rts[0].Action != "create" {
// 			t.Fatalf("want first action to be created, got %+v\n", rts[0].Action)
// 		}
// 		if rts[1].Kind != "user" {
// 			t.Fatalf("want second kind to be user, got %+v\n", rts[1].Kind)
// 		}
// 		if rts[1].Action != "alias" {
// 			t.Fatalf("want second action to be alias, got %+v\n", rts[1].Action)
// 		}
// 	})

// 	// TODO: test merge

// 	// MergeUserSessionsFunc: func(ctx context.Context, workspace *entity.Workspace, fromUserID, toUserID string, tx *sql.Tx) error {
// 	// 	return nil
// 	// },
// 	// MergeUserPageviewsFunc: func(ctx context.Context, workspace *entity.Workspace, fromUserID, toUserID string, tx *sql.Tx) error {
// 	// 	return nil
// 	// },
// 	// MergeUserCartsFunc: func(ctx context.Context, workspace *entity.Workspace, fromUserID, toUserID string, tx *sql.Tx) error {
// 	// 	return nil
// 	// },
// 	// MergeUserOrdersFunc: func(ctx context.Context, workspace *entity.Workspace, fromUserID, toUserID string, tx *sql.Tx) error {
// 	// 	return nil
// 	// },
// }
