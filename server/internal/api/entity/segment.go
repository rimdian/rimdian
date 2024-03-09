package entity

import (
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/rimdian/rimdian/internal/api/common"
	"github.com/rotisserie/eris"
)

var (
	TaskKindRecomputeSegment = "recompute_segment"
	TaskNameRecomputeSegment = "Recompute segment"

	SegmentStatusActive   = "active"
	SegmentStatusDeleted  = "deleted"
	SegmentStatusBuilding = "building"

	SegmentAllUsersID = "_all"

	ErrSegmentAlreadyExists = eris.New("segment already exists")

	DefaultAnonymousSegment = Segment{
		ID:    "anonymous",
		Name:  "Anonymous",
		Color: "default",
		Tree: SegmentTreeNode{
			Kind: "branch",
			Branch: &SegmentTreeNodeBranch{
				Operator: "and",
				Leaves: []*SegmentTreeNode{
					{
						Kind: "leaf",
						Leaf: &SegmentTreeNodeLeaf{
							Table: "user",
							Filters: []*SegmentDimensionFilter{
								{
									FieldName:    "is_authenticated",
									FieldType:    "number",
									Operator:     "equals",
									NumberValues: []float64{0},
								},
							},
						},
					},
				},
			},
		},
		Timezone:      "UTC",
		Version:       1,
		GeneratedSQL:  "SELECT COUNT(user__auth.id) FROM `user` user__auth WHERE user__auth.is_authenticated = ?",
		GeneratedArgs: []interface{}{0},
		Status:        SegmentStatusActive,
	}

	DefaultAuthenticatedSegment = Segment{
		ID:    "authenticated",
		Name:  "Authenticated",
		Color: "blue",
		Tree: SegmentTreeNode{
			Kind: "branch",
			Branch: &SegmentTreeNodeBranch{
				Operator: "and",
				Leaves: []*SegmentTreeNode{
					{
						Kind: "leaf",
						Leaf: &SegmentTreeNodeLeaf{
							Table: "user",
							Filters: []*SegmentDimensionFilter{
								{
									FieldName:    "is_authenticated",
									FieldType:    "number",
									Operator:     "equals",
									NumberValues: []float64{1},
								},
							},
						},
					},
				},
			},
		},
		Timezone:      "UTC",
		Version:       1,
		GeneratedSQL:  "SELECT COUNT(user__auth.id) FROM `user` user__auth WHERE user__auth.is_authenticated = ?",
		GeneratedArgs: []interface{}{1},
		Status:        SegmentStatusActive,
	}

	RecomputeSegmentStepMatchUsers    = "match_users"
	RecomputeSegmentStepEnterUsers    = "enter_users"
	RecomputeSegmentStepExitUsers     = "exit_users"
	RecomputeSegmentStepEnterDataLogs = "enter_data_logs"
	RecomputeSegmentStepExitDataLogs  = "exit_data_logs"
	RecomputeSegmentStepEnqueueJobs   = "enqueue_jobs"
	RecomputeSegmentStepFinalize      = "finalize"
	RecomputeSegmentStepDone          = "done"
)

type Segment struct {
	ID              string            `db:"id" json:"id"`
	Name            string            `db:"name" json:"name"`
	Color           string            `db:"color" json:"color"`
	ParentSegmentID *string           `db:"parent_segment_id" json:"parent_segment_id,omitempty"`
	Tree            SegmentTreeNode   `db:"tree" json:"tree"`
	Timezone        string            `db:"timezone" json:"timezone"`
	Version         int               `db:"version" json:"version"`
	GeneratedSQL    string            `db:"generated_sql" json:"generated_sql"`
	GeneratedArgs   ArrayOfInterfaces `db:"generated_args" json:"generated_args"`
	Status          string            `db:"status" json:"status"`
	CreatedAt       time.Time         `db:"db_created_at" json:"db_created_at"`
	UpdatedAt       time.Time         `db:"db_updated_at" json:"db_updated_at"`

	// server-side JOINED fields:
	UsersCount int `db:"users_count" json:"users_count"`
}

func (s *Segment) Validate(existingSegments []*Segment, schemasMap map[string]*CubeJSSchema) (err error) {
	if s.ID == "" {
		return eris.New("segment: id is required")
	}
	if s.Name == "" {
		return eris.New("segment: name is required")
	}
	if err = common.ValidateColor(s.Color); err != nil {
		return eris.Wrap(err, "invalid color")
	}
	if s.ParentSegmentID != nil && (*s.ParentSegmentID != "authenticated" && *s.ParentSegmentID != SegmentAllUsersID) {
		return eris.Errorf("invalid parent segment id: %s", *s.ParentSegmentID)
	}
	if err = s.Tree.Validate(schemasMap); err != nil {
		return eris.Wrap(err, "invalid tree")
	}
	if !govalidator.IsIn(s.Timezone, common.Timezones...) {
		return eris.Errorf("invalid timezone: %s", s.Timezone)
	}
	if s.Version < 1 {
		return eris.Errorf("invalid version: %d", s.Version)
	}
	if s.Status != SegmentStatusActive && s.Status != SegmentStatusBuilding {
		return eris.Errorf("invalid status: %s", s.Status)
	}

	return
}

func GenerateDefaultSegments() (segments []Segment) {
	segments = append(segments, DefaultAnonymousSegment)
	segments = append(segments, DefaultAuthenticatedSegment)
	return
}

var SegmentSchema string = `CREATE ROWSTORE TABLE IF NOT EXISTS segment (
	id VARCHAR(60) NOT NULL,
	name VARCHAR(255) NOT NULL,
	color VARCHAR(20) NOT NULL,
	parent_segment_id VARCHAR(60),
	timezone VARCHAR(255) NOT NULL,
	tree JSON NOT NULL,
	version INT DEFAULT 0,
	generated_sql VARCHAR(10000) NOT NULL,
	generated_args JSON NOT NULL,
	status VARCHAR(20) NOT NULL,
	db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (id)
  ) COLLATE utf8mb4_general_ci;`

var SegmentSchemaMYSQL string = `CREATE TABLE IF NOT EXISTS segment (
	id VARCHAR(60) NOT NULL,
	name VARCHAR(255) NOT NULL,
	color VARCHAR(20) NOT NULL,
	parent_segment_id VARCHAR(60),
	timezone VARCHAR(255) NOT NULL,
	tree JSON NOT NULL,
	version INT DEFAULT 0,
	generated_sql VARCHAR(10000) NOT NULL,
	generated_args JSON NOT NULL,
	status VARCHAR(20) NOT NULL,
	db_created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	db_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (id)
  ) COLLATE utf8mb4_general_ci;`
