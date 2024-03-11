package service

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rimdian/rimdian/internal/common/dto"
	"github.com/tidwall/gjson"
)

// extract the data from the data_pipe.Logger.item JSON and validate it
func (pipe *DataLogPipeline) ExtractAndValidateItem() {

	// verify Origin is allowed in workspace a web domain, for client-side batches only
	// and set the matching domain_id to items
	if pipe.DataLogInQueue.Origin == dto.DataLogOriginClient {

		host, ok := pipe.DataLogInQueue.Context.HeadersAndParams["Origin"]

		if !ok {
			// grab referer
			host, ok = pipe.DataLogInQueue.Context.HeadersAndParams["Referer"]

			if !ok {
				// report error but continue to persist the data_log in DB
				pipe.SetError("context.headersAndParams.Origin", "Origin and Referer not found in headers and params", false)
				return
			}
		}

		// parse host
		result, err := url.Parse(host)

		if err != nil {
			// report error but continue to persist the data_log in DB
			pipe.SetError("context.headersAndParams.Origin", fmt.Sprintf("doDataLog: error parsing Origin: %v", err), false)
			return
		}

		hostIsValid := false

		for _, dom := range pipe.Workspace.Domains {
			if dom.Type == entity.DomainWeb {
				for _, domainHost := range dom.Hosts {
					if domainHost.Host == result.Host {
						hostIsValid = true
						// enrich data_log with domain_id, that will be used to enrich items (session, pageview, etc.)
						pipe.DataLog.DomainID = &dom.ID
					}
				}
			}
		}

		if !hostIsValid {
			// report error but continue to persist the data_log in DB
			pipe.SetError("context.host", fmt.Sprintf("doDataLog: host not allowed: %v", result.Host), false)
			return
		}
	}

	// verify that item is a valid JSON
	if !gjson.Valid(pipe.DataLog.Item) {
		pipe.Logger.Printf("doDataLog: item is not a valid JSON: %v", pipe.DataLog.Item)
		// replace item with a valid JSON
		replace := struct {
			Item string `json:"invalid_json"`
		}{pipe.DataLog.Item}

		jsonString, err := json.Marshal(replace)
		if err != nil {
			pipe.DataLog.Item = `{"invalid_json": true}`
			pipe.SetError("item", fmt.Sprintf("error replacing invalid json: %v", err), false)
			return
		}

		pipe.DataLog.Item = string(jsonString)
		pipe.SetError("item", "item is not a valid JSON", false)
		return
	}

	switch pipe.DataLog.Kind {
	case "user":
		pipe.ExtractUserFromDataLogItem()
	case "user_alias":
		pipe.ExtractUserAliasFromDataLogItem()
	case "device":
		// device requires a user
		pipe.ExtractUserFromDataLogItem()
		if pipe.HasError() {
			return
		}
		pipe.ExtractDeviceFromDataLogItem(pipe.DataLog.UpsertedUser.ID)
	case "pageview":
		// pageview requires a user
		pipe.ExtractUserFromDataLogItem()
		if pipe.HasError() {
			return
		}
		// pageview requires a session
		pipe.ExtractSessionFromDataLogItem(true)
		if pipe.HasError() {
			return
		}
		pipe.ExtractPageviewFromDataLogItem()
	case "order":
		// order requires a user
		pipe.ExtractUserFromDataLogItem()
		if pipe.HasError() {
			return
		}
		// order requires a session
		pipe.ExtractSessionFromDataLogItem(false)
		if pipe.HasError() {
			return
		}
		pipe.ExtractOrderFromDataLogItem()
	case "cart":
		// cart requires a user
		pipe.ExtractUserFromDataLogItem()
		if pipe.HasError() {
			return
		}
		// cart requires a session
		pipe.ExtractSessionFromDataLogItem(false)
		if pipe.HasError() {
			return
		}
		pipe.ExtractCartFromDataLogItem()
	case "session":
		// session requires a user
		pipe.ExtractUserFromDataLogItem()
		if pipe.HasError() {
			return
		}
		pipe.ExtractSessionFromDataLogItem(true)
	case "postview":
		// postview requires a user
		pipe.ExtractUserFromDataLogItem()
		if pipe.HasError() {
			return
		}
		pipe.ExtractPostviewFromDataLogItem()
	case "custom_event":
		// custom_event requires a user
		pipe.ExtractUserFromDataLogItem()
		if pipe.HasError() {
			return
		}
		// custom_event requires a session
		pipe.ExtractSessionFromDataLogItem(false)
		if pipe.HasError() {
			return
		}
		pipe.ExtractCustomEventFromDataLogItem()
	default:
		pipe.ExtractAppItemFromDataLogItem()
	}
}

func (pipe *DataLogPipeline) ExtractUserFromDataLogItem() {

	var err error
	user, err := entity.NewUserFromDataLog(pipe.DataLog, pipe.DataLogInQueue.Context.ClockDifference, pipe.Workspace, pipe.DataLogInQueue.Origin, pipe.Config.SECRET_KEY)

	if err != nil {
		pipe.SetError("user", err.Error(), false)
		return
	}

	pipe.DataLog.UpsertedUser = user

	// add user_id to the list of users to lock
	pipe.UsersLock.AddUser(pipe.DataLog.UpsertedUser.ID)

	pipe.DataLog.UserID = pipe.DataLog.UpsertedUser.ID

	if pipe.DataLog.Kind == "user" {
		pipe.DataLog.ItemID = pipe.DataLog.UpsertedUser.ID
		pipe.DataLog.ItemExternalID = pipe.DataLog.UpsertedUser.ExternalID
		pipe.DataLog.EventAt = *pipe.DataLog.UpsertedUser.UpdatedAt
		pipe.DataLog.EventAtTrunc = pipe.DataLog.EventAt.Truncate(time.Hour)
	}
}

func (pipe *DataLogPipeline) ExtractCustomEventFromDataLogItem() {

	var err error
	customEvent, err := entity.NewCustomEventFromDataLog(pipe.DataLog, pipe.DataLogInQueue.Context.ClockDifference, pipe.Workspace)

	if err != nil {
		pipe.SetError("customEvent", err.Error(), false)
		return
	}

	pipe.DataLog.UpsertedCustomEvent = customEvent

	if pipe.DataLog.Kind == "custom_event" {
		pipe.DataLog.ItemID = pipe.DataLog.UpsertedCustomEvent.ID
		pipe.DataLog.ItemExternalID = pipe.DataLog.UpsertedCustomEvent.ExternalID
		pipe.DataLog.EventAt = *pipe.DataLog.UpsertedCustomEvent.UpdatedAt
		pipe.DataLog.EventAtTrunc = pipe.DataLog.UpsertedCustomEvent.UpdatedAt.Truncate(time.Hour)
	}
}

func (pipe *DataLogPipeline) ExtractPageviewFromDataLogItem() {

	var err error
	pageview, err := entity.NewPageviewFromDataLog(pipe.DataLog, pipe.DataLogInQueue.Context.ClockDifference, pipe.Workspace)

	if err != nil {
		pipe.SetError("pageview", err.Error(), false)
		return
	}

	pipe.DataLog.UpsertedPageview = pageview

	if pipe.DataLog.Kind == "pageview" {
		pipe.DataLog.ItemID = pipe.DataLog.UpsertedPageview.ID
		pipe.DataLog.ItemExternalID = pipe.DataLog.UpsertedPageview.ExternalID
		pipe.DataLog.EventAt = *pipe.DataLog.UpsertedPageview.UpdatedAt
		pipe.DataLog.EventAtTrunc = pipe.DataLog.UpsertedPageview.UpdatedAt.Truncate(time.Hour)
	}
}

func (pipe *DataLogPipeline) ExtractOrderFromDataLogItem() {

	var err error
	order, err := entity.NewOrderFromDataLog(pipe.DataLog, pipe.DataLogInQueue.Context.ClockDifference, pipe.Workspace)

	if err != nil {
		pipe.SetError("order", err.Error(), false)
		return
	}

	pipe.DataLog.UpsertedOrder = order

	if pipe.DataLog.Kind == "order" {
		pipe.DataLog.ItemID = pipe.DataLog.UpsertedOrder.ID
		pipe.DataLog.ItemExternalID = pipe.DataLog.UpsertedOrder.ExternalID
		pipe.DataLog.EventAt = *pipe.DataLog.UpsertedOrder.UpdatedAt
		pipe.DataLog.EventAtTrunc = pipe.DataLog.UpsertedOrder.UpdatedAt.Truncate(time.Hour)
	}
}

func (pipe *DataLogPipeline) ExtractCartFromDataLogItem() {

	var err error
	cart, err := entity.NewCartFromDataLog(pipe.DataLog, pipe.DataLogInQueue.Context.ClockDifference, pipe.Workspace)

	if err != nil {
		pipe.SetError("cart", err.Error(), false)
		return
	}

	pipe.DataLog.UpsertedCart = cart

	if pipe.DataLog.Kind == "cart" {
		pipe.DataLog.ItemID = pipe.DataLog.UpsertedCart.ID
		pipe.DataLog.ItemExternalID = pipe.DataLog.UpsertedCart.ExternalID
		pipe.DataLog.EventAt = *pipe.DataLog.UpsertedCart.UpdatedAt
		pipe.DataLog.EventAtTrunc = pipe.DataLog.UpsertedCart.UpdatedAt.Truncate(time.Hour)
	}
}

func (pipe *DataLogPipeline) ExtractUserAliasFromDataLogItem() {

	var err error
	userAlias, err := entity.NewUserAliasFromDataLog(pipe.DataLog, pipe.DataLogInQueue.Context.ClockDifference, pipe.DataLogInQueue.Context.DataSentAt, pipe.DataLogInQueue.Context.ReceivedAt)

	if err != nil {
		pipe.SetError("user_alias", err.Error(), false)
		return
	}

	pipe.DataLog.UserAlias = userAlias

	result := gjson.Get(pipe.DataLog.Item, "user_alias")
	if !result.Exists() {
		pipe.SetError("user_alias", "doDataLog: item has no user_alias object", false)
		return
	}

	toUserID := entity.ComputeUserID(pipe.DataLog.UserAlias.ToUserExternalID)

	pipe.DataLog.Action = "create"
	pipe.DataLog.UserID = toUserID
	pipe.DataLog.ItemID = toUserID
	pipe.DataLog.ItemExternalID = pipe.DataLog.UserAlias.ToUserExternalID
	pipe.DataLog.EventAt = *pipe.DataLog.UserAlias.ToUserCreatedAt
	pipe.DataLog.EventAtTrunc = pipe.DataLog.UserAlias.ToUserCreatedAt.Truncate(time.Hour)

	// add user_id to the list of users to lock
	pipe.UsersLock.AddUser(entity.ComputeUserID(pipe.DataLog.UserAlias.FromUserExternalID))
	pipe.UsersLock.AddUser(toUserID)
}

// so far the device object is provided together with a pageview
func (pipe *DataLogPipeline) ExtractDeviceFromDataLogItem(userID string) {

	var err error
	device, err := entity.NewDeviceFromDataLog(pipe.DataLog, pipe.DataLogInQueue.Context.ClockDifference, pipe.Workspace)

	if err != nil {
		pipe.SetError("device", err.Error(), false)
		return
	}

	pipe.DataLog.UpsertedDevice = device

	// extract browser/os/platform/type from user agent
	if pipe.DataLog.UpsertedDevice.ShouldParseUserAgent() {
		result, errUA := pipe.ParseUserAgent(pipe.DataLog.UpsertedDevice.UserAgent.String)
		if errUA != nil {
			pipe.SetError("user_agent", errUA.Error(), true)
			return
		}
		pipe.DataLog.UpsertedDevice.ProcessUserAgent(result)
	}

	if pipe.DataLog.Kind == "device" {
		pipe.DataLog.ItemID = pipe.DataLog.UpsertedDevice.ID
		pipe.DataLog.ItemExternalID = pipe.DataLog.UpsertedDevice.ExternalID
		pipe.DataLog.EventAt = *pipe.DataLog.UpsertedDevice.UpdatedAt
		pipe.DataLog.EventAtTrunc = pipe.DataLog.UpsertedDevice.UpdatedAt.Truncate(time.Hour)
	}
}

// so far the session object is provided together with an event (pageview,cart,order,custom_event,app item, etc.)
// the session can have a device attached to it
func (pipe *DataLogPipeline) ExtractSessionFromDataLogItem(isMandatory bool) {

	// extract device before session
	hasDevice := gjson.Get(pipe.DataLog.Item, "device")
	if hasDevice.Exists() {
		pipe.ExtractDeviceFromDataLogItem(pipe.DataLog.UpsertedUser.ID)
	}

	hasSession := gjson.Get(pipe.DataLog.Item, "session")
	if hasSession.Exists() {
		var err error
		session, err := entity.NewSessionFromDataLog(pipe.DataLog, pipe.DataLogInQueue.Context.ClockDifference, pipe.Workspace)

		if err != nil {
			pipe.SetError("session", err.Error(), false)
			return
		}

		pipe.DataLog.UpsertedSession = session
	}

	if isMandatory && pipe.DataLog.UpsertedSession == nil {
		pipe.SetError("session", "item has no session object", false)
		return
	}

	if pipe.DataLog.Kind == "session" {
		pipe.DataLog.ItemID = pipe.DataLog.UpsertedSession.ID
		pipe.DataLog.ItemExternalID = pipe.DataLog.UpsertedSession.ExternalID
		pipe.DataLog.EventAt = *pipe.DataLog.UpsertedSession.UpdatedAt
		pipe.DataLog.EventAtTrunc = pipe.DataLog.UpsertedSession.UpdatedAt.Truncate(time.Hour)
	}
}

func (pipe *DataLogPipeline) ExtractPostviewFromDataLogItem() {

	// extract device before postview
	hasDevice := gjson.Get(pipe.DataLog.Item, "device")
	if hasDevice.Exists() {
		pipe.ExtractDeviceFromDataLogItem(pipe.DataLog.UpsertedUser.ID)
	}

	var err error
	postview, err := entity.NewPostviewFromDataLog(pipe.DataLog, pipe.DataLogInQueue.Context.ClockDifference, pipe.Workspace)

	if err != nil {
		pipe.SetError("postview", err.Error(), false)
		return
	}

	pipe.DataLog.UpsertedPostview = postview

	if pipe.DataLog.Kind == "postview" {
		pipe.DataLog.ItemID = pipe.DataLog.UpsertedPostview.ID
		pipe.DataLog.ItemExternalID = pipe.DataLog.UpsertedPostview.ExternalID
		pipe.DataLog.EventAt = *pipe.DataLog.UpsertedPostview.UpdatedAt
		pipe.DataLog.EventAtTrunc = pipe.DataLog.UpsertedPostview.UpdatedAt.Truncate(time.Hour)
	}
}

func (pipe *DataLogPipeline) ExtractAppItemFromDataLogItem() {

	// check if kind starts with "app_"
	if !strings.HasPrefix(pipe.DataLog.Kind, "app_") && !strings.HasPrefix(pipe.DataLog.Kind, "appx_") {
		pipe.SetError("kind", fmt.Sprintf("item kind not supported: %v", pipe.DataLog.Kind), false)
		return
	}

	// extract eventual user
	hasUser := gjson.Get(pipe.DataLog.Item, "user")
	if hasUser.Exists() {
		pipe.ExtractUserFromDataLogItem()
	}

	var err error
	appItem, err := entity.NewAppItemFromDataLog(pipe.DataLog, pipe.DataLogInQueue.Context.ClockDifference, pipe.Workspace)

	if err != nil {
		pipe.SetError("appItem", err.Error(), false)
		return
	}

	pipe.DataLog.UpsertedAppItem = appItem

	// is an app item
	if strings.HasPrefix(pipe.DataLog.Kind, "app_") || strings.HasPrefix(pipe.DataLog.Kind, "appx_") {
		if pipe.DataLog.UpsertedUser != nil {
			pipe.DataLog.UserID = pipe.DataLog.UpsertedUser.ID
		} else {
			pipe.DataLog.UserID = entity.None
		}
		pipe.DataLog.ItemID = pipe.DataLog.UpsertedAppItem.ID
		pipe.DataLog.ItemExternalID = pipe.DataLog.UpsertedAppItem.ExternalID
		pipe.DataLog.EventAt = *pipe.DataLog.UpsertedAppItem.UpdatedAt
		pipe.DataLog.EventAtTrunc = pipe.DataLog.UpsertedAppItem.UpdatedAt.Truncate(time.Hour)
	}
}

// extract extra columns app_appname_fields from the JSON batch items
// and add them into the ExtraColumns map of concerned items
// if values are nullable the values are NullString/NullBool...
// if values dont match the custom dimension type, an error is set and the field is not written
func (pipe *DataLogPipeline) ExtractExtraColumnsFromItem(kind string) {

	// for each configured extra column, find if it exists in items or context
	for _, app := range pipe.Workspace.InstalledApps {
		if app.ExtraColumns != nil && len(app.ExtraColumns) == 0 {
			continue
		}
		for _, augTable := range app.ExtraColumns {

			if augTable.Kind != kind {
				continue
			}

			switch augTable.Kind {
			case "user":
				// extra columns are not implemented yet for user
			case "postview":
				if pipe.DataLog.UpsertedPostview != nil {
					// init map
					if pipe.DataLog.UpsertedPostview.ExtraColumns == nil {
						pipe.DataLog.UpsertedPostview.ExtraColumns = entity.AppItemFields{}
					}

					for _, col := range augTable.Columns {

						if result := gjson.Get(pipe.DataLog.Item, "postview."+col.Name); result.Exists() {

							fieldValue, err := entity.ExtractFieldValueFromGJSON(col, result, 0)
							if err != nil {
								pipe.DataLog.Errors[col.Name] = err.Error()
							} else {
								pipe.DataLog.UpsertedPostview.ExtraColumns[col.Name] = fieldValue
							}
						}
					}
				}
			case "pageview":
			case "custom_event":
				if pipe.DataLog.UpsertedCustomEvent != nil {
					// init map
					if pipe.DataLog.UpsertedCustomEvent.ExtraColumns == nil {
						pipe.DataLog.UpsertedCustomEvent.ExtraColumns = entity.AppItemFields{}
					}

					for _, col := range augTable.Columns {

						if result := gjson.Get(pipe.DataLog.Item, "custom_event."+col.Name); result.Exists() {

							fieldValue, err := entity.ExtractFieldValueFromGJSON(col, result, 0)
							if err != nil {
								pipe.DataLog.Errors[col.Name] = err.Error()
							} else {
								pipe.DataLog.UpsertedCustomEvent.ExtraColumns[col.Name] = fieldValue
							}
						}
					}
				}
			default:
				// do nothing
			}
		}
	}
}
