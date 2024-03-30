package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/internal/common/dto"
	"go.opencensus.io/trace"
)

func (pipe *DataLogPipeline) StepUpsertItem(ctx context.Context) {

	spanCtx, span := trace.StartSpan(ctx, "StepUpsertItem")
	defer span.End()

	// acquire users lock
	if err := pipe.EnsureUsersLock(spanCtx); err != nil {
		// locks could not be acquired, DB is busy... end here and should retry
		pipe.SetError("server", fmt.Sprintf("doDataLog: %v", err), true)
		return
	}

	// upsert items
	_, retryableErr := pipe.Repository.RunInTransactionForWorkspace(spanCtx, pipe.Workspace.ID, func(ctx context.Context, tx *sql.Tx) (txCode int, txErr error) {

		// upsert user before other items, has it might find a user_alias
		// and mutate its user_id
		if pipe.DataLog.UpsertedUser != nil {
			isChild := true
			if pipe.DataLog.Kind == "user" {
				isChild = false
			}

			if txErr = pipe.UpsertUser(spanCtx, isChild, tx); txErr != nil {
				return 500, txErr
			}
		}

		if pipe.DataLog.UserAlias != nil {
			isChild := true
			if pipe.DataLog.Kind == "user_alias" {
				isChild = false
			}

			if txErr = pipe.UpsertUserAlias(spanCtx, isChild, tx); txErr != nil {
				return 500, txErr
			}
		}

		if pipe.DataLog.UpsertedPageview != nil {
			isChild := true
			if pipe.DataLog.Kind == "pageview" {
				isChild = false
			}

			// user_id might have been updated by user_alias
			// so we need to update it in the pageview
			pipe.DataLog.UpsertedPageview.UserID = pipe.DataLog.UpsertedUser.ID

			if txErr = pipe.UpsertPageview(spanCtx, isChild, tx); txErr != nil {
				return 500, txErr
			}
		}

		if pipe.DataLog.UpsertedOrder != nil {
			isChild := true
			if pipe.DataLog.Kind == "order" {
				isChild = false
			}

			// user_id might have been updated by user_alias
			// so we need to update it in the pageview
			pipe.DataLog.UpsertedOrder.UserID = pipe.DataLog.UpsertedUser.ID
			// update order_items
			for _, orderItem := range pipe.DataLog.UpsertedOrder.Items {
				orderItem.UserID = pipe.DataLog.UpsertedUser.ID
			}

			if txErr = pipe.UpsertOrder(spanCtx, isChild, tx); txErr != nil {
				return 500, txErr
			}
		}

		if pipe.DataLog.UpsertedCart != nil {
			isChild := true
			if pipe.DataLog.Kind == "cart" {
				isChild = false
			}

			// user_id might have been updated by user_alias
			// so we need to update it in the pageview
			pipe.DataLog.UpsertedCart.UserID = pipe.DataLog.UpsertedUser.ID
			// update cart_items
			for _, cartItem := range pipe.DataLog.UpsertedCart.Items {
				cartItem.UserID = pipe.DataLog.UpsertedUser.ID
			}

			if txErr = pipe.UpsertCart(spanCtx, isChild, tx); txErr != nil {
				return 500, txErr
			}
		}

		if pipe.DataLog.UpsertedCustomEvent != nil {
			isChild := true
			if pipe.DataLog.Kind == "custom_event" {
				isChild = false
			}

			// user_id might have been updated by user_alias
			// so we need to update it in the pageview
			pipe.DataLog.UpsertedCustomEvent.UserID = pipe.DataLog.UpsertedUser.ID

			if txErr = pipe.UpsertCustomEvent(spanCtx, isChild, tx); txErr != nil {
				return 500, txErr
			}
		}

		if pipe.DataLog.UpsertedDevice != nil {
			isChild := true
			if pipe.DataLog.Kind == "device" {
				isChild = false
			}

			// user_id might have been updated by user_alias
			// so we need to update it in the pageview
			pipe.DataLog.UpsertedDevice.UserID = pipe.DataLog.UpsertedUser.ID

			if txErr = pipe.UpsertDevice(spanCtx, isChild, tx); txErr != nil {
				return 500, txErr
			}
		}

		if pipe.DataLog.UpsertedSession != nil {
			isChild := true
			if pipe.DataLog.Kind == "session" {
				isChild = false
			}

			// user_id might have been updated by user_alias
			// so we need to update it in the pageview
			pipe.DataLog.UpsertedSession.UserID = pipe.DataLog.UpsertedUser.ID

			if txErr = pipe.UpsertSession(spanCtx, isChild, tx); txErr != nil {
				return 500, txErr
			}
		}

		if pipe.DataLog.UpsertedPostview != nil {
			isChild := true
			if pipe.DataLog.Kind == "postview" {
				isChild = false
			}

			// user_id might have been updated by user_alias
			// so we need to update it in the pageview
			pipe.DataLog.UpsertedPostview.UserID = pipe.DataLog.UpsertedUser.ID

			if txErr = pipe.UpsertPostview(spanCtx, isChild, tx); txErr != nil {
				return 500, txErr
			}
		}

		if pipe.DataLog.UpsertedSubscriptionListUser != nil {
			isChild := true
			if pipe.DataLog.Kind == "subscription_list_user" {
				isChild = false
			}

			// user_id might have been updated by user_alias
			// so we need to update it in the pageview
			pipe.DataLog.UpsertedSubscriptionListUser.UserID = pipe.DataLog.UpsertedUser.ID

			if txErr = pipe.UpsertSubscriptionListUser(spanCtx, isChild, tx); txErr != nil {
				return 500, txErr
			}

			// send double opt-in message

			if pipe.DataLog.Action == "create" &&
				pipe.DataLog.UpsertedSubscriptionListUser.SubscriptionList != nil &&
				pipe.DataLog.UpsertedSubscriptionListUser.SubscriptionList.DoubleOptIn &&
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
		}

		if pipe.DataLog.UpsertedAppItem != nil {
			isChild := true
			if strings.HasPrefix(pipe.DataLog.Kind, "app_") || strings.HasPrefix(pipe.DataLog.Kind, "appx_") {
				isChild = false
			}

			// user_id might have been updated by user_alias
			// so we need to update it in the pageview
			// could also be "none" for app_item
			if pipe.DataLog.UpsertedUser != nil {
				pipe.DataLog.UpsertedAppItem.UserID = pipe.DataLog.UpsertedUser.ID
			}

			if txErr = pipe.UpsertAppItem(spanCtx, isChild, tx); txErr != nil {
				return 500, txErr
			}
		}

		return 200, nil
	})

	if retryableErr != nil {
		pipe.SetError("server", fmt.Sprintf("will retry: %v", retryableErr), true)
		return
	}

	// for user_alias we delete old user data after the merge
	if pipe.DataLog.UserAlias != nil {
		if err := pipe.Repository.CleanAfterUserAlias(pipe.Workspace.ID, pipe.DataLog.UserAlias.FromUserExternalID); err != nil {
			// just log error and continue
			pipe.Logger.Println(err)
		}
	}

	// set status
	pipe.DataLog.Checkpoint = entity.DataLogCheckpointItemUpserted
}
