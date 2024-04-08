package service

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (svc *ServiceImpl) BroadcastCampaignUpsert(ctx context.Context, accountID string, data *dto.BroadcastCampaign) (code int, err error) {

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, data.WorkspaceID, accountID)

	if err != nil {
		return code, eris.Wrap(err, "BroadcastCampaignUpsert")
	}

	// fetch existing campaign
	if data.ID != "" {
		campaign, err := svc.Repo.GetBroadcastCampaign(ctx, workspace.ID, data.ID)

		if err != nil {
			if err == sql.ErrNoRows {
				return 400, eris.New("BroadcastCampaign not found")
			}
			return 500, eris.Wrap(err, "BroadcastCampaignUpsert")
		}

		// update existing campaign if is not launched yet
		if campaign.Status != entity.BroadcastCampaignStatusDraft {
			return 400, eris.New("Cannot update launched campaign")
		}

		campaign.Name = data.Name
		campaign.Channel = data.Channel
		campaign.Templates = data.Templates
		campaign.Status = data.Status
		campaign.SubscriptionLists = data.SubscriptionLists
		campaign.UTMSource = data.UTMSource
		campaign.UTMMedium = data.UTMMedium
		campaign.ScheduledAt = data.ScheduledAt

		if err = campaign.Validate(); err != nil {
			return 400, eris.Wrap(err, "BroadcastCampaignUpsert")
		}

		if err = svc.Repo.UpdateBroadcastCampaign(ctx, workspace.ID, campaign); err != nil {
			return 500, eris.Wrap(err, "BroadcastCampaignUpsert")
		}

		return 200, nil
	}

	// create new campaign
	campaign := &entity.BroadcastCampaign{
		ID:                data.ID,
		Name:              data.Name,
		Channel:           data.Channel,
		Templates:         data.Templates,
		Status:            data.Status,
		SubscriptionLists: data.SubscriptionLists,
		UTMSource:         data.UTMSource,
		UTMMedium:         data.UTMMedium,
		ScheduledAt:       data.ScheduledAt,
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
