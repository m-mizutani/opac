package opac

import "context"

type Client interface {
	Query(ctx context.Context, in interface{}, out interface{}) error
}
