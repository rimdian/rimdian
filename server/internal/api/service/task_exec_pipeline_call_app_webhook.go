package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rimdian/rimdian/internal/api/entity"
	commonDTO "github.com/rimdian/rimdian/internal/common/dto"
	"go.opencensus.io/trace"
)

// receives a task request from the task orchestrator
// and should send the payload to the external webhook endpoint
func TaskExecCallAppWebhook(ctx context.Context, pipe *TaskExecPipeline) (result *entity.TaskExecResult) {

	spanCtx, span := trace.StartSpan(ctx, "TaskExecCallAppWebhook")
	defer span.End()

	result = &entity.TaskExecResult{
		// keep current state by default
		UpdatedWorkerState: pipe.TaskExec.State.Workers[pipe.TaskExecPayload.WorkerID],
	}

	select {
	case <-spanCtx.Done():
		result.SetError("task_exec timeout", false)
		return
	default:
	}

	// webhook data is in the task state
	if pipe.TaskExec.State.Workers[pipe.TaskExecPayload.WorkerID] == nil {
		result.SetError("task_exec worker state not found", true)
		return
	}

	if !strings.HasPrefix(pipe.TaskExec.TaskID, "app_") && !strings.HasPrefix(pipe.TaskExec.TaskID, "appx_") {
		result.SetError("task_exec kind not supported", true)
		return
	}

	// task.Kind is like "app_shopify_sync", we extract "app_shopify" to get the app_id

	bits := strings.Split(pipe.TaskExec.TaskID, "_")

	if len(bits) < 3 {
		result.SetError("task_exec kind not supported", true)
		return
	}

	appID := bits[0] + "_" + bits[1]

	// get app
	app, err := pipe.Repository.GetApp(spanCtx, pipe.Workspace.ID, appID)

	if err != nil {
		result.SetError("task_exec app not found", true)
		return
	}

	if app.Status != entity.AppStatusActive {
		result.SetError("task_exec app not active", true)
		return
	}

	payloadSent := entity.AppWebhookPayload{
		APIEndpoint:       pipe.Config.API_ENDPOINT,
		CollectorEndpoint: pipe.Config.COLLECTOR_ENDPOINT,
		WorkspaceID:       pipe.Workspace.ID,
		AppID:             app.ID,
		AppState:          app.State,
		Kind:              entity.AppWebhookKindTaskExec,
		TaskExecWorker: &entity.TaskExecWorker{
			TaskID:            pipe.TaskExec.TaskID,
			TaskName:          pipe.TaskExec.Name,
			TaskExecID:        pipe.TaskExec.ID,
			TaskExecCreatedAt: *pipe.TaskExec.DBCreatedAt,
			WorkerID:          pipe.TaskExecPayload.WorkerID,
			WorkerState:       pipe.TaskExec.State.Workers[pipe.TaskExecPayload.WorkerID],
			RetryCount:        pipe.TaskExec.RetryCount,
		},
	}

	data, err := json.Marshal(payloadSent)

	if err != nil {
		result.SetError("task_exec invalid payload", true)
		return
	}

	appKey, err := app.GetAppSecretKey(pipe.Config.SECRET_KEY)

	if err != nil {
		result.SetError("task_exec invalid app key", true)
		return
	}

	// compute HMAC signature of the payload
	h := hmac.New(sha256.New, []byte(appKey))
	h.Write([]byte(data))
	hmac256 := hex.EncodeToString(h.Sum(nil))

	// send request to external webhook
	req, _ := http.NewRequest("POST", app.Manifest.WebhookEndpoint, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(commonDTO.HeaderSignature, hmac256)

	res, err := pipe.NetClient.Do(req)

	if err != nil {
		result.SetError(fmt.Sprintf("task_exec call app endpoint err: %v, retry in %vs", err, entity.TaskRetryDelayInSecs), false)
		result.DelayNextRequestInSecs = &entity.TaskRetryDelayInSecs
		return
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	if err != nil {
		result.SetError(fmt.Sprintf("task_exec call app endpoint response err: %v, retry in %vs", err, entity.TaskRetryDelayInSecs), false)
		result.DelayNextRequestInSecs = &entity.TaskRetryDelayInSecs
		return
	}

	// unmarschal response
	appResult := &entity.TaskExecResult{}

	if err := json.Unmarshal(body, &appResult); err != nil {
		result.SetError(fmt.Sprintf("task_exec call app invalid response err: %v", err), true)
		return
	}

	if appResult == nil {
		result.SetError(fmt.Sprintf("task_exec call app invalid response err: %v, body: %v", err, string(body)), true)
		return
	}

	// use the app endpoint response has result
	result = appResult

	// reset fields to avoid json injection
	result.DelayNextRequestInSecs = nil
	result.WorkerID = pipe.TaskExecPayload.WorkerID

	if res.StatusCode != 200 {

		if res.StatusCode == 400 {
			result.SetError(fmt.Sprintf("task_exec call app endpoint status: %v, body: %v", res.StatusCode, string(body)), true)
			return
		}

		// error 500, retry later
		result.SetError(fmt.Sprintf("task_exec call app endpoint status: %v, body: %v, retry in %vs", res.StatusCode, string(body), entity.TaskRetryDelayInSecs), false)
		result.DelayNextRequestInSecs = &entity.TaskRetryDelayInSecs
		return
	}

	return
}
