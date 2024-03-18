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
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/rimdian/rimdian/internal/api/entity"
	commonDTO "github.com/rimdian/rimdian/internal/common/dto"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"go.opencensus.io/trace"
)

// execute the on_validation hook
func (pipe *DataLogPipeline) StepPending(ctx context.Context) {

	spanCtx, span := trace.StartSpan(ctx, "StepPending")
	defer span.End()

	hooksToRun := []*entity.DataHook{}

	// init hooks for this data log
	for _, hook := range pipe.Workspace.DataHooks {

		// only on_validation hooks here, action will be empty for on_validation
		if hook.On == entity.DataHookKindOnValidation && hook.Enabled && pipe.DataLog.HasError == 0 && hook.MatchesDataLogKind(pipe.DataLog.Kind, pipe.DataLog.Action) {
			hooksToRun = append(hooksToRun, hook)
		}
	}

	// exec hooks in series
	for _, hook := range hooksToRun {

		// set default hook state if not existsxw
		if _, ok := pipe.DataLog.Hooks[hook.ID]; !ok {
			pipe.DataLog.Hooks[hook.ID] = &entity.DataHookState{
				Done: false,
			}
		}

		// skip if already done
		if pipe.DataLog.Hooks[hook.ID].Done {
			continue
		}

		// get app
		var appFound *entity.App
		for _, app := range pipe.Apps {

			if app.ID == hook.AppID {
				appFound = app
			}
		}

		if appFound == nil {
			pipe.Logger.Printf("Error getting app for hook: %v", hook.AppID)
			continue
		}

		payload := &entity.AppWebhookPayload{
			APIEndpoint:       pipe.Config.API_ENDPOINT,
			CollectorEndpoint: pipe.Config.COLLECTOR_ENDPOINT,
			WorkspaceID:       pipe.Workspace.ID,
			AppID:             hook.AppID,
			AppState:          appFound.State,
			Kind:              entity.AppWebhookKindDataHook,
			DataHook: &entity.DataHookPayload{
				DataHookID:           hook.ID,
				DataHookName:         hook.Name,
				DataHookOn:           hook.On,
				DataLogID:            pipe.DataLog.ID,
				DataLogItem:          pipe.DataLog.Item,
				DataLogKind:          pipe.DataLog.Kind,
				DataLogUpdatedFields: pipe.DataLog.UpdatedFields,
				User:                 pipe.DataLog.UpsertedUser,
			},
		}

		// restrict the roundtripper to 5 seconds
		ctxWithTimeout, cancel := context.WithTimeout(spanCtx, 5*time.Second)
		defer cancel()

		data, err := json.Marshal(payload)

		if err != nil {
			pipe.Logger.Printf("Error marshalling payload: %v", err.Error())
			pipe.SetError("server", "Error marshalling payload", true)
			return
		}

		// compute HMAC signature of the payload
		appKey, err := appFound.GetAppSecretKey(pipe.Config.SECRET_KEY)

		if err != nil {
			pipe.Logger.Printf("Error getting app secret key: %v", err.Error())
			pipe.SetError("server", "Error getting app secret key", true)
			return
		}

		h := hmac.New(sha256.New, []byte(appKey))
		h.Write([]byte(data))
		hmac256 := hex.EncodeToString(h.Sum(nil))

		// send request to external webhook
		req, _ := http.NewRequestWithContext(ctxWithTimeout, "POST", appFound.Manifest.WebhookEndpoint, bytes.NewBuffer(data))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set(commonDTO.HeaderSignature, hmac256)

		res, err := pipe.NetClient.Do(req)

		if err != nil {
			pipe.Logger.Printf("Error sending request to external webhook: %v", err.Error())
			pipe.SetError("server", "Error sending request to external webhook", true)
			return
		}

		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)

		if err != nil {
			pipe.DataLog.Hooks[hook.ID].Done = true
			pipe.DataLog.Hooks[hook.ID].IsError = true

			// check if its timeout
			if ctxWithTimeout.Err() == context.DeadlineExceeded {
				pipe.Logger.Printf("Error on_validation timeout after 5 secs for app %v", appFound.ID)
				pipe.DataLog.Hooks[hook.ID].Message = "Timeout after 5 secs"
				continue
			}

			pipe.Logger.Printf("Error reading response body: %v", err.Error())
			pipe.DataLog.Hooks[hook.ID].Message = "Error reading webhook response"
			continue
		}

		if res.StatusCode >= 200 {
			pipe.DataLog.Hooks[hook.ID].Done = true
			pipe.DataLog.Hooks[hook.ID].IsError = true
			pipe.DataLog.Hooks[hook.ID].Message = fmt.Sprintf("Reponse status != 200, got: %v", res.Status)
			continue
		}

		// returned a message to attach to the hook result
		var msg string
		messageResult := gjson.Get(string(body), "message")
		if messageResult.Exists() {
			msg = messageResult.String()
		}

		result := gjson.Get(string(body), "action")

		// item rejected
		if result.String() == entity.RejectItem {
			pipe.DataLog.Hooks[hook.ID].Done = true
			pipe.DataLog.Hooks[hook.ID].Message = "item rejected"
			if msg == "" {
				msg = fmt.Sprintf("item rejected by hook %v", hook.ID)
			}
			pipe.SetError("on_validation", msg, false)
			// end here
			return
		}

		// item updated
		if result.String() == entity.UpdateItem {
			pipe.DataLog.Hooks[hook.ID].Done = true
			if msg == "" {
				msg = "item updated"
			}
			pipe.DataLog.Hooks[hook.ID].Message = msg
			newItemResult := gjson.Get(string(body), "updated_item")

			if newItemResult.Exists() {
				// check if valid json
				if govalidator.IsJSON(newItemResult.String()) {
					pipe.DataLog.Item = newItemResult.String()
				} else {
					pipe.Logger.Printf("Error updating item: %v", newItemResult.String())
					pipe.DataLog.Hooks[hook.ID].IsError = true
					pipe.DataLog.Hooks[hook.ID].Message = fmt.Sprintf("invalid updated_item json: %v", string(body))
					continue
				}
			} else {
				pipe.Logger.Printf("Error updating item: %v", newItemResult.String())
				pipe.DataLog.Hooks[hook.ID].IsError = true
				pipe.DataLog.Hooks[hook.ID].Message = fmt.Sprintf("missing updated_item in response: %v", string(body))
				continue
			}
		}
	}

	var err error

	// check if has item.user
	hasUser := gjson.Get(pipe.DataLog.Item, "user")
	if hasUser.Exists() {

		// Extract Google Geo headers
		if _, ok := pipe.DataLog.Context.HeadersAndParams["X-Client-Geo-Country"]; ok {
			country := strings.ToUpper(pipe.DataLog.Context.HeadersAndParams["X-Client-Geo-Country"])
			if country != "" && govalidator.IsISO3166Alpha2(country) {
				// set user country in item json
				pipe.DataLog.Item, err = sjson.Set(pipe.DataLog.Item, "user.country", country)
				if err != nil {
					// log error and ignore
					pipe.Logger.Printf("Error setting user.country: %v", err.Error())
					return
				}
			}
		}

		if _, ok := pipe.DataLog.Context.HeadersAndParams["X-Client-Geo-Latlon"]; ok {
			latLon := pipe.DataLog.Context.HeadersAndParams["X-Client-Geo-Latlon"]

			if strings.Contains(latLon, ",") {

				// split LatLon
				parts := strings.Split(latLon, ",")

				latitude, latErr := strconv.ParseFloat(parts[0], 64)
				longitude, lonErr := strconv.ParseFloat(parts[1], 64)

				if latErr != nil {
					pipe.Logger.Printf("parse X-Client-Geo-Latlon latitude: %v", latErr.Error())
				} else if lonErr != nil {
					pipe.Logger.Printf("parse X-Client-Geo-Latlon longitude: %v", lonErr.Error())
				}

				// set user latitude in item json
				pipe.DataLog.Item, err = sjson.Set(pipe.DataLog.Item, "user.latitude", latitude)
				if err != nil {
					// log error and ignore
					pipe.Logger.Printf("Error setting user.latitude: %v", err.Error())
					return
				}
				pipe.DataLog.Item, err = sjson.Set(pipe.DataLog.Item, "user.longitude", longitude)
				if err != nil {
					// log error and ignore
					pipe.Logger.Printf("Error setting user.longitude: %v", err.Error())
					return
				}
			}
		}
	}

	pipe.DataLog.Checkpoint = entity.DataLogCheckpointHookOnValidationExecuted
}
