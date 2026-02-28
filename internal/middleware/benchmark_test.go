package g_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/go-chi/chi/v5"
	g "github.com/yandex-practicum/shorten-url/internal/middleware"
)

var gzPool = sync.Pool{
	New: func() interface{} {
		w, _ := gzip.NewWriterLevel(nil, gzip.BestSpeed)
		return w
	},
}

func BenchmarkGzipMiddleware(b *testing.B) {
	router := chi.NewRouter()
	router.Use(g.GzipMiddleware)

	router.Post("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result":"http://localhost:8081/very-long-unique-short-code-for-testing"}`))
	})

	srv := httptest.NewServer(router)
	defer srv.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", srv.URL+"/test", bytes.NewBufferString(`{"url":"https://yandex.ru"}`))
		req.Header.Set("Accept-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			b.Fatal(err)
		}

		// Обязательно вычитываем тело, чтобы закрыть запрос корректно
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// BenchmarkGzipWriterDirect проверяет только накладные расходы на создание Writer
func BenchmarkGzipWriterDirect(b *testing.B) {
	data := []byte(`{"result":"http://localhost:8081/Gf4xjZe"}`)
	b.Run("NewWriter_EveryTime", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer(nil)
			gz := gzip.NewWriter(buf)
			gz.Write(data)
			gz.Close()
		}
	})

	b.Run("SyncPool_Reset", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer(nil)
			gz := gzPool.Get().(*gzip.Writer)
			gz.Reset(buf)
			gz.Write(data)
			gz.Close()
			gzPool.Put(gz)
		}
	})
}
