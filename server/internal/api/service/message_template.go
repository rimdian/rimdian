package service

import (
	"context"
	"database/sql"

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

func (svc *ServiceImpl) MessageTemplateUpsert(ctx context.Context, accountID string, data *dto.MessageTemplate) (code int, err error) {

	workspace, code, err := svc.GetWorkspaceForAccount(ctx, data.WorkspaceID, accountID)

	if err != nil {
		return code, eris.Wrap(err, "MessageTemplateUpsert")
	}

	// create message template from data

	messageTemplate := &entity.MessageTemplate{
		ID:              data.ID,
		Name:            data.Name,
		Channel:         data.Channel,
		Engine:          data.Engine,
		Email:           data.Email,
		TemplateMacroID: data.TemplateMacroID,
		UTMSource:       data.UTMSource,
		UTMMedium:       data.UTMMedium,
		UTMCampaign:     data.UTMCampaign,
		UTMContent:      data.UTMContent,
		Settings:        data.Settings,
		TestData:        data.TestData,
	}

	// validate
	if err := messageTemplate.Validate(); err != nil {
		return 400, err
	}

	code, err = svc.Repo.RunInTransactionForWorkspace(ctx, workspace.ID, func(ctx context.Context, tx *sql.Tx) (code int, err error) {
		// find existing message template
		existing, err := svc.Repo.GetMessageTemplate(ctx, workspace.ID, messageTemplate.ID, nil, tx)

		if err != nil {
			return 500, eris.Wrapf(err, "InsertMessageTemplate GetMessageTemplate")
		}

		if existing != nil {
			// if we want to create a new template, and a template with this ID already exists
			if messageTemplate.ID == "" {
				return 400, eris.New("A template with this ID already exists")
			}
			messageTemplate.Version = existing.Version + 1
		} else {
			messageTemplate.Version = 0
		}

		err = svc.Repo.InsertMessageTemplate(ctx, workspace.ID, messageTemplate, tx)

		if err != nil {
			return 500, err
		}

		return 201, nil
	})

	return code, err
}