package opac

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/m-mizutani/goerr"
	"github.com/m-mizutani/zlog"
)

// Remote sends a HTTP/HTTPS request to OPA server.
type Remote struct {
	url         string
	httpHeaders map[string][]string
	httpClient  HTTPClient
	logger      *zlog.Logger
}

// RemoteOption is Option of functional option pattern for Remote
type RemoteOption func(x *Remote)

// HTTPClient is interface of http.Client for replancement with owned HTTP client
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewRemote creates a new Local client.
func NewRemote(url string, options ...RemoteOption) (*Remote, error) {
	if err := validation.Validate(url, validation.Required, is.URL); err != nil {
		return nil, goerr.Wrap(err, "invalid URL for remote policy")
	}

	client := &Remote{
		url:         url,
		httpHeaders: make(map[string][]string),
		httpClient:  http.DefaultClient,
		logger:      zlog.New(),
	}
	for _, opt := range options {
		opt(client)
	}

	client.logger.
		With("url", client.url).
		With("headers", client.httpHeaders).
		Debug("created remote client")

	return client, nil
}

// WithHTTPClient replaces `http.DefaultClient` with own `HTTPClient` instance.
func WithHTTPClient(client HTTPClient) RemoteOption {
	return func(x *Remote) {
		x.httpClient = client
	}
}

// WithHTTPHeader adds HTTP header. It can be added multiply.
func WithHTTPHeader(name, value string) RemoteOption {
	return func(x *Remote) {
		x.httpHeaders[name] = append(x.httpHeaders[name], value)
	}
}

// WithLoggingRemote enables logger for debug
func WithLoggingRemote() RemoteOption {
	return func(x *Remote) {
		x.logger = zlog.New(zlog.WithLogLevel("debug"))
	}
}

type httpInput struct {
	Input interface{} `json:"input"`
}

type httpOutput struct {
	Result interface{} `json:"result"`
}

// Query evaluates policy with `input` data. The result will be written to `out`. `out` must be pointer of instance.
func (x *Remote) Query(ctx context.Context, in interface{}, out interface{}, options ...QueryOption) error {
	x.logger.With("in", in).Debug("start Remote.Query")

	cfg := newQueryConfig(options...)
	if cfg.pkgSuffix != "" {
		return goerr.Wrap(ErrInvalidQueryOption, "suffix is not supported for remote inquiry")
	}

	input := httpInput{
		Input: in,
	}
	rawInput, err := json.Marshal(input)
	if err != nil {
		return goerr.Wrap(err, "fail to marshal rego input for remote inquiry")
	}

	req, err := http.NewRequest(http.MethodPost, x.url, bytes.NewReader(rawInput))
	if err != nil {
		return goerr.Wrap(err, "fail to create a http request for remote inquiry")
	}
	req.Header.Add("Content-Type", "application/json")

	for name, values := range x.httpHeaders {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	resp, err := x.httpClient.Do(req)
	if err != nil {
		return goerr.Wrap(err, "fail http request to OPA server").With("url", x.url)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return goerr.Wrap(err, "unexpected http code from OPA server").
			With("code", resp.StatusCode).
			With("body", string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return goerr.Wrap(err, "fail to read body from OPA server")
	}

	var output httpOutput
	if err := json.Unmarshal(body, &output); err != nil {
		return goerr.Wrap(err, "fail to parse OPA server result").With("body", string(body))
	}

	rawOutput, err := json.Marshal(output.Result)
	if err != nil {
		return goerr.Wrap(err, "fail to re-marshal result filed in OPA response")
	}

	if err := json.Unmarshal(rawOutput, out); err != nil {
		return goerr.Wrap(err, "fail to unmarshal OPA server result to out")
	}

	x.logger.With("response body", string(body)).Debug("done Remote.Query")

	return nil
}
