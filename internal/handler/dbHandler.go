package handler

import (
	"net/http"

	"github.com/yandex-practicum/shorten-url/internal/service"
)

type DBHandler struct {
	dsn string
}

func NewDBHandler(dsn string) *DBHandler {
	return &DBHandler{dsn: dsn}
}

func (h *DBHandler) Ping(w http.ResponseWriter, r *http.Request) {
	err := service.Ping(h.dsn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
