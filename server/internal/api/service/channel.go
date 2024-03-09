package service

import (
	"context"
	"time"

	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (svc *ServiceImpl) ChannelCreate(ctx context.Context, accountID string, channelDTO *dto.Channel) (updatedWorkspace *entity.Workspace, code int, err error) {

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, channelDTO.WorkspaceID, accountID)

	if err != nil {
		return nil, code, eris.Wrap(err, "ChannelCreate")
	}

	now := time.Now().UTC()

	// convert DTO to entity
	channel := &entity.Channel{
		ID:        channelDTO.ID,
		Name:      channelDTO.Name,
		GroupID:   channelDTO.GroupID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	channel.EnsureID(workspace.Channels)

	channel.Origins = []*entity.ChannelOrigin{}
	channel.VoucherCodes = []*entity.VoucherCode{}

	if channelDTO.Origins != nil {

		for _, x := range channelDTO.Origins {
			channel.Origins = append(channel.Origins, &entity.ChannelOrigin{
				ID:            x.ID,
				MatchOperator: x.MatchOperator,
				UTMSource:     x.UTMSource,
				UTMMedium:     x.UTMMedium,
				UTMCampaign:   x.UTMCampaign,
			})
		}
	}

	if channelDTO.VoucherCodes != nil {

		for _, x := range channelDTO.VoucherCodes {
			channel.VoucherCodes = append(channel.VoucherCodes, &entity.VoucherCode{
				Code:           x.Code,
				OriginID:       x.OriginID,
				SetUTMCampaign: x.SetUTMCampaign,
				SetUTMContent:  x.SetUTMContent,
				Description:    x.Description,
			})
		}
	}

	if err := channel.Validate(workspace.Channels, workspace.ChannelGroups); err != nil {
		return nil, 400, eris.Wrap(err, "ChannelCreate")
	}

	workspace.Channels = append(workspace.Channels, channel)
	workspace.OutdatedConversionsAttribution = true

	if err := workspace.Validate(); err != nil {
		return nil, 400, eris.Wrap(err, "ChannelCreate")
	}

	if err := svc.Repo.CreateChannel(ctx, workspace, channel); err != nil {
		return nil, 500, eris.Wrap(err, "ChannelCreate")
	}

	return workspace, 200, nil

}

func (svc *ServiceImpl) ChannelUpdate(ctx context.Context, accountID string, channelDTO *dto.Channel) (updatedWorkspace *entity.Workspace, code int, err error) {

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, channelDTO.WorkspaceID, accountID)

	if err != nil {
		return nil, code, eris.Wrap(err, "ChannelUpdate")
	}

	now := time.Now().UTC()

	// find and update channel
	var updatedChannel *entity.Channel

	for _, ch := range workspace.Channels {
		if ch.ID == channelDTO.ID {
			ch.Name = channelDTO.Name
			ch.GroupID = channelDTO.GroupID
			ch.UpdatedAt = now
			ch.Origins = []*entity.ChannelOrigin{}
			ch.VoucherCodes = []*entity.VoucherCode{}

			// replace paths
			if channelDTO.Origins != nil {

				for _, x := range channelDTO.Origins {
					ch.Origins = append(ch.Origins, &entity.ChannelOrigin{
						ID:            x.ID,
						MatchOperator: x.MatchOperator,
						UTMSource:     x.UTMSource,
						UTMMedium:     x.UTMMedium,
						UTMCampaign:   x.UTMCampaign,
					})
				}
			}

			// replace vouchers
			if channelDTO.VoucherCodes != nil {

				for _, x := range channelDTO.VoucherCodes {
					ch.VoucherCodes = append(ch.VoucherCodes, &entity.VoucherCode{
						Code:           x.Code,
						OriginID:       x.OriginID,
						SetUTMCampaign: x.SetUTMCampaign,
						SetUTMContent:  x.SetUTMContent,
						Description:    x.Description,
					})
				}
			}

			updatedChannel = ch
		}
	}

	if updatedChannel == nil {
		return nil, 400, eris.Wrap(entity.ErrChannelIDInvalid, "ChannelUpdate")
	}

	if err := updatedChannel.Validate(workspace.Channels, workspace.ChannelGroups); err != nil {
		return nil, 400, eris.Wrap(err, "ChannelUpdate")
	}

	workspace.OutdatedConversionsAttribution = true

	if err := workspace.Validate(); err != nil {
		return nil, 400, eris.Wrap(err, "ChannelUpdate")
	}

	if err := svc.Repo.UpdateChannel(ctx, workspace, updatedChannel); err != nil {
		return nil, 500, eris.Wrap(err, "ChannelUpdate")
	}

	return workspace, 200, nil

}

func (svc *ServiceImpl) ChannelDelete(ctx context.Context, accountID string, deleteChannelDTO *dto.DeleteChannel) (updatedWorkspace *entity.Workspace, code int, err error) {

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, deleteChannelDTO.WorkspaceID, accountID)

	if err != nil {
		return nil, code, eris.Wrap(err, "ChannelDelete")
	}

	var deletedChannel *entity.Channel
	updatedChannels := []*entity.Channel{}

	for _, x := range workspace.Channels {
		if x.ID == deleteChannelDTO.ID {
			deletedChannel = x
		} else {
			updatedChannels = append(updatedChannels, x)
		}
	}

	if deletedChannel == nil {
		return nil, 400, entity.ErrChannelIDInvalid
	}

	workspace.Channels = updatedChannels
	workspace.OutdatedConversionsAttribution = true

	if err := workspace.Validate(); err != nil {
		return nil, 400, eris.Wrap(err, "ChannelDelete")
	}

	if err := svc.Repo.DeleteChannel(ctx, workspace, deletedChannel.ID); err != nil {
		return nil, 500, eris.Wrap(err, "ChannelDelete")
	}

	return workspace, 200, nil

}
