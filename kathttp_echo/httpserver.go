package kathttp_echo

import (
	"context"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mobiletoly/gokatana/katapp"
	"log/slog"
	"net/http"
)

func Start(
	ctx context.Context,
	cfg *katapp.ServerConfig,
	logger *slog.Logger,
	setup func(e *echo.Echo),
) *echo.Echo {
	inTest := katapp.RunningInTest(ctx)

	e := echo.New()
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		DisableStackAll:   false,
		DisablePrintStack: false,
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			katapp.Logger(c.Request().Context()).Errorf("[PANIC RECOVER] err=%v\nstack=%s", err, stack)
			return err
		},
	}))
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

	return e
}
