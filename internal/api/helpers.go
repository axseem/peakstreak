package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to write json response", "error", err)
	}
}

func errorResponse(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func readJSON(r *http.Request, dst any) error {
	return json.NewDecoder(r.Body).Decode(dst)
}

func validationErrorResponse(w http.ResponseWriter, err error) {
	var verr validator.ValidationErrors
	if ok := errors.As(err, &verr); ok {
		errors := make(map[string]string)
		for _, fe := range verr {
			errors[fe.Field()] = "failed on the '" + fe.Tag() + "' tag"
		}
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "validation failed", "details": errors})
		return
	}
	errorResponse(w, http.StatusBadRequest, "invalid request body")
}
