package handler

import (
	"net/http"

	"github.com/yandex-practicum/shorten-url/internal/service"
)

type DBHandler struct {
	db *service.DBService
}

func NewDBHandler(db *service.DBService) *DBHandler {
	return &DBHandler{db}
}

func (h *DBHandler) Ping(w http.ResponseWriter, r *http.Request) {
	err := h.db.Ping()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	return
}
