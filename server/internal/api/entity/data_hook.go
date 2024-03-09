package entity

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"strings"

	"github.com/rotisserie/eris"
)

var (
	DataHookKindOnValidation = "on_validation" // before validation
	DataHookKindOnSuccess    = "on_success"
	AppIDSystem              = "system"
)

type DataHook struct {
	ID          string   `json:"id"`
	AppID       string   `json:"app_id"` // system or app_id_xxx
	Name        string   `json:"name"`
	On          string   `json:"on"`           // on_validation | on_success
	Kind        []string `json:"kind"`         // data log kind
	Action      []string `json:"action"`       // data log action
	JS          *string  `json:"js,omitempty"` // javascript code to execute, otherwise app endpoint
	Enabled     bool     `json:"enabled"`      // is enabled
	DBCreatedAt string   `json:"db_created_at"`
	DBUpdatedAt string   `json:"db_updated_at"`
}

// is this data hook for the given event kind
func (x *DataHook) MatchesDataLog(dataLogKind string, dataLogAction string) bool {

	kindMatches := false

	for _, kind := range x.Kind {
		if kind == dataLogKind || kind == "*" {
			kindMatches = true
			break
		}
	}

	if !kindMatches {
		return false
	}

	actionMatches := false

	for _, action := range x.Action {
		if action == dataLogAction || action == "*" {
			actionMatches = true
			break
		}
	}

	if !actionMatches {
		return false
	}

	return true
}

// validate data hook
func (x *DataHook) Validate(installedApps InstalledApps) error {

	if x.ID == "" {
		return eris.New("invalid data hook id")
	}

	if x.AppID == "" {
		return eris.New("invalid data hook app id")
	}

	// id should start with app_ or system
	if !strings.HasPrefix(x.ID+"_", x.AppID) && x.AppID != AppIDSystem {
		return eris.New("invalid data hook id, should start with app_id_")
	}

	// verify that app exists
	if x.ID != AppIDSystem {
		appFound := false
		for _, app := range installedApps {
			if app.ID == x.AppID {
				appFound = true
				break
			}
		}
		if !appFound {
			return eris.Errorf("invalid data hook app id: %s", x.AppID)
		}
	}

	if x.On != DataHookKindOnValidation && x.On != DataHookKindOnSuccess {
		return eris.Errorf("invalid data hook on: %s", x.On)
	}

	if len(x.Kind) == 0 {
		return eris.New("invalid data hook kind")
	}

	if len(x.Action) == 0 {
		return eris.New("invalid data hook action")
	}

	if x.Name == "" {
		return eris.New("invalid data hook name")
	}

	if x.AppID == AppIDSystem && (x.JS == nil || *x.JS == "") {
		return eris.New("data hook js is required")
	}

	allowedKind := []string{
		"*",
		"user",
		"session",
		"order",
		"order_item",
		"pageview",
		"cart",
		"cart_item",
		"postview",
		"app_",  // all app events
		"appx_", // all app events
	}

	for _, event := range x.Kind {
		found := false
		for _, allowedEvent := range allowedKind {
			if allowedEvent == "*" {
				found = true
				break
			}

			if event == allowedEvent {
				found = true
				break
			}

			if (allowedEvent == "app_" || allowedEvent == "appx_") && strings.HasPrefix(event, allowedEvent) {
				found = true
				break
			}
		}
		if !found {
			return eris.Errorf("invalid data hook event: %s", event)
		}
	}

	allowedAction := []string{
		"*",
		"create",
		"update",
		"noop",
		"enter",
		"exit",
	}

	for _, action := range x.Action {
		found := false
		for _, allowedAction := range allowedAction {
			if allowedAction == "*" {
				found = true
				break
			}

			if action == allowedAction {
				found = true
				break
			}
		}
		if !found {
			return eris.Errorf("invalid data hook action: %s", action)
		}
	}

	return nil
}

type DataHooks []*DataHook

func (x *DataHooks) Scan(val interface{}) error {

	var data []byte

	if b, ok := val.([]byte); ok {
		// VERY IMPORTANT: we need to clone the bytes here
		// The sql driver will reuse the same bytes RAM slots for future queries
		// Thank you St Antoine De Padoue for helping me find this bug
		data = bytes.Clone(b)
	} else if s, ok := val.(string); ok {
		data = []byte(s)
	} else if val == nil {
		return nil
	}

	return json.Unmarshal(data, x)
}

func (x DataHooks) Value() (driver.Value, error) {
	return json.Marshal(x)
}

type DataHooksState map[string]*DataHookState

type DataHookState struct {
	Done       bool   `json:"done"`
	IsError    bool   `json:"err,omitempty"`
	Message    string `json:"msg,omitempty"`
	TriedCount int    `json:"tried,omitempty"`
}

func (x *DataHooksState) Scan(val interface{}) error {

	var data []byte

	if b, ok := val.([]byte); ok {
		// VERY IMPORTANT: we need to clone the bytes here
		// The sql driver will reuse the same bytes RAM slots for future queries
		// Thank you St Antoine De Padoue for helping me find this bug
		data = bytes.Clone(b)
	} else if s, ok := val.(string); ok {
		data = []byte(s)
	} else if val == nil {
		return nil
	}

	return json.Unmarshal(data, x)
}

func (x DataHooksState) Value() (driver.Value, error) {
	return json.Marshal(x)
}
