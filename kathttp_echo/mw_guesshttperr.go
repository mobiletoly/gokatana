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
			katapp.Logger(ctx).Errorf("HTTP error reported: %v", err)
			return err
		}
		return nil
	}
}
