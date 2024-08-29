package dto

import (
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/rotisserie/eris"
)

type DBSelectParams struct {
	WorkspaceID string        `json:"workspace_id"`
	From        string        `json:"from"`
	Columns     []string      `json:"columns"`
	Where       string        `json:"where"`
	Args        []interface{} `json:"args"`
	OrderBy     string        `json:"order_by"`
	GroupBy     string        `json:"group_by"`
	Offset      int64         `json:"offset"`
	Limit       int64         `json:"limit"`
}

type DBSelectWhere struct {
	Predicate string        `json:"predicate"`
	Args      []interface{} `json:"args"`
}

var (
	DBSelectFroms = []string{
		"user",
		"order",
		"order_item",
		"cart",
		"cart_item",
		"custom_event",
		"device",
		"session",
		"pageview",
		"postview",
	}
)

func (params *DBSelectParams) Validate() (err error) {

	if params.WorkspaceID == "" {
		return eris.Errorf("missing workspace_id")
	}

	if params.From == "" {
		return eris.Errorf("missing from")
	}
	if !govalidator.IsIn(params.From, DBSelectFroms...) && !strings.HasPrefix(params.From, "app_") && !strings.HasPrefix(params.From, "appx_") {
		return eris.Errorf("invalid from: %s", params.From)
	}

	if len(params.Columns) == 0 {
		return eris.Errorf("missing columns")
	}

	// default limit
	if params.Limit == 0 {
		params.Limit = 50
	}
	if params.Limit < 1 || params.Limit > 100 {
		return eris.Errorf("invalid limit: %d", params.Limit)
	}

	return nil
}
