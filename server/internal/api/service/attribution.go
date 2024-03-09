package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/internal/api/repository"
	common "github.com/rimdian/rimdian/internal/common/dto"
	"github.com/rotisserie/eris"
	"go.opencensus.io/trace"
)

func TaskExecReattributeConversions(ctx context.Context, pipe *TaskExecPipeline) (result *entity.TaskExecResult) {

	spanCtx, span := trace.StartSpan(ctx, "TaskExecReattributeConversions")
	defer span.End()

	result = &entity.TaskExecResult{
		// keep current state by default
		UpdatedWorkerState: pipe.TaskExec.State.Workers[pipe.TaskExecPayload.WorkerID],
	}

	select {
	case <-ctx.Done():
		result.SetError("task execution timeout", false)
		return
	default:
	}

	// log time taken
	startedAt := time.Now()
	defer func() {
		log.Printf("TaskReattributeConversions: workspace %s, task %s, worker %d, took %s", pipe.Workspace.ID, pipe.TaskExec.ID, pipe.TaskExecPayload.WorkerID, time.Since(startedAt))
	}()

	bgCtx := context.Background()

	mainState := pipe.TaskExec.State.Workers[0]

	if len(mainState) == 0 {
		mainState = map[string]interface{}{
			"current_step":    "system_reattribute_conversions",
			"users_processed": 0.0,
		}
	}

	switch mainState["current_step"].(string) {
	case "system_reattribute_conversions":
		// get the number of users processed
		processed := mainState["users_processed"].(float64)

		limit := 10

		// get 10 users with orders to reattribute
		userIDs, err := pipe.Repository.FindUserIDsWithOrdersToReattribute(bgCtx, pipe.Workspace.ID, limit)

		if err != nil {
			result.SetError(err.Error(), false)
			return
		}

		// add users to lock
		for _, userID := range userIDs {
			pipe.UsersLock.AddUser(userID)
		}

		// ensure users are locked
		if err := pipe.EnsureUsersLock(spanCtx); err != nil {
			result.SetError(err.Error(), false)
			return
		}

		// log.Printf("TaskReattributeConversions: got %+v users to process", userIDs)

		// process users
		pipe.ReattributeUsersOrders(spanCtx)

		if pipe.HasError() {
			isDone := true
			if pipe.QueueResult.QueueShouldRetry {
				isDone = false
			}
			result.SetError(pipe.QueueResult.Error, isDone)
			return
		}

		if len(userIDs) < limit {
			// go to next step
			mainState["current_step"] = "finalize"
		}

		// increment user processed
		processed += float64(len(userIDs))

		// update the state of the main thread
		mainState["users_processed"] = processed

	case "finalize":
		// respawn eventual data_logs to process workflows+hooks
		var withNextToken *string

		if _, ok := mainState["next_token"]; ok {
			withNextToken = entity.StringPtr(mainState["next_token"].(string))
		}

		// get 10 rows and enqueue them, and do it again until we have no more rows or we have less than 5 secs remaining

		shouldContinue := true
		hasMoreRows := true
		limit := 50 //  50 rows = 5 secs with 100ms enqueing latency per row

		for shouldContinue {

			// check if the we have less than 5 secs remaining
			if deadline, _ := spanCtx.Deadline(); time.Until(deadline) < 5*time.Second {
				shouldContinue = false
				log.Printf("TaskExecReattributeConversions: deadline ellapsed, should continue = false")
				continue
			}

			// fetch 11 rows but will only enqueue 10
			// the last row will be used to determine if we have more rows
			rows, err := pipe.Repo().ListDataLogsToRespawn(spanCtx, pipe.Workspace.ID,
				common.DataLogOriginInternalTaskExec,
				pipe.TaskExec.ID,
				entity.DataLogCheckpointWorkflowsTriggered,
				limit+1,
				withNextToken,
			)

			if err != nil {
				if sqlscan.NotFound(err) {
					result.IsDone = true
					mainState["current_step"] = "done"
					result.UpdatedWorkerState = mainState
					return
				}

				result.SetError(fmt.Sprintf("ListDataLogsToRespawn err %v", err), false)
				return
			}

			for _, row := range rows {
				replayID := row.ID // copy the value
				if err := DataLogEnqueue(ctx, pipe.Config, pipe.NetClient, &replayID, common.DataLogOriginInternalTaskExec, pipe.TaskExec.ID, pipe.Workspace.ID, []string{""}, false); err != nil {
					result.SetError(fmt.Sprintf("TaskExecReattributeConversions err %v", err), false)
					return
				}
			}

			// we have no more rows
			if len(rows) < limit+1 {
				shouldContinue = false
				hasMoreRows = false
				continue
			} else {
				withNextToken = entity.StringPtr(dto.EncodePaginationToken(rows[len(rows)-1].ID, rows[len(rows)-1].EventAt))
			}
		}

		// compute next token if we have more rows
		if hasMoreRows {
			mainState["next_token"] = withNextToken
		} else {
			// delete the next token
			delete(mainState, "next_token")
			mainState["current_step"] = "done"

			result.UpdatedWorkerState = mainState
			result.IsDone = true
			return
		}
	default:
		// unknown step
		result.SetError("unknown step", true)
		return
	}

	result.UpdatedWorkerState = mainState
	return
}

// retrieves orders and touchpoints (sessions, impressions) for a given user
// computes attribution for each touchpoint and KPIs for the user profile
func ReattributeUsersOrders(ctx context.Context, pipe Pipeline) {

	spanCtx, span := trace.StartSpan(ctx, "ReattributeUsersOrders")
	defer span.End()

	// abort if no user ids found
	if len(pipe.GetUserIDs()) == 0 {
		return
	}

	// fetch segments
	// var segments []*entity.Segment
	// segments, err = svc.Repo.ListSegments(ctx, pipe.Workspace.ID, false)

	// if err != nil {
	// 	err = eris.Wrap(err, "ReattributeUsersOrders")
	// 	return
	// }

	// create a wait group and process each user in a goroutine
	wg := &sync.WaitGroup{}

	type result struct {
		err error
	}

	results := make([]*result, len(pipe.GetUserIDs()))

	for i, userID := range pipe.GetUserIDs() {

		wg.Add(1)

		go func(i int, uID string) {
			defer wg.Done()

			results[i] = &result{}

			// create a transaction for each user
			_, err := pipe.Repo().RunInTransactionForWorkspace(spanCtx, pipe.GetWorkspace().ID, func(ctx context.Context, tx *sql.Tx) (txCode int, txErr error) {

				// fetch user in TX
				user, err := pipe.Repo().FindUserByID(ctx, pipe.GetWorkspace(), uID, tx)

				if err != nil {
					return 500, eris.Wrap(err, "ReattributeUsersOrders")
				}

				if user == nil {
					// user doesnt exist, it has been aliased/merged, end here
					return 200, nil
				}

				// abort if user is merged
				if user.IsMerged {
					return 200, nil
				}

				// fetch user devices
				devices, err := pipe.Repo().ListDevicesForUser(ctx, pipe.GetWorkspace(), user.ID, "created_at ASC", tx)

				if err != nil {
					return 500, eris.Wrap(err, "ReattributeUsersOrders")
				}

				// fetch existing orders for this user
				var orders []*entity.Order
				orders, txErr = pipe.Repo().ListOrdersForUser(ctx, pipe.GetWorkspace(), user.ID, "created_at ASC", tx)

				if txErr != nil {
					return 500, eris.Wrap(txErr, "ReattributeUsersOrders")
				}

				// log.Printf("got %v orders for user %v", len(orders), user.ID)

				// abort if had no orders
				if len(orders) == 0 {
					return 200, nil
				}

				// fetch existing sessions for this user
				sessions, txErr := pipe.Repo().ListSessionsForUser(ctx, pipe.GetWorkspace(), user.ID, "created_at ASC", tx)

				if txErr != nil {
					txErr = eris.Wrap(txErr, "ReattributeUsersOrders")
					return
				}

				// fetch existing postviews for this user
				postviews, txErr := pipe.Repo().ListPostviewsForUser(ctx, pipe.GetWorkspace(), user.ID, "created_at ASC", tx)
				if txErr != nil {
					txErr = eris.Wrap(txErr, "ReattributeUsersOrders")
					return
				}

				// reevaluate which order is the first
				var firstOrder *entity.Order

				for _, order := range orders {
					if firstOrder == nil || order.CreatedAt.Before(firstOrder.CreatedAt) {
						firstOrder = order
					}
				}

				for _, order := range orders {
					if firstOrder.ID == order.ID {
						order.IsFirstConversion = true
					} else {
						order.IsFirstConversion = false
					}
				}

				// attribute each order
				var previousOrderAt *time.Time

				for i, order := range orders {

					// only keep sessions and impressions for this order
					orderSessions := []*entity.Session{}
					orderPostviews := []*entity.Postview{}

					previousOrders := []*entity.Order{}

					if i > 0 {
						previousOrders = orders[:i]

						// set previous order at
						previousOrderAt = &orders[i-1].CreatedAt
					}

					for _, session := range sessions {
						if order.CreatedAt.IsZero() || session.CreatedAt.IsZero() {
							continue
						}
						// filter out sessions that are before the previous order
						if previousOrderAt != nil && session.CreatedAt.Before(*previousOrderAt) {
							continue
						}
						// filter out sessions that are after conversion
						if session.CreatedAt.After(order.CreatedAt) {
							continue
						}
						// add row to timeline
						orderSessions = append(orderSessions, session)
					}

					for _, pv := range postviews {
						if order.CreatedAt.IsZero() || pv.CreatedAt.IsZero() {
							continue
						}
						// filter out impressions that are before the previous order
						if previousOrderAt != nil && pv.CreatedAt.Before(*previousOrderAt) {
							continue
						}
						// filter out impressions that are after conversion
						if pv.CreatedAt.After(order.CreatedAt) {
							continue
						}
						// add row to timeline
						orderPostviews = append(orderPostviews, pv)
					}

					pipe.AttributeOrder(spanCtx, order, orderSessions, orderPostviews, previousOrders, devices, tx)

					if pipe.HasError() {
						return
					}
				}

				// compute user KPIs
				// init
				ordersCount := int64(0)
				ordersLTV := int64(0)
				ordersAvgCart := int64(0)
				var lastOrderAt *time.Time
				repeatLTV := int64(0)
				repeatCount := int64(0)
				repeatTTC := int64(0)
				avgRepeatCart := int64(0)
				avgRepeatOrderTTC := int64(0)

				for _, order := range orders {
					// ignore cancelled orders
					if order.CancelledAt != nil {
						continue
					}
					if lastOrderAt == nil {
						lastOrderAt = &order.CreatedAt
					}
					if order.CreatedAt.After(*lastOrderAt) {
						lastOrderAt = &order.CreatedAt
					}
					ordersCount++
					if order.SubtotalPrice != nil {
						ordersLTV += order.SubtotalPrice.Int64
					}

					if !order.IsFirstConversion {
						// set last order at
						if lastOrderAt == nil || order.CreatedAt.After(*lastOrderAt) {
							lastOrderAt = &order.CreatedAt
						}
						repeatCount++
						if order.SubtotalPrice != nil {
							repeatLTV += order.SubtotalPrice.Int64
						}
						if order.TimeToConversion != nil {
							repeatTTC += *order.TimeToConversion
						}
					}
				}

				// compute ordersAvgCart
				if ordersCount > 0 && ordersLTV > 0 {
					ordersAvgCart = int64(ordersLTV / ordersCount)
				}

				// compute avgRepeatCart
				if repeatCount > 0 && repeatLTV > 0 {
					avgRepeatCart = int64(repeatLTV / repeatCount)
				}

				// compute avgRepeatOrderTTC
				if repeatCount > 0 && repeatTTC > 0 {
					avgRepeatOrderTTC = int64(repeatTTC / repeatCount)
				}

				now := time.Now()

				updatedFields := []*entity.UpdatedField{}
				if fieldUpdate := user.SetOrdersCount(ordersCount, now); fieldUpdate != nil {
					updatedFields = append(updatedFields, fieldUpdate)
				}
				if fieldUpdate := user.SetOrdersLTV(ordersLTV, now); fieldUpdate != nil {
					updatedFields = append(updatedFields, fieldUpdate)
				}
				if fieldUpdate := user.SetOrdersAvgCart(ordersAvgCart, now); fieldUpdate != nil {
					updatedFields = append(updatedFields, fieldUpdate)
				}

				if firstOrder != nil {

					// extract domain type
					firstDomainType := entity.DomainWeb // default
					for _, domain := range pipe.GetWorkspace().Domains {
						if domain.ID == firstOrder.DomainID {
							firstDomainType = domain.Type
							break
						}
					}

					if fieldUpdate := user.SetFirstOrderAt(&firstOrder.CreatedAt, now); fieldUpdate != nil {
						updatedFields = append(updatedFields, fieldUpdate)
					}
					if fieldUpdate := user.SetFirstOrderDomainID(&firstOrder.DomainID, now); fieldUpdate != nil {
						updatedFields = append(updatedFields, fieldUpdate)
					}
					if fieldUpdate := user.SetFirstOrderDomainType(&firstDomainType, now); fieldUpdate != nil {
						updatedFields = append(updatedFields, fieldUpdate)
					}
					if firstOrder.SubtotalPrice != nil {
						if fieldUpdate := user.SetFirstOrderSubtotal(firstOrder.SubtotalPrice.Int64, now); fieldUpdate != nil {
							updatedFields = append(updatedFields, fieldUpdate)
						}
					}
					if firstOrder.TimeToConversion != nil {
						if fieldUpdate := user.SetFirstOrderTTC(*firstOrder.TimeToConversion, now); fieldUpdate != nil {
							updatedFields = append(updatedFields, fieldUpdate)
						}
					}
				} else {
					if fieldUpdate := user.SetFirstOrderAt(nil, now); fieldUpdate != nil {
						updatedFields = append(updatedFields, fieldUpdate)
					}
					if fieldUpdate := user.SetFirstOrderSubtotal(0, now); fieldUpdate != nil {
						updatedFields = append(updatedFields, fieldUpdate)
					}
					if fieldUpdate := user.SetFirstOrderTTC(0, now); fieldUpdate != nil {
						updatedFields = append(updatedFields, fieldUpdate)
					}
				}

				if fieldUpdate := user.SetLastOrderAt(lastOrderAt, now); fieldUpdate != nil {
					updatedFields = append(updatedFields, fieldUpdate)
				}
				if fieldUpdate := user.SetAvgRepeatCart(avgRepeatCart, now); fieldUpdate != nil {
					updatedFields = append(updatedFields, fieldUpdate)
				}
				if fieldUpdate := user.SetAvgRepeatOrderTTC(avgRepeatOrderTTC, now); fieldUpdate != nil {
					updatedFields = append(updatedFields, fieldUpdate)
				}

				user.UpdatedAt = &now

				// save user with new computed KPIs

				if len(updatedFields) > 0 {
					if err = pipe.Repo().UpdateUser(spanCtx, user, tx); err != nil {
						if eris.Is(err, repository.ErrRowNotUpdated) {
							// debug updatedFields
							updatedFieldsJSON, _ := json.Marshal(updatedFields)
							log.Printf("user row not updated with fields: %s", string(updatedFieldsJSON))
							return 200, nil
						}
						return 500, eris.Wrap(err, "ReattributeUsersOrders")
					}

					if err := pipe.InsertChildDataLog(spanCtx, "user", "update", user.ID, user.ID, user.ExternalID, updatedFields, *user.UpdatedAt, tx); err != nil {
						return 500, eris.Wrap(err, "ReattributeUsersOrders")
					}
				}

				return 200, nil
			})

			if err != nil {
				results[i].err = err
				return
			}

		}(i, userID)
	}

	// wait for all goroutines to finish
	wg.Wait()

	// check results
	for _, result := range results {
		if result.err != nil {
			pipe.SetError("reattribute orders", result.err.Error(), true)
			return
		}
	}
}
