package handler

import (
	"encoding/json"
	"net/http"

	chimw "github.com/go-chi/chi/v5/middleware"
)

type ErrorResponse struct {
	Error     string `json:"error"`
	Code      string `json:"code"`
	RequestID string `json:"request_id,omitempty"`
}

func respondError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Error:     message,
		Code:      code,
		RequestID: chimw.GetReqID(r.Context()),
	})
}
