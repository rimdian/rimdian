package service

import (
	"context"

	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (svc *ServiceImpl) SubscriptionListList(ctx context.Context, accountID string, params *dto.SubscriptionListListParams) (result []*entity.SubscriptionList, code int, err error) {

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, params.WorkspaceID, accountID)

	if err != nil {
		return nil, code, eris.Wrap(err, "SubscriptionListList")
	}

	// fetch lists
	result, err = svc.Repo.ListSubscriptionLists(ctx, workspace.ID, params.WithUsersCount)

	if err != nil {
		return nil, 500, err
	}

	return result, 200, nil
}
