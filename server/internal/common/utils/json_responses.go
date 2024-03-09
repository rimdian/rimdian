package utils

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

func ReturnJSONError(w http.ResponseWriter, logger *logrus.Logger, code int, err error) {
	logger.Error(err)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Code: code, Message: err.Error()})
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
