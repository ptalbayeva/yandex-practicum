package model

type URL struct {
	Code     string
	Original string
}

type Request struct {
	Url string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

func NewURL(code, original string) *URL {
	return &URL{
		Code:     code,
		Original: original,
	}
}
