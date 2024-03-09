package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/common/auth"
	"go.opencensus.io/plugin/ochttp"
)

// OrganizationAccountCreateServiceAccount
func (api *API) OrganizationAccountCreateServiceAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ochttp.SetRoute(ctx, "/api/organizationAccount.createServiceAccount")

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

	payload := &dto.OrganizationAccountCreateServiceAccount{}

	// drop data if payload doesnt match
	if err := json.Unmarshal(body, payload); err != nil {
		api.ReturnJSONError(w, http.StatusBadRequest, err)
		return
	}

	// log.Printf("payload %+v\n", payload)

	code, err := api.Svc.OrganizationAccountCreateServiceAccount(ctx, claims.AccountID, payload)

	if err != nil {
		api.ReturnJSONError(w, code, err)
		return
	}

	ReturnJSON(w, http.StatusOK, map[string]bool{"success": true})
}
