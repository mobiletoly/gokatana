package kathttp_echo

import (
	"compress/gzip"
	"github.com/mobiletoly/gokatana/katapp"
	"io"

	"github.com/labstack/echo/v4"
)

// gzipDecompressMiddleware provides custom middleware to decompress optional gzip payloads
func gzipDecompressMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Request().Header.Get("content-encoding") == "gzip" {
			gz, err := gzip.NewReader(c.Request().Body)
			if err != nil {
				return ReportHTTPError(katapp.NewErr(katapp.ErrInvalidInput, "cannot decompress gzip payload"))
			}
			defer gz.Close()
			// Replace the request body with the decompressed data
			c.Request().Body = io.NopCloser(gz)
		}
		return next(c)
	}
}
