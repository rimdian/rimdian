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
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/rimdian/rimdian/internal/api/entity"
	commonDTO "github.com/rimdian/rimdian/internal/common/dto"
	"go.opencensus.io/trace"
)

type hookResult struct {
	DataLogID     string
	HookID        string
	IsDone        bool
	ErrorFromHook bool
	Message       string
	InternalError *string
	ShouldRetry   bool
}

func (pipe *DataLogPipeline) StepHookFinalize(ctx context.Context) {

	spanCtx, span := trace.StartSpan(ctx, "StepHookFinalize")
	defer span.End()

	hookPayloads := []*entity.AppWebhookPayload{}

	// get app states
	apps, err := pipe.Repository.ListApps(ctx, pipe.Workspace.ID)

	if err != nil {
		pipe.SetError("hook", fmt.Sprintf("error listing apps: %s", err), true)
		return
	}

	// children item is the parent item
	parentDataLogItem := pipe.DataLog.Item

	for _, hook := range pipe.Workspace.DataHooks {

		// get hooks for this data log and its children
		for _, dataLog := range pipe.GetDataLogsGenerated() {

			// parent data_log hooks might not be initialized yet
			if dataLog.ID == pipe.DataLog.ID && len(dataLog.Hooks) == 0 {
				for _, hook := range pipe.Workspace.DataHooks {
					// only on_success hooks here
					if (hook.On == entity.DataHookKindOnSuccess && dataLog.HasError == 0) &&
						hook.MatchesDataLog(dataLog.Kind, dataLog.Action) {
						// set default hook state
						dataLog.Hooks[hook.ID] = &entity.DataHookState{
							Done: false,
						}
					}
				}
			}

			if _, ok := dataLog.Hooks[hook.ID]; ok {

				// check that hook is not done already
				if dataLog.Hooks[hook.ID].Done {
					continue
				}

				// default app state (for system hooks)
				appState := entity.MapOfInterfaces{}

				// get app state
				for _, app := range apps {
					if app.ID == hook.AppID {
						appState = app.State
						break
					}
				}

				hookPayloads = append(hookPayloads, &entity.AppWebhookPayload{
					APIEndpoint:       pipe.Config.API_ENDPOINT,
					CollectorEndpoint: pipe.Config.COLLECTOR_ENDPOINT,
					WorkspaceID:       pipe.Workspace.ID,
					AppID:             hook.AppID,
					AppState:          appState,
					Kind:              entity.AppWebhookKindDataHook,
					DataHook: &entity.DataHookPayload{
						DataHookID:            hook.ID,
						DataHookName:          hook.Name,
						DataHookOn:            hook.On,
						DataLogID:             dataLog.ID,
						DataLogKind:           dataLog.Kind,
						DataLogAction:         dataLog.Action,
						DataLogItem:           parentDataLogItem,
						DataLogItemID:         dataLog.ItemID,
						DataLogItemExternalID: dataLog.ItemExternalID,
						DataLogUpdatedFields:  dataLog.UpdatedFields,
						User:                  dataLog.UpsertedUser,
					},
				})
			}
		}
	}

	total := len(hookPayloads)

	if total == 0 {
		pipe.DataLog.Checkpoint = entity.DataLogCheckpointHooksFinalizeExecuted
		return
	}

	// process hooks in parallel
	var wg sync.WaitGroup

	// buffered channel to receive errors from goroutines
	// buffered channel are non-blocking, we can write into it and read values later
	// as long as we don't write more values than its length
	resultsChan := make(chan hookResult, total)

	// restrict the roundtripper to 7 seconds
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	for i := range hookPayloads {

		wg.Add(1)

		payloadRef := hookPayloads[i]

		go func(payload entity.AppWebhookPayload) {
			defer wg.Done()

			// call webhooks with a timeout of 5 seconds
			_, cancel := context.WithTimeout(spanCtx, 5*time.Second)
			defer cancel()

			if payload.AppID == "system" {
				// TODO: call internal function
				log.Println("TODO: implement system hook")
				return
			} else {

				// get app endpoint
				var app *entity.App

				for _, a := range apps {
					if a.ID == payload.AppID && a.Status == entity.AppStatusActive {
						app = a
						break
					}
				}

				if app == nil {
					resultsChan <- hookResult{
						DataLogID:     payload.DataHook.DataLogID,
						HookID:        payload.DataHook.DataHookID,
						IsDone:        true,
						Message:       "app not found/active",
						InternalError: nil,
						ShouldRetry:   false,
					}
					return
				}

				data, err := json.Marshal(payload)

				if err != nil {
					msg := fmt.Sprintf("error marshalling hook payload: %s", err)
					resultsChan <- hookResult{
						DataLogID:     payload.DataHook.DataLogID,
						HookID:        payload.DataHook.DataHookID,
						IsDone:        false,
						Message:       msg,
						InternalError: &msg,
						ShouldRetry:   false,
					}
					return
				}

				// compute HMAC signature of the payload
				appKey, err := app.GetAppSecretKey(pipe.Config.SECRET_KEY)

				if err != nil {
					msg := fmt.Sprintf("error getting app key: %s", err)
					resultsChan <- hookResult{
						DataLogID:     payload.DataHook.DataLogID,
						HookID:        payload.DataHook.DataHookID,
						IsDone:        false,
						Message:       msg,
						InternalError: &msg,
						ShouldRetry:   false,
					}
					return
				}

				h := hmac.New(sha256.New, []byte(appKey))
				h.Write([]byte(data))
				hmac256 := hex.EncodeToString(h.Sum(nil))

				// send request to external webhook
				req, _ := http.NewRequestWithContext(ctxWithTimeout, "POST", app.Manifest.WebhookEndpoint, bytes.NewBuffer(data))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set(commonDTO.HeaderSignature, hmac256)

				res, err := pipe.NetClient.Do(req)

				if err != nil {
					msg := fmt.Sprintf("error calling app endpoint: %s", err)
					resultsChan <- hookResult{
						DataLogID:     payload.DataHook.DataLogID,
						HookID:        payload.DataHook.DataHookID,
						IsDone:        false,
						Message:       msg,
						InternalError: &msg,
						ShouldRetry:   true,
					}
					return
				}

				defer res.Body.Close()

				body, err := io.ReadAll(res.Body)

				if err != nil {
					// check if its timeout
					if ctxWithTimeout.Err() == context.DeadlineExceeded {

						// should retry on timeout
						resultsChan <- hookResult{
							DataLogID:     payload.DataHook.DataLogID,
							HookID:        payload.DataHook.DataHookID,
							IsDone:        false,
							ErrorFromHook: true,
							Message:       "timeout after 7 secs",
							InternalError: nil,
							ShouldRetry:   true,
						}
						return
					}

					msg := fmt.Sprintf("error reading app response: %s", err)
					resultsChan <- hookResult{
						DataLogID:     payload.DataHook.DataLogID,
						HookID:        payload.DataHook.DataHookID,
						IsDone:        true,
						Message:       msg,
						InternalError: nil,
						ShouldRetry:   false,
					}
					pipe.SetError("hook", fmt.Sprintf("error reading app response: %s", err), true)
					return
				}

				// return body as message if status code is not 2xx
				if res.StatusCode >= 400 {
					resultsChan <- hookResult{
						DataLogID:     payload.DataHook.DataLogID,
						HookID:        payload.DataHook.DataHookID,
						IsDone:        true,
						ErrorFromHook: true,
						Message:       fmt.Sprintf("app returned: %d - %v", res.StatusCode, string(body)),
						InternalError: nil,
						ShouldRetry:   false,
					}
					return
				}

				// webhooks return a TaskExecResult
				var taskExecResult *entity.TaskExecResult
				msg := string(body)

				// extract message from task exec result
				if err = json.Unmarshal(body, &taskExecResult); err == nil && taskExecResult.Message != nil {
					msg = *taskExecResult.Message
				}

				resultsChan <- hookResult{
					DataLogID:     payload.DataHook.DataLogID,
					HookID:        payload.DataHook.DataHookID,
					IsDone:        true,
					Message:       msg,
					InternalError: nil,
					ShouldRetry:   false,
				}

				return
			}
		}(*payloadRef)
	}

	// wait for all goroutines to finish
	wg.Wait()

	// close the channel, data written into it are still available
	close(resultsChan)

	dataLogChildrenToUpdate := []*entity.DataLog{}

	// check for errors
	for result := range resultsChan {

		if result.InternalError != nil {
			pipe.SetError("hook", fmt.Sprintf("error executing hook: %s", *result.InternalError), result.ShouldRetry)
		}

		// find datalog and update hook state
		for _, dataLog := range pipe.GetDataLogsGenerated() {
			if dataLog.ID == result.DataLogID {

				// init hooks state if not exists
				if _, ok := dataLog.Hooks[result.HookID]; !ok {
					dataLog.Hooks[result.HookID] = &entity.DataHookState{
						Done:    false,
						Message: "",
					}
				}

				dataLog.Hooks[result.HookID].Done = result.IsDone
				dataLog.Hooks[result.HookID].IsError = result.ErrorFromHook
				dataLog.Hooks[result.HookID].Message = result.Message

				if result.ShouldRetry {
					dataLog.Hooks[result.HookID].TriedCount++
				}

				// abort hook after 3 retries
				if dataLog.Hooks[result.HookID].TriedCount >= 3 {
					dataLog.Hooks[result.HookID].Done = true
					dataLog.Hooks[result.HookID].IsError = true
					dataLog.Hooks[result.HookID].Message = fmt.Sprintf("hook aborted after %d retries, err: %v", dataLog.Hooks[result.HookID].TriedCount, result.Message)
				}

				// if is children
				if dataLog.ID != pipe.DataLog.ID {
					dataLogChildrenToUpdate = append(dataLogChildrenToUpdate, dataLog)
				}
			}
		}
	}

	if len(dataLogChildrenToUpdate) > 0 {
		// update children
		for _, dataLog := range dataLogChildrenToUpdate {

			// check if all hooks are done for this data log
			allDone := true
			for _, hookState := range dataLog.Hooks {
				if !hookState.Done {
					allDone = false
					break
				}
			}

			// if all hooks are done, mark data log as done
			if allDone {
				dataLog.Checkpoint = entity.DataLogCheckpointDone
			}

			if err = pipe.Repository.UpdateDataLog(ctx, pipe.Workspace.ID, dataLog); err != nil {
				pipe.SetError("hook", fmt.Sprintf("error updating data log: %s", err), true)
				return
			}
		}
	}

	if !pipe.HasError() {
		pipe.DataLog.Checkpoint = entity.DataLogCheckpointHooksFinalizeExecuted
	}
}
