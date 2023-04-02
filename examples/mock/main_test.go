package mock_test

import (
	"testing"

	"github.com/m-mizutani/opac"
	"github.com/m-mizutani/opac/examples/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithMock(t *testing.T) {
	foo := mock.NewWithMock(func(input interface{}, options ...opac.QueryOption) (interface{}, error) {
		in, ok := input.(*mock.Input)
		require.True(t, ok)
		return &mock.Result{Allow: in.User == "blue"}, nil
	})

	assert.True(t, foo.IsAllow("blue"))
	assert.False(t, foo.IsAllow("orange"))
}
