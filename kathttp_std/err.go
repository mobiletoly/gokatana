package kathttp_std

import (
	"encoding/json"
	"github.com/mobiletoly/gokatana/kathttp"
	"net/http"
)

// ReportHTTPError writes an error response to the HTTP response writer
func ReportHTTPError(w http.ResponseWriter, err error) {
	errResp := kathttp.GuessHTTPError(err)
	// Set content type and status code
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errResp.HTTPStatusCode)

	// Write the error response as JSON
	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
