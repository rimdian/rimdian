package api

import (
	"net/http"
)

func (api *API) DevExecTaskWithWorkers(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	code, err := api.Svc.DevExecTaskWithWorkers(ctx, r.FormValue("workspace_id"))

	if err != nil {
		api.ReturnJSONError(w, code, err)
		return
	}

	// auto refresh
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html><html><head><meta charset="UTF-8"><meta http-equiv="refresh" content="1"><title>auto refresh</title></head><body>done</body></html>`))
}
