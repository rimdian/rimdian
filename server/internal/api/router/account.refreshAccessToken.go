package api

import (
	"net/http"

	"github.com/rimdian/rimdian/internal/common/auth"
	"go.opencensus.io/plugin/ochttp"
)

// AccountRefreshAccessToken
func (api *API) AccountRefreshAccessToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ochttp.SetRoute(ctx, "/api/account.refreshAccessToken")

	claims, err := auth.GetAccountTokenClaimsFromContext(ctx)
	if err != nil {
		api.ReturnJSONError(w, http.StatusUnauthorized, err)
		return
	}

	loginResult, code, err := api.Svc.AccountRefreshAccessToken(ctx, claims.AccountID, claims.SessionID)

	if err != nil {
		api.ReturnJSONError(w, code, err)
		return
	}

	ReturnJSON(w, http.StatusOK, loginResult)
}
