package handler

import (
	"io"
	"net/http"
	"strings"

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
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "text/plain" {
		w.WriteHeader(http.StatusBadRequest)
	}

	body, err := io.ReadAll(r.Body)
	originalURL := strings.TrimSpace(string(body))

	if originalURL == "" {
		http.Error(w, "Missing URL", http.StatusBadRequest)
		return
	}

	u, err := h.shortener.Shorten(string(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	short := "http://localhost:8080/" + u.Code + "/"

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	_, err = w.Write([]byte(short))

	if err != nil {
		http.Error(w, "Error shortening URL", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	code := r.PathValue("id")

	u, err := h.shortener.Resolve(code)
	if err != nil {
		http.Error(w, "Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, u.Original, http.StatusTemporaryRedirect)
}
