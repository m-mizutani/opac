package opaclient

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/m-mizutani/goerr"
)

type Client struct {
	baseURL    string
	httpClient HTTPClient
}

func New(baseURL string, options ...Option) (*Client, error) {
	client := &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{},
	}

	for _, opt := range options {
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	return client, nil
}

func (x *Client) request(ctx context.Context, method, url string, data io.Reader, dst interface{}) error {
	req, err := http.NewRequestWithContext(ctx, method, url, data)
	if err != nil {
		return ErrInvalidInput.Wrap(err)
	}

	if data != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	httpResp, err := x.httpClient.Do(req)
	if err != nil {
		return ErrRequestFailed.Wrap(err)
	}

	defer httpResp.Body.Close()
	if httpResp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(httpResp.Body)
		return goerr.Wrap(ErrRequestFailed, "status code is not OK").
			With("code", httpResp.StatusCode).
			With("body", string(body))
	}

	raw, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return ErrUnexpectedResp.Wrap(err).With("body", string(raw))
	}

	if err := json.Unmarshal(raw, dst); err != nil {
		return ErrUnexpectedResp.Wrap(err).With("body", string(raw))
	}

	return nil
}
