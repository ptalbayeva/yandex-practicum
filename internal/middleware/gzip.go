package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type compressWriter struct {
	w  http.ResponseWriter
	zr *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	zr := gzip.NewWriter(w)

	return &compressWriter{
		w:  w,
		zr: zr,
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(b []byte) (int, error) {
	return c.zr.Write(b)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	return c.zr.Close()
}

type compressReader struct {
	r  io.Reader
	zr *gzip.Reader
}

func newCompressReader(r io.Reader) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{r: r, zr: zr}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c compressReader) Close() error {
	return c.zr.Close()
}

func GzipHandler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			contentType := req.Header.Get("Content-Type")
			if !strings.Contains(contentType, "text/html") &&
				!strings.Contains(contentType, "application/json") {
				next.ServeHTTP(w, req)
				return
			}

			ow := w
			acceptEncoding := req.Header.Get("Accept-Encoding")
			supportsGzip := strings.Contains(acceptEncoding, "gzip")
			if supportsGzip {
				cw := newCompressWriter(w)
				ow = cw
				defer cw.Close()
			}

			contentEncoding := req.Header.Get("Content-Encoding")
			supportsCompress := strings.Contains(contentEncoding, "gzip")
			if supportsCompress {
				cr, err := newCompressReader(req.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				req.Body = cr
				defer cr.Close()
			}

			next.ServeHTTP(ow, req)
		})
	}
}
