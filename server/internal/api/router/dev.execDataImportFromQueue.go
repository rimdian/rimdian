package api

import (
	"net/http"
	"strconv"
)

// gets a dataImport stored in-memory to process it
// if concurrency=n parameter is provided, it will launch a semaphore to proccess them concurently
func (api *API) DevExecDataImportFromQueue(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	concurrency := 1

	if r.FormValue("concurrency") != "" {
		var err error
		concurrency, err = strconv.Atoi(r.FormValue("concurrency"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
	}

	code, err := api.Svc.DevExecDataImportFromQueue(ctx, concurrency)

	if err != nil {
		api.ReturnJSONError(w, code, err)
		return
	}

	// auto refresh
	w.WriteHeader(http.StatusOK)
	if r.FormValue("auto_refresh") != "" {
		w.Write([]byte(`<!DOCTYPE html><html><head><meta charset="UTF-8"><meta http-equiv="refresh" content="1"><title>auto refresh</title></head><body>done</body></html>`))
	} else {
		w.Write([]byte(`done`))
	}
}
