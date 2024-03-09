package api

import (
	"net/http"

	"go.opencensus.io/plugin/ochttp"
)

func (api *API) TaskWakeUpCron(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	ochttp.SetRoute(ctx, "/api/task.wakeUpCron")

	code, err := api.Svc.TaskWakeUpCron(ctx)

	if err != nil {
		api.ReturnJSONError(w, code, err)
		return
	}

	ReturnJSON(w, http.StatusOK, nil)
}
