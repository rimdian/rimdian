package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/rimdian/rimdian/internal/api/dto"
	"go.opencensus.io/plugin/ochttp"
)

// OrganizationInvitationRead
func (api *API) OrganizationInvitationRead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ochttp.SetRoute(ctx, "/api/organizationInvitation.read")

	// Read body
	body, err := io.ReadAll(r.Body)

	if err != nil {
		api.ReturnJSONError(w, http.StatusUnprocessableEntity, err)
		return
	}

	payload := &dto.OrganizationInvitationRead{}

	// drop data if payload doesnt match
	if err := json.Unmarshal(body, payload); err != nil {
		api.ReturnJSONError(w, http.StatusBadRequest, err)
		return
	}

	result, code, err := api.Svc.OrganizationInvitationRead(ctx, payload.Token)

	if err != nil {
		api.ReturnJSONError(w, code, err)
		return
	}

	ReturnJSON(w, http.StatusOK, result)
}
