package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/rimdian/rimdian/internal/api/dto"
	"go.opencensus.io/plugin/ochttp"
)

// AccountLogin
func (api *API) AccountLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ochttp.SetRoute(ctx, "/api/account.login")

	// Read body
	body, err := io.ReadAll(r.Body)

	if err != nil {
		api.ReturnJSONError(w, http.StatusUnprocessableEntity, err)
		return
	}

	loginDTO := &dto.AccountLogin{}

	// drop data if payload doesnt match
	if err := json.Unmarshal(body, loginDTO); err != nil {
		api.ReturnJSONError(w, http.StatusBadRequest, err)
		return
	}

	loginDTO.UserAgent = r.Header.Get("User-Agent")
	loginDTO.ClientIP = r.RemoteAddr

	loginResult, code, err := api.Svc.AccountLogin(ctx, loginDTO)

	if err != nil {
		api.ReturnJSONError(w, code, err)
		return
	}

	ReturnJSON(w, http.StatusOK, loginResult)
}
