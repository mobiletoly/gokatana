package kathttp

import (
	"context"
	"errors"
	"fmt"
	"github.com/mobiletoly/gokatana/katapp"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Run(
	ctx context.Context,
	cfg *katapp.ServerConfig,
	logger *slog.Logger,
	setup func(e *echo.Echo),
) {
	inTest := katapp.RunningInTest(ctx)

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())
	if cfg.ResponseCompression == "gzip" {
		e.Use(middleware.GzipWithConfig(middleware.GzipConfig{}))
	}
	if cfg.RequestDecompression == "request-gzip" {
		e.Use(gzipDecompressMiddleware)
	}

	setup(e)
	e.Use(reqContextMiddleware(logger, inTest))

	listenAddr := fmt.Sprintf("%s:%d", cfg.Addr, cfg.Port)
	katapp.Logger(ctx).InfoContext(ctx, fmt.Sprintf("Starting server on %s", listenAddr))
	go func() {
		if err := e.Start(listenAddr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			katapp.Logger(ctx).InfoContext(ctx, "shutting down the server: ", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	katapp.Logger(ctx).InfoContext(ctx, "ServerConfig is shutting down")
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		katapp.Logger(ctx).WarnContext(ctx, "Could not gracefully shutdown the server")
	}
	katapp.Logger(ctx).InfoContext(ctx, "ServerConfig stopped")
}
