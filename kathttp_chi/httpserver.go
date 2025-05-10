package kathttp_chi

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/mobiletoly/gokatana/katapp"
	"github.com/mobiletoly/gokatana/kathttp_std"
	"log/slog"
	"net/http"
)

// Start initializes and starts an HTTP server using the Chi router
func Start(
	ctx context.Context,
	cfg *katapp.ServerConfig,
	logger *slog.Logger,
	setup func(r *chi.Mux) http.Handler,
) *http.Server {
	inTest := katapp.RunningInTest(ctx)

	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Add CORS middleware
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc: func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	if cfg.ResponseCompression == "gzip" {
		r.Use(middleware.Compress(5))
	}
	if cfg.RequestDecompression == "request-gzip" {
		r.Use(kathttp_std.GzipDecompressMiddleware)
	}

	// Setup routes
	handler := setup(r)

	// Add request context middleware
	r.Use(reqContextMiddleware(logger, inTest))

	// Create the HTTP server
	listenAddr := fmt.Sprintf("%s:%d", cfg.Addr, cfg.Port)
	katapp.Logger(ctx).InfoContext(ctx, fmt.Sprintf("Starting server on %s", listenAddr))

	server := &http.Server{
		Addr:    listenAddr,
		Handler: handler,
	}

	// Start the server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			katapp.Logger(ctx).InfoContext(ctx, "shutting down the server: ", "error", err)
		}
	}()

	return server
}

// Shutdown gracefully shuts down the server
func Shutdown(ctx context.Context, server *http.Server) error {
	return server.Shutdown(ctx)
}
