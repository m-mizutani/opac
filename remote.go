package opac

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/open-policy-agent/opa/ast"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type RemoteOption func(*remoteSource)

func WithHTTPClient(client HTTPClient) RemoteOption {
	return func(r *remoteSource) {
		r.httpClient = client
	}
}

type remoteSource struct {
	httpClient HTTPClient
	logger     *slog.Logger
	rawURL     string
	url        *url.URL
	options    []RemoteOption
}

// AnnotationSet implements Source.
func (r *remoteSource) AnnotationSet() *ast.AnnotationSet {
	return &ast.AnnotationSet{}
}

// Configure implements Source.
func (r *remoteSource) Configure(cfg *config) error {
	tgtURL, err := url.Parse(r.rawURL)
	if err != nil {
		return fmt.Errorf("invalid remote base URL: %w", err)
	}

	for _, opt := range r.options {
		opt(r)
	}

	r.logger = cfg.logger
	r.url = tgtURL

	return nil
}

// Query implements Source.
func (r *remoteSource) Query(ctx context.Context, query string, input any, output any, opt queryOptions) error {
	type httpInput struct {
		Input any `json:"input"`
	}

	type httpOutput struct {
		Result any `json:"result"`
	}

	inputData := httpInput{Input: input}

	inputBody, err := json.Marshal(inputData)
	if err != nil {
		return fmt.Errorf("failed to marshal input: %w", err)
	}

	reqURL := r.url
	queryPath := strings.ReplaceAll(query, ".", "/")
	reqURL.Path = path.Join(reqURL.Path, queryPath)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL.String(), bytes.NewReader(inputBody))
	if err != nil {
		return fmt.Errorf("failed to create request to OPA server: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	r.logger.Debug("Sending request to OPA server", "url", req.URL.String(), "body", string(inputBody))
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to OPA server: %w", err)
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	r.logger.Debug("Received response from OPA server", "status", resp.StatusCode, "body", string(body), "headers", resp.Header)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code from OPA server: %d msg='%s'", resp.StatusCode, string(body))
	}
	if readErr != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var outputData httpOutput
	if err := json.Unmarshal(body, &outputData); err != nil {
		return fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	if outputData.Result == nil {
		return ErrNoEvalResult
	}

	raw, err := json.Marshal(outputData.Result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}
	if err := json.Unmarshal(raw, output); err != nil {
		return fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return nil
}

var _ Source = (*remoteSource)(nil)

func Remote(baseURL string, options ...RemoteOption) *remoteSource {
	return &remoteSource{
		httpClient: http.DefaultClient,
		rawURL:     baseURL,
		options:    options,
	}
}
