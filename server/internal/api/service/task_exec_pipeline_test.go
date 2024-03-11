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

	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/internal/api/repository"
	"github.com/rimdian/rimdian/internal/common/httpClient"
	"github.com/rimdian/rimdian/internal/common/taskorchestrator"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestServiceImpl_TaskPipeline(t *testing.T) {

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

	// create logger
	logger := logrus.New()

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

	t.Run("should reject invalid task_exec.kind", func(t *testing.T) {

		repoMock := &repository.RepositoryMock{
			GetTaskExecFunc: func(ctx context.Context, workspaceID, taskID string) (*entity.TaskExec, error) {
				return &entity.TaskExec{
					ID:              "test",
					TaskID:          "bad",
					Name:            "test",
					MultipleExecKey: nil,
					OnMultipleExec:  entity.OnMultipleExecAllow,
					State:           entity.NewTaskState(),
					RetryCount:      0,
					Message:         nil,
					Status:          entity.TaskExecStatusProcessing,
					DBCreatedAt:     entity.TimePtr(time.Now()),
				}, nil
			},
			ReleaseUsersLockFunc: func(workspaceID string, lock *entity.UsersLock) error {
				return nil
			},
			SetTaskExecErrorFunc: func(ctx context.Context, workspaceID, taskExecID string, workerID, status int, message string) error {
				return nil
			},
		}

		taskOrchestratorMock := &taskorchestrator.ClientMock{
			PostRequestFunc: func(ctx context.Context, taskRequest *taskorchestrator.TaskRequest) error {
				return nil
			},
		}

		props := &TaskExecPipelineProps{
			Config:           cfg,
			Logger:           logger,
			NetClient:        netClientMock,
			Repository:       repoMock,
			Workspace:        demoWorkspace,
			TaskOrchestrator: taskOrchestratorMock,
			TaskExecPayload: &dto.TaskExecRequestPayload{
				TaskExecID: "test",
				WorkerID:   0, // 0 = parent thread
				JobID:      "job_1",
			},
		}

		pipeline := NewTaskExecPipeline(props)
		pipeline.Execute(context.Background())

		result := pipeline.GetQueueResult()

		assert.NotNil(t, result)
		assert.True(t, result.HasError)
		assert.False(t, result.QueueShouldRetry)
		assert.Containsf(t, result.Error, "unknown task kind", "error should contain: unknown task kind")
	})

	t.Run("should persist done", func(t *testing.T) {

		var gotResult *entity.TaskExecResult

		repoMock := &repository.RepositoryMock{
			GetTaskExecFunc: func(ctx context.Context, workspaceID, taskID string) (*entity.TaskExec, error) {
				return &entity.TaskExec{
					ID:              "test",
					TaskID:          entity.TaskKindTestingDone,
					Name:            "test",
					MultipleExecKey: nil,
					OnMultipleExec:  entity.OnMultipleExecAllow,
					State:           entity.NewTaskState(),
					RetryCount:      0,
					Message:         nil,
					Status:          entity.TaskExecStatusProcessing,
					DBCreatedAt:     entity.TimePtr(time.Now()),
				}, nil
			},
			ReleaseUsersLockFunc: func(workspaceID string, lock *entity.UsersLock) error {
				return nil
			},
			RunInTransactionForWorkspaceFunc: func(ctx context.Context, workspaceID string, f func(context.Context, *sql.Tx) (int, error)) (int, error) {
				return f(ctx, nil)
			},
			UpdateTaskExecFromResultFunc: func(ctx context.Context, taskExecRequestPayload *dto.TaskExecRequestPayload, taskExecResult *entity.TaskExecResult, tx *sql.Tx) error {
				gotResult = taskExecResult
				return nil
			},
		}

		taskOrchestratorMock := &taskorchestrator.ClientMock{
			PostRequestFunc: func(ctx context.Context, taskRequest *taskorchestrator.TaskRequest) error {
				return nil
			},
		}

		props := &TaskExecPipelineProps{
			Config:           cfg,
			Logger:           logger,
			NetClient:        netClientMock,
			Repository:       repoMock,
			Workspace:        demoWorkspace,
			TaskOrchestrator: taskOrchestratorMock,
			TaskExecPayload: &dto.TaskExecRequestPayload{
				TaskExecID: "test",
				WorkerID:   0, // 0 = parent thread
				JobID:      "job_1",
			},
		}

		pipeline := NewTaskExecPipeline(props)
		pipeline.Execute(context.Background())

		result := pipeline.GetQueueResult()

		assert.NotNil(t, result)
		assert.False(t, result.HasError)
		assert.False(t, result.QueueShouldRetry)

		assert.NotNil(t, gotResult)
		assert.True(t, gotResult.IsDone)
		assert.False(t, gotResult.IsError)
		assert.Equal(t, 1, len(repoMock.UpdateTaskExecFromResultCalls()))
		assert.Equal(t, 1, len(repoMock.ReleaseUsersLockCalls()))
		assert.Equal(t, 0, len(taskOrchestratorMock.PostRequestCalls()))
	})

	t.Run("should persist state if not done", func(t *testing.T) {

		var gotResult *entity.TaskExecResult

		repoMock := &repository.RepositoryMock{
			GetTaskExecFunc: func(ctx context.Context, workspaceID, taskID string) (*entity.TaskExec, error) {
				return &entity.TaskExec{
					ID:              "test",
					TaskID:          entity.TaskKindTestingNotDone,
					Name:            "test",
					MultipleExecKey: nil,
					OnMultipleExec:  entity.OnMultipleExecAllow,
					State:           entity.NewTaskState(),
					RetryCount:      0,
					Message:         nil,
					Status:          entity.TaskExecStatusProcessing,
					DBCreatedAt:     entity.TimePtr(time.Now()),
				}, nil
			},
			ReleaseUsersLockFunc: func(workspaceID string, lock *entity.UsersLock) error {
				return nil
			},
			RunInTransactionForWorkspaceFunc: func(ctx context.Context, workspaceID string, f func(context.Context, *sql.Tx) (int, error)) (int, error) {
				return f(ctx, nil)
			},
			UpdateTaskExecFromResultFunc: func(ctx context.Context, taskExecRequestPayload *dto.TaskExecRequestPayload, taskExecResult *entity.TaskExecResult, tx *sql.Tx) error {
				gotResult = taskExecResult
				return nil
			},
			AddJobToTaskExecFunc: func(ctxWithTimeout context.Context, taskExecID, newJobID string, tx *sql.Tx) error {
				return nil
			},
		}

		taskOrchestratorMock := &taskorchestrator.ClientMock{
			PostRequestFunc: func(ctx context.Context, taskRequest *taskorchestrator.TaskRequest) error {
				return nil
			},
		}

		props := &TaskExecPipelineProps{
			Config:           cfg,
			Logger:           logger,
			NetClient:        netClientMock,
			Repository:       repoMock,
			Workspace:        demoWorkspace,
			TaskOrchestrator: taskOrchestratorMock,
			TaskExecPayload: &dto.TaskExecRequestPayload{
				TaskExecID: "test",
				WorkerID:   0, // 0 = parent thread
				JobID:      "job_1",
			},
		}

		pipeline := NewTaskExecPipeline(props)
		pipeline.Execute(context.Background())

		result := pipeline.GetQueueResult()

		assert.NotNil(t, result)
		assert.False(t, result.HasError)
		assert.False(t, result.QueueShouldRetry)

		assert.NotNil(t, gotResult)
		assert.False(t, gotResult.IsDone)
		assert.False(t, gotResult.IsError)

		assert.Equal(t, 1, len(repoMock.UpdateTaskExecFromResultCalls()))
		assert.Equal(t, 1, len(repoMock.AddJobToTaskExecCalls()))
		assert.Equal(t, 1, len(repoMock.ReleaseUsersLockCalls()))
		assert.Equal(t, 1, len(taskOrchestratorMock.PostRequestCalls()))
	})

	t.Run("should persist state on panic", func(t *testing.T) {

		repoMock := &repository.RepositoryMock{
			GetTaskExecFunc: func(ctx context.Context, workspaceID, taskID string) (*entity.TaskExec, error) {
				return &entity.TaskExec{
					ID:              "test",
					TaskID:          entity.TaskKindTestingPanic,
					Name:            "test",
					MultipleExecKey: nil,
					OnMultipleExec:  entity.OnMultipleExecAllow,
					State:           entity.NewTaskState(),
					RetryCount:      0,
					Message:         nil,
					Status:          entity.TaskExecStatusProcessing,
					DBCreatedAt:     nil, // IMPORTANT: not persisted yet
				}, nil
			},
			ReleaseUsersLockFunc: func(workspaceID string, lock *entity.UsersLock) error {
				return nil
			},
			RunInTransactionForWorkspaceFunc: func(ctx context.Context, workspaceID string, f func(context.Context, *sql.Tx) (int, error)) (int, error) {
				return f(ctx, nil)
			},
			SetTaskExecErrorFunc: func(ctx context.Context, workspaceID, taskExecID string, workerID, status int, message string) error {
				return nil
			},
		}

		taskOrchestratorMock := &taskorchestrator.ClientMock{
			PostRequestFunc: func(ctx context.Context, taskRequest *taskorchestrator.TaskRequest) error {
				return nil
			},
		}

		props := &TaskExecPipelineProps{
			Config:           cfg,
			Logger:           logger,
			NetClient:        netClientMock,
			Repository:       repoMock,
			Workspace:        demoWorkspace,
			TaskOrchestrator: taskOrchestratorMock,
			TaskExecPayload: &dto.TaskExecRequestPayload{
				TaskExecID: "test",
				WorkerID:   0, // 0 = parent thread
				JobID:      "job_1",
			},
		}

		pipeline := NewTaskExecPipeline(props)
		pipeline.Execute(context.Background())

		result := pipeline.GetQueueResult()

		assert.NotNil(t, result)
		assert.True(t, result.HasError)
		assert.True(t, result.QueueShouldRetry)

		assert.Equal(t, 1, len(repoMock.SetTaskExecErrorCalls()))
		assert.Equal(t, 1, len(repoMock.ReleaseUsersLockCalls()))
		assert.Equal(t, 0, len(repoMock.UpdateTaskExecFromResultCalls()))
		assert.Equal(t, 0, len(repoMock.AddJobToTaskExecCalls()))
		assert.Equal(t, 0, len(taskOrchestratorMock.PostRequestCalls()))
		assert.Containsf(t, result.Error, "recover", "error should contain: recover")
	})

	t.Run("should persist state on context timeout", func(t *testing.T) {

		repoMock := &repository.RepositoryMock{
			GetTaskExecFunc: func(ctx context.Context, workspaceID, taskID string) (*entity.TaskExec, error) {
				return &entity.TaskExec{
					ID:              "test",
					TaskID:          entity.TaskKindTestingTimeout,
					Name:            "test",
					MultipleExecKey: nil,
					OnMultipleExec:  entity.OnMultipleExecAllow,
					State:           entity.NewTaskState(),
					RetryCount:      0,
					Message:         nil,
					Status:          entity.TaskExecStatusProcessing,
					DBCreatedAt:     entity.TimePtr(time.Now()),
				}, nil
			},
			ReleaseUsersLockFunc: func(workspaceID string, lock *entity.UsersLock) error {
				return nil
			},
			RunInTransactionForWorkspaceFunc: func(ctx context.Context, workspaceID string, f func(context.Context, *sql.Tx) (int, error)) (int, error) {
				return f(ctx, nil)
			},
			AddJobToTaskExecFunc: func(ctxWithTimeout context.Context, taskExecID, newJobID string, tx *sql.Tx) error {
				return nil
			},
			SetTaskExecErrorFunc: func(ctx context.Context, workspaceID, taskExecID string, workerID, status int, message string) error {
				return nil
			},
		}

		taskOrchestratorMock := &taskorchestrator.ClientMock{
			PostRequestFunc: func(ctx context.Context, taskRequest *taskorchestrator.TaskRequest) error {
				return nil
			},
		}

		// context with timeout of 1 sec
		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		props := &TaskExecPipelineProps{
			Config:           cfg,
			Logger:           logger,
			NetClient:        netClientMock,
			Repository:       repoMock,
			Workspace:        demoWorkspace,
			TaskOrchestrator: taskOrchestratorMock,
			TaskExecPayload: &dto.TaskExecRequestPayload{
				TaskExecID: "test",
				WorkerID:   0, // 0 = parent thread
				JobID:      "job_1",
			},
		}

		pipeline := NewTaskExecPipeline(props)

		cancel()

		pipeline.Execute(ctxWithTimeout)

		result := pipeline.GetQueueResult()

		assert.NotNil(t, result)
		assert.True(t, result.HasError)
		assert.True(t, result.QueueShouldRetry)

		assert.Equal(t, 1, len(repoMock.SetTaskExecErrorCalls()))
		assert.Equal(t, 1, len(repoMock.ReleaseUsersLockCalls()))
		assert.Equal(t, 0, len(repoMock.UpdateTaskExecFromResultCalls()))
		assert.Equal(t, 0, len(repoMock.AddJobToTaskExecCalls()))
		assert.Equal(t, 0, len(taskOrchestratorMock.PostRequestCalls()))
	})
}
