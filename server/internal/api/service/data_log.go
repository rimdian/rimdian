package service

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/common"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	commonDTO "github.com/rimdian/rimdian/internal/common/dto"
	"github.com/rimdian/rimdian/internal/common/httpClient"
	"github.com/rotisserie/eris"
)

type CollectorPayload struct {
	WorkspaceID string    `json:"workspace_id"`
	Items       []*string `json:"items"`
}

// enqueue a new data_log to the collector for processing
func DataLogEnqueue(ctx context.Context, cfg *entity.Config, netClient httpClient.HTTPClient, replayID *string, origin int, originID string, workspaceID string, jsonItems []string, isSync bool) (err error) {
	// import data if any
	itemsCount := len(jsonItems)

	if itemsCount == 0 {
		return nil
	}

	// in dev mode, bypass queue and push data back to the API one by one
	if cfg.ENV == entity.ENV_DEV {
		isSync = true
	}

	if isSync && itemsCount > 1 {
		for _, jsonItem := range jsonItems {
			err = DataLogEnqueue(ctx, cfg, netClient, replayID, origin, originID, workspaceID, []string{jsonItem}, isSync)

			if err != nil {
				return eris.Wrap(err, "DataLogEnqueue")
			}
		}
		return
	}

	if itemsCount > 0 {
		// cut items into batches of 250 and import them internally
		const maxBatchSize int = 250
		skip := 0

		batchCount := int(math.Ceil(float64(itemsCount) / float64(maxBatchSize)))

		for i := 1; i <= batchCount; i++ {

			lowerBound := skip
			upperBound := skip + maxBatchSize

			if upperBound > itemsCount {
				upperBound = itemsCount
			}

			skip += maxBatchSize

			batch := jsonItems[lowerBound:upperBound]

			body := fmt.Sprintf(`{
				"workspace_id": "%v",
				"items": [%v]
			}`, workspaceID, strings.Join(batch, ","))

			// POST the batch to the collector

			collectorEndpoint := cfg.COLLECTOR_ENDPOINT + "/data"

			if isSync {
				collectorEndpoint = cfg.COLLECTOR_ENDPOINT + "/bypass"
			}

			req, _ := http.NewRequest("POST", collectorEndpoint, bytes.NewBuffer([]byte(body)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set(commonDTO.HeaderSignature, common.ComputeHMAC256([]byte(body), cfg.SECRET_KEY))
			req.Header.Set(commonDTO.HeaderOrigin, fmt.Sprintf("%v", origin))
			req.Header.Set(commonDTO.HeaderOriginID, fmt.Sprintf("%v", originID))

			if replayID != nil {
				req.Header.Set(commonDTO.HeaderReplayID, *replayID)
			}

			// inject OpenCensus span context into the request
			req = req.WithContext(ctx)

			_, err = netClient.Do(req)

			if err != nil {
				return eris.Wrap(err, "DataLogEnqueue")
			}

		}
	}

	return nil
}

// receives a DataLogInQueue from the Queue and process it
// it can be a replay of a previous data_log
// or a new data_log import
func (svc *ServiceImpl) DataLogImportFromQueue(ctx context.Context, dataLogInQueue *commonDTO.DataLogInQueue) (result *commonDTO.ResponseForTaskQueue) {

	// fetch workspace from DB
	workspace, err := svc.Repo.GetWorkspace(ctx, dataLogInQueue.Context.WorkspaceID)
	// svc.Logger.Printf("workspace %v", workspace)

	if err != nil {
		// check if not found
		if sqlscan.NotFound(err) {
			return &commonDTO.ResponseForTaskQueue{
				HasError:         true,
				Error:            fmt.Sprintf("DataLogImportFromQueue: workspace not found: %v", dataLogInQueue.Context.WorkspaceID),
				QueueShouldRetry: false,
			}
		}

		return &commonDTO.ResponseForTaskQueue{
			HasError:         true,
			Error:            fmt.Sprintf("DataLogImportFromQueue: %v", err),
			QueueShouldRetry: true,
		}
	}

	props := &DataPipelineProps{
		Config:           svc.Config,
		Logger:           svc.Logger,
		NetClient:        svc.NetClient,
		Repository:       svc.Repo,
		TaskOrchestrator: svc.TaskOrchestrator,
		Workspace:        workspace,
		DataLogInQueue:   dataLogInQueue,
	}

	pipeline := NewDataPipeline(props)
	pipeline.Execute(ctx)

	return pipeline.GetQueueResult()
}

// takes a  data import from the DB and  into the queue
func (svc *ServiceImpl) DataLogReprocessOne(ctx context.Context, accountID string, payload *dto.DataLogReprocessOne) (result *commonDTO.ResponseForTaskQueue, code int, err error) {

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, payload.WorkspaceID, accountID)

	if err != nil {
		return nil, code, eris.Wrap(err, "DataLogReprocessOne")
	}

	// fetch dataLog from DB
	dataLog, err := svc.Repo.GetDataLog(ctx, workspace.ID, payload.DataLogID)

	if err != nil {
		// check if not found
		if sqlscan.NotFound(err) {
			return nil, 400, fmt.Errorf("DataLogReprocess: id not found: %v", payload.DataLogID)
		}

		return nil, 500, eris.Wrap(err, "DataLogReprocess")
	}

	dataLogInQueue := &commonDTO.DataLogInQueue{
		ID:       dataLog.ID,
		Context:  dataLog.Context,
		Origin:   dataLog.Origin,
		OriginID: dataLog.OriginID,
		Item:     dataLog.Item,
		IsReplay: true,
	}

	props := &DataPipelineProps{
		Config:           svc.Config,
		Logger:           svc.Logger,
		NetClient:        svc.NetClient,
		Repository:       svc.Repo,
		TaskOrchestrator: svc.TaskOrchestrator,
		Workspace:        workspace,
		DataLogInQueue:   dataLogInQueue,
	}

	pipeline := NewDataPipeline(props)
	pipeline.Execute(ctx)

	return pipeline.GetQueueResult(), 200, nil
}

func (svc *ServiceImpl) DataLogList(ctx context.Context, accountID string, params *dto.DataLogListParams) (result *dto.DataLogListResult, code int, err error) {

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, params.WorkspaceID, accountID)

	if err != nil {
		return nil, code, eris.Wrap(err, "DataLogList")
	}

	// fetch tasks
	result = &dto.DataLogListResult{}

	result.DataLogs, result.NextToken, code, err = svc.Repo.ListDataLogs(ctx, workspace.ID, params)

	if err != nil {
		return nil, code, err
	}

	return result, 200, nil
}
