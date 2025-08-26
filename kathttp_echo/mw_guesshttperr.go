package kathttp_echo

import (
	"github.com/labstack/echo/v4"
	"github.com/mobiletoly/gokatana/katapp"
)

// GuessHTTPErrorMiddleware provides middleware for guessing HTTP errors
func GuessHTTPErrorMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := next(c)
		if err != nil {
			ctx := c.Request().Context()
			err = ReportHTTPError(err)
			katapp.Logger(ctx).Error("HTTP error reported", "error", err,
				"URL", c.Request().URL, "method", c.Request().Method)
			return err
		}
		return nil
	}
}
