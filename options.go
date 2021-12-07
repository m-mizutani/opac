package opaclient

import (
	"net/http"
)

type Option func(client *Client) error

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func WithHTTPClient(c HTTPClient) Option {
	return func(client *Client) error {
		client.client = c
		return nil
	}
}
