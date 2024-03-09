package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/common/auth"
	"go.opencensus.io/plugin/ochttp"
)

func (api *API) SegmentPreview(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	ochttp.SetRoute(ctx, "/api/segment.preview")

	claims, err := auth.GetAccountTokenClaimsFromContext(ctx)
	if err != nil {
		api.ReturnJSONError(w, http.StatusUnauthorized, err)
		return
	}

	// Read body

	body, err := io.ReadAll(r.Body)

	if err != nil {
		api.ReturnJSONError(w, http.StatusUnprocessableEntity, err)
		return
	}

	payload := &dto.SegmentPreviewParams{}

	// drop data if payload doesnt match
	if err := json.Unmarshal(body, payload); err != nil {
		api.ReturnJSONError(w, http.StatusBadRequest, err)
		return
	}

	result, code, err := api.Svc.SegmentPreview(ctx, claims.AccountID, payload)

	if err != nil {
		api.ReturnJSONError(w, code, err)
		return
	}

	ReturnJSON(w, http.StatusOK, result)
}
