package service

import (
	"context"
	"fmt"

	"github.com/rimdian/rimdian/internal/api/entity"
	"go.opencensus.io/trace"
)

func (pipe *DataLogPipeline) StepAttribution(ctx context.Context) {

	spanCtx, span := trace.StartSpan(ctx, "StepAttribution")
	defer span.End()

	// only do attribution it items contain a new session, a new order or a session with an updated channelID
	shouldAttribute := false

	for _, dataLog := range pipe.DataLogsGenerated {
		// acquire users locks if the import started from this point
		pipe.UsersLock.AddUser(dataLog.UserID)

		if dataLog.Kind == "user_alias" {
			shouldAttribute = true
			break
		}
		if dataLog.Kind == "session" || dataLog.Kind == "postview" {
			if dataLog.Action == "create" || dataLog.Action == "delete" {
				shouldAttribute = true
				break
			}
			if dataLog.Action == "update" {
				// check if updated field have channelID
				for _, field := range dataLog.UpdatedFields {
					if field.Field == "channel_id" {
						shouldAttribute = true
						break
					}
				}
			}
		}

		if dataLog.Kind == "order" {
			if dataLog.Action == "create" || dataLog.Action == "delete" {
				shouldAttribute = true
				break
			}

			if dataLog.Action == "update" {
				// check if subtotal_price has been update
				for _, field := range dataLog.UpdatedFields {
					if field.Field == "subtotal_price" {
						shouldAttribute = true
						break
					}
				}
			}
		}
	}

	if shouldAttribute {
		// ensure users are locked
		if err := pipe.EnsureUsersLock(spanCtx); err != nil {
			// locks could not be acquired, DB is busy... end here and should retry
			pipe.SetError("server", fmt.Sprintf("doDataLog: %v", err), true)
			pipe.ProcessNextStep(spanCtx)
			return
		}

		pipe.ReattributeUsersOrders(spanCtx)
	}

	// set status
	if !pipe.HasError() {
		pipe.DataLog.Checkpoint = entity.DataLogCheckpointConversionsAttributed
	}
}
