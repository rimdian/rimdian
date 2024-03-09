package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/rimdian/rimdian/internal/api/dto"
	"go.opencensus.io/plugin/ochttp"
)

// AccountResetPassword
func (api *API) AccountResetPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ochttp.SetRoute(ctx, "/api/account.resetPassword")

	// Read body
	body, err := io.ReadAll(r.Body)

	if err != nil {
		api.ReturnJSONError(w, http.StatusUnprocessableEntity, err)
		return
	}

	resetpasswordDTO := &dto.AccountResetPassword{}

	// drop data if payload doesnt match
	if err := json.Unmarshal(body, resetpasswordDTO); err != nil {
		api.ReturnJSONError(w, http.StatusBadRequest, err)
		return
	}

	code, err := api.Svc.AccountResetPassword(ctx, resetpasswordDTO)

	// ignore 4xx errors to avoid leaking email info
	if err != nil && code >= 500 {
		api.ReturnJSONError(w, code, err)
		return
	}

	ReturnJSON(w, http.StatusOK, map[string]string{})
}
