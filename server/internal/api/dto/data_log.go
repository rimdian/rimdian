package dto

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rimdian/rimdian/internal/api/entity"
	"github.com/rotisserie/eris"
)

type DataLogReprocessOne struct {
	DataLogID   string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
}

type DataLogListResult struct {
	DataLogs  []*entity.DataLog `json:"data_logs"`
	NextToken string            `json:"next_token,omitempty"`
	// data_log table has no "reverse" index, so we can only "fetch more" like logs
	// PreviousToken string               `json:"previous_token,omitempty"`
}

type DataLogListParams struct {
	WorkspaceID string  `json:"workspace_id"`
	Limit       int     `json:"limit"`
	NextToken   *string `json:"next_token,omitempty"`
	// PreviousToken *string `json:"previous_token,omitempty"`
	// filters:
	DataLogID    *string    `json:"id,omitempty"`
	Origin       *int       `json:"origin,omitempty"`
	OriginID     *string    `json:"origin_id,omitempty"`
	EventAtSince *time.Time `json:"event_at_since,omitempty"`
	EventAtUntil *time.Time `json:"event_at_until,omitempty"`
	HasError     *int       `json:"has_error,omitempty"`
	Checkpoint   *int       `json:"checkpoint,omitempty"`
	Kind         *string    `json:"kind,omitempty"`
	UserID       *string    `json:"user_id,omitempty"`
	ItemID       *string    `json:"item_id,omitempty"`
	// pagination computed server side:
	NextID       string
	NextDate     time.Time
	PreviousID   string
	PreviousDate time.Time
}

var (
	ErrDataLogListNextInvalid         error = eris.New("data_log list: next is not valid")
	ErrDataLogListPreviousInvalid     error = eris.New("data_log list: previous is not valid")
	ErrDataLogListLimitInvalid        error = eris.New("data_log list: limit is not valid")
	ErrDataLogListOriginInvalid       error = eris.New("data_log list: origin is not valid")
	ErrDataLogListCheckpointInvalid   error = eris.New("data_log list: checkpoint is not valid")
	ErrDataLogListEventAtSinceInvalid error = eris.New("data_log list: event_at_since is not valid")
	ErrDataLogListEventAtUntilInvalid error = eris.New("data_log list: event_at_until is not valid")

	MicrosecondLayout = "2006-01-02T15:04:05.999999Z" // .99 = trick for microseconds
)

func EncodePaginationToken(id string, date time.Time) string {
	return base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%v~%v", id, date.Format(MicrosecondLayout))))
}

func DecodePaginationToken(token string) (id string, date time.Time, err error) {
	var decoded []byte
	decoded, err = base64.URLEncoding.DecodeString(token)
	if err != nil {
		return
	}

	bits := strings.Split(string(decoded), "~")

	if len(bits) != 2 {
		err = eris.New("invalid pagination token")
		return
	}

	id = bits[0]

	date, err = time.Parse(MicrosecondLayout, bits[1])

	return
}

func (params *DataLogListParams) FromRequest(r *http.Request) (err error) {

	params.WorkspaceID = r.FormValue("workspace_id")

	// parse pagination token, either next_token or previous_token

	if r.FormValue("next_token") != "" {
		params.NextID, params.NextDate, err = DecodePaginationToken(r.FormValue("next_token"))
		if err != nil {
			return eris.Wrapf(ErrDataLogListNextInvalid, "err: %v", err)
		}
	} else if r.FormValue("previous_token") != "" {
		params.PreviousID, params.PreviousDate, err = DecodePaginationToken(r.FormValue("previous_token"))
		if err != nil {
			return eris.Wrapf(ErrDataLogListPreviousInvalid, "err: %v", err)
		}
	}

	// default limit
	limit := 25

	if r.FormValue("limit") != "" {
		limit, err = strconv.Atoi(r.FormValue("limit"))
		if err != nil {
			return eris.Wrapf(ErrDataLogListLimitInvalid, "err: %v", err)
		}
	}

	if limit < 1 || limit > 100 {
		return ErrDataLogListLimitInvalid
	}

	params.Limit = limit

	// filters

	if r.FormValue("event_at_since") != "" {
		receivedAtSince, err := time.Parse(MicrosecondLayout, r.FormValue("event_at_since"))
		if err != nil {
			return eris.Wrapf(ErrDataLogListEventAtSinceInvalid, "err: %v", err)
		}
		params.EventAtSince = &receivedAtSince
	}

	if r.FormValue("event_at_until") != "" {
		receivedAtUntil, err := time.Parse(MicrosecondLayout, r.FormValue("event_at_until"))
		if err != nil {
			return eris.Wrapf(ErrDataLogListEventAtUntilInvalid, "err: %v", err)
		}
		params.EventAtUntil = &receivedAtUntil
	}

	// import id
	importID := r.FormValue("id")
	if importID != "" {
		params.DataLogID = &importID
	}

	// Origin
	originString := r.FormValue("origin")
	if originString != "" {
		// convert to int
		origin, err := strconv.Atoi(originString)
		if err != nil {
			return eris.Wrapf(ErrDataLogListOriginInvalid, "err: %v", err)
		}
		if origin < 0 || origin > 4 {
			return ErrDataLogListOriginInvalid
		}
		params.Origin = &origin
	}

	// Origin ID
	originID := r.FormValue("origin_id")
	if originID != "" {
		params.OriginID = &originID
	}

	// has_error
	hasError := r.FormValue("has_error")
	if hasError != "" {
		hasErrorInt := 0
		hasErrorInt, err = strconv.Atoi(hasError)
		if err != nil {
			return eris.New("has_error is not valid")
		}

		if hasErrorInt != entity.DataLogHasErrorNone && hasErrorInt != entity.DataLogHasErrorRetryable && hasErrorInt != entity.DataLogHasErrorNotRetryable {
			return eris.New("has_error is not valid")
		}

		params.HasError = &hasErrorInt
	}

	// Code returned
	checkpoint := r.FormValue("checkpoint")
	if r.FormValue("checkpoint") != "" {
		checkpointInt := 0
		checkpointInt, err = strconv.Atoi(checkpoint)
		if err != nil {
			return eris.Wrapf(ErrDataLogListCheckpointInvalid, "err: %v", err)
		}

		if checkpointInt != entity.DataLogCheckpointPending &&
			checkpointInt != entity.DataLogCheckpointHookOnValidationExecuted &&
			checkpointInt != entity.DataLogCheckpointPersisted &&
			checkpointInt != entity.DataLogCheckpointItemUpserted &&
			checkpointInt != entity.DataLogCheckpointConversionsAttributed &&
			checkpointInt != entity.DataLogCheckpointSegmentsRecomputed &&
			checkpointInt != entity.DataLogCheckpointWorkflowsTriggered &&
			checkpointInt != entity.DataLogCheckpointHooksFinalizeExecuted &&
			checkpointInt != entity.DataLogCheckpointDone {
			return ErrDataLogListCheckpointInvalid
		}

		params.Checkpoint = &checkpointInt
	}

	// user id
	userID := r.FormValue("user_id")
	if userID != "" {
		params.UserID = &userID
	}

	// kind
	kind := r.FormValue("kind")
	if kind != "" {
		params.Kind = &kind
	}

	// item id
	itemID := r.FormValue("item_id")
	if itemID != "" {
		params.ItemID = &itemID
	}

	// if item_id is set, kind or origin_id or user_id must be set to match the index
	if params.ItemID != nil && (params.Kind == nil && params.OriginID == nil && params.UserID == nil) {
		return eris.Errorf("item_id is set, but kind or origin_id or user_id is not set")
	}

	return nil
}

type DataLogToRespawn struct {
	ID      string
	EventAt time.Time // used to compute the next token for pagination
}
