package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yandex-practicum/shorten-url/internal/model"
	"github.com/yandex-practicum/shorten-url/internal/repository"
	"github.com/yandex-practicum/shorten-url/internal/service"
)

type Handler struct {
	shortener *service.ShortenerService
	db        *sql.DB
}

func NewHandler(s *service.ShortenerService, db *sql.DB) *Handler {
	return &Handler{shortener: s, db: db}
}

func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "content type must be text/plain", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	originalURL := strings.TrimSpace(string(body))
	if originalURL == "" {
		http.Error(w, "missing url", http.StatusBadRequest)
		return
	}

	var fullURL string

	u, err := h.shortener.Shorten(originalURL)

	w.Header().Set("Content-Type", "text/plain")
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			w.WriteHeader(http.StatusConflict)
			fullURL = fmt.Sprintf("%s/%s", h.shortener.BaseURL, u.Code)

			w.Write([]byte(fullURL))
			return
		}

		http.Error(w, "failed to shorten", http.StatusInternalServerError)
		return
	}

	fullURL = fmt.Sprintf("%s/%s", h.shortener.BaseURL, u.Code)
	w.WriteHeader(http.StatusCreated)

	_, err = w.Write([]byte(fullURL))

	if err != nil {
		return
	}
}

func (h *Handler) ShortenJSON(w http.ResponseWriter, r *http.Request) {
	h.validateJSONMethod(w, r)

	var request model.Request
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&request); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	response := model.Response{}
	w.Header().Set("Content-Type", "application/json")

	result, err := h.shortener.Shorten(request.URL)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			response.Result = fmt.Sprintf("%s/%s", h.shortener.BaseURL, result.Code)

			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(response)

			return
		}

		http.Error(w, "failed to shorten", http.StatusUnprocessableEntity)
		return
	}

	response.Result = fmt.Sprintf("%s/%s", h.shortener.BaseURL, result.Code)

	w.WriteHeader(http.StatusCreated)

	enc := json.NewEncoder(w)

	if fail := enc.Encode(response); fail != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) BatchShorten(w http.ResponseWriter, r *http.Request) {
	h.validateJSONMethod(w, r)

	var reqs []model.BatchURLRequest
	if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	results, err := h.shortener.ShortenBatch(reqs)
	if err != nil {
		http.Error(w, "failed to shorten", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if fail := json.NewEncoder(w).Encode(results); fail != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "id")

	u, err := h.shortener.Resolve(code)
	if err != nil || u == nil {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, u.Original, http.StatusTemporaryRedirect)
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := h.db.PingContext(ctx)

	if err != nil {
		http.Error(w, "failed to ping database", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) validateJSONMethod(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "content type must be application/json", http.StatusBadRequest)
		return
	}
}
