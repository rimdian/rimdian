package service

import (
	"context"
	"database/sql"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/internal/api/repository"
	"github.com/rotisserie/eris"
	"go.opencensus.io/trace"
)

func (pipe *DataLogPipeline) UpsertOrder(ctx context.Context, isChild bool, tx *sql.Tx) (err error) {

	spanCtx, span := trace.StartSpan(ctx, "UpsertOrder")
	defer span.End()

	// find eventual existing order
	var existingOrder *entity.Order
	updatedFields := []*entity.UpdatedField{}

	existingOrder, err = pipe.Repository.FindOrderByID(spanCtx, pipe.Workspace, pipe.DataLog.UpsertedOrder.ID, pipe.DataLog.UpsertedOrder.UserID, tx)

	if err != nil {
		return eris.Wrap(err, "OrderUpsert")
	}

	// insert new order
	if existingOrder == nil {

		// just for insert: clear fields timestamp if object is new, to avoid storing extra data
		pipe.DataLog.UpsertedOrder.FieldsTimestamp = entity.FieldsTimestamp{}

		if err = pipe.Repository.InsertOrder(spanCtx, pipe.DataLog.UpsertedOrder, tx); err != nil {
			return
		}

		if isChild {
			if err := pipe.InsertChildDataLog(spanCtx, "order", "create", pipe.DataLog.UpsertedUser.ID, pipe.DataLog.UpsertedOrder.ID, pipe.DataLog.UpsertedOrder.ExternalID, updatedFields, *pipe.DataLog.UpsertedOrder.UpdatedAt, tx); err != nil {
				return err
			}
		} else {
			pipe.DataLog.Action = "create"
		}

		// insert order items
		for _, orderItem := range pipe.DataLog.UpsertedOrder.Items {
			err = pipe.Repository.InsertOrderItem(spanCtx, orderItem, tx)

			if err != nil {
				if eris.Is(err, repository.ErrRowAlreadyExists) {
					// ignore if already exists, should not happend though...
					continue
				}
				return eris.Wrap(err, "OrderUpsert")
			}

			// items are always children
			if err := pipe.InsertChildDataLog(spanCtx, "order_item", "create", pipe.DataLog.UpsertedUser.ID, orderItem.ID, orderItem.ExternalID, entity.UpdatedFields{}, *pipe.DataLog.UpsertedOrder.UpdatedAt, tx); err != nil {
				return err
			}
		}

		return
	}

	// keep a copy of the upserted order items, before overwriting them
	upsertedOrderItems := pipe.DataLog.UpsertedOrder.Items

	// merge fields if order already exists
	updatedFields = pipe.DataLog.UpsertedOrder.MergeInto(existingOrder)
	pipe.DataLog.UpsertedOrder = existingOrder
	pipe.DataLog.UpsertedOrder.Items = upsertedOrderItems

	// dont abort if no fields were updated, we still need to upsert the order items
	if len(updatedFields) == 0 {
		if !isChild {
			pipe.DataLog.Action = "noop"
		}
	}

	if len(updatedFields) > 0 {

		if !isChild {
			pipe.DataLog.Action = "update"
			pipe.DataLog.UpdatedFields = updatedFields
		}

		// persist order changes
		if err = pipe.Repository.UpdateOrder(spanCtx, pipe.DataLog.UpsertedOrder, tx); err != nil {
			return eris.Wrap(err, "OrderUpsert")
		}

		if isChild {
			if err := pipe.InsertChildDataLog(spanCtx, "order", "update", pipe.DataLog.UpsertedUser.ID, pipe.DataLog.UpsertedOrder.ID, pipe.DataLog.UpsertedOrder.ExternalID, updatedFields, *pipe.DataLog.UpsertedOrder.UpdatedAt, tx); err != nil {
				return err
			}
		}
	}

	// upsert order items
	// fetch items
	var existingOrderItems entity.OrderItems
	if existingOrderItems, err = pipe.Repository.FindOrderItemsByOrderID(spanCtx, pipe.Workspace.ID, existingOrder.ID, existingOrder.UserID, tx); err != nil {
		return eris.Wrap(err, "OrderUpsert")
	}

	// delete items that are not in the new order
	for _, existingOrderItem := range existingOrderItems {

		existsInNewOrder := false
		for _, orderItem := range pipe.DataLog.UpsertedOrder.Items {
			if orderItem.ID == existingOrderItem.ID {
				existsInNewOrder = true
				break
			}
		}
		if !existsInNewOrder {
			if err = pipe.Repository.DeleteOrderItem(spanCtx, existingOrderItem.ID, existingOrderItem.UserID, tx); err != nil {
				return eris.Wrap(err, "OrderUpsert")
			}

			// items are always children
			if err := pipe.InsertChildDataLog(spanCtx, "order_item", "delete", pipe.DataLog.UpsertedUser.ID, existingOrderItem.ID, existingOrderItem.ExternalID, entity.UpdatedFields{}, *pipe.DataLog.UpsertedOrder.UpdatedAt, tx); err != nil {
				return err
			}
		}
	}

	// update modified order items
	for _, orderItem := range upsertedOrderItems {

		// find existing order item in the order
		var foundOrderItem *entity.OrderItem
		for _, x := range existingOrderItems {
			if x.ID == orderItem.ID {
				foundOrderItem = x
			}
		}

		if foundOrderItem == nil {
			// insert new order item
			if err = pipe.Repository.InsertOrderItem(spanCtx, orderItem, tx); err != nil {
				return eris.Wrap(err, "OrderUpsert")
			}
			// items are always children
			if err := pipe.InsertChildDataLog(spanCtx, "order_item", "create", pipe.DataLog.UpsertedUser.ID, orderItem.ID, orderItem.ExternalID, entity.UpdatedFields{}, *pipe.DataLog.UpsertedOrder.UpdatedAt, tx); err != nil {
				return err
			}
		} else {

			// merge fields if order already exists
			updatedFields := orderItem.MergeInto(foundOrderItem)
			orderItem = foundOrderItem

			if len(updatedFields) > 0 {
				if err = pipe.Repository.UpdateOrderItem(spanCtx, orderItem, tx); err != nil {
					return eris.Wrap(err, "OrderUpsert")
				}

				if err := pipe.InsertChildDataLog(spanCtx, "order_item", "create", pipe.DataLog.UpsertedUser.ID, foundOrderItem.ID, foundOrderItem.ExternalID, updatedFields, *pipe.DataLog.UpsertedOrder.UpdatedAt, tx); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
