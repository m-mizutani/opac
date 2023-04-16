package opac

import (
	"context"
	"io"
)

type Client interface {
	Query(ctx context.Context, in interface{}, out interface{}, options ...QueryOption) error
}

type queryConfig struct {
	pkgSuffix   string
	printWriter io.Writer
}

type QueryOption func(*queryConfig)

func WithPackageSuffix(suffix string) QueryOption {
	return func(x *queryConfig) {
		x.pkgSuffix = suffix
	}
}

func WithPrintWriter(w io.Writer) QueryOption {
	return func(x *queryConfig) {
		x.printWriter = w
	}
}

func newQueryConfig(options ...QueryOption) *queryConfig {
	cfg := &queryConfig{}

	for _, opt := range options {
		opt(cfg)
	}

	return cfg
}
