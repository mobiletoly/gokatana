package kathttp_echo

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
)

func QueryOptionalUUID(c echo.Context, key string) (*uuid.UUID, error) {
	value := c.QueryParam(key)
	if value == "" {
		return nil, nil
	}
	id, err := uuid.Parse(value)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid '%s' format", key))
	}
	return &id, nil
}

func QueryMandatoryUUID(c echo.Context, key string) (uuid.UUID, error) {
	if id, err := QueryOptionalUUID(c, key); err != nil {
		return uuid.Nil, err
	} else if id == nil {
		return uuid.Nil, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("missing mandatory query param '%s'", key))
	} else {
		return *id, nil
	}
}

func UUIDParam(c echo.Context, key string) (uuid.UUID, error) {
	value := c.Param(key)
	if value == "" {
		return uuid.Nil, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("missing mandatory path param '%s'", key))
	}
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid path param '%s' format", key))
	}
	return id, nil
}
