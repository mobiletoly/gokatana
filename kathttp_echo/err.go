package kathttp_echo

import (
	"github.com/labstack/echo/v4"
	"github.com/mobiletoly/gokatana/kathttp"
)

// ReportHTTPError writes an error response to the HTTP response writer
func ReportHTTPError(err error) *echo.HTTPError {
	var errResp = kathttp.GuessHTTPError(err)
	return echo.NewHTTPError(errResp.HTTPStatusCode, errResp)
}
