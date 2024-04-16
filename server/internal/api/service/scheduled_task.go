package service

import (
	"context"

	"github.com/rimdian/rimdian/internal/api/entity"
	common "github.com/rimdian/rimdian/internal/common/dto"
	"github.com/rimdian/rimdian/internal/common/taskorchestrator"
	"github.com/rotisserie/eris"
	"go.opencensus.io/trace"
)

func (svc *ServiceImpl) ScheduledTaskDo(ctx context.Context, scheduledTask *entity.ScheduledTask) (result *common.ResponseForTaskQueue) {
	// TODO: post the tasks exec
	return nil
}

func (svc *ServiceImpl) ScheduledTaskPost(ctx context.Context, scheduledTask entity.ScheduledTask) (err error) {

	spanCtx, span := trace.StartSpan(ctx, "ScheduledTaskPost")
	defer span.End()

	scheduledTask.TaskExec.EnsureID()

	if err = scheduledTask.Validate(); err != nil {
		return eris.Wrap(err, "ScheduledTaskPost")
	}

	googleTaskQueueJob := &taskorchestrator.TaskRequest{
		UniqueID:          &scheduledTask.TaskExec.ID,
		QueueLocation:     svc.Config.TASK_QUEUE_LOCATION,
		QueueName:         entity.ScheduledTasksQueueName,
		PostEndpoint:      svc.Config.API_ENDPOINT + entity.ScheduledTaskEndpoint,
		TaskTimeoutInSecs: &entity.TaskTimeoutInSecs,
		ScheduleTime:      &scheduledTask.ScheduledAt,
		Payload:           scheduledTask,
	}

	if err := svc.TaskOrchestrator.PostRequest(spanCtx, googleTaskQueueJob); err != nil {
		return eris.Wrap(err, "ScheduledTaskPost")
	}

	return nil
}
