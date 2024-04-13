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

func (svc *ServiceImpl) UserShow(ctx context.Context, accountID string, params *dto.UserShowParams) (user *entity.User, code int, err error) {

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, params.WorkspaceID, accountID)

	if err != nil {
		return nil, code, eris.Wrap(err, "UserShow")
	}

	// compute user id from external id
	userID := entity.ComputeUserID(params.UserExternalID)

	// fetch user
	user, err = svc.Repo.FindUserByID(ctx, workspace, userID, nil, params.UserWith)

	if err != nil {
		return nil, 500, eris.Wrap(err, "UserShow")
	}

	if user == nil {
		return nil, 400, eris.New("user not found")
	}

	return user, 200, nil
}
