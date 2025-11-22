package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/yandex-practicum/shorten-url/internal/model"
	"github.com/yandex-practicum/shorten-url/internal/service"
)

type Handler struct {
	shortener *service.ShortenerService
}

func NewHandler(s *service.ShortenerService) *Handler {
	return &Handler{shortener: s}
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

	u, err := h.shortener.Shorten(originalURL)
	if err != nil {
		http.Error(w, "failed to shorten", http.StatusInternalServerError)
		return
	}

	fullURL := fmt.Sprintf("%s/%s", h.shortener.BaseURL, u.Code)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(fullURL))

	if err != nil {
		return
	}
}

func (h *Handler) ShortenJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "content type must be application/json", http.StatusBadRequest)
		return
	}

	var request model.Request
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&request); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	result, err := h.shortener.Shorten(request.URL)
	if err != nil || result == nil {
		http.Error(w, "failed to shorten", http.StatusUnprocessableEntity)
		return
	}

	response := model.Response{
		Result: fmt.Sprintf("%s/%s", h.shortener.BaseURL, result.Code),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	enc := json.NewEncoder(w)

	if fail := enc.Encode(response); fail != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
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
