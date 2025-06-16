package kathttp_echo

import (
	"github.com/labstack/echo/v4"
	"github.com/mobiletoly/gokatana/kathttp"
	"net/http"
)

// ReportHTTPError writes an error response to the HTTP response writer
func ReportHTTPError(err error) *echo.HTTPError {
	var errResp = kathttp.GuessHTTPError(err)
	return echo.NewHTTPError(errResp.HTTPStatusCode, errResp)
}

func ReportBadRequest(err error) *echo.HTTPError {
	return echo.NewHTTPError(http.StatusBadRequest, kathttp.NewBadRequestErrResponse(err))
}

func ReportNotFound(err error) *echo.HTTPError {
	return echo.NewHTTPError(http.StatusNotFound, kathttp.NewNotFoundErrResponse(err))
}

func ReportConflict(err error) *echo.HTTPError {
	return echo.NewHTTPError(http.StatusConflict, kathttp.NewConflictErrResponse(err))
}

func ReportBadGateway(err error) *echo.HTTPError {
	return echo.NewHTTPError(http.StatusBadGateway, kathttp.NewBadGatewayErrResponse(err))
}

func ReportUnauthorized(err error) *echo.HTTPError {
	return echo.NewHTTPError(http.StatusUnauthorized, kathttp.NewUnauthorizedErrResponse(err))
}

func ReportForbidden(err error) *echo.HTTPError {
	return echo.NewHTTPError(http.StatusForbidden, kathttp.NewForbiddenErrResponse(err))
}
