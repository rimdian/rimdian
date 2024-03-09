package collector

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rimdian/rimdian/internal/common/auth"
	"github.com/rimdian/rimdian/internal/common/dto"
	"github.com/rimdian/rimdian/internal/common/httpClient"
	"github.com/rimdian/rimdian/internal/common/taskorchestrator"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestCollector(t *testing.T) {

	viper.SetDefault("ENV", ENV_DEV)
	viper.SetDefault("COLLECTOR_PORT", "8080")

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../../")

	// load config.yaml
	err := viper.ReadInConfig()
	if err != nil {
		t.Fatal(err)
	}

	cfg := &Config{
		ENV:                     viper.GetString("ENV"),
		COLLECTOR_PORT:          viper.GetString("COLLECTOR_PORT"),
		API_ENDPOINT:            viper.GetString("API_ENDPOINT"),
		SECRET_KEY:              viper.GetString("SECRET_KEY"),
		GCLOUD_PROJECT:          viper.GetString("GCLOUD_PROJECT"),
		GCLOUD_JSON_CREDENTIALS: viper.GetString("GCLOUD_JSON_CREDENTIALS"),
		TASK_QUEUE_LOCATION:     viper.GetString("TASK_QUEUE_LOCATION"),
		DEV_SSL_CERT:            viper.GetString("DEV_SSL_CERT"),
		DEV_SSL_KEY:             viper.GetString("DEV_SSL_KEY"),
	}

	if err := ValidateConfig(cfg); err != nil {
		t.Fatal(err)
	}

	item := `{"kind": "app_test"}`
	payload := fmt.Sprintf(`{
		"workspace_id": "acme_testing", 
		"context": {"data_sent_at": "%v"},
		"items": [%v]
	}`, time.Now().Format(time.RFC3339), item)

	t.Run("should error on empty body", func(t *testing.T) {

		collector := *NewCollector(cfg, nil, nil, logrus.New())
		srv := httptest.NewServer(collector.Handler)
		defer srv.Close()

		req, err := http.NewRequest(http.MethodPost, srv.URL+"/data", bytes.NewReader([]byte(" ")))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("failed to do request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "status code")
		assert.Equal(t, ErrEmptyBody.Error(), string(readAllAndTrim(t, resp.Body)), "body")
	})

	t.Run("should send request to data Queue with a Token", func(t *testing.T) {

		httpClientMoq := &httpClient.HTTPClientMock{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					Status:     "200 OK",
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader("ok")),
				}, nil
			},
		}

		taskClientMoq := &taskorchestrator.ClientMock{
			PostRequestFunc: func(ctx context.Context, job *taskorchestrator.TaskRequest) error {
				return nil
			},
			GetLiveQueueNameForWorkspaceFunc: func(workspaceID string) string {
				return workspaceID + "-data-imports-live"
			},
			GetHistoricalQueueNameForWorkspaceFunc: func(workspaceID string) string {
				return workspaceID + "-data-imports"
			},
		}

		collector := *NewCollector(cfg, taskClientMoq, httpClientMoq, logrus.New())

		srv := httptest.NewServer(collector.Handler)
		defer srv.Close()

		accessToken, err := auth.CreateAccountToken(cfg.SECRET_KEY, cfg.API_ENDPOINT, time.Now(), time.Now().AddDate(0, 0, 1), auth.TypeAccessToken, "root", "")

		if err != nil {
			t.Fatalf("create access token err %v", err)
		}

		req, err := http.NewRequest(http.MethodPost, srv.URL+"/data", bytes.NewReader([]byte(payload)))
		req.Header.Add("Authorization", "Bearer "+accessToken)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("failed to do request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "status code")
		assert.Equal(t, ResponseQueued, string(readAllAndTrim(t, resp.Body)), "body")

		assert.Equal(t, len(taskClientMoq.PostRequestCalls()), 1, "task created")
		assert.Equal(t, taskClientMoq.PostRequestCalls()[0].TaskRequest.QueueName, "acme_testing-data-imports", "queue name")

		assert.Equal(t, len(httpClientMoq.DoCalls()), 0, "bypass queue requests sent")

		// check data task object
		assert.Equal(t, item, taskClientMoq.PostRequestCalls()[0].TaskRequest.Payload.(*dto.DataLogInQueue).Item, "body content")
		assert.Equal(t, dto.DataLogOriginToken, taskClientMoq.PostRequestCalls()[0].TaskRequest.Payload.(*dto.DataLogInQueue).Origin, "origin")
	})

	t.Run("should bypass Task Queue", func(t *testing.T) {

		httpClientMoq := &httpClient.HTTPClientMock{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					Status:     "200 OK",
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader("ok")),
				}, nil
			},
		}

		taskClientMoq := &taskorchestrator.ClientMock{
			PostRequestFunc: func(ctx context.Context, job *taskorchestrator.TaskRequest) error {
				return nil
			},
			GetLiveQueueNameForWorkspaceFunc: func(workspaceID string) string {
				return workspaceID + "-data-imports-live"
			},
			GetHistoricalQueueNameForWorkspaceFunc: func(workspaceID string) string {
				return workspaceID + "-data-imports"
			},
		}

		collector := *NewCollector(cfg, taskClientMoq, httpClientMoq, logrus.New())

		srv := httptest.NewServer(collector.Handler)
		defer srv.Close()

		req, err := http.NewRequest(http.MethodPost, srv.URL+"/bypass", bytes.NewReader([]byte(payload)))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("failed to do request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, "ok", string(readAllAndTrim(t, resp.Body)), "body")
		assert.Equal(t, http.StatusOK, resp.StatusCode, "status code")
		assert.Equal(t, len(taskClientMoq.PostRequestCalls()), 0, "task created")
		assert.Equal(t, len(httpClientMoq.DoCalls()), 1, "bypass queue requests sent")
	})
}

func readAllAndTrim(t *testing.T, r io.Reader) []byte {
	t.Helper()

	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read all from reader: %v", err)
	}

	return bytes.TrimSpace(b)
}
