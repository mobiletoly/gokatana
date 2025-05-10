package kathttp_std

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// gzipResponseWriter implements http.ResponseWriter and adds gzip compression
type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code and delegates to the underlying ResponseWriter
func (w *gzipResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Write compresses the data and writes it to the underlying ResponseWriter
func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// Status returns the HTTP status code of the response
func (w *gzipResponseWriter) Status() int {
	if w.statusCode == 0 {
		return http.StatusOK
	}
	return w.statusCode
}

// GzipDecompressMiddleware provides custom middleware to decompress optional gzip payloads
func GzipDecompressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Cannot decompress gzip payload", http.StatusBadRequest)
				return
			}
			defer gz.Close()
			// Replace the request body with the decompressed data
			r.Body = io.NopCloser(gz)
		}
		next.ServeHTTP(w, r)
	})
}

// gzipCompressMiddleware provides response compression using gzip
func gzipCompressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts gzip encoding
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Set content encoding header
		w.Header().Set("Content-Encoding", "gzip")

		// Create gzip writer
		gz, err := gzip.NewWriterLevel(w, gzip.DefaultCompression)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		defer gz.Close()

		// Create gzip response writer
		gzw := &gzipResponseWriter{
			Writer:         gz,
			ResponseWriter: w,
			statusCode:     0,
		}

		// Call the next handler with the gzip response writer
		next.ServeHTTP(gzw, r)
	})
}
