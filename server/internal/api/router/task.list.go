package api

import (
	"net/http"

	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/common/auth"
	"go.opencensus.io/plugin/ochttp"
)

func (api *API) TaskList(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	ochttp.SetRoute(ctx, "/api/task.list")

	claims, err := auth.GetAccountTokenClaimsFromContext(ctx)
	if err != nil {
		api.ReturnJSONError(w, http.StatusUnauthorized, err)
		return
	}

	params := &dto.TaskListParams{}

	if err := params.FromRequest(r); err != nil {
		api.ReturnJSONError(w, http.StatusBadRequest, err)
		return
	}

	result, code, err := api.Svc.TaskList(ctx, claims.AccountID, params)

	if err != nil {
		api.ReturnJSONError(w, code, err)
		return
	}

	ReturnJSON(w, http.StatusOK, result)
}
