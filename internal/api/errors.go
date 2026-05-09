package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func respondError(w http.ResponseWriter, message string, statusCode int, err error) {
	errorMsg := message
	if err != nil {
		errorMsg = fmt.Sprintf("%s: %v", message, err)
	}

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   message,
		Message: errorMsg,
		Code:    statusCode,
	})
}
