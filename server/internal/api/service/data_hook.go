package service

import (
	"context"

	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

func (svc *ServiceImpl) DataHookUpdate(ctx context.Context, accountID string, dataHookDTO *dto.DataHook) (workspace *entity.Workspace, code int, err error) {

	workspace, code, err = svc.GetWorkspaceForAccount(ctx, dataHookDTO.WorkspaceID, accountID)

	if err != nil {
		return nil, code, eris.Wrap(err, "DataHookUpdate")
	}

	// find data hook
	var dataHook *entity.DataHook

	for _, dh := range workspace.DataHooks {
		if dh.ID == dataHookDTO.ID {
			dataHook = dh
			break
		}
	}

	if dataHook == nil {
		return nil, 400, eris.New("data hook not found")
	}

	dataHook.Enabled = dataHookDTO.Enabled

	if err := svc.Repo.UpdateWorkspace(ctx, workspace, nil); err != nil {
		return nil, 500, eris.Wrap(err, "DataHookUpdate")
	}

	return
}
