package service

import (
	"context"

	"github.com/rimdian/rimdian/internal/api/entity"
	"go.opencensus.io/trace"
)

func (pipe *DataLogPipeline) StepSegmentation(ctx context.Context) {

	spanCtx, span := trace.StartSpan(ctx, "StepSegmentation")
	defer span.End()

	pipe.ComputeSegmentsForGivenUsers(spanCtx)

	// set status
	if !pipe.HasError() {
		pipe.DataLog.Checkpoint = entity.DataLogCheckpointSegmentsRecomputed
		return
	}
}
