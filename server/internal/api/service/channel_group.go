package service

import (
	"context"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (svc *ServiceImpl) ChannelGroupUpsert(ctx context.Context, accountID string, channelGroupDTO *dto.ChannelGroup) (updatedWorkspace *entity.Workspace, code int, err error) {

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, channelGroupDTO.WorkspaceID, accountID)

	if err != nil {
		return nil, code, eris.Wrap(err, "ChannelGroupUpsert")
	}

	now := time.Now().UTC()

	// convert DTO to entity
	group := &entity.ChannelGroup{
		ID:        channelGroupDTO.ID,
		Name:      channelGroupDTO.Name,
		Color:     channelGroupDTO.Color,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := group.Validate(workspace.ChannelGroups); err != nil {
		return nil, 400, eris.Wrap(err, "ChannelGroupUpsert")
	}

	// upsert channel group in workspace

	isInsert := true

	for _, gr := range workspace.ChannelGroups {
		if gr.ID == group.ID {
			// is update
			gr.Name = group.Name
			gr.Color = group.Color
			gr.UpdatedAt = group.UpdatedAt
			isInsert = false
		}
	}

	if isInsert {
		workspace.ChannelGroups = append(workspace.ChannelGroups, group)
	}

	if err := workspace.Validate(); err != nil {
		return nil, 400, eris.Wrap(err, "ChannelGroupUpsert")
	}

	if err := svc.Repo.UpdateWorkspace(ctx, workspace, nil); err != nil {
		return nil, 500, eris.Wrap(err, "ChannelGroupUpsert")
	}

	return workspace, 200, nil
}

func (svc *ServiceImpl) ChannelGroupDelete(ctx context.Context, accountID string, deleteChannelGroupDTO *dto.DeleteChannelGroup) (updatedWorkspace *entity.Workspace, code int, err error) {

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, deleteChannelGroupDTO.WorkspaceID, accountID)

	if err != nil {
		return nil, code, eris.Wrap(err, "ChannelGroupDelete")
	}

	// exclude reserved IDs (not-mapped...)
	if govalidator.IsIn(deleteChannelGroupDTO.ID, entity.ReservedChannelGroupIDs...) {
		return nil, 400, entity.ErrChannelGroupIDReserved
	}

	var deletedChannelGroup *entity.ChannelGroup
	updatedChannelGroups := []*entity.ChannelGroup{}

	for _, x := range workspace.ChannelGroups {
		if x.ID == deleteChannelGroupDTO.ID {
			deletedChannelGroup = x
		} else {
			updatedChannelGroups = append(updatedChannelGroups, x)
		}
	}

	if deletedChannelGroup == nil {
		return nil, 400, entity.ErrChannelIDInvalid
	}

	// check that this group has no more channels attached to it
	hasChannels := false
	for _, ch := range workspace.Channels {
		if ch.GroupID == deletedChannelGroup.ID {
			hasChannels = true
		}
	}

	if hasChannels {
		return nil, 400, entity.ErrChannelGroupStillHasChannels
	}

	workspace.ChannelGroups = updatedChannelGroups

	if err := workspace.Validate(); err != nil {
		return nil, 400, eris.Wrap(err, "ChannelGroupDelete")
	}

	if err := svc.Repo.UpdateWorkspace(ctx, workspace, nil); err != nil {
		return nil, 500, eris.Wrap(err, "ChannelGroupDelete")
	}

	return workspace, 200, nil

}
