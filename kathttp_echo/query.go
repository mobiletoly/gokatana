package kathttp_echo

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
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

func QueryOptionalDate(c echo.Context, key string) (*time.Time, error) {
	value := c.QueryParam(key)
	if value == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid '%s' format", key))
	}
	return &t, nil
}

func QueryMandatoryDate(c echo.Context, key string) (time.Time, error) {
	if t, err := QueryOptionalDate(c, key); err != nil {
		return time.Time{}, err
	} else if t == nil {
		return time.Time{}, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("missing mandatory query param '%s'", key))
	} else {
		return *t, nil
	}
}

func FormOptionalDate(c echo.Context, key string, endOfDay bool) (*time.Time, error) {
	value := c.FormValue(key)
	if value == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid '%s' format", key))
	}
	if endOfDay {
		t = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, time.UTC)
	} else {
		t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	}
	return &t, nil
}

func FormMandatoryDate(c echo.Context, key string, endOfDay bool) (time.Time, error) {
	if t, err := FormOptionalDate(c, key, endOfDay); err != nil {
		return time.Time{}, err
	} else if t == nil {
		return time.Time{}, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("missing mandatory form param '%s'", key))
	} else {
		return *t, nil
	}
}
