package service

import (
	"context"
	"net/http"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

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

		if err = svc.Repo.UpdateBroadcastCampaign(ctx, workspace.ID, campaign); err != nil {
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
