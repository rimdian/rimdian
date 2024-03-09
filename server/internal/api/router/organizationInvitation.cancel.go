package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/common/auth"
	"go.opencensus.io/plugin/ochttp"
)

// OrganizationInvitationCancel
func (api *API) OrganizationInvitationCancel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ochttp.SetRoute(ctx, "/api/organizationInvitation.cancel")

	claims, err := auth.GetAccountTokenClaimsFromContext(ctx)
	if err != nil {
		api.ReturnJSONError(w, http.StatusUnauthorized, err)
		return
	}

	// Read body

	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "cannot read body", http.StatusUnprocessableEntity)
		return
	}

	accountDeleteInvitation := &dto.OrganizationInvitationCancel{}

	// drop data if payload doesnt match
	if err := json.Unmarshal(body, accountDeleteInvitation); err != nil {
		api.ReturnJSONError(w, http.StatusBadRequest, err)
		return
	}

	code, err := api.Svc.OrganizationInvitationCancel(ctx, claims.AccountID, accountDeleteInvitation)

	if err != nil {
		api.ReturnJSONError(w, code, err)
		return
	}

	ReturnJSON(w, http.StatusOK, map[string]bool{"success": true})
}
