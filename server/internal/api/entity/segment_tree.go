package entity

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/rotisserie/eris"
)

var (
	SegmentTreeNodeKindBranch = "branch"
	SegmentTreeNodeKindLeaf   = "leaf"

	ErrSegmentLeavesEmpty               = eris.New("children leaves are empty")
	ErrSegmentNodeKindInvalid           = eris.New("node kind is not valid")
	ErrSegmentNodeBranchEmpty           = eris.New("branch is empty")
	ErrSegmentNodeLeafEmpty             = eris.New("leaf is empty")
	ErrSegmentNodeTableEmpty            = eris.New("node table is empty")
	ErrSegmentNodeTableInvalid          = eris.New("node table is invalid")
	ErrSegmentFilterFieldNameEmpty      = eris.New("filter field_name is empty")
	ErrSegmentFilterFieldTypeRequired   = eris.New("filter field type is required")
	ErrSegmentFilterTypeOperatorInvalid = eris.New("filter operator is not valid")

	SegmentFilterOperators = []string{
		"is_set",
		"is_not_set",
		"equals",
		"not_equals",
		"contains",
		"not_contains",
		// | 'starts_with'
		// | 'not_starts_with'
		// | 'ends_with'
		// | 'not_ends_with'
		"gt",
		"gte",
		"lt",
		"lte",
		"in_date_range",
		"not_in_date_range",
		"before_date",
		"after_date",
	}

	SegmentFilterOperatorsString = []string{
		"is_set",
		"is_not_set",
		"equals",
		"not_equals",
		"contains",
		"not_contains",
	}

	SegmentFilterOperatorsNumber = []string{
		"is_set",
		"is_not_set",
		"equals",
		"not_equals",
		"contains",
		"not_contains",
		"gt",
		"gte",
		"lt",
		"lte",
	}

	SegmentFilterOperatorsTime = []string{
		"is_set",
		"is_not_set",
		"in_date_range",
		"not_in_date_range",
		"before_date",
		"after_date",
	}
)

// tree of filters
type SegmentTreeNode struct {
	Kind   string                 `json:"kind"` // branch | leaf
	Branch *SegmentTreeNodeBranch `json:"branch,omitempty"`
	Leaf   *SegmentTreeNodeLeaf   `json:"leaf,omitempty"`
}

func (x *SegmentTreeNode) Scan(val interface{}) error {

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

func (x SegmentTreeNode) Value() (driver.Value, error) {
	return json.Marshal(x)
}

type SegmentTreeNodeBranch struct {
	Operator string             `json:"operator"` // and | or
	Leaves   []*SegmentTreeNode `json:"leaves"`
}

type SegmentTreeNodeLeaf struct {
	Table   string                    `json:"table"`
	Filters []*SegmentDimensionFilter `json:"filters"`
	Action  *SegmentActionCondition   `json:"action,omitempty"`
}

type SegmentDimensionFilter struct {
	FieldName    string    `json:"field_name"`
	FieldType    string    `json:"field_type"` // string | number | time
	Operator     string    `json:"operator"`
	StringValues []string  `json:"string_values,omitempty"`
	NumberValues []float64 `json:"number_values,omitempty"`
}

type SegmentActionCondition struct {
	CountOperator     string   `json:"count_operator"` // at_least | at_most | exactly
	CountValue        int      `json:"count_value"`
	TimeframeOperator string   `json:"timeframe_operator"` // 'anytime' | 'in_date_range' | 'before_date' | 'after_date' | 'in_the_last_days'
	TimeframeValues   []string `json:"timeframe_values"`
}

type SegmentConditionStringType struct {
	Operator string  `json:"operator"` // 'exists' | 'notExists' | 'equals' | 'notEquals' | 'contains' | 'notContains' | 'regex'
	Value    *string `json:"value,omitempty"`
}

func (node *SegmentTreeNode) Validate(schemasMap map[string]*CubeJSSchema) error {

	if node.Kind != SegmentTreeNodeKindBranch && node.Kind != SegmentTreeNodeKindLeaf {
		return ErrSegmentNodeKindInvalid
	}

	// Branch
	if node.Kind == SegmentTreeNodeKindBranch {
		if node.Branch == nil {
			return ErrSegmentNodeBranchEmpty
		}

		if node.Branch.Leaves == nil || len(node.Branch.Leaves) == 0 {
			return ErrSegmentLeavesEmpty
		}

		// validate children
		for _, child := range node.Branch.Leaves {
			if err := child.Validate(schemasMap); err != nil {
				return err
			}
		}
	}

	// Leaf
	if node.Kind == SegmentTreeNodeKindLeaf {
		if node.Leaf == nil {
			return ErrSegmentNodeLeafEmpty
		}

		node.Leaf.Table = strings.TrimSpace(node.Leaf.Table)
		if node.Leaf.Table == "" {
			return ErrSegmentNodeTableEmpty
		}

		// validate table + field againts cubejs schema

		// uc first
		tableTitle := strings.ToUpper(node.Leaf.Table[:1]) + node.Leaf.Table[1:]
		if _, ok := schemasMap[tableTitle]; !ok {
			return eris.Wrapf(ErrSegmentNodeTableInvalid, "got: %v", node.Leaf.Table)
		}

		for _, filter := range node.Leaf.Filters {
			filter.FieldName = strings.TrimSpace(filter.FieldName)
			if filter.FieldName == "" {
				return ErrSegmentFilterFieldNameEmpty
			}

			filter.FieldType = strings.TrimSpace(filter.FieldType)
			if filter.FieldType == "" {
				return ErrSegmentFilterFieldTypeRequired
			}

			switch filter.FieldType {
			case "string":
				if !govalidator.IsIn(filter.Operator, SegmentFilterOperatorsString...) {
					return eris.Wrapf(ErrSegmentFilterTypeOperatorInvalid, "got: %v", filter.Operator)
				}
				if (filter.Operator == "equals" || filter.Operator == "not_equals" || filter.Operator == "contains" || filter.Operator == "not_contains") && len(filter.StringValues) == 0 {
					return eris.Errorf("string_values must be > 0, got %v", len(filter.StringValues))
				}
			case "number":
				if !govalidator.IsIn(filter.Operator, SegmentFilterOperatorsNumber...) {
					return eris.Wrapf(ErrSegmentFilterTypeOperatorInvalid, "got: %v", filter.Operator)
				}
				if (filter.Operator == "equals" || filter.Operator == "not_equals" || filter.Operator == "gt" || filter.Operator == "gte" || filter.Operator == "lt" || filter.Operator == "lte") && len(filter.NumberValues) == 0 {
					return eris.Errorf("number_values must be > 0, got %v", len(filter.NumberValues))
				}

			case "time":
				if !govalidator.IsIn(filter.Operator, SegmentFilterOperatorsTime...) {
					return eris.Wrapf(ErrSegmentFilterTypeOperatorInvalid, "got: %v", filter.Operator)
				}
				if (filter.Operator == "in_date_range" || filter.Operator == "not_in_date_range") && len(filter.StringValues) != 2 {
					return eris.Errorf("string_values must be 2, got %v", len(filter.StringValues))
				}
				if (filter.Operator == "before_date" || filter.Operator == "after_date") && len(filter.StringValues) != 1 {
					return eris.Errorf("string_values must be 1, got %v", len(filter.StringValues))
				}

			default:
				return eris.Errorf("field type %v is not valid", filter.FieldType)
			}

			fieldExists := false
			foundType := ""

			// find field in measures
			if _, ok := schemasMap[tableTitle].Measures[filter.FieldName]; ok {
				fieldExists = true
				if filter.FieldType == "count" || filter.FieldType == "count_distinct" || filter.FieldType == "avg" || filter.FieldType == "sum" || filter.FieldType == "min" || filter.FieldType == "max" {
					foundType = "number"
				} else {
					foundType = filter.FieldType
				}
			}

			// find field in dimensions
			if _, ok := schemasMap[tableTitle].Dimensions[filter.FieldName]; ok {
				fieldExists = true
				foundType = schemasMap[tableTitle].Dimensions[filter.FieldName].Type
			}

			if !fieldExists {
				return eris.Errorf("field %v not found in table %v", filter.FieldName, node.Leaf.Table)
			}

			if foundType != filter.FieldType {
				return eris.Errorf("field %v in table %v is of type %v, got %v", filter.FieldName, node.Leaf.Table, foundType, filter.FieldType)
			}
		}

		// is an action
		if node.Leaf.Table != "user" {
			if node.Leaf.Action == nil {
				return eris.Errorf("action is required")
			}

			if !govalidator.IsIn(node.Leaf.Action.CountOperator, "at_least", "at_most", "exactly") {
				return eris.Errorf("count_operator must be one of %v, got %v", SegmentFilterOperatorsNumber, node.Leaf.Action.CountOperator)
			}

			if !govalidator.IsIn(node.Leaf.Action.TimeframeOperator, "anytime", "in_date_range", "before_date", "after_date", "in_the_last_days") {
				return eris.Errorf("timeframe_operator must be one of %v, got %v", SegmentFilterOperatorsTime, node.Leaf.Action.TimeframeOperator)
			}

			if node.Leaf.Action.CountValue <= 0 {
				return eris.Errorf("count_value must be > 0, got %v", node.Leaf.Action.CountValue)
			}

			if node.Leaf.Action.TimeframeOperator == "in_date_range" && len(node.Leaf.Action.TimeframeValues) != 2 {
				return eris.Errorf("timeframe_values must be 2, got %v", len(node.Leaf.Action.TimeframeValues))
			}

			if (node.Leaf.Action.TimeframeOperator == "before_date" || node.Leaf.Action.TimeframeOperator == "after_date" || node.Leaf.Action.TimeframeOperator == "in_the_last_days") && len(node.Leaf.Action.TimeframeValues) != 1 {
				return eris.Errorf("timeframe_values must be 1, got %v", len(node.Leaf.Action.TimeframeValues))
			}

		}
	}

	return nil
}
