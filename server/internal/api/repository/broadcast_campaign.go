package repository

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
)

func (repo *RepositoryImpl) UpdateBroadcastCampaign(ctx context.Context, workspaceID string, campaign *entity.BroadcastCampaign) (err error) {

	conn, errConn := repo.GetWorkspaceConnection(ctx, workspaceID)

	if errConn != nil {
		return
	}

	defer conn.Close()

	sql, args, err := squirrel.Update("broadcast_campaign").
		Set("name", campaign.Name).
		Set("channel", campaign.Channel).
		Set("message_templates", campaign.MessageTemplates).
		Set("status", campaign.Status).
		Set("subscription_lists", campaign.SubscriptionLists).
		Set("utm_source", campaign.UTMSource).
		Set("utm_medium", campaign.UTMMedium).
		Set("scheduled_at", campaign.ScheduledAt).
		Where("id = ?", campaign.ID).
		ToSql()

	if err != nil {
		return err
	}

	_, err = conn.ExecContext(ctx, sql, args...)

	return

}

func (repo *RepositoryImpl) InsertBroadcastCampaign(ctx context.Context, workspaceID string, campaign *entity.BroadcastCampaign) (err error) {
	conn, errConn := repo.GetWorkspaceConnection(ctx, workspaceID)

	if errConn != nil {
		return
	}

	defer conn.Close()

	sql, args, err := squirrel.Insert("broadcast_campaign").
		Columns(
			"id",
			"name",
			"channel",
			"message_templates",
			"status",
			"subscription_lists",
			"utm_source",
			"utm_medium",
			"scheduled_at",
		).
		Values(campaign.ID,
			campaign.Name,
			campaign.Channel,
			campaign.MessageTemplates,
			campaign.Status,
			campaign.SubscriptionLists,
			campaign.UTMSource,
			campaign.UTMMedium,
			campaign.ScheduledAt,
		).
		ToSql()

	if err != nil {
		return err
	}

	_, err = conn.ExecContext(ctx, sql, args...)

	return
}

func (repo *RepositoryImpl) GetBroadcastCampaign(ctx context.Context, workspaceID string, campaignID string) (campaign *entity.BroadcastCampaign, err error) {
	conn, errConn := repo.GetWorkspaceConnection(ctx, workspaceID)

	if errConn != nil {
		return
	}

	defer conn.Close()

	err = sqlscan.Get(ctx, conn, &campaign, "SELECT * FROM broadcast_campaign WHERE id = ?", campaignID)

	return
}

func (repo *RepositoryImpl) ListBroadcastCampaigns(ctx context.Context, workspaceID string, params *dto.BroadcastCampaignListParams) (campaigns []*entity.BroadcastCampaign, err error) {
	conn, errConn := repo.GetWorkspaceConnection(ctx, workspaceID)

	if errConn != nil {
		return
	}

	defer conn.Close()

	err = sqlscan.Select(ctx, conn, &campaigns, "SELECT * FROM broadcast_campaign")

	if err != nil {
		return
	}

	if campaigns == nil {
		campaigns = []*entity.BroadcastCampaign{}
	}

	return
}
