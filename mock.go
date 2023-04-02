package opac

import (
	"context"
	"encoding/json"

	"github.com/m-mizutani/goerr"
)

type Mock struct {
	mockFunc MockFunc
}

type MockFunc func(in interface{}, options ...QueryOption) (interface{}, error)

func NewMock(f MockFunc) *Mock {
	return &Mock{
		mockFunc: f,
	}
}

func (x *Mock) Query(ctx context.Context, in interface{}, out interface{}, options ...QueryOption) error {
	result, err := x.mockFunc(in, options...)
	if err != nil {
		return err
	}

	raw, err := json.Marshal(result)
	if err != nil {
		return goerr.Wrap(err)
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return goerr.Wrap(err)
	}

	return nil
}
