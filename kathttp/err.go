package kathttp

import (
	"context"
	"errors"
	"github.com/mobiletoly/gokatana/katapp"
	"net/http"

	"github.com/labstack/echo/v4"
)

// ErrResponse renderer for HTTP failed response
type ErrResponse struct {
	Err            error  `json:"-"`               // low-level runtime error
	HTTPStatusCode int    `json:"-"`               // http response status code
	StatusText     string `json:"status"`          // user-level status message
	AppCode        int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText      string `json:"error,omitempty"` // application-level error message, for debugging
}

func NewInternalServerErrResponse(err error) *ErrResponse {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusInternalServerError,
		StatusText:     "Internal server error",
		// TODO most likely we don't want to show actual error in case of Internal ServerConfig error to
		// 	the user as it can potentially reveal sensitive information
		ErrorText: err.Error(),
	}
}

func NewBadRequestErrResponse(err error) *ErrResponse {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusBadRequest,
		StatusText:     "Bad request",
		ErrorText:      err.Error(),
	}
}

func NewNotFoundErrResponse(err error) *ErrResponse {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusNotFound,
		StatusText:     "Not found",
		ErrorText:      err.Error(),
	}
}

func NewConflictErrResponse(err error) *ErrResponse {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusConflict,
		StatusText:     "Conflict",
		ErrorText:      err.Error(),
	}
}

func NewBadGatewayErrResponse(err error) *ErrResponse {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusBadGateway,
		StatusText:     "Bad Gateway",
		// TODO most likely we don't want to show actual error in case of Internal ServerConfig error to
		// 	the user as it can potentially reveal sensitive information
		ErrorText: err.Error(),
	}
}

func NewUnauthorizedErrResponse(err error) *ErrResponse {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusUnauthorized,
		StatusText:     "Unauthorized",
		ErrorText:      err.Error(),
	}
}

func GuessHTTPError(ctx context.Context, err error) *echo.HTTPError {
	var appErr *katapp.Err
	if errors.As(err, &appErr) {
		switch appErr.Scope {
		case katapp.ErrUnknown, katapp.ErrInternal:
			return echo.NewHTTPError(http.StatusInternalServerError, NewInternalServerErrResponse(err))
		case katapp.ErrInvalidInput:
			return echo.NewHTTPError(http.StatusBadRequest, NewBadRequestErrResponse(err))
		case katapp.ErrNotFound:
			return echo.NewHTTPError(http.StatusNotFound, NewNotFoundErrResponse(err))
		case katapp.ErrDuplicate:
			return echo.NewHTTPError(http.StatusConflict, NewConflictErrResponse(err))
		case katapp.ErrFailedExternalService:
			return echo.NewHTTPError(http.StatusBadGateway, NewBadGatewayErrResponse(err))
		case katapp.ErrUnauthorized:
			return echo.NewHTTPError(http.StatusUnauthorized, NewUnauthorizedErrResponse(err))
		}
	}
	katapp.Logger(ctx).ErrorContext(ctx, "Unhandled application error", "cause", err)
	return echo.NewHTTPError(http.StatusInternalServerError, NewInternalServerErrResponse(err))
}
