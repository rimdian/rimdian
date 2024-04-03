package service

import (
	"context"
	"database/sql"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/internal/api/repository"
	"github.com/rotisserie/eris"
	"go.opencensus.io/trace"
)

func (pipe *DataLogPipeline) UpsertCart(ctx context.Context, isChild bool, tx *sql.Tx) (err error) {

	spanCtx, span := trace.StartSpan(ctx, "UpsertCart")
	defer span.End()

	// find eventual existing cart
	var existingCart *entity.Cart
	updatedFields := []*entity.UpdatedField{}

	existingCart, err = pipe.Repository.FindCartByID(spanCtx, pipe.Workspace.ID, pipe.DataLog.UpsertedCart.ID, pipe.DataLog.UpsertedCart.UserID, tx)

	if err != nil {
		return eris.Wrap(err, "CartUpsert")
	}

	// insert new cart
	if existingCart == nil {

		// just for insert: clear fields timestamp if object is new, to avoid storing extra data
		pipe.DataLog.UpsertedCart.FieldsTimestamp = entity.FieldsTimestamp{}

		if err = pipe.Repository.InsertCart(spanCtx, pipe.DataLog.UpsertedCart, tx); err != nil {
			return
		}

		if isChild {
			if err := pipe.InsertChildDataLog(spanCtx, entity.ChildDataLog{
				Kind:           "cart",
				Action:         "update",
				UserID:         pipe.DataLog.UpsertedUser.ID,
				ItemID:         pipe.DataLog.UpsertedCart.ID,
				ItemExternalID: pipe.DataLog.UpsertedCart.ExternalID,
				UpdatedFields:  updatedFields,
				EventAt:        *pipe.DataLog.UpsertedCart.UpdatedAt,
				Tx:             tx,
			}); err != nil {
				return err
			}
		} else {
			pipe.DataLog.Action = "create"
		}

		// insert cart items
		for _, cartItem := range pipe.DataLog.UpsertedCart.Items {
			err = pipe.Repository.InsertCartItem(spanCtx, cartItem, tx)

			if err != nil {
				if eris.Is(err, repository.ErrRowAlreadyExists) {
					// ignore if already exists, should not happend though...
					continue
				}
				return eris.Wrap(err, "CartUpsert (insert)")
			}

			// items are always children
			if err := pipe.InsertChildDataLog(spanCtx, entity.ChildDataLog{
				Kind:           "cart_item",
				Action:         "create",
				UserID:         pipe.DataLog.UpsertedUser.ID,
				ItemID:         cartItem.ID,
				ItemExternalID: cartItem.ExternalID,
				UpdatedFields:  entity.UpdatedFields{},
				EventAt:        *pipe.DataLog.UpsertedCart.UpdatedAt,
				Tx:             tx,
			}); err != nil {
				return err
			}
		}

		return
	}

	// keep a copy of the upserted cart items, before overwriting them
	upsertedCartItems := pipe.DataLog.UpsertedCart.Items

	// merge fields if cart already exists
	updatedFields = pipe.DataLog.UpsertedCart.MergeInto(existingCart)
	pipe.DataLog.UpsertedCart = existingCart
	pipe.DataLog.UpsertedCart.Items = upsertedCartItems

	// dont abort if no fields were updated, we still need to upsert the cart items
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

		// persist cart changes
		if err = pipe.Repository.UpdateCart(spanCtx, pipe.DataLog.UpsertedCart, tx); err != nil {
			return eris.Wrap(err, "CartUpsert (update)")
		}

		if isChild {
			if err := pipe.InsertChildDataLog(spanCtx, entity.ChildDataLog{
				Kind:           "cart",
				Action:         "update",
				UserID:         pipe.DataLog.UpsertedUser.ID,
				ItemID:         pipe.DataLog.UpsertedCart.ID,
				ItemExternalID: pipe.DataLog.UpsertedCart.ExternalID,
				UpdatedFields:  updatedFields,
				EventAt:        *pipe.DataLog.UpsertedCart.UpdatedAt,
				Tx:             tx,
			}); err != nil {
				return err
			}
		}
	}

	// upsert cart items
	// fetch items
	var existingCartItems entity.CartItems
	if existingCartItems, err = pipe.Repository.FindCartItemsByCartID(spanCtx, pipe.Workspace.ID, existingCart.ID, existingCart.UserID, tx); err != nil {
		return eris.Wrap(err, "CartUpsert (update)")
	}

	// delete items that are not in the new cart
	for _, existingCartItem := range existingCartItems {

		existsInNewCart := false
		for _, cartItem := range pipe.DataLog.UpsertedCart.Items {
			if cartItem.ID == existingCartItem.ID {
				existsInNewCart = true
				break
			}
		}
		if !existsInNewCart {
			if err = pipe.Repository.DeleteCartItem(spanCtx, existingCartItem.ID, existingCartItem.UserID, tx); err != nil {
				return eris.Wrap(err, "CartUpsert (update)")
			}

			// items are always children
			if err := pipe.InsertChildDataLog(spanCtx, entity.ChildDataLog{
				Kind:           "cart_item",
				Action:         "delete",
				UserID:         pipe.DataLog.UpsertedUser.ID,
				ItemID:         existingCartItem.ID,
				ItemExternalID: existingCartItem.ExternalID,
				UpdatedFields:  entity.UpdatedFields{},
				EventAt:        *pipe.DataLog.UpsertedCart.UpdatedAt,
				Tx:             tx,
			}); err != nil {
				return err
			}
		}
	}

	// update modified cart items
	for _, cartItem := range upsertedCartItems {

		// find existing cart item in the cart
		var foundCartItem *entity.CartItem
		for _, x := range existingCartItems {
			if x.ID == cartItem.ID {
				foundCartItem = x
			}
		}

		if foundCartItem == nil {
			// insert new cart item
			if err = pipe.Repository.InsertCartItem(spanCtx, cartItem, tx); err != nil {
				return eris.Wrap(err, "CartUpsert (update)")
			}
			// items are always children
			if err := pipe.InsertChildDataLog(spanCtx, entity.ChildDataLog{
				Kind:           "cart_item",
				Action:         "create",
				UserID:         pipe.DataLog.UpsertedUser.ID,
				ItemID:         cartItem.ID,
				ItemExternalID: cartItem.ExternalID,
				UpdatedFields:  entity.UpdatedFields{},
				EventAt:        *pipe.DataLog.UpsertedCart.UpdatedAt,
				Tx:             tx,
			}); err != nil {
				return err
			}
		} else {

			// merge fields if cart already exists
			updatedFields := cartItem.MergeInto(foundCartItem)
			cartItem = foundCartItem

			if len(updatedFields) > 0 {
				if err = pipe.Repository.UpdateCartItem(spanCtx, cartItem, tx); err != nil {
					return eris.Wrap(err, "CartUpsert")
				}

				if err := pipe.InsertChildDataLog(spanCtx, entity.ChildDataLog{
					Kind:           "cart_item",
					Action:         "update",
					UserID:         pipe.DataLog.UpsertedUser.ID,
					ItemID:         foundCartItem.ID,
					ItemExternalID: foundCartItem.ExternalID,
					UpdatedFields:  updatedFields,
					EventAt:        *pipe.DataLog.UpsertedCart.UpdatedAt,
					Tx:             tx,
				}); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
