package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/common/auth"
	"go.opencensus.io/plugin/ochttp"
)

// OrganizationInvitationCreate
func (api *API) OrganizationInvitationCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ochttp.SetRoute(ctx, "/api/organizationInvitation.create")

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

	accountInvitation := &dto.OrganizationInvitation{}

	// drop data if payload doesnt match
	if err := json.Unmarshal(body, accountInvitation); err != nil {
		api.ReturnJSONError(w, http.StatusBadRequest, err)
		return
	}

	accountInvitation.FromAccountID = claims.AccountID

	code, err := api.Svc.OrganizationInvitationCreate(ctx, accountInvitation)

	if err != nil {
		api.ReturnJSONError(w, code, err)
		return
	}

	ReturnJSON(w, http.StatusOK, map[string]bool{"success": true})
}
