package dto

import (
	"net/http"
	"strconv"
	"time"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

type UserShowParams struct {
	WorkspaceID    string `json:"workspace_id"`
	UserID         string `json:"id"`
	UserExternalID string `json:"external_id"`
	*UserWith
}

func (params *UserShowParams) FromRequest(r *http.Request) (err error) {
	params.WorkspaceID = r.FormValue("workspace_id")
	params.UserID = r.FormValue("id")
	params.UserExternalID = r.FormValue("external_id")

	if params.UserID == "" && params.UserExternalID == "" {
		return eris.New("id or external_id is required")
	}

	params.UserWith = &UserWith{}
	params.UserWith.FromRequest(r)

	return nil
}

type UserWith struct {
	Devices           bool `json:"with_devices"`
	Segments          bool `json:"with_segments"`
	Aliases           bool `json:"with_aliases"`
	SubscriptionLists bool `json:"with_subscription_lists"`
}

func (params *UserWith) FromRequest(r *http.Request) {
	params.Devices = r.FormValue("with_devices") == "true"
	params.Segments = r.FormValue("with_segments") == "true"
	params.Aliases = r.FormValue("with_aliases") == "true"
	params.SubscriptionLists = r.FormValue("with_subscription_lists") == "true"
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
	ListID    *string `json:"list_id,omitempty"`
	// enrichments:
	WithSegments          bool `json:"with_segments"`
	WithSubscriptionLists bool `json:"with_subscription_lists"`
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

	// check if list_id contains a value if provided
	listID := r.FormValue("list_id")
	if listID != "" {
		params.ListID = &listID
	}

	// check if with_segments contains a value if provided
	params.WithSegments = r.FormValue("with_segments") == "true"

	// check if with_subscription_lists contains a value if provided
	params.WithSubscriptionLists = r.FormValue("with_subscription_lists") == "true"

	return nil
}
