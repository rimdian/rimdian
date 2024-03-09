package api

import (
	"net/http"
)

// Resets the DB in dev env
func (api *API) DevResetDB(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := api.Svc.DevResetDB(ctx)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("done"))
}
