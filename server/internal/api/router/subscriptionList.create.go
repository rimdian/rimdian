package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/rimdian/rimdian/internal/api/dto"
	"github.com/rimdian/rimdian/internal/common/auth"
	"go.opencensus.io/plugin/ochttp"
)

func (api *API) SubscriptionListCreate(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	ochttp.SetRoute(ctx, "/api/subscriptionList.create")

	claims, err := auth.GetAccountTokenClaimsFromContext(ctx)
	if err != nil {
		api.ReturnJSONError(w, http.StatusUnauthorized, err)
		return
	}

	// Read body

	body, err := io.ReadAll(r.Body)

	if err != nil {
		api.ReturnJSONError(w, http.StatusUnprocessableEntity, err)
		return
	}

	data := &dto.SubscriptionListCreate{}

	// drop data if payload doesnt match
	if err := json.Unmarshal(body, data); err != nil {
		api.ReturnJSONError(w, http.StatusBadRequest, err)
		return
	}

	code, err := api.Svc.SubscriptionListCreate(ctx, claims.AccountID, data)

	if err != nil {
		api.ReturnJSONError(w, code, err)
		return
	}

	ReturnJSON(w, http.StatusCreated, nil)
}
