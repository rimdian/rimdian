package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/common/auth"
	"go.opencensus.io/plugin/ochttp"
)

// OrganizationInvitationConsume
func (api *API) OrganizationInvitationConsume(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ochttp.SetRoute(ctx, "/api/organizationInvitation.consume")

	// Read body
	body, err := io.ReadAll(r.Body)

	if err != nil {
		api.ReturnJSONError(w, http.StatusUnprocessableEntity, err)
		return
	}

	payload := &dto.OrganizationInvitationConsume{}

	// drop data if payload doesnt match
	if err := json.Unmarshal(body, payload); err != nil {
		api.ReturnJSONError(w, http.StatusBadRequest, err)
		return
	}

	// reset account ID to avoid injection
	payload.AccountID = nil

	eventualAccountToken := r.Context().Value(auth.AccountTokenContextKey)

	// log.Printf("eventualAccountToken %+v\n", eventualAccountToken)

	if eventualAccountToken != nil {

		claims, err := auth.GetAccountTokenClaimsFromContext(ctx)
		if err != nil {
			api.ReturnJSONError(w, http.StatusUnauthorized, err)
			return
		}

		payload.AccountID = &claims.AccountID
	}

	payload.UserAgent = r.Header.Get("User-Agent")
	payload.ClientIP = r.RemoteAddr

	accountLoginResult, code, err := api.Svc.OrganizationInvitationConsume(ctx, payload)

	if err != nil {
		api.ReturnJSONError(w, code, err)
		return
	}

	ReturnJSON(w, http.StatusOK, accountLoginResult)
}
