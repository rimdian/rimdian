package service

import (
	"context"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (svc *ServiceImpl) CubeJSSchemas(ctx context.Context, accountID string, workspaceID string) (schemas dto.CubeJSSchemas, code int, err error) {

	// fetch workspace
	workspace, err := svc.Repo.GetWorkspace(ctx, workspaceID)

	if err != nil {
		if sqlscan.NotFound(err) {
			return schemas, 400, err
		}
		return schemas, 500, eris.Wrap(err, "CubeJSSchemas")
	}

	// verify that workspace belongs to its organization
	isAccount, code, err := svc.IsAccountOfOrganization(ctx, accountID, workspace.OrganizationID)

	if err != nil {
		return schemas, code, eris.Wrap(err, "CubeJSSchemas")
	}

	if !isAccount {
		return nil, 400, eris.New("account is not part of the organization")
	}

	files := []dto.CubeJSSchemaFile{}
	schemasMap := entity.GenerateSchemas(workspace.InstalledApps)

	for fileName, schema := range schemasMap {
		files = append(files, dto.CubeJSSchemaFile{FileName: fileName, Content: schema.BuildContent(fileName)})
	}

	return files, 200, nil
}
