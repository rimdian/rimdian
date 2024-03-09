package integrationtests_test

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	commonDTO "github.com/rimdian/rimdian/internal/common/dto"
	"github.com/rimdian/rimdian/package/api"
	"github.com/sirupsen/logrus"
	"github.com/teris-io/shortid"
	"github.com/tidwall/sjson"
)

var (
	accountID   = "root"
	orgID       = "acme"
	workspaceID = "acme_testing"
	workspace   *entity.Workspace
	userAgent   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36"

	sid, _ = shortid.New(1, shortid.DefaultABC, 777)

	testJSON = entity.M{
		"json_string":  "string",
		"json_number":  123.456,
		"json_boolean": true,
		"json_null":    nil,
		"json_array":   []string{"a", "b", "c"},
		"json_object": entity.M{
			"a": 1,
			"b": 2,
		},
	}
	testJSONByte, _ = json.Marshal(testJSON)
	testDate        = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	testDateTime    = time.Date(2020, 1, 1, 12, 34, 56, 0, time.UTC)
)

func TestIntegration_InitDB(t *testing.T) {
	ctx := context.Background()
	log := logrus.New()

	// instantiate a repo + service + router and assemble a web server
	repo, cfg, err := api.NewRepository(ctx, log)
	if err != nil {
		t.Fatalf("cannot create repo: %v", err)
	}

	svc, err := api.NewService(ctx, log, cfg, repo)
	if err != nil {
		t.Fatalf("cannot create svc: %v", err)
	}

	t.Run("should reset DB before testing", func(t *testing.T) {
		err = svc.DevResetDB(ctx)
		if err != nil {
			t.Fatalf("cannot reset DB: %v", err)
		}
	})
}

func TestIntegration_CreateWorkspace(t *testing.T) {
	ctx := context.Background()
	log := logrus.New()

	// instantiate a repo + service + router and assemble a web server
	repo, cfg, err := api.NewRepository(ctx, log)
	if err != nil {
		t.Fatalf("cannot create repo: %v", err)
	}

	svc, err := api.NewService(ctx, log, cfg, repo)
	if err != nil {
		t.Fatalf("cannot create a svc: %v", err)
	}

	t.Run("should create workspace: testing", func(t *testing.T) {

		workspace, _, err = svc.WorkspaceCreate(ctx, accountID, &dto.WorkspaceCreate{
			ID:                  "testing",
			Name:                "testing",
			WebsiteURL:          "https://testing.com",
			PrivacyPolicyURL:    "https://testing.com",
			Industry:            "other",
			Currency:            "AUD", // use non-EUR currency for testing fx convertions
			OrganizationID:      orgID,
			DefaultUserTimezone: "Europe/Paris",
			DefaultUserCountry:  "FR",
			DefaultUserLanguage: "en",
		})

		if err != nil {
			t.Errorf("cannot create workspace: %v", err)
		}

		if workspace == nil {
			t.Errorf("workspace is nil")
		}
	})

	t.Run("should configure workspace domain", func(t *testing.T) {
		workspace, _, err = svc.DomainUpsert(ctx, accountID, &dto.Domain{
			ID:              "web",
			WorkspaceID:     workspace.ID,
			Type:            entity.DomainWeb,
			Name:            "website",
			Hosts:           []*entity.DomainHost{{Host: "testing.com"}},
			ParamsWhitelist: []string{"category"},
		})

		if err != nil {
			t.Errorf("cannot create workspace: %v", err)
		}
	})
}

func TestIntegration_InstallApp(t *testing.T) {
	ctx := context.Background()
	log := logrus.New()

	// instantiate a repo + service + router and assemble a web server
	repo, cfg, err := api.NewRepository(ctx, log)
	if err != nil {
		t.Fatalf("cannot create repo: %v", err)
	}

	svc, err := api.NewService(ctx, log, cfg, repo)
	if err != nil {
		t.Fatalf("cannot create a svc: %v", err)
	}

	t.Run("should install app test", func(t *testing.T) {
		_, _, err := svc.AppInstall(ctx, accountID, &dto.AppInstall{
			WorkspaceID: workspaceID,
			Manifest:    &entity.AppManifestTest,
		})

		if err != nil {
			t.Errorf("cannot install app test: %v", err)
		}
	})
	t.Run("should activate app test", func(t *testing.T) {
		_, err := svc.AppActivate(ctx, accountID, &dto.AppActivate{
			WorkspaceID: workspaceID,
			ID:          entity.AppManifestTest.ID,
		})

		if err != nil {
			t.Errorf("cannot activate app test: %v", err)
		}
	})
}

func TestIntegration_ImportAppData(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()

	// instantiate a repo + service + router and assemble a web server
	repo, cfg, err := api.NewRepository(ctx, logger)
	if err != nil {
		t.Fatalf("cannot create repo: %v", err)
	}

	svc, err := api.NewService(ctx, logger, cfg, repo)
	if err != nil {
		t.Fatalf("cannot create svc: %v", err)
	}

	t.Run("should import data", func(t *testing.T) {

		createdAt := time.Now().AddDate(0, 0, -1).Format(time.RFC3339)
		updatedAt := time.Now().Format(time.RFC3339)
		timestamp := time.Now().Unix()

		items := []string{
			fmt.Sprintf(`{
				"kind": "%v",
				"%v": {
					"external_id": "xyz",
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
					"external_id": "user_pv",
					"is_authenticated": false,
					"created_at": "%v"
				}
			}`,
				entity.AppManifestTest.AppTables[0].Name,
				entity.AppManifestTest.AppTables[0].Name,
				createdAt,
				updatedAt,
				timestamp,
				createdAt,
			),

			fmt.Sprintf(`{
					"kind": "%v",
					"%v": {
						"external_id": "xyz",
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
						},
						"optional_varchar": "updated"
					},
					"user": {
						"external_id": "user_pv",
						"is_authenticated": false,
						"created_at": "%v"
					}
			}`,
				entity.AppManifestTest.AppTables[0].Name,
				entity.AppManifestTest.AppTables[0].Name,
				createdAt,
				updatedAt,
				timestamp,
				createdAt,
			),
		}

		for _, item := range items {
			dataLogInqueue := &commonDTO.DataLogInQueue{
				Origin: commonDTO.DataLogOriginClient,
				Context: commonDTO.DataLogContext{
					WorkspaceID:      workspaceID,
					HeadersAndParams: commonDTO.MapOfStrings{"Origin": "https://testing.com"},
					IP:               "0.0.0.0",
					ReceivedAt:       time.Now().UTC(),
				},
				Item: item,
			}

			// compute ID from body
			dataLogInqueue.ID = commonDTO.ComputeDataLogID(cfg.SECRET_KEY, commonDTO.DataLogOriginClient, item)

			result := svc.DataLogImportFromQueue(ctx, dataLogInqueue)

			if result == nil {
				t.Fatalf("result is nil")
			}

			if result.HasError || result.Error != "" {
				t.Errorf("error: %v", result.Error)
			}
		}
	})
}

func TestIntegration_ImportData(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()

	// instantiate a repo + service + router and assemble a web server
	repo, cfg, err := api.NewRepository(ctx, logger)
	if err != nil {
		t.Fatalf("cannot create repo: %v", err)
	}

	svc, err := api.NewService(ctx, logger, cfg, repo)
	if err != nil {
		t.Fatalf("cannot create svc: %v", err)
	}

	t.Run("should import data", func(t *testing.T) {

		batchTime := time.Now()
		// substract 5 secs
		batchTime = batchTime.Add(-5 * time.Second)

		domainID := "web"
		userExternalID := sid.MustGenerate()
		deviceExternalID := sid.MustGenerate()
		pageviewExternalID := sid.MustGenerate()
		sessionExternalID := "session-test-extra-columns"
		orderExternalID := sid.MustGenerate()
		// cartExternalID := sid.MustGenerate()

		pageview := fmt.Sprintf(`{
				"kind": "pageview",
				"pageview": {
					"external_id": "%v",
					"session_external_id": "%v",
					"domain_id": "%v",
					"page_id": "https://testing.com/page-1",
					"created_at": "%v",
					"title": "Page 1",
					"referrer": "https://google.com/"
				},
				"user": {
					"external_id": "%v",
					"is_authenticated": false,
					"created_at": "%v"
				},
				"device": {
					"external_id": "%v",
					"created_at": "%v",
					"user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:121.0) Gecko/20100101 Firefox/121.0",
					"resolution": "1980x1024",
					"language": "en"
				},
				"session": {
					"external_id": "%v",
					"created_at": "%v",
					"timezone": "Europe/Paris",
					"device_external_id": "%v",
					"domain_id": "%v",
					"landing_page": "https://testing.com/page-1",
					"referrer": "https://google.com/",
					"duration": 15,
					"pageviews_count": 1,
					"interactions_count": 1,
					"utm_source": "google.com",
					"utm_medium": "ads",
					"utm_id": "123",
					"utm_id_from": "gclid",
					"utm_campaign": "campaign-1",
					"app_test_string": "abc",
					"app_test_bool": true,
					"app_test_number": 123.456,
					"app_test_date": "%v",
					"app_test_datetime": "%v",
					"app_test_longtext": "abcd",
					"app_test_json": %v
				}
			}`,
			pageviewExternalID,
			sessionExternalID,
			domainID,
			batchTime.Format(time.RFC3339),
			userExternalID,
			batchTime.Format(time.RFC3339),
			deviceExternalID,
			batchTime.Format(time.RFC3339),
			sessionExternalID,
			batchTime.Format(time.RFC3339),
			deviceExternalID,
			domainID,
			testDate.Format("2006-01-02"),
			testDateTime.Format(time.RFC3339),
			string(testJSONByte),
		)

		pageviewWithDuration := pageview
		// second pageview has duration
		pageviewWithDuration, err := sjson.Set(pageviewWithDuration, "pageview.duration", 30)
		if err != nil {
			t.Fatalf("cannot set pageview duration: %v", err)
		}
		pageviewWithDuration, err = sjson.Set(pageviewWithDuration, "session.duration", 60)
		if err != nil {
			t.Fatalf("cannot set session duration: %v", err)
		}
		pageviewWithDuration, err = sjson.Set(pageviewWithDuration, "session.pageviews_count", 2)
		if err != nil {
			t.Fatalf("cannot set pageviews_count: %v", err)
		}
		pageviewWithDuration, err = sjson.Set(pageviewWithDuration, "session.interactions_count", 2)
		if err != nil {
			t.Fatalf("cannot set interactions_count: %v", err)
		}
		pageviewWithDuration, err = sjson.Set(pageviewWithDuration, "session.updated_at", batchTime.Add(2*time.Second).Format(time.RFC3339))
		if err != nil {
			t.Fatalf("cannot set session.updated_at: %v", err)
		}

		orderWithoutSession := fmt.Sprintf(`{
				"kind": "order",
				"order": {
					"external_id": "%v",
					"domain_id": "%v",
					"created_at": "%v",
					"subtotal_price": 15000,
					"total_price": 30000,
					"discount_codes": ["CODE1", "CODE2"],
					"currency": "GBP",
					"items": [
						{
							"external_id": "%v",
							"name": "Product 1",
							"product_external_id": "1",
							"price": 15000,
							"quantity": 1
						},
						{
							"external_id": "%v",
							"name": "Product 2",
							"product_external_id": "2",
							"price": 7500,
							"quantity": 2
						}
					]
				},
				"user": {
					"external_id": "%v",
					"is_authenticated": true,
					"signed_up_at": "%v",
					"created_at": "%v"
				}
			}`,
			orderExternalID,
			domainID,
			batchTime.Format(time.RFC3339),
			sid.MustGenerate(),
			sid.MustGenerate(),
			userExternalID,
			batchTime.Format(time.RFC3339),
			batchTime.Format(time.RFC3339),
		)

		for _, item := range []string{pageview, pageviewWithDuration, orderWithoutSession} {
			dataLogInqueue := &commonDTO.DataLogInQueue{
				Origin: commonDTO.DataLogOriginClient,
				Context: commonDTO.DataLogContext{
					WorkspaceID:      workspaceID,
					HeadersAndParams: commonDTO.MapOfStrings{"Origin": "https://testing.com"},
					IP:               "0.0.0.0",
					ReceivedAt:       time.Now().UTC(),
				},
				Item: item,
			}
			// compute ID from body
			dataLogInqueue.ComputeID(cfg.SECRET_KEY)

			result := svc.DataLogImportFromQueue(ctx, dataLogInqueue)

			if result == nil {
				t.Fatalf("result is nil")
			}

			if result.HasError || result.Error != "" {
				t.Fatalf("error: %v", result.Error)
			}
		}
	})

	t.Run("should fetch session with extra columns", func(t *testing.T) {

		sessionExternalID := "session-test-extra-columns"

		// fetch workspace
		workspace, err := repo.GetWorkspace(ctx, workspaceID)
		if err != nil {
			t.Fatalf("cannot fetch workspace: %v", err)
		}

		selectBuilder := squirrel.Select("*").From(entity.ItemKindSession).Where(squirrel.Eq{"external_id": sessionExternalID})

		sessions, err := repo.FetchSessions(ctx, workspace, selectBuilder, nil)

		if err != nil {
			t.Fatalf("cannot fetch sessions: %v", err)
		}

		if sessions == nil {
			t.Fatalf("sessions is nil")
		}

		if len(sessions) != 1 {
			t.Fatalf("sessions is not correct: %+v\n", sessions)
		}

		for _, session := range sessions {
			for k, v := range session.ExtraColumns {
				t.Logf("extra column: %v: %+v\n", k, v)
			}
		}
		session := sessions[0]

		if session.ExternalID != sessionExternalID {
			t.Errorf("session.ExternalID is not correct: %v", session.ExternalID)
		}

		// check session duration
		if session.Duration.Int64 != 60 {
			t.Errorf("session.Duration is not correct: %v, expect 60", session.Duration.Int64)
		}

		// check session interactions count
		if *session.PageviewsCount != 2 {
			t.Errorf("session.PageviewsCount is not correct: %v, expect 2", session.PageviewsCount)
		}
		if *session.InteractionsCount != 2 {
			t.Errorf("session.InteractionsCount is not correct: %v, expect 2", session.InteractionsCount)
		}

		// check "app_test_string"
		if _, ok := session.ExtraColumns["app_test_string"]; !ok {
			t.Errorf("session.ExtraColumns[\"app_test_string\"] not found in: %+v\n", session.ExtraColumns)
		}

		if session.ExtraColumns["app_test_string"].StringValue.String != "abc" {
			t.Errorf("app_test_string is not correct, got %+v, expected %v", session.ExtraColumns["app_test_string"].StringValue, "abc")
		}
		// check "app_test_bool"
		if _, ok := session.ExtraColumns["app_test_bool"]; !ok {
			t.Errorf("session.ExtraColumns[\"app_test_bool\"] not found in: %+v\n", session.ExtraColumns)
		}

		if session.ExtraColumns["app_test_bool"].BoolValue.Bool != true {
			t.Errorf("app_test_bool is not correct, got %+v, expected %v", session.ExtraColumns["app_test_bool"].BoolValue, true)
		}
		// check "app_test_number"
		if _, ok := session.ExtraColumns["app_test_number"]; !ok {
			t.Errorf("session.ExtraColumns[\"app_test_number\"] not found in: %+v\n", session.ExtraColumns)
		}

		if session.ExtraColumns["app_test_number"].Float64Value.Float64 != 123.456 {
			t.Errorf("app_test_number is not correct, got %+v, expected %v", session.ExtraColumns["app_test_number"].Float64Value, 123.456)
		}
		// check "app_test_date"
		if _, ok := session.ExtraColumns["app_test_date"]; !ok {
			t.Errorf("session.ExtraColumns[\"app_test_date\"] not found in: %+v\n", session.ExtraColumns)
		}

		if !session.ExtraColumns["app_test_date"].TimeValue.Time.Equal(testDate) {
			t.Errorf("app_test_date is not correct, got %+v, expected %v", session.ExtraColumns["app_test_date"].TimeValue, time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
		}
		// check "app_test_datetime"
		if _, ok := session.ExtraColumns["app_test_datetime"]; !ok {
			t.Errorf("session.ExtraColumns[\"app_test_datetime\"] not found in: %+v\n", session.ExtraColumns)
		}

		if !session.ExtraColumns["app_test_datetime"].TimeValue.Time.Equal(testDateTime) {
			t.Errorf("app_test_datetime is not correct, got %+v, expected %v", session.ExtraColumns["app_test_datetime"].TimeValue, testDateTime)
		}
		// check "app_test_timestamp"
		if _, ok := session.ExtraColumns["app_test_timestamp"]; !ok {
			t.Errorf("session.ExtraColumns[\"app_test_timestamp\"] not found in: %+v\n", session.ExtraColumns)
		}

		// check "app_test_longtext"
		if _, ok := session.ExtraColumns["app_test_longtext"]; !ok {
			t.Errorf("session.ExtraColumns[\"app_test_longtext\"] not found in: %+v \n", session.ExtraColumns)
		}

		if session.ExtraColumns["app_test_longtext"].StringValue.String != "abcd" {
			t.Errorf("app_test_longtext is not correct, got %+v, expected %v", session.ExtraColumns["app_test_longtext"].StringValue, "abcd")
		}
		// check "app_test_json"
		if _, ok := session.ExtraColumns["app_test_json"]; !ok {
			t.Errorf("session.ExtraColumns[\"app_test_json\"] not found in: %+v \n", session.ExtraColumns)
		}

		if !entity.AreEqualJSON(session.ExtraColumns["app_test_json"].JSONValue.JSON, testJSONByte) {
			t.Errorf("app_test_json is not correct, got %+v, expected %v", string(session.ExtraColumns["app_test_json"].JSONValue.JSON), string(testJSONByte))
		}

	})
}

func TestIntegration_DBSelectQuery(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()

	// instantiate a repo + service + router and assemble a web server
	repo, cfg, err := api.NewRepository(ctx, logger)
	if err != nil {
		t.Fatalf("cannot create repo: %v", err)
	}

	svc, err := api.NewService(ctx, logger, cfg, repo)
	if err != nil {
		t.Fatalf("cannot create a svc: %v", err)
	}

	t.Run("should select data with custom query", func(t *testing.T) {

		selectBuilder := squirrel.Select(`
			s.external_id,
			s.user_id,
			s.created_at,
			u.external_id as user_extid,
			u.is_authenticated,
			u.created_at as user_at,
			u.orders_ltv,
			u.country`).
			From("session as s").
			Join("user as u ON s.user_id = u.id").
			Limit(10)

		sqlQuery, args, err := selectBuilder.ToSql()
		if err != nil {
			t.Fatalf("cannot build sql query: %v", err)
		}

		jsonData, err := svc.DoDBSelect(workspaceID, sqlQuery, args)

		if err != nil {
			t.Errorf("cannot select data: %v", err)
		}

		data := []struct {
			ExternalID      string    `json:"external_id"`
			UserID          string    `json:"user_id"`
			CreatedAt       time.Time `json:"created_at"`
			UserExternalID  string    `json:"user_extid"`
			IsAuthenticated int       `json:"is_authenticated"`
			UserCreatedAt   time.Time `json:"user_at"`
			OrdersLTV       float64   `json:"orders_ltv"`
			Country         string    `json:"country"`
		}{}

		if err = json.Unmarshal(jsonData, &data); err != nil {
			t.Errorf("cannot unmarshal data: %v", err)
		}

		if data == nil {
			t.Errorf("data is nil")
		}

		if len(data) == 0 {
			t.Errorf("data is empty")
		}

		firstRow := data[0]
		// t.Logf("firstRow: %+v\n", firstRow)

		// strings are returned as []byte (string)
		if firstRow.ExternalID == "" {
			t.Errorf("cannot extract external_id")
		}
		if firstRow.CreatedAt.IsZero() {
			t.Errorf("cannot extract created_at")
		}
	})
}

func TestIntegration_ImportPageviewData(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()

	// instantiate a repo + service + router and assemble a web server
	repo, cfg, err := api.NewRepository(ctx, logger)
	if err != nil {
		t.Fatalf("cannot create repo: %v", err)
	}

	svc, err := api.NewService(ctx, logger, cfg, repo)
	if err != nil {
		t.Fatalf("cannot create svc: %v", err)
	}

	t.Run("should import pageview data and update session", func(t *testing.T) {

		createdAt := time.Now().AddDate(0, 0, -1).Format(time.RFC3339)
		updatedAt := time.Now().Add(-5 * time.Second).Format(time.RFC3339)

		// first pageview has no duration yet
		item := fmt.Sprintf(`{
					"kind": "%v",
					"pageview": {
						"external_id": "bbb",
						"session_external_id": "sss",
						"domain_id": "web",
						"page_id": "home",
						"created_at": "%v",
						"updated_at": "%v",
						"title": "Home",
						"duration": 15
					},
					"user": {
						"external_id": "user_pv",
						"is_authenticated": false,
						"created_at": "%v"
					},
					"session": {
						"external_id": "sss",
						"domain_id": "web",
						"landing_page": "https://web.com",
						"created_at": "%v",
						"updated_at": "%v",
						"duration": 15,
						"pageviews_count": 1,
						"interactions_count": 1,
						"utm_source": "test",
						"utm_medium": "test",
						"utm_campaign": "test-campaign"
					}
				
		}`,
			entity.ItemKindPageview,
			createdAt,
			updatedAt,
			createdAt,
			createdAt,
			updatedAt,
		)

		dataLogInqueue := &commonDTO.DataLogInQueue{
			Origin: commonDTO.DataLogOriginClient,
			Context: commonDTO.DataLogContext{
				WorkspaceID:      workspaceID,
				HeadersAndParams: commonDTO.MapOfStrings{"Origin": "https://testing.com"},
				IP:               "0.0.0.0",
				ReceivedAt:       time.Now().UTC(),
			},
			Item: item,
		}
		// compute ID from body
		dataLogInqueue.ID = commonDTO.ComputeDataLogID(cfg.SECRET_KEY, commonDTO.DataLogOriginClient, item)

		result := svc.DataLogImportFromQueue(ctx, dataLogInqueue)

		if result == nil {
			t.Fatalf("result is nil")
		}

		if result.HasError || result.Error != "" {
			t.Errorf("error: %v", result.Error)
		}

		updatedAt = time.Now().Format(time.RFC3339)

		// add pageview duration
		item = fmt.Sprintf(`{
				"kind": "%v",
				"pageview": {
					"external_id": "bbb",
					"session_external_id": "sss",
					"domain_id": "web",
					"page_id": "home",
					"created_at": "%v",
					"updated_at": "%v",
					"title": "Home",
					"duration": 30
				},
				"user": {
					"external_id": "user_pv",
					"is_authenticated": false,
					"created_at": "%v"
				},
				"session": {
					"external_id": "sss",
					"domain_id": "web",
					"landing_page": "https://web.com",
					"created_at": "%v",
					"updated_at": "%v",
					"utm_source": "test",
					"utm_medium": "test",
					"duration": 30,
					"pageviews_count": 2,
					"interactions_count": 2
				}
			
		}`,
			entity.ItemKindPageview,
			createdAt,
			updatedAt,
			createdAt,
			createdAt,
			updatedAt,
		)

		dataLogInqueue = &commonDTO.DataLogInQueue{
			Origin: commonDTO.DataLogOriginClient,
			Context: commonDTO.DataLogContext{
				WorkspaceID:      workspaceID,
				HeadersAndParams: commonDTO.MapOfStrings{"Origin": "https://testing.com"},
				IP:               "0.0.0.0",
				ReceivedAt:       time.Now().UTC(),
			},
			Item: item,
		}
		// compute ID from item
		dataLogInqueue.ID = commonDTO.ComputeDataLogID(cfg.SECRET_KEY, commonDTO.DataLogOriginClient, item)

		result = svc.DataLogImportFromQueue(ctx, dataLogInqueue)

		if result == nil {
			t.Fatalf("result is nil")
		}

		if result.HasError || result.Error != "" {
			t.Errorf("error: %v", result.Error)
		}

		// fetch session and check that campaign is not empty

		workspace, err := repo.GetWorkspace(ctx, workspaceID)
		if err != nil {
			t.Fatalf("cannot fetch workspace: %v", err)
		}

		userID := entity.ComputeUserID("user_pv")
		session, err := repo.FindSessionByID(ctx, workspace, entity.ComputeSessionID("sss"), userID, nil)

		if err != nil {
			t.Fatalf("cannot fetch session: %v", err)
		}

		if session.Duration == nil || session.Duration.IsNull || session.Duration.Int64 != 30 {
			t.Errorf("Duration is not correct, got %+v, expected %v", session.Duration, "test-campaign")
		}

		if session.UTMCampaign == nil || session.UTMCampaign.IsNull || session.UTMCampaign.String != "test-campaign" {
			t.Errorf("UTMCampaign is not correct, got %+v, expected %v", session.UTMCampaign, "test-campaign")
		}

	})
}
func TestIntegration_ImportCustomEventData(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()

	// instantiate a repo + service + router and assemble a web server
	repo, cfg, err := api.NewRepository(ctx, logger)
	if err != nil {
		t.Fatalf("cannot create repo: %v", err)
	}

	svc, err := api.NewService(ctx, logger, cfg, repo)
	if err != nil {
		t.Fatalf("cannot create svc: %v", err)
	}

	t.Run("should import custom event data and update session", func(t *testing.T) {

		createdAt := time.Now().AddDate(0, 0, -1).Format(time.RFC3339)
		updatedAt := time.Now().Format(time.RFC3339)

		body := fmt.Sprintf(`{
				"kind": "%v",
				"custom_event": {
					"external_id": "bbb",
					"session_external_id": "eventsession",
					"domain_id": "web",
					"label": "event 1",
					"string_value": "abc",
					"number_value": 123.456,
					"boolean_value": true,
					"created_at": "%v",
					"updated_at": "%v"
				},
				"user": {
					"external_id": "user_pv",
					"is_authenticated": false,
					"created_at": "%v"
				},
				"session": {
					"external_id": "eventsession",
					"domain_id": "web",
					"landing_page": "https://web.com",
					"created_at": "%v",
					"updated_at": "%v",
					"utm_source": "test",
					"utm_medium": "test"
				}
			
		}`,
			entity.ItemKindCustomEvent,
			createdAt,
			updatedAt,
			createdAt,
			createdAt,
			updatedAt,
		)

		dataLogInqueue := &commonDTO.DataLogInQueue{
			Origin: commonDTO.DataLogOriginClient,
			Context: commonDTO.DataLogContext{
				WorkspaceID:      workspaceID,
				HeadersAndParams: commonDTO.MapOfStrings{"Origin": "https://testing.com"},
				IP:               "0.0.0.0",
				ReceivedAt:       time.Now().UTC(),
			},
			Item: body,
		}
		// compute ID from body
		dataLogInqueue.ID = commonDTO.ComputeDataLogID(cfg.SECRET_KEY, commonDTO.DataLogOriginClient, body)

		result := svc.DataLogImportFromQueue(ctx, dataLogInqueue)

		if result == nil {
			t.Fatalf("result is nil")
		}

		if result.HasError || result.Error != "" {
			t.Errorf("error: %v", result.Error)
		}
	})
}

func TestIntegration_MergeUsers(t *testing.T) {
	ctx := context.Background()
	logger := logrus.New()

	// instantiate a repo + service + router and assemble a web server
	repo, cfg, err := api.NewRepository(ctx, logger)
	if err != nil {
		t.Fatalf("cannot create repo: %v", err)
	}

	svc, err := api.NewService(ctx, logger, cfg, repo)
	if err != nil {
		t.Fatalf("cannot create svc: %v", err)
	}

	sid, _ = shortid.New(1, shortid.DefaultABC, 777)
	id := sid.MustGenerate()

	t.Run("should import and merge 2 users", func(t *testing.T) {

		user := fmt.Sprintf(`{
					"kind": "user",
					"user": {
						"external_id": "anon-%v",
						"is_authenticated": false,
						"created_at": "2023-05-20T10:14:39+00:00"
					}
				}`, id)

		user2 := fmt.Sprintf(`{
					"kind": "user",
					"user": {
						"external_id": "auth-%v",
						"is_authenticated": true,
						"created_at": "2023-05-26T11:14:39+00:00",
						"signed_up_at": "2023-05-26T11:14:39+12:00",
						"gender": "female",
						"country": "CA",
						"birthday": "1995-08-14",
						"language": "en",
						"latitude": 43.85162,
						"timezone": "America/Toronto",
						"first_name": "clara",
						"last_name": "clark",
						"email": "clara123456789@gmail.com",
						"longitude": -79.487554,
						"photo_url": "https://randomuser.me/api/portraits/med/women/53.jpg",
						"consent_all": true
					}
				}`, id)

		userAlias := fmt.Sprintf(`{
					"kind": "user_alias",
					"user_alias": {
						"from_user_external_id": "anon-%v",
						"to_user_created_at": "2023-05-26T11:14:39+00:00",
						"to_user_external_id": "auth-%v",
						"to_user_is_authenticated": true
					}
				}`, id, id)

		pageview := fmt.Sprintf(`{
					"kind": "pageview",
					"pageview": {
						"external_id": "pageview-%v",
						"domain_id": "web",
						"page_id": "/pageview-%v",
						"title": "pageview title",
						"created_at": "2023-05-27T11:14:39+00:00",
						"duration": 10
					},
					"session": {
						"external_id": "pageviewsession-%v",
						"domain_id": "web",
						"landing_page": "https://web.com",
						"created_at": "2023-05-27T11:14:39+00:00",
						"utm_source": "test",
						"utm_medium": "test"
					},
					"user": {
						"external_id": "anon-%v",
						"is_authenticated": false,
						"created_at": "2023-05-26T11:14:39+00:00"
					}
				}`, id, id, id, id)

		items := []string{user, user2, userAlias, pageview}

		for _, item := range items {
			dataLogInqueue := &commonDTO.DataLogInQueue{
				Origin: commonDTO.DataLogOriginClient,
				Context: commonDTO.DataLogContext{
					WorkspaceID:      workspaceID,
					HeadersAndParams: commonDTO.MapOfStrings{"Origin": "https://testing.com"},
					IP:               "0.0.0.0",
					ReceivedAt:       time.Now().UTC(),
				},
				Item: item,
			}
			// compute ID from item
			dataLogInqueue.ID = commonDTO.ComputeDataLogID(cfg.SECRET_KEY, commonDTO.DataLogOriginClient, item)

			result := svc.DataLogImportFromQueue(ctx, dataLogInqueue)
			if result == nil {
				t.Fatalf("result is nil")
			}

			if result.HasError || result.Error != "" {
				t.Errorf("error: %v", result.Error)
			}
		}

		// fetch auth user
		authExternalID := fmt.Sprintf("auth-%v", id)
		authID := fmt.Sprintf("%x", sha1.Sum([]byte(authExternalID)))
		authUser, err := repo.FindUserByID(context.Background(), workspace, authID, nil)

		if err != nil {
			t.Fatalf("cannot fetch auth user: %v", err)
		}

		if authUser == nil {
			t.Fatalf("auth user is nil, ext id: %v, id: %v", authExternalID, authID)
		}

		// user first_name should be Clara
		if authUser.FirstName == nil || authUser.FirstName.IsNull || authUser.FirstName.String != "clara" {
			t.Errorf("auth user first_name is not clara, got %+v", authUser.FirstName)
		}
	})
}

// func TestIntegration_Task(t *testing.T) {
// 	ctx := context.Background()
// 	logger := logrus.New()

// 	// instantiate a repo + service + router and assemble a web server
// 	repo, cfg, err := api.NewRepository(ctx, logger)
// 	if err != nil {
// 		t.Fatalf("cannot create repo: %v", err)
// 	}

// 	svc, err := api.NewService(ctx, logger, cfg, repo)
// 	if err != nil {
// 		t.Fatalf("cannot create svc: %v", err)
// 	}

// 	t.Run("should create a task", func(t *testing.T) {

// 	})
// }
