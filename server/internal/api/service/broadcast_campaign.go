package service

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (svc *ServiceImpl) BroadcastCampaignLaunch(ctx context.Context, accountID string, data *dto.BroadcastCampaignLaunchParams) (code int, err error) {

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, data.WorkspaceID, accountID)

	if err != nil {
		return code, eris.Wrap(err, "BroadcastCampaignLaunch")
	}

	campaign, err := svc.Repo.GetBroadcastCampaign(ctx, workspace.ID, data.ID)

	if err != nil {
		return 500, eris.Wrap(err, "BroadcastCampaignLaunch")
	}

	if campaign == nil {
		return 404, eris.New("Campaign not found")
	}

	if campaign.Status != entity.BroadcastCampaignStatusDraft && campaign.Status != entity.BroadcastCampaignStatusScheduled {
		return 400, eris.New("Campaign already launched")
	}

	// campaign.ScheduledAt = data.ScheduledAt
	// campaign.Timezone = data.Timezone
	// if err = campaign.Validate(); err != nil {
	// 	return 400, eris.Wrap(err, "BroadcastCampaignLaunch")
	// }

	// workspace tx
	code, err = svc.Repo.RunInTransactionForWorkspace(ctx, workspace.ID, func(ctx context.Context, tx *sql.Tx) (int, error) {

		campaign.LaunchNow()

		if err = svc.Repo.UpdateBroadcastCampaign(ctx, workspace.ID, campaign, tx); err != nil {
			return 500, eris.Wrap(err, "BroadcastCampaignLaunch")
		}

		// create a task to send the campaign
		state := entity.NewTaskState()
		state.Workers[0] = entity.TaskWorkerState{
			"broadcast_campaign_id": campaign.ID,
		}

		taskExec := &entity.TaskExec{
			TaskID:          entity.TaskKindLaunchBroadcastCampaign,
			Name:            fmt.Sprintf("Launch %v campaign %v", campaign.Channel, campaign.Name),
			MultipleExecKey: entity.StringPtr(campaign.ID),   // deduplicate tasks by campaign ID
			OnMultipleExec:  entity.OnMultipleExecDiscardNew, // aborting new task if campaign is already launched
			State:           state,
		}

		return svc.doTaskCreate(ctx, workspace.ID, taskExec)
	})

	return code, err
}

func (svc *ServiceImpl) BroadcastCampaignUpsert(ctx context.Context, accountID string, data *dto.BroadcastCampaign) (code int, err error) {

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, data.WorkspaceID, accountID)

	if err != nil {
		return code, eris.Wrap(err, "BroadcastCampaignUpsert")
	}

	// fetch eventual existing campaign
	campaign, err := svc.Repo.GetBroadcastCampaign(ctx, workspace.ID, data.ID)

	if err != nil && !sqlscan.NotFound(err) {
		return 500, eris.Wrap(err, "BroadcastCampaignUpsert")
	}

	if campaign != nil {
		// update existing campaign if is not launched yet
		if campaign.Status == entity.BroadcastCampaignStatusLaunched || campaign.Status == entity.BroadcastCampaignStatusSent || campaign.Status == entity.BroadcastCampaignStatusFailed {
			return 400, eris.New("Cannot update launched campaign")
		}

		campaign.Name = data.Name
		campaign.MessageTemplates = data.MessageTemplates
		campaign.SubscriptionLists = data.SubscriptionLists
		campaign.UTMSource = data.UTMSource
		campaign.UTMMedium = data.UTMMedium
		campaign.ScheduledAt = data.ScheduledAt
		campaign.Timezone = data.Timezone

		if err = campaign.Validate(); err != nil {
			return 400, eris.Wrap(err, "BroadcastCampaignUpsert")
		}

		// cancel eventual scheduled_tasks from the queue
		taskID := fmt.Sprintf("%v-%v", entity.TaskKindLaunchBroadcastCampaign, campaign.ID)

		if err = svc.TaskOrchestrator.DeleteTask(ctx, svc.Config.TASK_QUEUE_LOCATION, entity.ScheduledTasksQueueName, taskID); err != nil {
			return 500, eris.Wrap(err, "BroadcastCampaignUpsert")
		}

		// post a scheduled_tasks if scheduled_at is set
		if campaign.Status == entity.BroadcastCampaignStatusScheduled {
			state := entity.NewTaskState()
			state.Workers[0] = entity.TaskWorkerState{
				"broadcast_campaign_id": campaign.ID,
			}

			scheduledAt, err := campaign.GetScheduledAt()

			if err != nil {
				return 500, eris.Wrap(err, "BroadcastCampaignUpsert")
			}

			scheduledTask := entity.NewScheduledTask(workspace.ID, entity.TaskExec{
				ID:              taskID,
				TaskID:          entity.TaskKindLaunchBroadcastCampaign,
				Name:            fmt.Sprintf("Launch %v campaign %v", campaign.Channel, campaign.Name),
				State:           state,
				MultipleExecKey: entity.StringPtr(campaign.ID),   // deduplicate tasks by campaign ID
				OnMultipleExec:  entity.OnMultipleExecDiscardNew, // aborting new task if campaign is already launched
			}, *scheduledAt)

			// enqueue job
			if err := svc.ScheduledTaskPost(ctx, scheduledTask); err != nil {
				return 500, eris.Wrap(err, "BroadcastCampaignUpsert")
			}
		}

		if err = svc.Repo.UpdateBroadcastCampaign(ctx, workspace.ID, campaign, nil); err != nil {
			return 500, eris.Wrap(err, "BroadcastCampaignUpsert")
		}

		return 200, nil
	}

	// create new campaign
	campaign = &entity.BroadcastCampaign{
		ID:                data.ID,
		Name:              data.Name,
		Channel:           data.Channel,
		MessageTemplates:  data.MessageTemplates,
		Status:            data.Status,
		SubscriptionLists: data.SubscriptionLists,
		UTMSource:         data.UTMSource,
		UTMMedium:         data.UTMMedium,
		ScheduledAt:       data.ScheduledAt,
		Timezone:          data.Timezone,
	}

	if err = campaign.Validate(); err != nil {
		return 400, eris.Wrap(err, "BroadcastCampaignUpsert")
	}

	if err = svc.Repo.InsertBroadcastCampaign(ctx, workspace.ID, campaign); err != nil {
		return 500, eris.Wrap(err, "BroadcastCampaignUpsert")
	}

	return 200, nil
}

func (svc *ServiceImpl) BroadcastCampaignList(ctx context.Context, accountID string, params *dto.BroadcastCampaignListParams) (broadcasts []*entity.BroadcastCampaign, code int, err error) {

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, params.WorkspaceID, accountID)

	if err != nil {
		return nil, code, eris.Wrap(err, "BroadcastCampaignList")
	}

	broadcasts, err = svc.Repo.ListBroadcastCampaigns(ctx, workspace.ID, params)

	if err != nil {
		return nil, http.StatusInternalServerError, eris.Wrap(err, "BroadcastCampaignList")
	}

	return broadcasts, http.StatusOK, nil
}
