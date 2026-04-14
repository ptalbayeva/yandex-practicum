package model

import "github.com/google/uuid"

// URL объект урла
type URL struct {
	UID       string `json:"uid"`
	Code      string `json:"short"`
	Original  string `json:"original"`
	UserID    string `json:"user_id"`
	IsDeleted bool   `json:"is_deleted"`
}

// DeleteURLTask удаляемая запись
type DeleteURLTask struct {
	UserID string `json:"user_id"`
	Short  string `json:"short"`
}

// Request объект запроса
type Request struct {
	URL string `json:"url"`
}

// Response объект ответа
type Response struct {
	Result string `json:"result"`
}

// BatchURLRequest объект несколько запросов на сокращение
type BatchURLRequest struct {
	CorrelationID *string `json:"correlation_id"`
	OriginalURL   string  `json:"original_url"`
}

// BatchURLResponse объект ответа несколько сокращенных url-ов
type BatchURLResponse struct {
	CorrelationID *string `json:"correlation_id"`
	ShortenURL    string  `json:"short_url"`
}

// URLResponse ответ сокращенного запроса
type URLResponse struct {
	ShortenURL  string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// StatsResponse ответ статистики по сокращенному урлу и пользователей в сервисе
type StatsResponse struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}

// NewURL создание объекта url-а
func NewURL(code, original string, uid *string, userID string, isDeleted bool) *URL {
	url := &URL{
		Code:      code,
		Original:  original,
		UID:       uuid.NewString(),
		UserID:    userID,
		IsDeleted: isDeleted,
	}

	if uid != nil {
		url.UID = *uid
	}

	return url
}
