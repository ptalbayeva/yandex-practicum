package g

import (
	"compress/gzip"
	"net/http"
	"strings"
	"sync"
)

// Глобальный пул для переиспользования тяжелых объектов gzip.Writer
var gzPool = sync.Pool{
	New: func() interface{} {
		w, _ := gzip.NewWriterLevel(nil, gzip.BestSpeed)
		return w
	},
}

type gzipResponseWriter struct {
	http.ResponseWriter
	gzWriter *gzip.Writer
}

func (g *gzipResponseWriter) WriteHeader(statusCode int) {
	if (statusCode >= 300 && statusCode < 400) || statusCode == http.StatusNoContent || statusCode >= 400 {
		g.Header().Del("Content-Encoding")
	}
	g.ResponseWriter.WriteHeader(statusCode)
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	if g.Header().Get("Content-Encoding") != "gzip" {
		return g.ResponseWriter.Write(b)
	}

	if g.gzWriter == nil {
		gz := gzPool.Get().(*gzip.Writer)
		gz.Reset(g.ResponseWriter)
		g.gzWriter = gz
	}
	return g.gzWriter.Write(b)
}

func (g *gzipResponseWriter) Close() {
	if g.gzWriter != nil {
		g.gzWriter.Close()
		gzPool.Put(g.gzWriter)
		g.gzWriter = nil
	}
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gzReader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "invalid gzip body", http.StatusBadRequest)
				return
			}
			defer gzReader.Close()
			r.Body = gzReader
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		gzResponse := &gzipResponseWriter{
			ResponseWriter: w,
		}

		defer gzResponse.Close()

		next.ServeHTTP(gzResponse, r)
	})
}
