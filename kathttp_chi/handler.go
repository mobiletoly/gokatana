package kathttp_chi

import (
	"github.com/go-chi/chi/v5"
	"github.com/mobiletoly/gokatana/kathttp_std"
	"net/http"
)

// Bind binds the request body to the given struct
func Bind(r *http.Request, v interface{}) error {
	return kathttp_std.Bind(r, v)
}

// WriteJSON writes a JSON response to the HTTP response writer
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	return kathttp_std.WriteJSON(w, statusCode, data)
}

// URLParam gets a URL parameter by name
func URLParam(r *http.Request, name string) string {
	return chi.URLParam(r, name)
}
