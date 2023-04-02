package opac

import "context"

type Client interface {
	Query(ctx context.Context, in interface{}, out interface{}, options ...QueryOption) error
}

type queryConfig struct {
	pkgSuffix string
}

type QueryOption func(*queryConfig)

func WithPackageSuffix(suffix string) QueryOption {
	return func(x *queryConfig) {
		x.pkgSuffix = suffix
	}
}

func newQueryConfig(options ...QueryOption) *queryConfig {
	cfg := &queryConfig{}

	for _, opt := range options {
		opt(cfg)
	}

	return cfg
}
