package opac

import (
	"context"
	"fmt"
	"log/slog"
)

// Client is the main interface to interact with the opac library.
type Client struct {
	query queryFunc
}

type queryFunc func(ctx context.Context, query string, input, output any) error

type config struct {
	logger *slog.Logger
}

// Source is a function that returns the policy data. It is used to provide the policy data to the client.
type Source func(cfg *config) (queryFunc, error)

// Option is a function that configures the client.
type Option func(*config)

// WithLogger sets the logger for the client.
func WithLogger(logger *slog.Logger) Option {
	return func(cfg *config) {
		cfg.logger = logger
	}
}

// New creates a new opac client. It returns an error if neither policy data nor configuration is provided.
func New(src Source, options ...Option) (*Client, error) {
	cfg := &config{
		logger: slog.Default(),
	}

	for _, opt := range options {
		opt(cfg)
	}

	query, err := src(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &Client{
		query: query,
	}, nil
}

// Query evaluates the given query with the provided input and output. The query is evaluated against the policy data provided during client creation.
func (c *Client) Query(ctx context.Context, query string, input, output any) error {
	return c.query(ctx, query, input, output)
}
