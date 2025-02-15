package api

import (
	"net/http"

	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/common/auth"
	"go.opencensus.io/plugin/ochttp"
)

func (api *API) AppGet(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	ochttp.SetRoute(ctx, "/api/app.get")

	claims, err := auth.GetAccountTokenClaimsFromContext(ctx)
	if err != nil {
		api.ReturnJSONError(w, http.StatusUnauthorized, err)
		return
	}

	params := &dto.AppGetParams{}

	if err := params.FromRequest(r); err != nil {
		api.ReturnJSONError(w, http.StatusBadRequest, err)
		return
	}

	result, code, err := api.Svc.AppGet(ctx, claims.AccountID, params)

	if err != nil {
		api.ReturnJSONError(w, code, err)
		return
	}

	ReturnJSON(w, http.StatusOK, result)
}
