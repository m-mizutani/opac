package opac

import (
	"net/http"

	"github.com/m-mizutani/zlog"
)

type Option func(client *Client) error

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func WithHTTPClient(httpClient HTTPClient) Option {
	return func(client *Client) error {
		client.httpClient = httpClient
		return nil
	}
}

func WithZLog(l *zlog.Logger) Option {
	return func(client *Client) error {
		logger = l
		return nil
	}
}
