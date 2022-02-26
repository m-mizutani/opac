package opac

import (
	"context"
	"encoding/json"

	"github.com/m-mizutani/goerr"
)

type Mock struct {
	mockFunc MockFunc
}

type MockFunc func(in interface{}) (interface{}, error)

func NewMock(f MockFunc) *Mock {
	return &Mock{
		mockFunc: f,
	}
}

func (x *Mock) Query(ctx context.Context, in interface{}, out interface{}) error {
	result, err := x.mockFunc(in)
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
