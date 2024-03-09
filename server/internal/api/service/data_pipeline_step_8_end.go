package service

import (
	"context"

	"github.com/rimdian/rimdian/internal/api/entity"
	"go.opencensus.io/trace"
)

func (pipe *DataLogPipeline) StepEnd(ctx context.Context) {

	_, span := trace.StartSpan(ctx, "StepEnd")
	defer span.End()

	if !pipe.HasError() {
		pipe.DataLog.Checkpoint = entity.DataLogCheckpointDone
	}
}
