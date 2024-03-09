package service

import (
	"context"

	"github.com/rimdian/rimdian/internal/api/entity"
	"go.opencensus.io/trace"
)

func (pipe *DataLogPipeline) StepWorkflows(ctx context.Context) {

	_, span := trace.StartSpan(ctx, "StepWorkflows")
	defer span.End()

	// to implement
	if !pipe.HasError() {
		pipe.DataLog.Checkpoint = entity.DataLogCheckpointWorkflowsTriggered
	}
}
