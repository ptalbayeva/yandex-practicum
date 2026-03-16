package g

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/yandex-practicum/shorten-url/cmd/pool"
)

// pooledWriter wraps gzip.Writer to satisfy the pool.Resetter interface
type pooledWriter struct {
	*gzip.Writer
}

func (p *pooledWriter) Reset() {
	p.Writer.Reset(io.Discard) // Detach from the previous response writer
}

// Initialize the generic pool
var gzPool = pool.New[gzip.Writer](func() *pooledWriter {
	w, _ := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
	return &pooledWriter{Writer: w}
})

type gzipResponseWriter struct {
	http.ResponseWriter
	gzWriter *pooledWriter
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	if g.Header().Get("Content-Encoding") != "gzip" {
		return g.ResponseWriter.Write(b)
	}

	if g.gzWriter == nil {
		g.gzWriter = gzPool.Get()
		g.gzWriter.Writer.Reset(g.ResponseWriter)
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

// GzipMiddleware сжатие
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
		gzResponse := &gzipResponseWriter{ResponseWriter: w}
		defer gzResponse.Close()

		next.ServeHTTP(gzResponse, r)
	})
}
