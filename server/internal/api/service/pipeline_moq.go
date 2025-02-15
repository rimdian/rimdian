// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package service

import (
	"context"
	"database/sql"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/internal/api/repository"
	common "github.com/rimdian/rimdian/internal/common/dto"
	"github.com/rimdian/rimdian/internal/common/httpClient"
	"github.com/sirupsen/logrus"
	"sync"
)

// Ensure, that PipelineMock does implement Pipeline.
// If this is not the case, regenerate this file with moq.
var _ Pipeline = &PipelineMock{}

// PipelineMock is a mock implementation of Pipeline.
//
//	func TestSomethingThatUsesPipeline(t *testing.T) {
//
//		// make and configure a mocked Pipeline
//		mockedPipeline := &PipelineMock{
//			AddDataLogGeneratedFunc: func(dataLog *entity.DataLog)  {
//				panic("mock out the AddDataLogGenerated method")
//			},
//			AttributeOrderFunc: func(ctx context.Context, order *entity.Order, orderSessions []*entity.Session, orderPostviews []*entity.Postview, previousOrders []*entity.Order, devices []*entity.Device, tx *sql.Tx)  {
//				panic("mock out the AttributeOrder method")
//			},
//			CfgFunc: func() *entity.Config {
//				panic("mock out the Cfg method")
//			},
//			DataLogEnqueueFunc: func(ctx context.Context, replayID *string, origin int, originID string, workspaceID string, jsonItems []string, isSync bool)  {
//				panic("mock out the DataLogEnqueue method")
//			},
//			EnsureUsersLockFunc: func(ctx context.Context) error {
//				panic("mock out the EnsureUsersLock method")
//			},
//			ExecuteFunc: func(ctx context.Context)  {
//				panic("mock out the Execute method")
//			},
//			GetDataLogsGeneratedFunc: func() []*entity.DataLog {
//				panic("mock out the GetDataLogsGenerated method")
//			},
//			GetQueueResultFunc: func() *common.ResponseForTaskQueue {
//				panic("mock out the GetQueueResult method")
//			},
//			GetUserIDsFunc: func() []string {
//				panic("mock out the GetUserIDs method")
//			},
//			GetWorkspaceFunc: func() *entity.Workspace {
//				panic("mock out the GetWorkspace method")
//			},
//			HasErrorFunc: func() bool {
//				panic("mock out the HasError method")
//			},
//			InsertChildDataLogFunc: func(ctx context.Context, data entity.ChildDataLog) error {
//				panic("mock out the InsertChildDataLog method")
//			},
//			LogFunc: func() *logrus.Logger {
//				panic("mock out the Log method")
//			},
//			NetFunc: func() httpClient.HTTPClient {
//				panic("mock out the Net method")
//			},
//			ProcessNextStepFunc: func(ctx context.Context)  {
//				panic("mock out the ProcessNextStep method")
//			},
//			ReattributeUsersOrdersFunc: func(ctx context.Context)  {
//				panic("mock out the ReattributeUsersOrders method")
//			},
//			ReleaseUsersLockFunc: func() error {
//				panic("mock out the ReleaseUsersLock method")
//			},
//			RepoFunc: func() repository.Repository {
//				panic("mock out the Repo method")
//			},
//			SetErrorFunc: func(key string, err string, retryable bool)  {
//				panic("mock out the SetError method")
//			},
//		}
//
//		// use mockedPipeline in code that requires Pipeline
//		// and then make assertions.
//
//	}
type PipelineMock struct {
	// AddDataLogGeneratedFunc mocks the AddDataLogGenerated method.
	AddDataLogGeneratedFunc func(dataLog *entity.DataLog)

	// AttributeOrderFunc mocks the AttributeOrder method.
	AttributeOrderFunc func(ctx context.Context, order *entity.Order, orderSessions []*entity.Session, orderPostviews []*entity.Postview, previousOrders []*entity.Order, devices []*entity.Device, tx *sql.Tx)

	// CfgFunc mocks the Cfg method.
	CfgFunc func() *entity.Config

	// DataLogEnqueueFunc mocks the DataLogEnqueue method.
	DataLogEnqueueFunc func(ctx context.Context, replayID *string, origin int, originID string, workspaceID string, jsonItems []string, isSync bool)

	// EnsureUsersLockFunc mocks the EnsureUsersLock method.
	EnsureUsersLockFunc func(ctx context.Context) error

	// ExecuteFunc mocks the Execute method.
	ExecuteFunc func(ctx context.Context)

	// GetDataLogsGeneratedFunc mocks the GetDataLogsGenerated method.
	GetDataLogsGeneratedFunc func() []*entity.DataLog

	// GetQueueResultFunc mocks the GetQueueResult method.
	GetQueueResultFunc func() *common.ResponseForTaskQueue

	// GetUserIDsFunc mocks the GetUserIDs method.
	GetUserIDsFunc func() []string

	// GetWorkspaceFunc mocks the GetWorkspace method.
	GetWorkspaceFunc func() *entity.Workspace

	// HasErrorFunc mocks the HasError method.
	HasErrorFunc func() bool

	// InsertChildDataLogFunc mocks the InsertChildDataLog method.
	InsertChildDataLogFunc func(ctx context.Context, data entity.ChildDataLog) error

	// LogFunc mocks the Log method.
	LogFunc func() *logrus.Logger

	// NetFunc mocks the Net method.
	NetFunc func() httpClient.HTTPClient

	// ProcessNextStepFunc mocks the ProcessNextStep method.
	ProcessNextStepFunc func(ctx context.Context)

	// ReattributeUsersOrdersFunc mocks the ReattributeUsersOrders method.
	ReattributeUsersOrdersFunc func(ctx context.Context)

	// ReleaseUsersLockFunc mocks the ReleaseUsersLock method.
	ReleaseUsersLockFunc func() error

	// RepoFunc mocks the Repo method.
	RepoFunc func() repository.Repository

	// SetErrorFunc mocks the SetError method.
	SetErrorFunc func(key string, err string, retryable bool)

	// calls tracks calls to the methods.
	calls struct {
		// AddDataLogGenerated holds details about calls to the AddDataLogGenerated method.
		AddDataLogGenerated []struct {
			// DataLog is the dataLog argument value.
			DataLog *entity.DataLog
		}
		// AttributeOrder holds details about calls to the AttributeOrder method.
		AttributeOrder []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Order is the order argument value.
			Order *entity.Order
			// OrderSessions is the orderSessions argument value.
			OrderSessions []*entity.Session
			// OrderPostviews is the orderPostviews argument value.
			OrderPostviews []*entity.Postview
			// PreviousOrders is the previousOrders argument value.
			PreviousOrders []*entity.Order
			// Devices is the devices argument value.
			Devices []*entity.Device
			// Tx is the tx argument value.
			Tx *sql.Tx
		}
		// Cfg holds details about calls to the Cfg method.
		Cfg []struct {
		}
		// DataLogEnqueue holds details about calls to the DataLogEnqueue method.
		DataLogEnqueue []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// ReplayID is the replayID argument value.
			ReplayID *string
			// Origin is the origin argument value.
			Origin int
			// OriginID is the originID argument value.
			OriginID string
			// WorkspaceID is the workspaceID argument value.
			WorkspaceID string
			// JsonItems is the jsonItems argument value.
			JsonItems []string
			// IsSync is the isSync argument value.
			IsSync bool
		}
		// EnsureUsersLock holds details about calls to the EnsureUsersLock method.
		EnsureUsersLock []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// Execute holds details about calls to the Execute method.
		Execute []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// GetDataLogsGenerated holds details about calls to the GetDataLogsGenerated method.
		GetDataLogsGenerated []struct {
		}
		// GetQueueResult holds details about calls to the GetQueueResult method.
		GetQueueResult []struct {
		}
		// GetUserIDs holds details about calls to the GetUserIDs method.
		GetUserIDs []struct {
		}
		// GetWorkspace holds details about calls to the GetWorkspace method.
		GetWorkspace []struct {
		}
		// HasError holds details about calls to the HasError method.
		HasError []struct {
		}
		// InsertChildDataLog holds details about calls to the InsertChildDataLog method.
		InsertChildDataLog []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Data is the data argument value.
			Data entity.ChildDataLog
		}
		// Log holds details about calls to the Log method.
		Log []struct {
		}
		// Net holds details about calls to the Net method.
		Net []struct {
		}
		// ProcessNextStep holds details about calls to the ProcessNextStep method.
		ProcessNextStep []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// ReattributeUsersOrders holds details about calls to the ReattributeUsersOrders method.
		ReattributeUsersOrders []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// ReleaseUsersLock holds details about calls to the ReleaseUsersLock method.
		ReleaseUsersLock []struct {
		}
		// Repo holds details about calls to the Repo method.
		Repo []struct {
		}
		// SetError holds details about calls to the SetError method.
		SetError []struct {
			// Key is the key argument value.
			Key string
			// Err is the err argument value.
			Err string
			// Retryable is the retryable argument value.
			Retryable bool
		}
	}
	lockAddDataLogGenerated    sync.RWMutex
	lockAttributeOrder         sync.RWMutex
	lockCfg                    sync.RWMutex
	lockDataLogEnqueue         sync.RWMutex
	lockEnsureUsersLock        sync.RWMutex
	lockExecute                sync.RWMutex
	lockGetDataLogsGenerated   sync.RWMutex
	lockGetQueueResult         sync.RWMutex
	lockGetUserIDs             sync.RWMutex
	lockGetWorkspace           sync.RWMutex
	lockHasError               sync.RWMutex
	lockInsertChildDataLog     sync.RWMutex
	lockLog                    sync.RWMutex
	lockNet                    sync.RWMutex
	lockProcessNextStep        sync.RWMutex
	lockReattributeUsersOrders sync.RWMutex
	lockReleaseUsersLock       sync.RWMutex
	lockRepo                   sync.RWMutex
	lockSetError               sync.RWMutex
}

// AddDataLogGenerated calls AddDataLogGeneratedFunc.
func (mock *PipelineMock) AddDataLogGenerated(dataLog *entity.DataLog) {
	if mock.AddDataLogGeneratedFunc == nil {
		panic("PipelineMock.AddDataLogGeneratedFunc: method is nil but Pipeline.AddDataLogGenerated was just called")
	}
	callInfo := struct {
		DataLog *entity.DataLog
	}{
		DataLog: dataLog,
	}
	mock.lockAddDataLogGenerated.Lock()
	mock.calls.AddDataLogGenerated = append(mock.calls.AddDataLogGenerated, callInfo)
	mock.lockAddDataLogGenerated.Unlock()
	mock.AddDataLogGeneratedFunc(dataLog)
}

// AddDataLogGeneratedCalls gets all the calls that were made to AddDataLogGenerated.
// Check the length with:
//
//	len(mockedPipeline.AddDataLogGeneratedCalls())
func (mock *PipelineMock) AddDataLogGeneratedCalls() []struct {
	DataLog *entity.DataLog
} {
	var calls []struct {
		DataLog *entity.DataLog
	}
	mock.lockAddDataLogGenerated.RLock()
	calls = mock.calls.AddDataLogGenerated
	mock.lockAddDataLogGenerated.RUnlock()
	return calls
}

// AttributeOrder calls AttributeOrderFunc.
func (mock *PipelineMock) AttributeOrder(ctx context.Context, order *entity.Order, orderSessions []*entity.Session, orderPostviews []*entity.Postview, previousOrders []*entity.Order, devices []*entity.Device, tx *sql.Tx) {
	if mock.AttributeOrderFunc == nil {
		panic("PipelineMock.AttributeOrderFunc: method is nil but Pipeline.AttributeOrder was just called")
	}
	callInfo := struct {
		Ctx            context.Context
		Order          *entity.Order
		OrderSessions  []*entity.Session
		OrderPostviews []*entity.Postview
		PreviousOrders []*entity.Order
		Devices        []*entity.Device
		Tx             *sql.Tx
	}{
		Ctx:            ctx,
		Order:          order,
		OrderSessions:  orderSessions,
		OrderPostviews: orderPostviews,
		PreviousOrders: previousOrders,
		Devices:        devices,
		Tx:             tx,
	}
	mock.lockAttributeOrder.Lock()
	mock.calls.AttributeOrder = append(mock.calls.AttributeOrder, callInfo)
	mock.lockAttributeOrder.Unlock()
	mock.AttributeOrderFunc(ctx, order, orderSessions, orderPostviews, previousOrders, devices, tx)
}

// AttributeOrderCalls gets all the calls that were made to AttributeOrder.
// Check the length with:
//
//	len(mockedPipeline.AttributeOrderCalls())
func (mock *PipelineMock) AttributeOrderCalls() []struct {
	Ctx            context.Context
	Order          *entity.Order
	OrderSessions  []*entity.Session
	OrderPostviews []*entity.Postview
	PreviousOrders []*entity.Order
	Devices        []*entity.Device
	Tx             *sql.Tx
} {
	var calls []struct {
		Ctx            context.Context
		Order          *entity.Order
		OrderSessions  []*entity.Session
		OrderPostviews []*entity.Postview
		PreviousOrders []*entity.Order
		Devices        []*entity.Device
		Tx             *sql.Tx
	}
	mock.lockAttributeOrder.RLock()
	calls = mock.calls.AttributeOrder
	mock.lockAttributeOrder.RUnlock()
	return calls
}

// Cfg calls CfgFunc.
func (mock *PipelineMock) Cfg() *entity.Config {
	if mock.CfgFunc == nil {
		panic("PipelineMock.CfgFunc: method is nil but Pipeline.Cfg was just called")
	}
	callInfo := struct {
	}{}
	mock.lockCfg.Lock()
	mock.calls.Cfg = append(mock.calls.Cfg, callInfo)
	mock.lockCfg.Unlock()
	return mock.CfgFunc()
}

// CfgCalls gets all the calls that were made to Cfg.
// Check the length with:
//
//	len(mockedPipeline.CfgCalls())
func (mock *PipelineMock) CfgCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockCfg.RLock()
	calls = mock.calls.Cfg
	mock.lockCfg.RUnlock()
	return calls
}

// DataLogEnqueue calls DataLogEnqueueFunc.
func (mock *PipelineMock) DataLogEnqueue(ctx context.Context, replayID *string, origin int, originID string, workspaceID string, jsonItems []string, isSync bool) {
	if mock.DataLogEnqueueFunc == nil {
		panic("PipelineMock.DataLogEnqueueFunc: method is nil but Pipeline.DataLogEnqueue was just called")
	}
	callInfo := struct {
		Ctx         context.Context
		ReplayID    *string
		Origin      int
		OriginID    string
		WorkspaceID string
		JsonItems   []string
		IsSync      bool
	}{
		Ctx:         ctx,
		ReplayID:    replayID,
		Origin:      origin,
		OriginID:    originID,
		WorkspaceID: workspaceID,
		JsonItems:   jsonItems,
		IsSync:      isSync,
	}
	mock.lockDataLogEnqueue.Lock()
	mock.calls.DataLogEnqueue = append(mock.calls.DataLogEnqueue, callInfo)
	mock.lockDataLogEnqueue.Unlock()
	mock.DataLogEnqueueFunc(ctx, replayID, origin, originID, workspaceID, jsonItems, isSync)
}

// DataLogEnqueueCalls gets all the calls that were made to DataLogEnqueue.
// Check the length with:
//
//	len(mockedPipeline.DataLogEnqueueCalls())
func (mock *PipelineMock) DataLogEnqueueCalls() []struct {
	Ctx         context.Context
	ReplayID    *string
	Origin      int
	OriginID    string
	WorkspaceID string
	JsonItems   []string
	IsSync      bool
} {
	var calls []struct {
		Ctx         context.Context
		ReplayID    *string
		Origin      int
		OriginID    string
		WorkspaceID string
		JsonItems   []string
		IsSync      bool
	}
	mock.lockDataLogEnqueue.RLock()
	calls = mock.calls.DataLogEnqueue
	mock.lockDataLogEnqueue.RUnlock()
	return calls
}

// EnsureUsersLock calls EnsureUsersLockFunc.
func (mock *PipelineMock) EnsureUsersLock(ctx context.Context) error {
	if mock.EnsureUsersLockFunc == nil {
		panic("PipelineMock.EnsureUsersLockFunc: method is nil but Pipeline.EnsureUsersLock was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockEnsureUsersLock.Lock()
	mock.calls.EnsureUsersLock = append(mock.calls.EnsureUsersLock, callInfo)
	mock.lockEnsureUsersLock.Unlock()
	return mock.EnsureUsersLockFunc(ctx)
}

// EnsureUsersLockCalls gets all the calls that were made to EnsureUsersLock.
// Check the length with:
//
//	len(mockedPipeline.EnsureUsersLockCalls())
func (mock *PipelineMock) EnsureUsersLockCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockEnsureUsersLock.RLock()
	calls = mock.calls.EnsureUsersLock
	mock.lockEnsureUsersLock.RUnlock()
	return calls
}

// Execute calls ExecuteFunc.
func (mock *PipelineMock) Execute(ctx context.Context) {
	if mock.ExecuteFunc == nil {
		panic("PipelineMock.ExecuteFunc: method is nil but Pipeline.Execute was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockExecute.Lock()
	mock.calls.Execute = append(mock.calls.Execute, callInfo)
	mock.lockExecute.Unlock()
	mock.ExecuteFunc(ctx)
}

// ExecuteCalls gets all the calls that were made to Execute.
// Check the length with:
//
//	len(mockedPipeline.ExecuteCalls())
func (mock *PipelineMock) ExecuteCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockExecute.RLock()
	calls = mock.calls.Execute
	mock.lockExecute.RUnlock()
	return calls
}

// GetDataLogsGenerated calls GetDataLogsGeneratedFunc.
func (mock *PipelineMock) GetDataLogsGenerated() []*entity.DataLog {
	if mock.GetDataLogsGeneratedFunc == nil {
		panic("PipelineMock.GetDataLogsGeneratedFunc: method is nil but Pipeline.GetDataLogsGenerated was just called")
	}
	callInfo := struct {
	}{}
	mock.lockGetDataLogsGenerated.Lock()
	mock.calls.GetDataLogsGenerated = append(mock.calls.GetDataLogsGenerated, callInfo)
	mock.lockGetDataLogsGenerated.Unlock()
	return mock.GetDataLogsGeneratedFunc()
}

// GetDataLogsGeneratedCalls gets all the calls that were made to GetDataLogsGenerated.
// Check the length with:
//
//	len(mockedPipeline.GetDataLogsGeneratedCalls())
func (mock *PipelineMock) GetDataLogsGeneratedCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockGetDataLogsGenerated.RLock()
	calls = mock.calls.GetDataLogsGenerated
	mock.lockGetDataLogsGenerated.RUnlock()
	return calls
}

// GetQueueResult calls GetQueueResultFunc.
func (mock *PipelineMock) GetQueueResult() *common.ResponseForTaskQueue {
	if mock.GetQueueResultFunc == nil {
		panic("PipelineMock.GetQueueResultFunc: method is nil but Pipeline.GetQueueResult was just called")
	}
	callInfo := struct {
	}{}
	mock.lockGetQueueResult.Lock()
	mock.calls.GetQueueResult = append(mock.calls.GetQueueResult, callInfo)
	mock.lockGetQueueResult.Unlock()
	return mock.GetQueueResultFunc()
}

// GetQueueResultCalls gets all the calls that were made to GetQueueResult.
// Check the length with:
//
//	len(mockedPipeline.GetQueueResultCalls())
func (mock *PipelineMock) GetQueueResultCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockGetQueueResult.RLock()
	calls = mock.calls.GetQueueResult
	mock.lockGetQueueResult.RUnlock()
	return calls
}

// GetUserIDs calls GetUserIDsFunc.
func (mock *PipelineMock) GetUserIDs() []string {
	if mock.GetUserIDsFunc == nil {
		panic("PipelineMock.GetUserIDsFunc: method is nil but Pipeline.GetUserIDs was just called")
	}
	callInfo := struct {
	}{}
	mock.lockGetUserIDs.Lock()
	mock.calls.GetUserIDs = append(mock.calls.GetUserIDs, callInfo)
	mock.lockGetUserIDs.Unlock()
	return mock.GetUserIDsFunc()
}

// GetUserIDsCalls gets all the calls that were made to GetUserIDs.
// Check the length with:
//
//	len(mockedPipeline.GetUserIDsCalls())
func (mock *PipelineMock) GetUserIDsCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockGetUserIDs.RLock()
	calls = mock.calls.GetUserIDs
	mock.lockGetUserIDs.RUnlock()
	return calls
}

// GetWorkspace calls GetWorkspaceFunc.
func (mock *PipelineMock) GetWorkspace() *entity.Workspace {
	if mock.GetWorkspaceFunc == nil {
		panic("PipelineMock.GetWorkspaceFunc: method is nil but Pipeline.GetWorkspace was just called")
	}
	callInfo := struct {
	}{}
	mock.lockGetWorkspace.Lock()
	mock.calls.GetWorkspace = append(mock.calls.GetWorkspace, callInfo)
	mock.lockGetWorkspace.Unlock()
	return mock.GetWorkspaceFunc()
}

// GetWorkspaceCalls gets all the calls that were made to GetWorkspace.
// Check the length with:
//
//	len(mockedPipeline.GetWorkspaceCalls())
func (mock *PipelineMock) GetWorkspaceCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockGetWorkspace.RLock()
	calls = mock.calls.GetWorkspace
	mock.lockGetWorkspace.RUnlock()
	return calls
}

// HasError calls HasErrorFunc.
func (mock *PipelineMock) HasError() bool {
	if mock.HasErrorFunc == nil {
		panic("PipelineMock.HasErrorFunc: method is nil but Pipeline.HasError was just called")
	}
	callInfo := struct {
	}{}
	mock.lockHasError.Lock()
	mock.calls.HasError = append(mock.calls.HasError, callInfo)
	mock.lockHasError.Unlock()
	return mock.HasErrorFunc()
}

// HasErrorCalls gets all the calls that were made to HasError.
// Check the length with:
//
//	len(mockedPipeline.HasErrorCalls())
func (mock *PipelineMock) HasErrorCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockHasError.RLock()
	calls = mock.calls.HasError
	mock.lockHasError.RUnlock()
	return calls
}

// InsertChildDataLog calls InsertChildDataLogFunc.
func (mock *PipelineMock) InsertChildDataLog(ctx context.Context, data entity.ChildDataLog) error {
	if mock.InsertChildDataLogFunc == nil {
		panic("PipelineMock.InsertChildDataLogFunc: method is nil but Pipeline.InsertChildDataLog was just called")
	}
	callInfo := struct {
		Ctx  context.Context
		Data entity.ChildDataLog
	}{
		Ctx:  ctx,
		Data: data,
	}
	mock.lockInsertChildDataLog.Lock()
	mock.calls.InsertChildDataLog = append(mock.calls.InsertChildDataLog, callInfo)
	mock.lockInsertChildDataLog.Unlock()
	return mock.InsertChildDataLogFunc(ctx, data)
}

// InsertChildDataLogCalls gets all the calls that were made to InsertChildDataLog.
// Check the length with:
//
//	len(mockedPipeline.InsertChildDataLogCalls())
func (mock *PipelineMock) InsertChildDataLogCalls() []struct {
	Ctx  context.Context
	Data entity.ChildDataLog
} {
	var calls []struct {
		Ctx  context.Context
		Data entity.ChildDataLog
	}
	mock.lockInsertChildDataLog.RLock()
	calls = mock.calls.InsertChildDataLog
	mock.lockInsertChildDataLog.RUnlock()
	return calls
}

// Log calls LogFunc.
func (mock *PipelineMock) Log() *logrus.Logger {
	if mock.LogFunc == nil {
		panic("PipelineMock.LogFunc: method is nil but Pipeline.Log was just called")
	}
	callInfo := struct {
	}{}
	mock.lockLog.Lock()
	mock.calls.Log = append(mock.calls.Log, callInfo)
	mock.lockLog.Unlock()
	return mock.LogFunc()
}

// LogCalls gets all the calls that were made to Log.
// Check the length with:
//
//	len(mockedPipeline.LogCalls())
func (mock *PipelineMock) LogCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockLog.RLock()
	calls = mock.calls.Log
	mock.lockLog.RUnlock()
	return calls
}

// Net calls NetFunc.
func (mock *PipelineMock) Net() httpClient.HTTPClient {
	if mock.NetFunc == nil {
		panic("PipelineMock.NetFunc: method is nil but Pipeline.Net was just called")
	}
	callInfo := struct {
	}{}
	mock.lockNet.Lock()
	mock.calls.Net = append(mock.calls.Net, callInfo)
	mock.lockNet.Unlock()
	return mock.NetFunc()
}

// NetCalls gets all the calls that were made to Net.
// Check the length with:
//
//	len(mockedPipeline.NetCalls())
func (mock *PipelineMock) NetCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockNet.RLock()
	calls = mock.calls.Net
	mock.lockNet.RUnlock()
	return calls
}

// ProcessNextStep calls ProcessNextStepFunc.
func (mock *PipelineMock) ProcessNextStep(ctx context.Context) {
	if mock.ProcessNextStepFunc == nil {
		panic("PipelineMock.ProcessNextStepFunc: method is nil but Pipeline.ProcessNextStep was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockProcessNextStep.Lock()
	mock.calls.ProcessNextStep = append(mock.calls.ProcessNextStep, callInfo)
	mock.lockProcessNextStep.Unlock()
	mock.ProcessNextStepFunc(ctx)
}

// ProcessNextStepCalls gets all the calls that were made to ProcessNextStep.
// Check the length with:
//
//	len(mockedPipeline.ProcessNextStepCalls())
func (mock *PipelineMock) ProcessNextStepCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockProcessNextStep.RLock()
	calls = mock.calls.ProcessNextStep
	mock.lockProcessNextStep.RUnlock()
	return calls
}

// ReattributeUsersOrders calls ReattributeUsersOrdersFunc.
func (mock *PipelineMock) ReattributeUsersOrders(ctx context.Context) {
	if mock.ReattributeUsersOrdersFunc == nil {
		panic("PipelineMock.ReattributeUsersOrdersFunc: method is nil but Pipeline.ReattributeUsersOrders was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockReattributeUsersOrders.Lock()
	mock.calls.ReattributeUsersOrders = append(mock.calls.ReattributeUsersOrders, callInfo)
	mock.lockReattributeUsersOrders.Unlock()
	mock.ReattributeUsersOrdersFunc(ctx)
}

// ReattributeUsersOrdersCalls gets all the calls that were made to ReattributeUsersOrders.
// Check the length with:
//
//	len(mockedPipeline.ReattributeUsersOrdersCalls())
func (mock *PipelineMock) ReattributeUsersOrdersCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockReattributeUsersOrders.RLock()
	calls = mock.calls.ReattributeUsersOrders
	mock.lockReattributeUsersOrders.RUnlock()
	return calls
}

// ReleaseUsersLock calls ReleaseUsersLockFunc.
func (mock *PipelineMock) ReleaseUsersLock() error {
	if mock.ReleaseUsersLockFunc == nil {
		panic("PipelineMock.ReleaseUsersLockFunc: method is nil but Pipeline.ReleaseUsersLock was just called")
	}
	callInfo := struct {
	}{}
	mock.lockReleaseUsersLock.Lock()
	mock.calls.ReleaseUsersLock = append(mock.calls.ReleaseUsersLock, callInfo)
	mock.lockReleaseUsersLock.Unlock()
	return mock.ReleaseUsersLockFunc()
}

// ReleaseUsersLockCalls gets all the calls that were made to ReleaseUsersLock.
// Check the length with:
//
//	len(mockedPipeline.ReleaseUsersLockCalls())
func (mock *PipelineMock) ReleaseUsersLockCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockReleaseUsersLock.RLock()
	calls = mock.calls.ReleaseUsersLock
	mock.lockReleaseUsersLock.RUnlock()
	return calls
}

// Repo calls RepoFunc.
func (mock *PipelineMock) Repo() repository.Repository {
	if mock.RepoFunc == nil {
		panic("PipelineMock.RepoFunc: method is nil but Pipeline.Repo was just called")
	}
	callInfo := struct {
	}{}
	mock.lockRepo.Lock()
	mock.calls.Repo = append(mock.calls.Repo, callInfo)
	mock.lockRepo.Unlock()
	return mock.RepoFunc()
}

// RepoCalls gets all the calls that were made to Repo.
// Check the length with:
//
//	len(mockedPipeline.RepoCalls())
func (mock *PipelineMock) RepoCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockRepo.RLock()
	calls = mock.calls.Repo
	mock.lockRepo.RUnlock()
	return calls
}

// SetError calls SetErrorFunc.
func (mock *PipelineMock) SetError(key string, err string, retryable bool) {
	if mock.SetErrorFunc == nil {
		panic("PipelineMock.SetErrorFunc: method is nil but Pipeline.SetError was just called")
	}
	callInfo := struct {
		Key       string
		Err       string
		Retryable bool
	}{
		Key:       key,
		Err:       err,
		Retryable: retryable,
	}
	mock.lockSetError.Lock()
	mock.calls.SetError = append(mock.calls.SetError, callInfo)
	mock.lockSetError.Unlock()
	mock.SetErrorFunc(key, err, retryable)
}

// SetErrorCalls gets all the calls that were made to SetError.
// Check the length with:
//
//	len(mockedPipeline.SetErrorCalls())
func (mock *PipelineMock) SetErrorCalls() []struct {
	Key       string
	Err       string
	Retryable bool
} {
	var calls []struct {
		Key       string
		Err       string
		Retryable bool
	}
	mock.lockSetError.RLock()
	calls = mock.calls.SetError
	mock.lockSetError.RUnlock()
	return calls
}
