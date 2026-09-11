package middleware

import (
	"encoding/json"
	"net/http"

	chimw "github.com/go-chi/chi/v5/middleware"
)

func RespondError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":      message,
		"code":       code,
		"request_id": chimw.GetReqID(r.Context()),
	})
}
