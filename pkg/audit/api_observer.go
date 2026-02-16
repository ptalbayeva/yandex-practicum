package audit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type ApiObserver struct {
	client *http.Client
	url    string
}

func NewApiObserver(url string) *ApiObserver {
	return &ApiObserver{
		client: &http.Client{
			Timeout: time.Second * 5,
		},
		url: url,
	}
}

func (a ApiObserver) Publish(event Event) error {
	data, err := json.Marshal(event)

	if err != nil {
		return err
	}

	resp, err := a.client.Post(a.url, "application/json", bytes.NewBuffer(data))

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}
