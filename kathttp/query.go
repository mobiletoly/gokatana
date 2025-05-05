package kathttp

import (
	"github.com/labstack/echo/v4"
	"github.com/mobiletoly/gokatana/katapp"
	"strconv"
)

func QueryInt(c echo.Context, key string, defaultValue int64) int64 {
	value := c.QueryParam(key)
	if value == "" {
		return defaultValue
	}
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		katapp.Logger(c.Request().Context()).WarnContext(c.Request().Context(),
			"failed to parse query int param", "key", key, "cause", err)
		return defaultValue
	}
	return n
}

func QueryFloat(c echo.Context, key string, defaultValue float64) float64 {
	value := c.QueryParam(key)
	if value == "" {
		return defaultValue
	}
	n, err := strconv.ParseFloat(value, 64)
	if err != nil {
		katapp.Logger(c.Request().Context()).WarnContext(c.Request().Context(),
			"failed to parse query float param", "key", key, "cause", err)
		return defaultValue
	}
	return n
}

func QueryPage(c echo.Context) Page {
	offset := QueryInt(c, "offset", 0)
	limit := QueryInt(c, "limit", 50)
	return Page{
		Offset: offset,
		Limit:  limit,
	}
}
