package service

import (
	"context"

	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (svc *ServiceImpl) MessageTemplateList(ctx context.Context, accountID string, params *dto.MessageTemplateListParams) (result []*entity.MessageTemplate, code int, err error) {

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, params.WorkspaceID, accountID)

	if err != nil {
		return nil, code, eris.Wrap(err, "MessageTemplateList")
	}

	// fetch
	result, err = svc.Repo.ListMessageTemplates(ctx, workspace.ID, params)

	if err != nil {
		return nil, 500, err
	}

	return result, 200, nil
}
