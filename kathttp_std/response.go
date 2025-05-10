package kathttp_std

import (
	"encoding/json"
	"net/http"
)

// WriteJSON writes a JSON response to the HTTP response writer
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Set status code
	w.WriteHeader(statusCode)

	// Encode and write the response
	return json.NewEncoder(w).Encode(data)
}
