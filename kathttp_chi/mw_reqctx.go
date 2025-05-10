package kathttp_chi

import (
	"github.com/go-chi/chi/v5/middleware"
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

			// Get request ID from header or chi middleware
			requestID := r.Header.Get(middleware.RequestIDHeader)
			if requestID == "" {
				requestID = middleware.GetReqID(ctx)
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
