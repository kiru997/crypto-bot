package http_client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

var DefaultClient = &http.Client{}

func JSONRequest(ctx context.Context, url, method string, body interface{}, headers map[string]string) (*http.Response, error) {
	var bBody []byte
	var err error
	if body != nil {
		bBody, err = json.Marshal(&body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(bBody))
	req.Header.Add("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	response, err := DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func RawRequest(ctx context.Context, url, method string, body io.Reader, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, errors.Wrap(err, "NewRequestWithContext")
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	response, err := DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "DefaultClient.Do")
	}

	return response, nil
}
