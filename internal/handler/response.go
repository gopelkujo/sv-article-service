// Package handler exposes HTTP handlers for the article REST API.
// Handlers translate HTTP requests/responses and delegate to services.
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gopelkujo/sv-article-service/internal/validator"
)

// Envelope is the consistent JSON response shape for all endpoints.
type Envelope struct {
	Success bool       `json:"success"`
	Data    any        `json:"data"`
	Error   *ErrorBody `json:"error"`
}

// ErrorBody describes a failed response.
type ErrorBody struct {
	Message string                 `json:"message"`
	Details []validator.FieldError `json:"details"`
}

// writeJSON writes v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeSuccess writes a successful envelope response.
func writeSuccess(w http.ResponseWriter, status int, data any) {
	writeJSON(w, status, Envelope{
		Success: true,
		Data:    data,
		Error:   nil,
	})
}

// writeError writes a failed envelope response.
func writeError(w http.ResponseWriter, status int, message string, details []validator.FieldError) {
	if details == nil {
		details = []validator.FieldError{}
	}
	writeJSON(w, status, Envelope{
		Success: false,
		Data:    nil,
		Error: &ErrorBody{
			Message: message,
			Details: details,
		},
	})
}
