package kathttp_std

import (
	"context"
	"errors"
	"fmt"
	"github.com/mobiletoly/gokatana/katapp"
	"log/slog"
	"net/http"
	"time"
)

// Start initializes and starts an HTTP server using the standard Go http package
func Start(
	ctx context.Context,
	cfg *katapp.ServerConfig,
	logger *slog.Logger,
	setup func(mux *http.ServeMux) http.Handler,
) *http.Server {
	inTest := katapp.RunningInTest(ctx)
	router := http.NewServeMux()

	// Setup routes
	handler := setup(router)

	// Add middleware in reverse order (last added is executed first)
	handler = reqContextMiddleware(logger, inTest)(handler)

	if cfg.RequestDecompression == "request-gzip" {
		handler = GzipDecompressMiddleware(handler)
	}
	if cfg.ResponseCompression == "gzip" {
		handler = gzipCompressMiddleware(handler)
	}

	handler = requestIDMiddleware(handler)
	handler = corsMiddleware(handler)
	handler = recoveryMiddleware(handler)

	// Create the HTTP server
	listenAddr := fmt.Sprintf("%s:%d", cfg.Addr, cfg.Port)
	server := &http.Server{
		Addr:    listenAddr,
		Handler: handler,
	}

	// Start the server in a goroutine
	katapp.Logger(ctx).InfoContext(ctx, fmt.Sprintf("Starting server on %s", listenAddr))
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			katapp.Logger(ctx).InfoContext(ctx, "shutting down the server: ", "error", err)
		}
	}()

	return server
}

// WaitForInterruptSignal waits for interrupt signal to gracefully shut down the server with a timeout.
func WaitForInterruptSignal(ctx context.Context, server *http.Server, timeout time.Duration) {
	katapp.WaitForInterruptSignal(ctx, timeout, func() error {
		return server.Shutdown(ctx)
	})
}

// recoveryMiddleware provides panic recovery similar to Echo's middleware.Recover()
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Internal Server Error")
				// Log the panic
				katapp.Logger(r.Context()).ErrorContext(r.Context(), "panic recovered", "error", err)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// corsMiddleware provides CORS support similar to Echo's middleware.CORS()
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// requestIDMiddleware adds a request ID to the response headers
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request already has an ID
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			// Generate a new request ID (in a real implementation, you'd use a UUID)
			requestID = fmt.Sprintf("%d", time.Now().UnixNano())
		}

		// Add the request ID to the response
		w.Header().Set("X-Request-ID", requestID)

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
