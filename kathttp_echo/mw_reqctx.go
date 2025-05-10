package kathttp_echo

import (
	"github.com/labstack/echo/v4"
	"github.com/mobiletoly/gokatana/katapp"
	"log/slog"
)

// reqContextMiddleware provides middleware for inserting required context values
func reqContextMiddleware(logger *slog.Logger, runInTest bool) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		fn := func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			ctx := req.Context()
			if runInTest {
				ctx = katapp.ContextWithRunInTest(ctx, true)
			}
			requestID := req.Header.Get(echo.HeaderXRequestID)
			if requestID == "" {
				requestID = res.Header().Get(echo.HeaderXRequestID)
			}
			ctx = katapp.ContextWithRequestLogger(ctx, logger, requestID)
			c.SetRequest(req.WithContext(ctx))
			return next(c)
		}
		return fn
	}
}
