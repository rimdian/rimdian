package service

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/rimdian/rimdian/internal/api/entity"
	"go.opencensus.io/trace"
)

// takes an order with its touchpoints and computes the attribution for each touchpoint
func AttributeOrder(ctx context.Context, pipe Pipeline, order *entity.Order, orderSessions []*entity.Session, orderPostviews []*entity.Postview, previousOrders []*entity.Order, devices []*entity.Device, tx *sql.Tx) (err error) {

	spanCtx, span := trace.StartSpan(ctx, "AttributeOrder")
	defer span.End()

	type Touchpoint struct {
		Session  *entity.Session
		Postview *entity.Postview
	}

	touchpoints := []Touchpoint{}

	// recompute if its first conversion
	order.IsFirstConversion = true

	if len(previousOrders) > 0 && previousOrders[0].CreatedAt.Before(order.CreatedAt) {
		order.IsFirstConversion = false
	}

	// compute time to conversion for touchpoints
	var orderTimeToConversion int64

	// orderSessions in sorted by ASC created_at
	for _, session := range orderSessions {

		// set the orderTimeToConversion (in secs) to the oldest hit
		ttc := int64(order.CreatedAt.Sub(session.CreatedAt).Seconds())
		if ttc > orderTimeToConversion {
			orderTimeToConversion = ttc
		}

		// add the session to the touchpoints
		touchpoints = append(touchpoints, Touchpoint{Session: session})
	}

	// update order session_id with the last timeline session if missing
	if order.SessionID == nil && len(orderSessions) > 0 {
		order.SessionID = &orderSessions[len(orderSessions)-1].ID
	}

	// orderPostviews in sorted by ASC created_at
	for _, pv := range orderPostviews {

		// set the orderTimeToConversion (in secs) to the oldest hit
		ttc := int64(order.CreatedAt.Sub(pv.CreatedAt).Seconds())
		if ttc > orderTimeToConversion {
			orderTimeToConversion = ttc
		}

		// add the impression to the touchpoints
		touchpoints = append(touchpoints, Touchpoint{Postview: pv})
	}

	totalTouchpoints := len(touchpoints)

	// attribute voucher code if exists
	if order.DiscountCodes != nil && len(*order.DiscountCodes) > 0 && totalTouchpoints > 0 && len(orderSessions) > 0 {
		// log.Printf("found order with discounts %v, %+v\n", order.ConversionExternalId, order.OrderData.Discounts)

		for _, discountCode := range *order.DiscountCodes {

			// we have a discount code, check if it is attributed to a channel
			for _, ch := range pipe.GetWorkspace().Channels {
				for _, voucher := range ch.VoucherCodes {

					if voucher.Code == discountCode {
						// we have a matching voucher
						// extract source and medium from channel voucher sourceMediumPath

						// log.Printf("channelOrigin %v", channelOrigin)

						lastSession := orderSessions[len(orderSessions)-1]

						// we need to reattribute the last session if its not yet the case
						if lastSession.ChannelOriginID != voucher.OriginID {

							// find channel origin that matches the voucher origin
							var channelOrigin *entity.ChannelOrigin
							for _, origin := range ch.Origins {
								if origin.ID == voucher.OriginID {
									channelOrigin = origin
									break
								}
							}

							// copy old source / medium / campaign to "via"
							lastSession.ViaUTMSource = lastSession.UTMSource
							lastSession.ViaUTMMedium = lastSession.UTMMedium
							if lastSession.UTMCampaign != nil {
								lastSession.ViaUTMCampaign = lastSession.UTMCampaign
							}

							// set new channel
							lastSession.UTMSource = &entity.NullableString{IsNull: false, String: channelOrigin.UTMSource}
							lastSession.UTMMedium = &entity.NullableString{IsNull: false, String: channelOrigin.UTMMedium}

							lastSession.UTMCampaign = &entity.NullableString{IsNull: true}
							if channelOrigin.UTMCampaign != nil {
								lastSession.UTMCampaign = &entity.NullableString{IsNull: false, String: *channelOrigin.UTMCampaign}
							}
							lastSession.ChannelID = ch.ID
							lastSession.ChannelGroupID = ch.GroupID

							// overwrite campaign if provided
							if voucher.SetUTMCampaign != nil && *voucher.SetUTMCampaign != "" {
								lastSession.UTMCampaign = &entity.NullableString{IsNull: false, String: *voucher.SetUTMCampaign}
							}

							// overwrite content if provided
							if voucher.SetUTMContent != nil && *voucher.SetUTMContent != "" {
								lastSession.ViaUTMContent = lastSession.UTMContent
								lastSession.UTMContent = &entity.NullableString{IsNull: false, String: *voucher.SetUTMContent}
							}

							// log.Printf("last timeline session %+v\n", timeline[len(timeline)-1])
						}
					}
				}
			}
		}
	}

	// filter postviews that we put in the conversion funnel according to the postview attribution
	postviewsInFunnel := []*entity.Postview{}

	// by default only the first impression is considered, we might introduce other models in the future...
	if len(orderPostviews) > 0 {
		postviewsInFunnel = append(postviewsInFunnel, orderPostviews[0])
	}

	// only keep attributed touchpoints
	touchpointsAttributed := []Touchpoint{}

	for _, hit := range touchpoints {
		// log.Printf("touchpoint %+v\n", touchpoint)

		// check if this impression is allowed in the conversion funnel
		if hit.Postview != nil {
			hitAllowed := false
			for _, imp := range postviewsInFunnel {
				if imp.ID == hit.Postview.ID {
					hitAllowed = true
				}
			}

			// impression not attributed, ignore it
			if !hitAllowed {
				continue
			}
		}

		touchpointsAttributed = append(touchpointsAttributed, hit)
	}

	// compute funnel roles
	for index := range touchpointsAttributed {

		var role int64 = 1 // 1: initiator

		if index == 0 {
			if totalTouchpoints == 1 {
				role = 0 // 0: alone in the funnel
			}
		} else {
			if totalTouchpoints == 2 || totalTouchpoints == index+1 {
				role = 3 // 3: closer
			} else {
				role = 2 // 2: assistant
			}
		}

		if touchpointsAttributed[index].Session != nil {
			touchpointsAttributed[index].Session.Role = &role
		}
		if touchpointsAttributed[index].Postview != nil {
			touchpointsAttributed[index].Postview.Role = &role
		}
	}

	// log.Printf("impressionsInFunnel %+v\n", impressionsInFunnel)

	// funnel sums up the conversion path
	devicesFunnel := []string{}
	devicesTypeMap := map[string]string{}
	domainsFunnel := []string{}
	domainsTypeFunnel := []string{}
	domainsTypeMap := map[string]string{}

	funnel := entity.ConversionFunnel{}

	for index, hit := range touchpointsAttributed {

		var deviceID *string
		var domainID *string

		var currentRole int64
		var currentChannelID string
		var currentChannelGroupID string
		var currentChannelOriginID string

		if hit.Session != nil {
			deviceID = hit.Session.DeviceID
			domainID = &hit.Session.DomainID
			currentChannelOriginID = hit.Session.ChannelOriginID
			currentRole = *hit.Session.Role
			currentChannelID = hit.Session.ChannelID
			currentChannelGroupID = hit.Session.ChannelGroupID
		}

		if hit.Postview != nil {
			deviceID = hit.Postview.DeviceID
			currentChannelOriginID = hit.Postview.ChannelOriginID
			currentRole = *hit.Postview.Role
			currentChannelID = hit.Postview.ChannelID
			currentChannelGroupID = hit.Postview.ChannelGroupID
		}

		// devices funnel
		if deviceID != nil {

			for _, device := range devices {

				// if device found
				if device.ID == *deviceID && device.DeviceType != nil && device.DeviceType.String != "" {

					deviceType := device.DeviceType.String

					// add device to funnel if the last device is different
					if len(devicesFunnel) == 0 || devicesFunnel[len(devicesFunnel)-1] != deviceType {
						devicesFunnel = append(devicesFunnel, deviceType) // = Computer , Tablet...

						// map of unique devices type
						if _, ok := devicesTypeMap[deviceType]; !ok {
							devicesTypeMap[deviceType] = deviceType
						}
					}
				}
			}
		}

		// domains funnel
		if domainID != nil {
			// find matching domain
			for _, dom := range pipe.GetWorkspace().Domains {
				if dom.DeletedAt == nil {

					// add domain to funnel if the last domain ID is different
					if len(domainsFunnel) == 0 || domainsFunnel[len(domainsFunnel)-1] != *domainID {

						if dom.ID == *domainID {
							domainsFunnel = append(domainsFunnel, *domainID)
							domainsTypeFunnel = append(domainsTypeFunnel, dom.Type)

							// map of unique domains type
							if _, ok := domainsTypeMap[dom.Type]; !ok {
								domainsTypeMap[dom.Type] = dom.Type
							}
						}
					}
				}
			}
		}

		// build funnel record

		currentFunnelLength := len(funnel)

		if currentFunnelLength == 0 {
			event := &entity.ConversionTouchpoint{
				Position:        index + 1,
				Role:            int(currentRole),
				Count:           1,
				ChannelID:       currentChannelID,
				ChannelGroupID:  currentChannelGroupID,
				ChannelOriginID: currentChannelOriginID,
			}
			if hit.Postview != nil {
				event.Postview = true
			}
			funnel = append(funnel, event)
		} else {
			// if its the same source/medium, increment
			if lastItem := funnel[currentFunnelLength-1]; lastItem.ChannelOriginID == currentChannelOriginID {
				funnel[currentFunnelLength-1].Count += 1

				// update role if its closing
				if lastItem.Role < int(currentRole) {
					funnel[currentFunnelLength-1].Role = int(currentRole)
				}
			} else {
				// add touchpoint event to the funnel
				event := &entity.ConversionTouchpoint{
					Position:        index + 1,
					Role:            int(currentRole),
					Count:           1,
					ChannelID:       currentChannelID,
					ChannelGroupID:  currentChannelGroupID,
					ChannelOriginID: currentChannelOriginID,
				}
				if hit.Postview != nil {
					event.Postview = true
				}
				funnel = append(funnel, event)
			}
		}
	}

	// if no touchpoint in funnel, add the device that generated the conversion
	if len(devicesFunnel) == 0 && order.SessionID != nil {

		// find the session that generated the conversion
		for _, hit := range touchpointsAttributed {
			if hit.Session != nil && hit.Session.ID == *order.SessionID && hit.Session.DeviceID != nil {
				// find the device that generated the conversion
				for _, device := range devices {
					if device.ID == *hit.Session.DeviceID && device.DeviceType != nil && device.DeviceType.String != "" {
						devicesFunnel = append(devicesFunnel, device.DeviceType.String)
						devicesTypeMap[device.DeviceType.String] = device.DeviceType.String
					}
				}
			}
		}
	}

	// if no touchpoint in funnel, add the domain that generated the conversion
	if len(domainsFunnel) == 0 {
		for _, dom := range pipe.GetWorkspace().Domains {
			if dom.DeletedAt == nil {
				// if domain found
				if dom.ID == order.DomainID {
					domainsFunnel = append(domainsFunnel, order.DomainID)
					domainsTypeFunnel = append(domainsTypeFunnel, dom.Type)
					domainsTypeMap[dom.Type] = dom.Type
				}
			}
		}
	}

	order.SetDevicesFunnel(strings.Join(devicesFunnel, "~"))
	order.SetDevicesTypeCount(int64(len(devicesTypeMap)))

	order.SetDomainsFunnel(strings.Join(domainsFunnel, "~"))
	order.SetDomainsTypeFunnel(strings.Join(domainsTypeFunnel, "~"))
	order.SetDomainsCount(int64(len(domainsTypeMap)))

	// log.Printf("order.DevicesFunnel = %v", order.DevicesFunnel)
	// log.Printf("order.DevicesTypeCount = %v", order.DevicesTypeCount)
	// log.Printf("order.DomainsFunnel = %v", order.DomainsFunnel)
	// log.Printf("order.DomainsCount = %v", order.DomainsCount)

	// compute funnel hash key for conversion paths aggregation

	funnelHashKey := ""

	for i, f := range funnel {
		funnelHashKey = fmt.Sprintf("%v%v%v%v", funnelHashKey, i, f.ChannelOriginID, f.Count)
		// fmt.Println(funnelHashKey)
	}

	now := time.Now()
	order.SetFunnel(funnel)
	order.SetFunnelHash(fmt.Sprintf("%x", sha1.Sum([]byte(funnelHashKey))))
	order.SetTimeToConversion(&orderTimeToConversion)
	order.SetAttributionUpdatedAt(&now)

	// log.Printf("funnel %+v\n", funnel)

	// update order attribution data
	if err = pipe.Repo().UpdateOrderAttribution(spanCtx, order, tx); err != nil {
		return
	}

	// update hits

	totalTouchpointsAttributed := len(touchpointsAttributed)

	var attributionLinearAmount int64

	if order.SubtotalPrice != nil && order.SubtotalPrice.Int64 > 0 && totalTouchpointsAttributed > 0 {
		attributionLinearAmount = order.SubtotalPrice.Int64 / int64(totalTouchpointsAttributed)
	} else {
		attributionLinearAmount = 0
	}

	var attributionLinearPercentage int64

	if totalTouchpointsAttributed > 0 {
		attributionLinearPercentage = int64(math.Floor(10000 / float64(totalTouchpointsAttributed)))
	} else {
		attributionLinearPercentage = 0
	}

	if len(orderSessions) > 0 {
		// reset eventual hits already attributed
		if err = pipe.Repo().ResetSessionsAttributedForConversion(spanCtx, order.UserID, order.ID, tx); err != nil {
			return
		}
	}

	if len(orderPostviews) > 0 {
		// reset eventual hits already attributed
		if err = pipe.Repo().ResetPostviewsAttributedForConversion(spanCtx, order.UserID, order.ID, tx); err != nil {
			return
		}
	}

	for _, hit := range touchpointsAttributed {

		// update hits
		if hit.Session != nil {

			hit.Session.SetConversionType(entity.StringPtr("order"))
			hit.Session.SetConversionID(&order.ID)
			hit.Session.SetConversionExternalID(&order.ExternalID)
			hit.Session.SetConversionAt(&order.CreatedAt)
			hit.Session.SetConversionAmount(nil)
			if order.SubtotalPrice != nil {
				hit.Session.SetConversionAmount(&order.SubtotalPrice.Int64)
			}
			hit.Session.SetLinearAmountAttributed(&attributionLinearAmount)
			hit.Session.SetLinearPercentageAttributed(&attributionLinearPercentage)
			hit.Session.SetTimeToConversion(&orderTimeToConversion)
			hit.Session.SetIsFirstConversion(&order.IsFirstConversion)

			// update hit attribution
			if err = pipe.Repo().UpdateSession(spanCtx, hit.Session, tx); err != nil {
				return
			}
		}

		if hit.Postview != nil {

			hit.Postview.SetConversionType(entity.StringPtr("order"))
			hit.Postview.SetConversionID(&order.ID)
			hit.Postview.SetConversionExternalID(&order.ExternalID)
			hit.Postview.SetConversionAt(&order.CreatedAt)
			hit.Postview.SetConversionAmount(nil)
			if order.SubtotalPrice != nil {
				hit.Postview.SetConversionAmount(&order.SubtotalPrice.Int64)
			}
			hit.Postview.SetLinearAmountAttributed(&attributionLinearAmount)
			hit.Postview.SetLinearPercentageAttributed(&attributionLinearPercentage)
			hit.Postview.SetTimeToConversion(&orderTimeToConversion)
			hit.Postview.SetIsFirstConversion(&order.IsFirstConversion)

			// update hit attribution
			if err = pipe.Repo().UpdatePostview(spanCtx, hit.Postview, tx); err != nil {
				return
			}
		}
	}

	// TODO: generate data_logs for affected touchpoints ?

	// log.Printf("attribute one order took: %v", time.Now().Sub(attributeStart))

	return
}
