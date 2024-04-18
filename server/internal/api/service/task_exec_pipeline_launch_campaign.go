package service

import (
	"context"
	"fmt"
	"time"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/entity"
	"go.opencensus.io/trace"
)

func TaskExecLaunchBroadcastCampaign(ctx context.Context, pipe *TaskExecPipeline) (result *entity.TaskExecResult) {

	spanCtx, span := trace.StartSpan(ctx, "TaskExecLaunchBroadcastCampaign")
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
		pipe.Logger.Printf("TaskExecLaunchBroadcastCampaign: workspace %s, taskExec %s, worker %d, took %s", pipe.Workspace.ID, pipe.TaskExec.ID, pipe.TaskExecPayload.WorkerID, time.Since(startedAt))
	}()

	// by default, keep current state
	mainState := pipe.TaskExec.State.Workers[0]

	bgCtx := context.Background()

	campaignID := ""

	if _, ok := mainState["broadcast_campaign_id"]; ok {
		campaignID = mainState["broadcast_campaign_id"].(string)
	}

	if campaignID == "" {
		result.SetError("broadcast_campaign_id is required", true)
		return
	}

	campaign, err := pipe.Repository.GetBroadcastCampaign(bgCtx, pipe.Workspace.ID, campaignID)

	if err != nil {
		if sqlscan.NotFound(err) {
			result.SetError("campaign not found", true)
			return
		}
		result.SetError(err.Error(), false)
		return
	}

	// get offset from state or start from 0
	var offset int64
	offset = 0

	if _, ok := mainState["offset"]; ok {
		offset = int64(mainState["offset"].(float64))
	}

	limit := int64(100)

	subscriptionListIDs := []string{}

	for _, subscriptionList := range campaign.SubscriptionLists {
		subscriptionListIDs = append(subscriptionListIDs, subscriptionList.ID)
	}

	// Fetch subscribers according to the offset
	subscribers, err := pipe.Repository.GetSubscriptionListUsersToMessage(bgCtx, pipe.Workspace.ID, subscriptionListIDs, offset, limit)

	if err != nil {
		result.SetError(err.Error(), false)
		return
	}

	// if no subscribers found, mark task as completed
	if len(subscribers) == 0 {
		result.IsDone = true
		return
	}

	// generate a message for each subscriber
	now := time.Now().Format(time.RFC3339)
	items := []string{}

	for _, subscriber := range subscribers {

		// default template
		messageTemplate := campaign.MessageTemplates[0]

		// AB testing with percentage
		totalTemplates := len(campaign.MessageTemplates)
		if totalTemplates > 1 {

			// random int between 1 and 100
			randInt := RandomInt(SeededRand, 1, 100)

			// matches a percentage
			accumulator := 0

			for i, template := range campaign.MessageTemplates {
				min := 1
				max := 1

				if i == 0 {
					max = template.Percentage
				} else {
					min = accumulator + 1
					max = accumulator + template.Percentage
				}

				if randInt >= min && randInt <= max {
					messageTemplate = template
					break
				}

				accumulator += template.Percentage
			}
		}

		// generate message external_id
		messageExternalID, err := entity.GenerateShortID()
		if err != nil {
			result.SetError(err.Error(), false)
			return
		}

		items = append(items, fmt.Sprintf(`{
			"kind": "message",
			"message": {
				"external_id": "%s",
				"created_at": "%s",
				"channel": "%s",
				"message_template_id": "%s",
				"subscription_list_id": "%s"
			},
			"user": {
				"external_id": "%s",
				"is_authenticated": %t,
				"created_at": "%s"
			}
		}`,
			messageExternalID,
			now,
			campaign.Channel,
			messageTemplate.ID,
			subscriber.SubscriptionListID,
			subscriber.UserExternalID,
			subscriber.UserIsAuthenticated,
			now, // not important, users already exist
		))
	}

	// update state
	offset += limit
	mainState["offset"] = offset
	result.UpdatedWorkerState = mainState

	result.Message = entity.StringPtr(fmt.Sprintf("campaign: %v, offset: %d", campaign.ID, offset))
	result.ItemsToImport = items

	return result
}
