package service

import (
	"context"

	"github.com/rimdian/rimdian/internal/api/dto"
)

func (svc *ServiceImpl) FolderFiles(ctx context.Context, accountID string, params *dto.FolderFilesParams) (result *dto.FolderFilesResult, code int, err error) {
	// TODO
	return nil, 0, nil
}
