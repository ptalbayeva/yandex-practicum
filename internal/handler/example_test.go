package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/yandex-practicum/shorten-url/internal/repository"
	"github.com/yandex-practicum/shorten-url/internal/service"
	"github.com/yandex-practicum/shorten-url/pkg/audit"
)

// UserURL Модель для парсинга ответа
type UserURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// BatchRequest Структура запроса
type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// BatchResponse Структура ответа
type BatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// Example Примеры запросов
func Example() {
	r := chi.NewRouter()

	repo := repository.NewMemoryRepo()
	deleteURL := service.NewDeleteURLService(repo, 100)
	urlHandler := &Handler{
		shortener:    service.NewShortenerService(repo, testC.BaseURL, *deleteURL),
		auditService: audit.NewNoopPublisher(),
	}

	r.Post("/api/shorten", urlHandler.ShortenURL)
	r.Get("/{id}", urlHandler.ExpandURL)
	r.Get("/api/user/urls", urlHandler.ListUserURLs)

	// 2. Создаем тестовый сервер
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Пример 1: Сокращение ссылки через JSON API
	jsonBody := `{"url": "https://google.com"}`
	resp, _ := http.Post(ts.URL+"/api/shorten", "application/json", strings.NewReader(jsonBody))
	fmt.Printf("POST /api/shorten: %d\n", resp.StatusCode)
	resp.Body.Close()

	// Создаем клиент, который НЕ переходит по ссылкам автоматически
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Возврат этой ошибки говорит клиенту: "Стоп, не иди дальше"
			return http.ErrUseLastResponse
		},
	}

	// Пример 2: Редирект (GET запрос)
	resp, _ = client.Get(ts.URL + "/BQRvJsg")
	fmt.Printf("GET /{id}: %d\n", resp.StatusCode)
	resp.Body.Close()

	// Пример 3: Получение урлов по пользователю без авторизации
	// Добавляем данные авторизации, но получим 401 так как невалидный токен
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/user/urls", nil)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "authorization", Value: "token"})

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("GET /api/user/urls: %d\n", resp.StatusCode)

	var urls []UserURL
	if err := json.NewDecoder(resp.Body).Decode(&urls); err == nil {
		if len(urls) > 0 {
			fmt.Println("Data received: yes")
		}
	}

	// Output:
	// POST /api/shorten: 201
	// GET /{id}: 307
	// GET /api/user/urls: 401
}

// Example_batchShorten Пример при сокращении несколько url
func Example_batchShorten() {
	r := chi.NewRouter()

	repo := repository.NewMemoryRepo()
	deleteURL := service.NewDeleteURLService(repo, 100)
	urlHandler := &Handler{
		shortener:    service.NewShortenerService(repo, testC.BaseURL, *deleteURL),
		auditService: audit.NewNoopPublisher(),
	}

	r.Post("/api/shorten/batch", urlHandler.BatchShorten)

	ts := httptest.NewServer(r)
	defer ts.Close()

	input := []BatchRequest{
		{CorrelationID: "1", OriginalURL: "https://google.com"},
		{CorrelationID: "2", OriginalURL: "https://yandex.ru"},
	}
	body, _ := json.Marshal(input)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Post(ts.URL+"/api/shorten/batch", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	var result []BatchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
		fmt.Printf("Status Code: %d\n", resp.StatusCode)
		fmt.Printf("Items received: %d\n", len(result))
		if len(result) > 0 {
			fmt.Printf("First CorrelationID: %s\n", result[0].CorrelationID)
		}
	}

	// Output:
	// Status Code: 201
	// Items received: 2
	// First CorrelationID: 1
}
