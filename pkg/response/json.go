package response

import (
	"encoding/json"
	"net/http"
)

// JSON writes data as a JSON response with the given status code.
func JSON(w http.ResponseWriter, _ *http.Request, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data) //nolint:errcheck
}

// Created writes a 201 JSON response and sets the Location header.
func Created(w http.ResponseWriter, r *http.Request, location string, data any) {
	w.Header().Set("Location", location)
	JSON(w, r, http.StatusCreated, data)
}
