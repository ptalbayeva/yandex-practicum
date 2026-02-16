package audit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type APIObserver struct {
	client *http.Client
	url    string
}

func NewAPIObserver(url string) *APIObserver {
	return &APIObserver{
		client: &http.Client{
			Timeout: time.Second * 5,
		},
		url: url,
	}
}

func (a APIObserver) Publish(event Event) error {
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
