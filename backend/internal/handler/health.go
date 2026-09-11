package handler

import (
	"encoding/json"
	"net/http"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
)

func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":     "ok",
		"service":    "tablebite-api",
		"time":       time.Now().UTC(),
		"request_id": chimw.GetReqID(r.Context()),
	})
}
