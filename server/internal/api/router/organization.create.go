package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/common/auth"
	"go.opencensus.io/plugin/ochttp"
)

// OrganizationCreate
func (api *API) OrganizationCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ochttp.SetRoute(ctx, "/api/organization.create")

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

	data := &dto.OrganizationCreate{}

	// drop data if payload doesnt match
	if err := json.Unmarshal(body, data); err != nil {
		api.ReturnJSONError(w, http.StatusBadRequest, err)
		return
	}

	// abort if account is not root
	if claims.AccountID != "root" {
		api.ReturnJSONError(w, http.StatusForbidden, err)
		return
	}

	org, code, err := api.Svc.OrganizationCreate(ctx, data)

	if err != nil {
		api.ReturnJSONError(w, code, err)
		return
	}

	ReturnJSON(w, http.StatusOK, org)
}
