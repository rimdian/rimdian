package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/internal/common/dto"
	"github.com/rotisserie/eris"
	"go.opencensus.io/trace"
)

func (pipe *DataLogPipeline) UpsertSubscriptionListUser(ctx context.Context, isChild bool, tx *sql.Tx) (err error) {

	spanCtx, span := trace.StartSpan(ctx, "UpsertSubscriptionListUser")
	defer span.End()

	// find eventual existing subscription_list_user
	var existingSubscriptionListUser *entity.SubscriptionListUser
	updatedFields := []*entity.UpdatedField{}

	existingSubscriptionListUser, err = pipe.Repository.FindSubscriptionListUser(spanCtx, pipe.DataLog.UpsertedSubscriptionListUser.SubscriptionListID, pipe.DataLog.UpsertedSubscriptionListUser.UserID, tx)

	if err != nil {
		return eris.Wrap(err, "SubscriptionListUserUpsert")
	}

	// insert new subscription_list_user
	if existingSubscriptionListUser == nil {

		// default status to active if not set
		if pipe.DataLog.UpsertedSubscriptionListUser.Status == nil {
			pipe.DataLog.UpsertedSubscriptionListUser.Status = entity.Int64Ptr(entity.SubscriptionListUserStatusActive)
		}

		// just for insert: clear fields timestamp if object is new, to avoid storing extra data
		pipe.DataLog.UpsertedSubscriptionListUser.FieldsTimestamp = entity.FieldsTimestamp{}

		if err = pipe.Repository.InsertSubscriptionListUser(spanCtx, pipe.DataLog.UpsertedSubscriptionListUser, tx); err != nil {
			return
		}

		if isChild {
			if err := pipe.InsertChildDataLog(spanCtx, "subscription_list_user", "create", pipe.DataLog.UpsertedUser.ID, pipe.DataLog.UpsertedSubscriptionListUser.SubscriptionListID, pipe.DataLog.UpsertedSubscriptionListUser.SubscriptionListID, updatedFields, *pipe.DataLog.UpsertedSubscriptionListUser.UpdatedAt, tx); err != nil {
				return err
			}
		} else {
			pipe.DataLog.Action = "create"
		}

		// send double opt-in message

		if pipe.DataLog.UpsertedSubscriptionListUser.SubscriptionList.DoubleOptIn &&
			*pipe.DataLog.UpsertedSubscriptionListUser.Status == entity.SubscriptionListUserStatusPaused {

			message := fmt.Sprintf(`{
					"kind": "message",
					"message": {
						"external_id": "%s",
						"created_at": "%s",
						"channel": "%s",
						"message_template_id": "%s",
						"subscription_list_id": "%s"
					},
					"user": {
						"external_id": "%s",
						"is_authenticated": %t,
						"created_at": "%s"
					}
				}`,
				fmt.Sprintf("double_opt_in_%s", pipe.DataLog.ID),
				pipe.DataLog.UpsertedSubscriptionListUser.CreatedAt.Format(time.RFC3339),
				pipe.DataLog.UpsertedSubscriptionListUser.SubscriptionList.Channel,
				*pipe.DataLog.UpsertedSubscriptionListUser.SubscriptionList.MessageTemplateID,
				pipe.DataLog.UpsertedSubscriptionListUser.SubscriptionList.ID,
				pipe.DataLog.UpsertedUser.ExternalID,
				pipe.DataLog.UpsertedUser.IsAuthenticated,
				pipe.DataLog.UpsertedUser.CreatedAt.Format("2006-01-02T15:04:05Z"),
			)

			pipe.DataLogEnqueue(spanCtx, nil, dto.DataLogOriginInternalDataLog, pipe.DataLog.ID, pipe.Workspace.ID, []string{message}, false)
		}

		return
	}

	// merge fields if subscription_list_user already exists
	updatedFields = pipe.DataLog.UpsertedSubscriptionListUser.MergeInto(existingSubscriptionListUser)
	pipe.DataLog.UpsertedSubscriptionListUser = existingSubscriptionListUser

	// abort if no fields were updated
	if len(updatedFields) == 0 {
		if !isChild {
			pipe.DataLog.Action = "noop"
		}
		return nil
	}

	if !isChild {
		pipe.DataLog.Action = "update"
		pipe.DataLog.UpdatedFields = updatedFields
	}

	// persist changes
	if err = pipe.Repository.UpdateSubscriptionListUser(spanCtx, pipe.DataLog.UpsertedSubscriptionListUser, tx); err != nil {
		return eris.Wrap(err, "SubscriptionListUserUpsert")
	}

	if isChild {
		if err := pipe.InsertChildDataLog(spanCtx, "subscription_list_user", "update", pipe.DataLog.UpsertedUser.ID, pipe.DataLog.UpsertedSubscriptionListUser.SubscriptionListID, pipe.DataLog.UpsertedSubscriptionListUser.SubscriptionListID, updatedFields, *pipe.DataLog.UpsertedSubscriptionListUser.UpdatedAt, tx); err != nil {
			return err
		}
	}

	return nil
}
