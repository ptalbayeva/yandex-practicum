package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	return g.writer.Write(b)
}

func (g *gzipResponseWriter) Close() error {
	return g.writer.Close()
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gzReader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "invalid gzip body", http.StatusBadRequest)
				return
			}
			r.Body = gzReader
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")

		gzWriter := gzip.NewWriter(w)
		defer gzWriter.Close()

		gzResponse := &gzipResponseWriter{
			ResponseWriter: w,
			writer:         gzWriter,
		}

		next.ServeHTTP(gzResponse, r)
	})
}
