package service

import (
	"context"

	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (svc *ServiceImpl) UserList(ctx context.Context, accountID string, params *dto.UserListParams) (result *dto.UserListResult, code int, err error) {

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, params.WorkspaceID, accountID)

	if err != nil {
		return nil, code, eris.Wrap(err, "UserList")
	}
	// fetch users
	result = &dto.UserListResult{}

	result.Users, result.NextToken, result.PreviousToken, err = svc.Repo.ListUsers(ctx, workspace, params)

	if err != nil {
		code = 500
		return
	}

	return result, 200, nil
}

func (svc *ServiceImpl) UserShow(ctx context.Context, workspaceID string, accountID string, userExternalID string) (result *dto.UserShowResult, code int, err error) {
	// init
	result = &dto.UserShowResult{}

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, workspaceID, accountID)

	if err != nil {
		return nil, code, eris.Wrap(err, "UserShow")
	}

	// compute user id from external id
	userID := entity.ComputeUserID(userExternalID)

	// fetch user
	user, err := svc.Repo.FindUserByID(ctx, workspace, userID, nil)

	if err != nil {
		return nil, 500, eris.Wrap(err, "UserShow")
	}

	if user == nil {
		return nil, 400, eris.New("user not found")
	}

	// populate result
	result.User = user

	// fetch user segments
	result.UserSegments, err = svc.Repo.ListUserSegments(ctx, workspace.ID, []string{user.ID}, nil)

	if err != nil {
		return nil, 500, eris.Wrap(err, "UserShow")
	}

	// fetch user devices
	result.Devices, err = svc.Repo.ListDevicesForUser(ctx, workspace, user.ID, "created_at ASC", nil)

	if err != nil {
		return nil, 500, eris.Wrap(err, "UserShow")
	}

	// svc.Logger.Printf("result.Devices = %+v", result.Devices)

	// fetch user aliases
	result.Aliases, err = svc.Repo.FindUsersAliased(ctx, workspace.ID, user.ExternalID)

	if err != nil {
		return nil, 500, eris.Wrap(err, "UserShow")
	}

	return result, 200, nil
}
