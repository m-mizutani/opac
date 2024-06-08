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
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type RemoteOption func(*remoteConfig)

func WithHTTPClient(client HTTPClient) RemoteOption {
	return func(cfg *remoteConfig) {
		cfg.httpClient = client
	}
}

type remoteConfig struct {
	httpClient HTTPClient
	logger     *slog.Logger
}

func Remote(baseURL string, options ...RemoteOption) Source {
	return func(cfg *config) (queryFunc, error) {
		tgtURL, err := url.Parse(baseURL)
		if err != nil {
			return nil, fmt.Errorf("invalid remote base URL: %w", err)
		}

		remoteCfg := &remoteConfig{
			httpClient: http.DefaultClient,
			logger:     cfg.logger,
		}

		for _, opt := range options {
			opt(remoteCfg)
		}

		return func(ctx context.Context, query string, input, output any, opt queryOptions) error {
			return remoteQuery(ctx, query, input, output, remoteCfg, tgtURL, opt)
		}, nil
	}
}

func remoteQuery(ctx context.Context, query string, input, output any, cfg *remoteConfig, tgtURL *url.URL, _ queryOptions) error {
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

	reqURL := tgtURL
	queryPath := strings.ReplaceAll(query, ".", "/")
	reqURL.Path = path.Join(reqURL.Path, queryPath)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL.String(), bytes.NewReader(inputBody))
	if err != nil {
		return fmt.Errorf("failed to create request to OPA server: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	cfg.logger.Debug("Sending request to OPA server", "url", req.URL.String(), "body", string(inputBody))
	resp, err := cfg.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to OPA server: %w", err)
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	cfg.logger.Debug("Received response from OPA server", "status", resp.StatusCode, "body", string(body), "headers", resp.Header)

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
