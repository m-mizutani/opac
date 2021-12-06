package opaclient

import (
	"context"
	"io"
	"net/http"
)

type Option func(client *Client) error

type HTTPRequest func(ctx context.Context, method, url string, data io.Reader) (*http.Response, error)

// OptHTTPRequest sets HTTPRequest sender
func OptHTTPRequest(f HTTPRequest) Option {
	return func(client *Client) error {
		client.httpRequest = f
		return nil
	}
}
