package kathttp_std

import (
	"github.com/mobiletoly/gokatana/katapp"
	"log/slog"
	"net/http"
)

// reqContextMiddleware provides middleware for inserting required context values
func reqContextMiddleware(logger *slog.Logger, runInTest bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Add run in test flag if needed
			if runInTest {
				ctx = katapp.ContextWithRunInTest(ctx, true)
			}

			// Get request ID from header
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = w.Header().Get("X-Request-ID")
			}

			// Add logger with request ID to context
			ctx = katapp.ContextWithRequestLogger(ctx, logger, requestID)

			// Create a new request with the updated context
			r = r.WithContext(ctx)

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}
