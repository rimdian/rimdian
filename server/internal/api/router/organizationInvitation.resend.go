package api

// import (
// 	"encoding/json"
// 	"io"
// 	"net/http"

// 	"github.com/rimdian/rimdian/internal/api/dto"
// 	"github.com/rimdian/rimdian/internal/common/auth"
// )

// // OrganizationInvitationCreate
// func (api *API) OrganizationInvitationResend(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()

// 	claims, err := auth.GetAccountTokenClaimsFromContext(ctx)
// 	if err != nil {
// 		api.ReturnJSONError(w, http.StatusUnauthorized, err)
// 		return
// 	}

// 	// Read body

// 	body, err := io.ReadAll(r.Body)

// 	if err != nil {
// 		http.Error(w, "cannot read body", http.StatusUnprocessableEntity)
// 		return
// 	}

// 	data := &dto.OrganizationInvitationResend{}

// 	// drop data if payload doesnt match
// 	if err := json.Unmarshal(body, data); err != nil {
// 		api.ReturnJSONError(w, http.StatusBadRequest, err)
// 		return
// 	}

// 	code, err := api.Svc.OrganizationInvitationResend(ctx, claims.AccountID, data)

// 	if err != nil {
// 		api.ReturnJSONError(w, code, err)
// 		return
// 	}

// 	ReturnJSON(w, http.StatusOK, map[string]bool{"success": true})
// }
