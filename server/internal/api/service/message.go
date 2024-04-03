package service

import (
	"context"

	"github.com/rimdian/rimdian/internal/api/dto"
	common "github.com/rimdian/rimdian/internal/common/dto"
)

func (svc *ServiceImpl) MessageSend(ctx context.Context, data *dto.SendMessage) (result *common.DataLogInQueueResult) {

	// TODO
	return
}
