package model

import "github.com/google/uuid"

type URL struct {
	UID      string
	Code     string
	Original string
	UserID   string
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

func NewURL(code, original string, uid *string, userID string) *URL {
	url := &URL{
		Code:     code,
		Original: original,
		UID:      uuid.NewString(),
		UserID:   userID,
	}

	if uid != nil {
		url.UID = *uid
	}

	return url
}
