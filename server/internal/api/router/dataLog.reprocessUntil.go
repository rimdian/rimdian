package api

import (
	"net/http"
	"time"

	"go.opencensus.io/plugin/ochttp"
)

func (api *API) DataLogReprocessUntil(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	ochttp.SetRoute(ctx, "/api/dataLog.reprocessUntil")

	// make sur we don't reprocess too soon, default to 24h
	minimumDelay := 24 * time.Hour

	// extract delay from query string, ex: ?delay=1h
	qsDelay := r.URL.Query().Get("delay")

	if qsDelay != "" {
		delay, err := time.ParseDuration(qsDelay)
		if err != nil {
			api.ReturnJSONError(w, http.StatusBadRequest, err)
			return
		}

		minimumDelay = delay
	}

	untilDate := time.Now().Add(-minimumDelay)

	code, err := api.Svc.DataLogReprocessUntil(ctx, untilDate)

	if err != nil {
		api.ReturnJSONError(w, code, err)
		return
	}

	ReturnJSON(w, http.StatusOK, nil)
}
