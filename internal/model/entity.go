package model

type URL struct {
	Code     string
	Original string
}

func New(code, original string) *URL {
	return &URL{
		Code:     code,
		Original: original,
	}
}
