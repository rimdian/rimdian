package dto

import (
	"net/http"
	"strconv"
	"time"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

type UserShowResult struct {
	User         *entity.User          `json:"user"`
	UserSegments []*entity.UserSegment `json:"user_segments"`
	Devices      []*entity.Device      `json:"devices"`
	Aliases      []*entity.UserAlias   `json:"aliases"`
}

type UserListResult struct {
	Users         []*entity.User `json:"users"`
	NextToken     string         `json:"next_token,omitempty"`
	PreviousToken string         `json:"previous_token,omitempty"`
}

type UserListParams struct {
	WorkspaceID   string  `json:"workspace_id"`
	Limit         int     `json:"limit"`
	NextToken     *string `json:"next_token,omitempty"`
	PreviousToken *string `json:"previous_token,omitempty"`
	// filters:
	UserIDs   *string `json:"user_ids,omitempty"` // comma separated user_ids
	SegmentID *string `json:"segment_id,omitempty"`
	// pagination computed server side:
	NextID       string
	NextDate     time.Time
	PreviousID   string
	PreviousDate time.Time
}

var (
	UserListLimitMax                 = 100
	ErrUserListNextInvalid     error = eris.New("user list: next is not valid")
	ErrUserListPreviousInvalid error = eris.New("user list: previous is not valid")
	ErrUserListLimitInvalid    error = eris.New("user list: limit is not valid")
)

func (params *UserListParams) FromRequest(r *http.Request) (err error) {

	params.WorkspaceID = r.FormValue("workspace_id")

	// parse pagination token, either next_token or previous_token

	if r.FormValue("next_token") != "" {
		params.NextID, params.NextDate, err = DecodePaginationToken(r.FormValue("next_token"))
		if err != nil {
			return eris.Errorf("next_token err: %v", err)
		}
	} else if r.FormValue("previous_token") != "" {
		params.PreviousID, params.PreviousDate, err = DecodePaginationToken(r.FormValue("previous_token"))
		if err != nil {
			return eris.Errorf("previous_token err: %v", err)
		}
	}

	// default limit
	limit := 25

	if r.FormValue("limit") != "" {
		limit, err = strconv.Atoi(r.FormValue("limit"))
		if err != nil {
			return eris.Wrapf(ErrUserListLimitInvalid, "err: %v", err)
		}
	}

	if limit < 1 || limit > UserListLimitMax {
		return ErrUserListLimitInvalid
	}

	params.Limit = limit

	// check if user_ids contains a value if provided
	userIDs := r.FormValue("user_ids")
	if userIDs != "" {
		params.UserIDs = &userIDs
	}

	// check if segment_id contains a value if provided
	segmentID := r.FormValue("segment_id")
	if segmentID != "" {
		params.SegmentID = &segmentID
	}

	return nil
}
