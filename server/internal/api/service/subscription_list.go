package service

import (
	"context"

	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (svc *ServiceImpl) SubscriptionListCreate(ctx context.Context, accountID string, data *dto.SubscriptionListCreate) (code int, err error) {

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, data.WorkspaceID, accountID)

	if err != nil {
		return code, eris.Wrap(err, "SubscriptionListCreate")
	}

	// create list from data
	list := &entity.SubscriptionList{
		ID:          data.ID,
		Name:        data.Name,
		Color:       data.Color,
		Channel:     data.Channel,
		DoubleOptIn: data.DoubleOptIn,
	}

	if data.DoubleOptIn && data.MessageTemplateID != nil {
		list.MessageTemplateID = data.MessageTemplateID
	}

	// validate list
	if err := list.Validate(); err != nil {
		return 400, err
	}

	// create list
	if err := svc.Repo.CreateSubscriptionList(ctx, workspace.ID, list); err != nil {
		if svc.Repo.IsDuplicateEntry(err) {
			return 400, eris.New("this list ID already exists")
		}
		return 500, err
	}

	return 200, nil
}

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
