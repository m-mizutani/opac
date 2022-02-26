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
)

type Remote struct {
	url         string
	httpHeaders map[string][]string
	httpClient  HTTPClient
}

type RemoteOption func(x *Remote)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewRemote(url string, options ...RemoteOption) (*Remote, error) {
	if err := validation.Validate(url, validation.Required, is.URL); err != nil {
		return nil, goerr.Wrap(err, "invalid URL for remote policy")
	}

	client := &Remote{
		url:         url,
		httpHeaders: make(map[string][]string),
		httpClient:  http.DefaultClient,
	}
	for _, opt := range options {
		opt(client)
	}

	return client, nil
}

func WithHTTPClient(client HTTPClient) RemoteOption {
	return func(x *Remote) {
		x.httpClient = client
	}
}

func WithHTTPHeader(name, value string) RemoteOption {
	return func(x *Remote) {
		x.httpHeaders[name] = append(x.httpHeaders[name], value)
	}
}

type httpInput struct {
	Input interface{} `json:"input"`
}

type httpOutput struct {
	Result interface{} `json:"result"`
}

func (x *Remote) Query(ctx context.Context, in interface{}, out interface{}) error {
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

	return nil
}
