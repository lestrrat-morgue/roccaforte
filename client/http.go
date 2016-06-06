package client

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/lestrrat/roccaforte/event"
	"github.com/pkg/errors"
)

type HTTP struct {
	endpoint string
	Client   *http.Client
}

func NewHTTP(endpoint string) *HTTP {
	return &HTTP{
		endpoint: endpoint,
		Client:   &http.Client{},
	}
}

func (cl *HTTP) Enqueue(events ...event.Event) error {
	buf := bytes.Buffer{}
	if err := json.NewEncoder(&buf).Encode(events); err != nil {
		return errors.Wrap(err, "failed to serialize data")
	}

	req, err := http.NewRequest("POST", cl.endpoint, &buf)
	if err != nil {
		return errors.Wrap(err, "failed to crete HTTP request")
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := cl.Client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to POST message")
	}

	if res.StatusCode >= 300 {
		body := bytes.Buffer{}
		io.Copy(&body, res.Body)
		return errors.Wrap(errors.New(body.String()), "server responded with error")
	}

	return nil
}
