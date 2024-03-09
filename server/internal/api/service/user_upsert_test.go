package service

// import (
// 	"context"
// 	"database/sql"
// 	"fmt"
// 	"io/ioutil"
// 	"net/http"
// 	"strings"
// 	"testing"
// 	"time"

// 	"github.com/rimdian/rimdian/internal/api/entity"
// 	"github.com/rimdian/rimdian/internal/api/repository"
// 	"github.com/rimdian/rimdian/internal/common/httpClient"
// )

// func TestServiceImpl_UserUpsert(t *testing.T) {

// 	// https://ozh.github.io/ascii-tables/
// 	// 	CASE	Upserted user is?	Reconciliation keys in payload	Matches with	Which user will remain after merging?	Overwrite upserted user id?
// 	// 1	authenticated	1 key	1 anonymous	the upserted user	no
// 	// 2	anonymous	1 key	1 anonymous	the matched anonymous user	yes
// 	// 3	anonymous	1 key	1 authenticated	the matched authenticated user	yes
// 	// 4	anonymous	2 keys (ie: email + phone)	1 authenticated + 1 anonymous	the matched authenticated user	yes
// 	// 5	anonymous	2 keys (ie: email + phone)	2 anonymous	the matched user with oldest profile	yes

// 	// +------+-------------------+--------------------------------+-------------------------------+---------------------------------------+-----------------------------+
// 	// | CASE | Upserted user is? | Reconciliation keys in payload |         Matches with          | Which user will remain after merging? | Overwrite upserted user id? |
// 	// +------+-------------------+--------------------------------+-------------------------------+---------------------------------------+-----------------------------+
// 	// |    1 | authenticated     | 1 key                          | 1 anonymous                   | the upserted user                     | no                          |
// 	// |    2 | anonymous         | 1 key                          | 1 anonymous                   | the matched anonymous user            | yes                         |
// 	// |    3 | anonymous         | 1 key                          | 1 authenticated               | the matched authenticated user        | yes                         |
// 	// |    4 | anonymous         | 2 keys (ie: email + phone)     | 1 authenticated + 1 anonymous | the matched authenticated user        | yes                         |
// 	// |    5 | anonymous         | 2 keys (ie: email + phone)     | 2 anonymous                   | the matched user with oldest profile  | yes                         |
// 	// +------+-------------------+--------------------------------+-------------------------------+---------------------------------------+-----------------------------+

// 	cfgSecretKey := "12345678901234567890123456789012"
// 	orgID := "testing"
// 	workspaceID := fmt.Sprintf("%v_%v", orgID, "demoecommerce")

// 	demoWorkspace, err := entity.GenerateDemoWorkspace(workspaceID, entity.WorkspaceDemoOrder, orgID, cfgSecretKey)
// 	if err != nil {
// 		t.Fatalf("generate demo workspace err %v", err)
// 	}

// 	t.Run("CASE 1 - upsert authenticated user that matches anonymous user", func(t *testing.T) {
// 		date := time.Now().AddDate(-1, 0, 0)
// 		existingUser := entity.NewUser("anon-user-ext-id", false, date, date, "Europe/Paris", "fr", "FR", nil)
// 		// set DB date to avoid inserting user
// 		existingUser.DBCreatedAt = date
// 		existingUser.DBUpdatedAt = date

// 		repoMock := &repository.RepositoryMock{
// 			FindUserAliasFunc: func(ctx context.Context, fromUserID string, tx *sql.Tx) (*entity.UserAlias, error) {
// 				return nil, nil
// 			},
// 			FindEventualUsersToMergeWithFunc: func(ctx context.Context, withUser *entity.User, withReconciliationKeys entity.MapOfInterfaces, tx *sql.Tx) ([]*entity.User, error) {
// 				return []*entity.User{
// 					// matches one anonymous user
// 					existingUser,
// 				}, nil
// 			},
// 			FindUserByIDFunc: func(ctx context.Context, workspaceID, userID string, tx *sql.Tx) (*entity.User, error) {
// 				return nil, nil
// 			},
// 			CreateUserAliasFunc: func(ctx context.Context, fromUserID, toUserID string, toUserIsAuthenticated bool, tx *sql.Tx) error {
// 				return nil
// 			},
// 			InsertUserFunc: func(ctx context.Context, user *entity.User, tx *sql.Tx) (err error) {
// 				return
// 			},
// 			InsertItemTimelineFunc: func(ctx context.Context, userResourceTimeline *entity.ItemTimeline, tx *sql.Tx) error {
// 				return nil
// 			},
// 		}

// 		clientMock := &httpClient.HTTPClientMock{
// 			DoFunc: func(req *http.Request) (*http.Response, error) {
// 				return &http.Response{
// 					StatusCode: http.StatusOK,
// 					Body:       ioutil.NopCloser(strings.NewReader("ok")),
// 				}, nil
// 			},
// 		}

// 		svc := &ServiceImpl{
// 			Config:    &entity.Config{SECRET_KEY: cfgSecretKey},
// 			Repo:      repoMock,
// 			NetClient: clientMock,
// 		}

// 		email := "auth-user@rimdian.com"

// 		authUserCreatedAt := time.Now().AddDate(0, -1, 0)

// 		upsertedUser := entity.NewUser("auth-user-ext-id", true, authUserCreatedAt, authUserCreatedAt, "Europe/Paris", "fr", "FR", nil)
// 		upsertedUser.Email = entity.NewNullableString(&email)

// 		var tx *sql.Tx

// 		rts, code, err := svc.UserUpsert(context.Background(), demoWorkspace, "data-import-id", nil, upsertedUser, tx)

// 		if err != nil || code != 200 {
// 			t.Fatalf("should not fail, got: %v, %v", code, err)
// 		}

// 		if len(repoMock.FindUserAliasCalls()) != 0 {
// 			t.Fatal("should not FindUserAlias")
// 		}

// 		if len(repoMock.FindEventualUsersToMergeWithCalls()) != 1 {
// 			t.Fatal("should call FindEventualUsersToMergeWith")
// 		}

// 		if len(repoMock.FindEventualUsersToMergeWithCalls()[0].WithReconciliationKeys) != 1 && repoMock.FindEventualUsersToMergeWithCalls()[0].WithReconciliationKeys["email"] != email {
// 			t.Fatalf("should pass 'email' reconciliation key, got %v", repoMock.FindEventualUsersToMergeWithCalls()[0].WithReconciliationKeys)
// 		}

// 		// check net client sent a user_alias data import
// 		if len(clientMock.DoCalls()) != 1 {
// 			t.Fatal("should call DataImportEnqueueInternal")
// 		}

// 		if len(repoMock.InsertUserCalls()) != 1 {
// 			t.Fatal("should call InsertUser")
// 		}

// 		if repoMock.InsertUserCalls()[0].User.ExternalID != upsertedUser.ExternalID {
// 			t.Fatalf("should insert user %v, got %v", upsertedUser.ExternalID, repoMock.InsertUserCalls()[0].User.ExternalID)
// 		}

// 		if len(rts) != 1 {
// 			t.Fatalf("should have 1 rts, got: %v", len(rts))
// 		}

// 		if rts[0].Kind != "user" {
// 			t.Fatalf("want first kind to be user, got %+v\n", rts[0].Kind)
// 		}
// 		if rts[0].ItemExternalID != upsertedUser.ExternalID {
// 			t.Fatalf("want urts user id %v, got %+v\n", upsertedUser.ExternalID, rts[0].ItemExternalID)
// 		}
// 		if rts[0].Action != "create" {
// 			t.Fatalf("want first action to be user new, got %+v\n", rts[0].Action)
// 		}
// 	})

// 	t.Run("CASE 2 - upsert anonymous user that matches anonymous user", func(t *testing.T) {

// 		email := "user@rimdian.com"

// 		date := time.Now().AddDate(-1, 0, 0)
// 		matchedAnon := entity.NewUser("existing-anon-user-ext-id", false, date, date, "Europe/Paris", "fr", "FR", nil)
// 		matchedAnon.Email = entity.NewNullableString(&email)
// 		matchedAnon.FieldsTimestamp["email"] = matchedAnon.CreatedAt
// 		matchedAnon.FirstName = entity.NewNullableString(entity.StringPtr("existing firstame"))
// 		matchedAnon.FieldsTimestamp["first_name"] = matchedAnon.CreatedAt
// 		// set DB date to avoid inserting user
// 		matchedAnon.DBCreatedAt = date
// 		matchedAnon.DBUpdatedAt = date

// 		repoMock := &repository.RepositoryMock{
// 			FindUserAliasFunc: func(ctx context.Context, fromUserID string, tx *sql.Tx) (*entity.UserAlias, error) {
// 				return nil, nil
// 			},
// 			FindEventualUsersToMergeWithFunc: func(ctx context.Context, withUser *entity.User, withReconciliationKeys entity.MapOfInterfaces, tx *sql.Tx) ([]*entity.User, error) {
// 				return []*entity.User{
// 					// matches one anonymous user
// 					matchedAnon,
// 				}, nil
// 			},
// 			FindUserByIDFunc: func(ctx context.Context, workspaceID, userID string, tx *sql.Tx) (*entity.User, error) {
// 				return nil, nil
// 			},
// 			CreateUserAliasFunc: func(ctx context.Context, fromUserID, toUserID string, toUserIsAuthenticated bool, tx *sql.Tx) error {
// 				return nil
// 			},
// 			InsertUserFunc: func(ctx context.Context, user *entity.User, tx *sql.Tx) (err error) {
// 				return
// 			},
// 			UpdateUserFunc: func(ctx context.Context, user *entity.User, tx *sql.Tx) (err error) {
// 				return
// 			},
// 			MergeUserSessionsFunc: func(ctx context.Context, workspace *entity.Workspace, fromUserID, toUserID string, tx *sql.Tx) error {
// 				return nil
// 			},
// 			MergeUserPageviewsFunc: func(ctx context.Context, workspace *entity.Workspace, fromUserID, toUserID string, tx *sql.Tx) error {
// 				return nil
// 			},
// 			MergeUserCartsFunc: func(ctx context.Context, workspace *entity.Workspace, fromUserID, toUserID string, tx *sql.Tx) error {
// 				return nil
// 			},
// 			MergeUserOrdersFunc: func(ctx context.Context, workspace *entity.Workspace, fromUserID, toUserID string, tx *sql.Tx) error {
// 				return nil
// 			},
// 			MergeUserCartItemsFunc: func(ctx context.Context, workspace *entity.Workspace, fromUserID, toUserID string, tx *sql.Tx) error {
// 				return nil
// 			},
// 			InsertItemTimelineFunc: func(ctx context.Context, userResourceTimeline *entity.ItemTimeline, tx *sql.Tx) error {
// 				return nil
// 			},
// 		}

// 		clientMock := &httpClient.HTTPClientMock{
// 			DoFunc: func(req *http.Request) (*http.Response, error) {
// 				return &http.Response{
// 					StatusCode: http.StatusOK,
// 					Body:       ioutil.NopCloser(strings.NewReader("ok")),
// 				}, nil
// 			},
// 		}

// 		svc := &ServiceImpl{Config: &entity.Config{SECRET_KEY: cfgSecretKey}, Repo: repoMock, NetClient: clientMock}

// 		upsertedAnonCreatedAt := time.Now().AddDate(0, -1, 0)

// 		upsertedUser := entity.NewUser("upserted-anon-user-ext-id", false, upsertedAnonCreatedAt, upsertedAnonCreatedAt, "Europe/Paris", "fr", "FR", nil)
// 		upsertedUser.Email = entity.NewNullableString(&email)
// 		upsertedUser.FieldsTimestamp["email"] = upsertedUser.CreatedAt
// 		upsertedUser.LastName = entity.NewNullableString(entity.StringPtr("upserted_lastname"))
// 		upsertedUser.FieldsTimestamp["last_name"] = upsertedUser.CreatedAt

// 		var tx *sql.Tx

// 		rts, code, err := svc.UserUpsert(context.Background(), demoWorkspace, "data-import-id", nil, upsertedUser, tx)

// 		if err != nil || code != 200 {
// 			t.Fatalf("should not fail, got: %v, %v", code, err)
// 		}

// 		if len(repoMock.FindUserAliasCalls()) != 1 {
// 			t.Fatal("should FindUserAlias once")
// 		}

// 		if len(repoMock.FindEventualUsersToMergeWithCalls()) != 1 {
// 			t.Fatal("should call FindEventualUsersToMergeWith")
// 		}

// 		if len(repoMock.FindEventualUsersToMergeWithCalls()[0].WithReconciliationKeys) != 1 && repoMock.FindEventualUsersToMergeWithCalls()[0].WithReconciliationKeys["email"] != email {
// 			t.Fatalf("should pass 'email' reconciliation key, got %v", repoMock.FindEventualUsersToMergeWithCalls()[0].WithReconciliationKeys)
// 		}

// 		// check net client sent a user_alias data import
// 		if len(clientMock.DoCalls()) != 1 {
// 			t.Fatal("should call DataImportEnqueueInternal")
// 		}

// 		if len(repoMock.InsertUserCalls()) != 1 {
// 			t.Fatal("should call InsertUser")
// 		}

// 		if repoMock.InsertUserCalls()[0].User.ExternalID != upsertedUser.ExternalID {
// 			t.Fatalf("should insert user %v, got %v", upsertedUser.ExternalID, repoMock.InsertUserCalls()[0].User.ExternalID)
// 		}

// 		if len(rts) != 1 {
// 			t.Fatalf("should have 1 rts, got: %v", len(rts))
// 		}

// 		if rts[0].Kind != "user" {
// 			t.Fatalf("want first change to be user, got %+v\n", rts[0].Kind)
// 		}
// 		if rts[0].ItemExternalID != upsertedUser.ExternalID {
// 			t.Fatalf("want urts user id %v, got %+v\n", upsertedUser.ExternalID, rts[0].ItemExternalID)
// 		}
// 		if rts[0].Action != "create" {
// 			t.Fatalf("want first action to be create, got %+v\n", rts[0].Action)
// 		}
// 	})

// 	t.Run("CASE 3 - upsert anonymous user that matches authenticated user", func(t *testing.T) {

// 		email := "user@rimdian.com"
// 		date := time.Now().AddDate(-1, 0, 0)

// 		matchedAuth := entity.NewUser("existing-auth-user-ext-id", true, date, date, "Europe/Paris", "fr", "FR", nil)
// 		matchedAuth.Email = entity.NewNullableString(&email)
// 		matchedAuth.FieldsTimestamp["email"] = matchedAuth.CreatedAt
// 		matchedAuth.FirstName = entity.NewNullableString(entity.StringPtr("existing firstame"))
// 		matchedAuth.FieldsTimestamp["first_name"] = matchedAuth.CreatedAt
// 		// set DB date to avoid inserting user
// 		matchedAuth.DBCreatedAt = date
// 		matchedAuth.DBUpdatedAt = date

// 		repoMock := &repository.RepositoryMock{
// 			FindUserAliasFunc: func(ctx context.Context, fromUserID string, tx *sql.Tx) (*entity.UserAlias, error) {
// 				return nil, nil
// 			},
// 			FindEventualUsersToMergeWithFunc: func(ctx context.Context, withUser *entity.User, withReconciliationKeys entity.MapOfInterfaces, tx *sql.Tx) ([]*entity.User, error) {
// 				return []*entity.User{
// 					// matches one auth user
// 					matchedAuth,
// 				}, nil
// 			},
// 			FindUserByIDFunc: func(ctx context.Context, workspaceID, userID string, tx *sql.Tx) (*entity.User, error) {
// 				return nil, nil
// 			},
// 			CreateUserAliasFunc: func(ctx context.Context, fromUserID, toUserID string, toUserIsAuthenticated bool, tx *sql.Tx) error {
// 				return nil
// 			},
// 			InsertUserFunc: func(ctx context.Context, user *entity.User, tx *sql.Tx) (err error) {
// 				return
// 			},
// 			UpdateUserFunc: func(ctx context.Context, user *entity.User, tx *sql.Tx) (err error) {
// 				return
// 			},
// 			MergeUserSessionsFunc: func(ctx context.Context, workspace *entity.Workspace, fromUserID, toUserID string, tx *sql.Tx) error {
// 				return nil
// 			},
// 			MergeUserPageviewsFunc: func(ctx context.Context, workspace *entity.Workspace, fromUserID, toUserID string, tx *sql.Tx) error {
// 				return nil
// 			},
// 			MergeUserCartsFunc: func(ctx context.Context, workspace *entity.Workspace, fromUserID, toUserID string, tx *sql.Tx) error {
// 				return nil
// 			},
// 			MergeUserOrdersFunc: func(ctx context.Context, workspace *entity.Workspace, fromUserID, toUserID string, tx *sql.Tx) error {
// 				return nil
// 			},
// 			MergeUserCartItemsFunc: func(ctx context.Context, workspace *entity.Workspace, fromUserID, toUserID string, tx *sql.Tx) error {
// 				return nil
// 			},
// 			InsertItemTimelineFunc: func(ctx context.Context, userResourceTimeline *entity.ItemTimeline, tx *sql.Tx) error {
// 				return nil
// 			},
// 		}

// 		clientMock := &httpClient.HTTPClientMock{
// 			DoFunc: func(req *http.Request) (*http.Response, error) {
// 				return &http.Response{
// 					StatusCode: http.StatusOK,
// 					Body:       ioutil.NopCloser(strings.NewReader("ok")),
// 				}, nil
// 			},
// 		}

// 		svc := &ServiceImpl{Config: &entity.Config{SECRET_KEY: cfgSecretKey}, Repo: repoMock, NetClient: clientMock}

// 		upsertedAnonCreatedAt := time.Now().AddDate(0, -1, 0)

// 		upsertedUser := entity.NewUser("upserted-anon-user-ext-id", false, upsertedAnonCreatedAt, upsertedAnonCreatedAt, "Europe/Paris", "fr", "FR", nil)
// 		upsertedUser.Email = entity.NewNullableString(&email)
// 		upsertedUser.FieldsTimestamp["email"] = upsertedUser.CreatedAt
// 		upsertedUser.LastName = entity.NewNullableString(entity.StringPtr("upserted_lastname"))
// 		upsertedUser.FieldsTimestamp["last_name"] = upsertedUser.CreatedAt

// 		var tx *sql.Tx

// 		rts, code, err := svc.UserUpsert(context.Background(), demoWorkspace, "data-import-id", nil, upsertedUser, tx)

// 		if err != nil || code != 200 {
// 			t.Fatalf("should not fail, got: %v, %v", code, err)
// 		}

// 		if len(repoMock.FindUserAliasCalls()) != 1 {
// 			t.Fatal("should FindUserAlias once")
// 		}

// 		if len(repoMock.FindEventualUsersToMergeWithCalls()) != 1 {
// 			t.Fatal("should call FindEventualUsersToMergeWith")
// 		}

// 		if len(repoMock.FindEventualUsersToMergeWithCalls()[0].WithReconciliationKeys) != 1 && repoMock.FindEventualUsersToMergeWithCalls()[0].WithReconciliationKeys["email"] != email {
// 			t.Fatalf("should pass 'email' reconciliation key, got %v", repoMock.FindEventualUsersToMergeWithCalls()[0].WithReconciliationKeys)
// 		}

// 		// check net client sent a user_alias data import
// 		if len(clientMock.DoCalls()) != 1 {
// 			t.Fatal("should call DataImportEnqueueInternal")
// 		}

// 		if len(repoMock.InsertUserCalls()) != 1 {
// 			t.Fatal("should call InsertUser")
// 		}

// 		if len(rts) != 1 {
// 			t.Fatalf("should have 1 rts, got: %v", len(rts))
// 		}

// 		if rts[0].Kind != "user" {
// 			t.Fatalf("want first change to be user, got %+v\n", rts[0].Kind)
// 		}
// 		if rts[0].ItemExternalID != upsertedUser.ExternalID {
// 			t.Fatalf("want urts user id %v, got %+v\n", upsertedUser.ExternalID, rts[0].ItemExternalID)
// 		}
// 		if rts[0].Action != "create" {
// 			t.Fatalf("want first action to be create, got %+v\n", rts[0].Action)
// 		}
// 	})

// 	t.Run("CASE 4 - upsert anonymous user that matches authenticated and anonymous user with 2 different reconciliation keys", func(t *testing.T) {

// 		email := "user@rimdian.com"
// 		telephone := "+33671827366"
// 		date := time.Now().AddDate(-1, 0, 0)

// 		matchedAuth := entity.NewUser("existing-auth-user-ext-id", true, date, date, "Europe/Paris", "fr", "FR", nil)
// 		matchedAuth.Email = entity.NewNullableString(&email)
// 		matchedAuth.FieldsTimestamp["email"] = matchedAuth.CreatedAt
// 		matchedAuth.FirstName = entity.NewNullableString(entity.StringPtr("existing firstame"))
// 		matchedAuth.FieldsTimestamp["first_name"] = matchedAuth.CreatedAt
// 		// set DB date to avoid inserting user
// 		matchedAuth.DBCreatedAt = date
// 		matchedAuth.DBUpdatedAt = date

// 		matchedAnon := entity.NewUser("existing-anon-user-ext-id", false, date, date, "Europe/Paris", "fr", "FR", nil)
// 		matchedAnon.Telephone = entity.NewNullableString(&telephone)
// 		matchedAnon.FieldsTimestamp["telephone"] = matchedAnon.CreatedAt
// 		matchedAnon.LastName = entity.NewNullableString(entity.StringPtr("existing last name"))
// 		matchedAnon.FieldsTimestamp["last_name"] = matchedAnon.CreatedAt
// 		// set DB date to avoid inserting user
// 		matchedAnon.DBCreatedAt = date
// 		matchedAnon.DBUpdatedAt = date

// 		repoMock := &repository.RepositoryMock{
// 			FindUserAliasFunc: func(ctx context.Context, fromUserID string, tx *sql.Tx) (*entity.UserAlias, error) {
// 				return nil, nil
// 			},
// 			FindEventualUsersToMergeWithFunc: func(ctx context.Context, withUser *entity.User, withReconciliationKeys entity.MapOfInterfaces, tx *sql.Tx) ([]*entity.User, error) {
// 				return []*entity.User{
// 					// matches one auth user and one anon user
// 					matchedAnon,
// 					matchedAuth,
// 				}, nil
// 			},
// 			FindUserByIDFunc: func(ctx context.Context, workspaceID, userID string, tx *sql.Tx) (*entity.User, error) {
// 				return nil, nil
// 			},
// 			CreateUserAliasFunc: func(ctx context.Context, fromUserID, toUserID string, toUserIsAuthenticated bool, tx *sql.Tx) error {
// 				return nil
// 			},
// 			InsertUserFunc: func(ctx context.Context, user *entity.User, tx *sql.Tx) (err error) {
// 				return
// 			},
// 			UpdateUserFunc: func(ctx context.Context, user *entity.User, tx *sql.Tx) (err error) {
// 				return
// 			},
// 			MergeUserSessionsFunc: func(ctx context.Context, workspace *entity.Workspace, fromUserID, toUserID string, tx *sql.Tx) error {
// 				return nil
// 			},
// 			MergeUserPageviewsFunc: func(ctx context.Context, workspace *entity.Workspace, fromUserID, toUserID string, tx *sql.Tx) error {
// 				return nil
// 			},
// 			MergeUserCartsFunc: func(ctx context.Context, workspace *entity.Workspace, fromUserID, toUserID string, tx *sql.Tx) error {
// 				return nil
// 			},
// 			MergeUserOrdersFunc: func(ctx context.Context, workspace *entity.Workspace, fromUserID, toUserID string, tx *sql.Tx) error {
// 				return nil
// 			},
// 			MergeUserCartItemsFunc: func(ctx context.Context, workspace *entity.Workspace, fromUserID, toUserID string, tx *sql.Tx) error {
// 				return nil
// 			},
// 			InsertItemTimelineFunc: func(ctx context.Context, userResourceTimeline *entity.ItemTimeline, tx *sql.Tx) error {
// 				return nil
// 			},
// 		}

// 		netClientMock := httpClient.HTTPClientMock{
// 			DoFunc: func(req *http.Request) (*http.Response, error) {
// 				return nil, nil
// 			},
// 		}

// 		svc := &ServiceImpl{Config: &entity.Config{SECRET_KEY: cfgSecretKey}, Repo: repoMock, NetClient: &netClientMock}

// 		upsertedAnonCreatedAt := time.Now().AddDate(0, -1, 0)
// 		upsertedUser := entity.NewUser("upserted-anon-user-ext-id", false, upsertedAnonCreatedAt, upsertedAnonCreatedAt, "Europe/Paris", "fr", "FR", nil)
// 		upsertedUser.Email = entity.NewNullableString(&email)
// 		upsertedUser.FieldsTimestamp["email"] = upsertedUser.CreatedAt
// 		upsertedUser.Telephone = entity.NewNullableString(&telephone)
// 		upsertedUser.FieldsTimestamp["telephone"] = upsertedUser.CreatedAt
// 		upsertedUser.Gender = entity.NewNullableString(entity.StringPtr("male"))
// 		upsertedUser.FieldsTimestamp["gender"] = upsertedUser.CreatedAt

// 		var tx *sql.Tx

// 		rts, code, err := svc.UserUpsert(context.Background(), demoWorkspace, "data-import-id", nil, upsertedUser, tx)

// 		if err != nil || code != 200 {
// 			t.Fatalf("should not fail, got: %v, %v", code, err)
// 		}

// 		if len(repoMock.FindUserAliasCalls()) != 1 {
// 			t.Fatal("should FindUserAlias once")
// 		}

// 		if len(repoMock.FindEventualUsersToMergeWithCalls()) != 1 {
// 			t.Fatal("should call FindEventualUsersToMergeWith")
// 		}

// 		if len(repoMock.FindEventualUsersToMergeWithCalls()[0].WithReconciliationKeys) != 2 && repoMock.FindEventualUsersToMergeWithCalls()[0].WithReconciliationKeys["email"] != email && repoMock.FindEventualUsersToMergeWithCalls()[0].WithReconciliationKeys["telephone"] != telephone {
// 			t.Fatalf("should pass 'email' and 'telephone' reconciliation keys, got %v", repoMock.FindEventualUsersToMergeWithCalls()[0].WithReconciliationKeys)
// 		}

// 		if len(repoMock.InsertUserCalls()) != 1 {
// 			t.Fatal("should call InsertUser")
// 		}

// 		if len(netClientMock.DoCalls()) != 1 {
// 			t.Fatalf("should send a data import call to merge the other user in async, got: %v", len(netClientMock.DoCalls()))
// 		}

// 		if len(rts) != 1 {
// 			t.Fatalf("should have 1 rts, got: %v", len(rts))
// 		}

// 		if rts[0].Kind != "user" {
// 			t.Fatalf("want first kind to be user, got %+v\n", rts[0].Kind)
// 		}
// 		if rts[0].ItemExternalID != upsertedUser.ExternalID {
// 			t.Fatalf("want urts user id %v, got %+v\n", upsertedUser.ExternalID, rts[0].ItemExternalID)
// 		}
// 		if rts[0].Action != "create" {
// 			t.Fatalf("want first action to be update, got %+v\n", rts[0].Action)
// 		}
// 	})

// 	t.Run("CASE 5 - upsert anonymous user that matches 2 anonymous users with 2 different reconciliation keys", func(t *testing.T) {

// 		email := "user@rimdian.com"
// 		telephone := "+33671827366"

// 		// anon1 is oldest, it will remain
// 		date1 := time.Now().AddDate(-2, 0, 0)
// 		matchedAnon1 := entity.NewUser("existing-anon1-user-ext-id", true, date1, date1, "Europe/Paris", "fr", "FR", nil)
// 		matchedAnon1.Email = entity.NewNullableString(&email)
// 		matchedAnon1.FieldsTimestamp["email"] = matchedAnon1.CreatedAt
// 		matchedAnon1.FirstName = entity.NewNullableString(entity.StringPtr("existing firstame"))
// 		matchedAnon1.FieldsTimestamp["first_name"] = matchedAnon1.CreatedAt
// 		// set DB date to avoid inserting user
// 		matchedAnon1.DBCreatedAt = date1
// 		matchedAnon1.DBUpdatedAt = date1

// 		date2 := time.Now().AddDate(-1, 0, 0)
// 		matchedAnon2 := entity.NewUser("existing-anon2-user-ext-id", false, date2, date2, "Europe/Paris", "fr", "FR", nil)
// 		matchedAnon2.Telephone = entity.NewNullableString(&telephone)
// 		matchedAnon2.FieldsTimestamp["telephone"] = matchedAnon2.CreatedAt
// 		matchedAnon2.LastName = entity.NewNullableString(entity.StringPtr("existing last name"))
// 		matchedAnon2.FieldsTimestamp["last_name"] = matchedAnon2.CreatedAt
// 		// set DB date to avoid inserting user
// 		matchedAnon2.DBCreatedAt = date2
// 		matchedAnon2.DBUpdatedAt = date2

// 		repoMock := &repository.RepositoryMock{
// 			FindUserAliasFunc: func(ctx context.Context, fromUserID string, tx *sql.Tx) (*entity.UserAlias, error) {
// 				return nil, nil
// 			},
// 			FindEventualUsersToMergeWithFunc: func(ctx context.Context, withUser *entity.User, withReconciliationKeys entity.MapOfInterfaces, tx *sql.Tx) ([]*entity.User, error) {
// 				return []*entity.User{
// 					// matches 2 anon users
// 					matchedAnon1,
// 					matchedAnon2,
// 				}, nil
// 			},
// 			FindUserByIDFunc: func(ctx context.Context, workspaceID, userID string, tx *sql.Tx) (*entity.User, error) {
// 				return nil, nil
// 			},
// 			CreateUserAliasFunc: func(ctx context.Context, fromUserID, toUserID string, toUserIsAuthenticated bool, tx *sql.Tx) error {
// 				return nil
// 			},
// 			InsertUserFunc: func(ctx context.Context, user *entity.User, tx *sql.Tx) (err error) {
// 				return
// 			},
// 			UpdateUserFunc: func(ctx context.Context, user *entity.User, tx *sql.Tx) (err error) {
// 				return
// 			},
// 			MergeUserSessionsFunc: func(ctx context.Context, workspace *entity.Workspace, fromUserID, toUserID string, tx *sql.Tx) error {
// 				return nil
// 			},
// 			MergeUserPageviewsFunc: func(ctx context.Context, workspace *entity.Workspace, fromUserID, toUserID string, tx *sql.Tx) error {
// 				return nil
// 			},
// 			MergeUserCartsFunc: func(ctx context.Context, workspace *entity.Workspace, fromUserID, toUserID string, tx *sql.Tx) error {
// 				return nil
// 			},
// 			MergeUserOrdersFunc: func(ctx context.Context, workspace *entity.Workspace, fromUserID, toUserID string, tx *sql.Tx) error {
// 				return nil
// 			},
// 			MergeUserCartItemsFunc: func(ctx context.Context, workspace *entity.Workspace, fromUserID, toUserID string, tx *sql.Tx) error {
// 				return nil
// 			},
// 			InsertItemTimelineFunc: func(ctx context.Context, userResourceTimeline *entity.ItemTimeline, tx *sql.Tx) error {
// 				return nil
// 			},
// 		}

// 		netClientMock := httpClient.HTTPClientMock{
// 			DoFunc: func(req *http.Request) (*http.Response, error) {
// 				return nil, nil
// 			},
// 		}

// 		svc := &ServiceImpl{Config: &entity.Config{SECRET_KEY: cfgSecretKey}, Repo: repoMock, NetClient: &netClientMock}

// 		upsertedAnonCreatedAt := time.Now().AddDate(0, -1, 0)

// 		upsertedUser := entity.NewUser("upserted-anon-user-ext-id", false, upsertedAnonCreatedAt, upsertedAnonCreatedAt, "Europe/Paris", "fr", "FR", nil)
// 		upsertedUser.Email = entity.NewNullableString(&email)
// 		upsertedUser.FieldsTimestamp["email"] = upsertedUser.CreatedAt
// 		upsertedUser.Telephone = entity.NewNullableString(&telephone)
// 		upsertedUser.FieldsTimestamp["telephone"] = upsertedUser.CreatedAt
// 		upsertedUser.Gender = entity.NewNullableString(entity.StringPtr("male"))
// 		upsertedUser.FieldsTimestamp["gender"] = upsertedUser.CreatedAt

// 		var tx *sql.Tx

// 		rts, code, err := svc.UserUpsert(context.Background(), demoWorkspace, "data-import-id", nil, upsertedUser, tx)

// 		if err != nil || code != 200 {
// 			t.Fatalf("should not fail, got: %v, %v", code, err)
// 		}

// 		if len(repoMock.FindUserAliasCalls()) != 1 {
// 			t.Fatal("should FindUserAlias once")
// 		}

// 		if len(repoMock.FindEventualUsersToMergeWithCalls()) != 1 {
// 			t.Fatal("should call FindEventualUsersToMergeWith")
// 		}

// 		if len(repoMock.FindEventualUsersToMergeWithCalls()[0].WithReconciliationKeys) != 2 && repoMock.FindEventualUsersToMergeWithCalls()[0].WithReconciliationKeys["email"] != email && repoMock.FindEventualUsersToMergeWithCalls()[0].WithReconciliationKeys["telephone"] != telephone {
// 			t.Fatalf("should pass 'email' and 'telephone' reconciliation keys, got %v", repoMock.FindEventualUsersToMergeWithCalls()[0].WithReconciliationKeys)
// 		}

// 		// check net client sent a user_alias data import
// 		if len(netClientMock.DoCalls()) != 1 {
// 			t.Fatal("should call DataImportEnqueueInternal")
// 		}

// 		if len(repoMock.InsertUserCalls()) != 1 {
// 			t.Fatal("should call InsertUser")
// 		}

// 		if len(netClientMock.DoCalls()) != 1 {
// 			t.Fatalf("should send a data import call to merge the other user in async, got: %v", len(netClientMock.DoCalls()))
// 		}

// 		if len(rts) != 1 {
// 			t.Fatalf("should have 1 rts, got: %v", len(rts))
// 		}

// 		if rts[0].Kind != "user" {
// 			t.Fatalf("want first change to be user, got %+v\n", rts[0].Kind)
// 		}
// 		if rts[0].ItemExternalID != upsertedUser.ExternalID {
// 			t.Fatalf("want urts user id %v, got %+v\n", upsertedUser.ExternalID, rts[0].ItemExternalID)
// 		}
// 		if rts[0].Action != "create" {
// 			t.Fatalf("want first action to be create, got %+v\n", rts[0].Action)
// 		}
// 	})
// }
