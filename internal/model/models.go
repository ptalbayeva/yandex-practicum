package model

import "github.com/google/uuid"

type URL struct {
	UID      string
	Code     string
	Original string
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
	ShortenURL    string  `json:"shorten_url"`
}

func NewURL(code, original string, uid *string) *URL {
	url := &URL{
		Code:     code,
		Original: original,
		UID:      uuid.NewString(),
	}

	if uid != nil {
		url.UID = *uid
	}

	return url
}
