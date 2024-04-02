//go:generate moq -out pipeline_moq.go . Pipeline
package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/internal/api/repository"
	common "github.com/rimdian/rimdian/internal/common/dto"
	"github.com/rimdian/rimdian/internal/common/httpClient"
	"github.com/sirupsen/logrus"
)

// Pipeline is the interface that wraps the basic methods to generate a data_log
// it could be a DataLogPipeline | TaskPipeline | WorkflowPipeline

type Pipeline interface {
	// Ctx() context.Context
	Cfg() *entity.Config
	Log() *logrus.Logger
	Net() httpClient.HTTPClient
	Repo() repository.Repository
	GetWorkspace() *entity.Workspace
	GetQueueResult() *common.DataLogInQueueResult
	Execute(ctx context.Context)
	ProcessNextStep(ctx context.Context)
	InsertChildDataLog(ctx context.Context, kind string, action string, userID string, itemID string, itemExternalID string, updatedFields entity.UpdatedFields, eventAt time.Time, tx *sql.Tx) error
	EnsureUsersLock(ctx context.Context) error
	ReleaseUsersLock() error
	GetUserIDs() []string
	AddDataLogGenerated(dataLog *entity.DataLog)
	GetDataLogsGenerated() []*entity.DataLog
	SetError(key string, err string, retryable bool)
	HasError() bool
	DataLogEnqueue(ctx context.Context, replayID *string, origin int, originID string, workspaceID string, jsonItems []string, isSync bool)
	ReattributeUsersOrders(ctx context.Context)
	AttributeOrder(ctx context.Context, order *entity.Order, orderSessions []*entity.Session, orderPostviews []*entity.Postview, previousOrders []*entity.Order, devices []*entity.Device, tx *sql.Tx)
}
