package kathttp

//
//import (
//	"fmt"
//	"github.com/labstack/echo/v4"
//	"github.com/mobiletoly/gokatana/katapp"
//	"strconv"
//)
//
//// QueryInt retrieves an integer query parameter by key from the provided Echo context.
//// Returns a pointer to the parsed int64 value or nil if the parameter is absent.
//// Returns katapp.ErrInvalidInput if the parameter fails to parse as an integer.
//func QueryInt(c echo.Context, key string) (*int64, error) {
//	value := c.QueryParam(key)
//	if value == "" {
//		return nil, nil
//	}
//	n, err := strconv.ParseInt(value, 10, 64)
//	if err != nil {
//		return nil, katapp.NewErr(
//			katapp.ErrInvalidInput, fmt.Sprintf("failed to parse query int param '%s': %s", key, err.Error()))
//	}
//	return &n, nil
//}
//
//func QueryFloat(c echo.Context, key string) (*float64, error) {
//	value := c.QueryParam(key)
//	if value == "" {
//		return nil, nil
//	}
//	n, err := strconv.ParseFloat(value, 64)
//	if err != nil {
//		return nil, katapp.NewErr(
//			katapp.ErrInvalidInput, fmt.Sprintf("failed to parse query float param '%s': %s", key, err.Error()))
//	}
//	return &n, nil
//}
