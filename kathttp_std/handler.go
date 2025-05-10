package kathttp_std

import (
	"encoding/json"
	"github.com/mobiletoly/gokatana/katapp"
	"io"
	"net/http"
	"strings"
)

// Bind binds the request body to the given struct
func Bind(r *http.Request, v interface{}) error {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return katapp.NewErr(katapp.ErrInternal, "failed to read request body")
	}
	defer r.Body.Close()

	// Check content type
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if err := json.Unmarshal(body, v); err != nil {
			return katapp.NewErr(katapp.ErrInvalidInput, "failed to parse JSON request body")
		}
		return nil
	}

	return katapp.NewErr(katapp.ErrInvalidInput, "unsupported content type")
}
