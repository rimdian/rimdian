package service

import (
	"context"
	"fmt"
	"time"

	"github.com/rimdian/rimdian/internal/api/entity"
	"go.opencensus.io/trace"
)

func TaskExecImportUsersToSubscriptionList(ctx context.Context, pipe *TaskExecPipeline) (result *entity.TaskExecResult) {

	spanCtx, span := trace.StartSpan(ctx, "TaskExecImportUsersToSubscriptionList")
	defer span.End()

	result = &entity.TaskExecResult{
		// keep current state by default
		UpdatedWorkerState: pipe.TaskExec.State.Workers[pipe.TaskExecPayload.WorkerID],
	}

	select {
	case <-spanCtx.Done():
		result.SetError("task execution timeout", false)
		return
	default:
	}

	// log time taken
	startedAt := time.Now()
	defer func() {
		pipe.Logger.Printf("TaskExecImportUsersToSubscriptionList: workspace %s, taskExec %s, worker %d, took %s", pipe.Workspace.ID, pipe.TaskExec.ID, pipe.TaskExecPayload.WorkerID, time.Since(startedAt))
	}()

	// by default, keep current state
	mainState := pipe.TaskExec.State.Workers[0]

	bgCtx := context.Background()

	source := ""

	if _, ok := mainState["source"]; ok {
		source = mainState["source"].(string)
	}

	// only segment source is supported for now
	if source != "segment" {
		result.SetError("source not supported", true)
		return
	}

	segmentID := ""

	if _, ok := mainState["segment_id"]; ok {
		segmentID = mainState["segment_id"].(string)
	}

	if segmentID == "" {
		result.SetError("segment_id is required", true)
		return
	}

	subscriptionListID := ""

	if _, ok := mainState["subscription_list_id"]; ok {
		subscriptionListID = mainState["subscription_list_id"].(string)
	}

	if subscriptionListID == "" {
		result.SetError("subscription_list_id is required", true)
		return
	}

	// get offset from state or start from 0
	var offset int64
	offset = 0

	if _, ok := mainState["offset"]; ok {
		offset = int64(mainState["offset"].(float64))
	}

	limit := int64(100)

	// fetch users who are not in the subscription list and belong to the segment
	users, err := pipe.Repository.GetUsersNotInSubscriptionList(bgCtx, pipe.Workspace.ID, segmentID, offset, limit, &segmentID)

	if err != nil {
		result.SetError(err.Error(), false)
		return
	}

	// if no users found, mark task as completed
	if len(users) == 0 {
		result.IsDone = true
		return
	}

	now := time.Now().Format(time.RFC3339)
	items := []string{}

	for _, user := range users {
		items = append(items, fmt.Sprintf(`{
			"kind": "subscription_list_user",
			"subscription_list_user": {
				"subscription_list_id": "%s",
				"user_id": "%s",
				"status": 1,
				"created_at": "%s"
			},
			"user": {
				"external_id": "%s",
				"is_authenticated": %t,
				"created_at": "%s"
			}
		}`,
			subscriptionListID,
			user.ExternalID,
			now,
			user.ExternalID,
			user.IsAuthenticated,
			user.CreatedAt.Format(time.RFC3339),
		))
	}

	// update state
	offset += limit
	mainState["offset"] = offset
	result.UpdatedWorkerState = mainState

	result.Message = entity.StringPtr(fmt.Sprintf("subscription list: %v, offset: %d", subscriptionListID, offset))
	result.ItemsToImport = items

	return result
}
