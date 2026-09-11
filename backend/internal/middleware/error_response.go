package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	chimw "github.com/go-chi/chi/v5/middleware"
)

type errorResponseWriter struct {
	http.ResponseWriter
	ctx    context.Context
	status int
}

func (w *errorResponseWriter) WriteHeader(status int) {
	w.status = status
	if status >= http.StatusInternalServerError || status >= http.StatusBadRequest {
		return
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *errorResponseWriter) Write(body []byte) (int, error) {
	contentType := w.Header().Get("Content-Type")
	if w.status >= http.StatusBadRequest && strings.HasPrefix(contentType, "text/plain") {
		message := strings.TrimSpace(string(bytes.TrimSuffix(body, []byte("\n"))))
		if w.status >= http.StatusInternalServerError {
			message = "internal server error"
		}
		w.Header().Set("Content-Type", "application/json")
		payload := map[string]string{
			"error":      message,
			"code":       codeForStatus(w.status),
			"request_id": chimw.GetReqID(w.ctx),
		}
		encoded, err := json.Marshal(payload)
		if err != nil {
			return 0, err
		}
		w.ResponseWriter.WriteHeader(w.status)
		return w.ResponseWriter.Write(append(encoded, '\n'))
	}
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(body)
}

func JSONErrors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(&errorResponseWriter{ResponseWriter: w, ctx: r.Context()}, r)
	})
}

func RequestIDHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requestID := chimw.GetReqID(r.Context()); requestID != "" {
			w.Header().Set("X-Request-ID", requestID)
		}
		next.ServeHTTP(w, r)
	})
}

func codeForStatus(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "bad_request"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	default:
		return "internal_error"
	}
}
