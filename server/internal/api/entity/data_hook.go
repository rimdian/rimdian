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

type DataHookFor struct {
	Kind   string `json:"kind"`
	Action string `json:"action"`
}

type DataHook struct {
	ID    string         `json:"id"`
	AppID string         `json:"app_id"` // system or app_id_xxx
	Name  string         `json:"name"`
	On    string         `json:"on"`            // on_validation | on_success
	For   []*DataHookFor `json:"for,omitempty"` // for which data log
	// Kind        []string `json:"kind"`         // data log kind
	// Action      []string `json:"action"`       // data log action
	JS          *string `json:"js,omitempty"` // javascript code to execute, otherwise app endpoint
	Enabled     bool    `json:"enabled"`      // is enabled
	DBCreatedAt string  `json:"db_created_at"`
	DBUpdatedAt string  `json:"db_updated_at"`
}

// is this data hook for the given event kind
func (x *DataHook) MatchesDataLog(dataLogKind string, dataLogAction string) bool {

	if x.For == nil {
		return false
	}

	for _, forKind := range x.For {
		if forKind.Kind == "*" || forKind.Kind == dataLogKind {
			if forKind.Action == "*" || forKind.Action == dataLogAction {
				return true
			}
		}
	}

	return false
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

	if x.For == nil || len(x.For) == 0 {
		return eris.New("data hook for is required")
	}

	if x.Name == "" {
		return eris.New("invalid data hook name")
	}

	if x.AppID == AppIDSystem && (x.JS == nil || *x.JS == "") {
		return eris.New("data hook js is required")
	}

	allowedKinds := []string{
		"*",
		"user",
		"session",
		"order",
		"order_item",
		"pageview",
		"cart",
		"cart_item",
		"postview",
	}

	allowedActions := []string{
		"*",
		"create",
		"update",
		"noop",
		"enter",
		"exit",
	}

	for _, forItem := range x.For {
		if forItem.Kind == "" {
			return eris.New("invalid data hook for kind")
		}

		if forItem.Action == "" {
			return eris.New("invalid data hook for action")
		}

		for _, allowedKind := range allowedKinds {
			if allowedKind == forItem.Kind || forItem.Kind == "*" {
				break
			}

			if strings.HasPrefix(forItem.Kind, "app_") || strings.HasPrefix(forItem.Kind, "appx_") {
				break
			}

			return eris.Errorf("invalid data hook for kind: %s", forItem.Kind)
		}

		for _, allowedAction := range allowedActions {
			if allowedAction == forItem.Action || forItem.Action == "*" {
				break
			}

			return eris.Errorf("invalid data hook for action: %s", forItem.Action)
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
