package model

import "github.com/google/uuid"

type URL struct {
	UID       string `json:"uid"`
	Code      string `json:"short"`
	Original  string `json:"original"`
	UserID    string `json:"user_id"`
	IsDeleted bool   `json:"is_deleted"`
}

type DeleteURLTask struct {
	UserID string `json:"user_id"`
	Short  string `json:"short"`
}

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

type BatchURLRequest struct {
	CorrelationID *string `json:"correlation_id"`
	OriginalURL   string  `json:"original_url"`
}

type BatchURLResponse struct {
	CorrelationID *string `json:"correlation_id"`
	ShortenURL    string  `json:"short_url"`
}

type URLResponse struct {
	ShortenURL  string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

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
