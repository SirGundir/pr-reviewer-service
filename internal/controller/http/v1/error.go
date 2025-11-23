package v1

import (
	"encoding/json"
	"net/http"
)

type errorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func newErrorResponse(code, message string) errorResponse {
	var resp errorResponse
	resp.Error.Code = code
	resp.Error.Message = message
	return resp
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, code, message string) {
	respondJSON(w, status, newErrorResponse(code, message))
}
