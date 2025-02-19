package opac

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/topdown/print"
)

// Client is the main interface to interact with the opac library.
type Client struct {
	src Source
}

type config struct {
	logger *slog.Logger
}

// Source is a function that returns the policy data. It is used to provide the policy data to the client.
type Source interface {
	Configure(cfg *config) error
	Query(ctx context.Context, query string, input, output any, opt queryOptions) error
	AnnotationSet() *ast.AnnotationSet
}

// Option is a function that configures the client.
type Option func(*config)

// WithLogger sets the logger for the client. The default logger is a no-op logger. The log message level is DEBUG and you should set LogLevel by own.
func WithLogger(logger *slog.Logger) Option {
	return func(cfg *config) {
		cfg.logger = logger
	}
}

type noopWriter struct{}

func (noopWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

// New creates a new opac client. It returns an error if neither policy data nor configuration is provided.
func New(src Source, options ...Option) (*Client, error) {
	cfg := &config{
		logger: slog.New(slog.NewTextHandler(&noopWriter{}, nil)),
	}

	for _, opt := range options {
		opt(cfg)
	}

	if err := src.Configure(cfg); err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &Client{
		src: src,
	}, nil
}

// Query evaluates the given query with the provided input and output. The query is evaluated against the policy data provided during client creation.
func (c *Client) Query(ctx context.Context, query string, input, output any, options ...QueryOption) error {
	opt := queryOptions{}
	for _, o := range options {
		o(&opt)
	}

	return c.src.Query(ctx, query, input, output, opt)
}

type queryOptions struct {
	printHook print.Hook
}

type QueryOption func(*queryOptions)

// WithPrintHook sets the print hook for the query. The print hook is used to capture the print statements in the policy evaluation.
func WithPrintHook(h print.Hook) QueryOption {
	return func(o *queryOptions) {
		o.printHook = h
	}
}

// Metadata returns the annotation set of the policy data. It works only for local policy data (File or Data).
func (c *Client) Metadata() ast.FlatAnnotationsRefSet {
	as := c.src.AnnotationSet()
	return as.Flatten()
}
