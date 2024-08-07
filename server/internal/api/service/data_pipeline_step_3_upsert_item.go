package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/rimdian/rimdian/internal/api/entity"
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
			// so we need to update it
			pipe.DataLog.UpsertedSubscriptionListUser.UserID = pipe.DataLog.UpsertedUser.ID

			if txErr = pipe.UpsertSubscriptionListUser(spanCtx, isChild, tx); txErr != nil {
				return 500, txErr
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

		// other items can emit messages, so we need to upsert them first
		// ex: new sessions that have a rimdian utm_id will update the message.first_click_at
		if pipe.DataLog.UpsertedMessage != nil {
			isChild := true
			if pipe.DataLog.Kind == "message" {
				isChild = false
			}

			// user_id might have been updated by user_alias
			// so we need to update it
			pipe.DataLog.UpsertedMessage.UserID = pipe.DataLog.UpsertedUser.ID

			if txErr = pipe.UpsertMessage(spanCtx, isChild, tx); txErr != nil {
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
