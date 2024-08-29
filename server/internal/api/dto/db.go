package dto

import (
	"encoding/json"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/rotisserie/eris"
)

// type DBAnalyticsQuery struct {
// 	Measures       []string      `json:"measures"`
// 	Dimensions     []string      `json:"dimensions"`
// 	Filters        []interface{} `json:"filters"`
// 	TimeDimensions []interface{} `json:"time_dimensions"`
// 	Segments       []string      `json:"segments"`
// 	Limit          *int64        `json:"limit,omitempty"`
// 	Offset         *int64        `json:"offset,omitempty"`
// 	Order          interface{}   `json:"order,omitempty"`
// 	Timezone       string        `json:"timezone"`
// 	RenewQuery     *bool         `json:"renew_query,omitempty"`
// 	Ungrouped      *bool         `json:"ungrouped,omitempty"`
// 	ResponseFormat *string       `json:"response_format,omitempty"`
// 	Total          *bool         `json:"total,omitempty"`
// }

type DBAnalyticsResult struct {
	SQL  string        `json:"sql"`
	Args []interface{} `json:"args"`
	Data interface{}   `json:"data"`
}

type DBAnalyticsParams struct {
	WorkspaceID string          `json:"workspace_id"`
	Query       json.RawMessage `json:"query"`
}

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
