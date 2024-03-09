package dto

import (
	"net/http"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

type Segment struct {
	WorkspaceID     string                 `json:"workspace_id"`
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Color           string                 `json:"color"`
	ParentSegmentID *string                `json:"parent_segment_id"`
	Tree            entity.SegmentTreeNode `json:"tree"`
	Timezone        string                 `json:"timezone"`
}

type DeleteSegment struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
}

type SegmentListResult struct {
	Segments []*entity.Segment `json:"segments"`
}

type SegmentListParams struct {
	WorkspaceID    string `json:"workspace_id"`
	WithUsersCount bool   `json:"with_users_count,omitempty"`
}

var (
	ErrSegmentListWorkspaceIDRequired error = eris.New("segment list: workspace_id is required")
)

func (params *SegmentListParams) FromRequest(r *http.Request) (err error) {

	params.WorkspaceID = r.FormValue("workspace_id")
	if params.WorkspaceID == "" {
		return ErrSegmentListWorkspaceIDRequired
	}

	if r.FormValue("with_users_count") == "true" {
		params.WithUsersCount = true
	}

	return nil
}

type SegmentPreviewParams struct {
	WorkspaceID     string                  `json:"workspace_id"`
	ParentSegmentID *string                 `json:"parent_segment_id"`
	Tree            *entity.SegmentTreeNode `json:"tree"`
	Timezone        string                  `json:"timezone"`
}

type SegmentPreviewResult struct {
	Count int64         `json:"count"`
	SQL   string        `json:"sql"`
	Args  []interface{} `json:"args"`
}
