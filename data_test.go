package opac_test

import (
	"context"
	"testing"

	opac "github.com/m-mizutani/opac"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type exampleRequest struct {
	User string `json:"user"`
}

type exampleResult struct {
	Allowed bool   `json:"allowed"`
	MyData  string `json:"mydata"`
}

func TestDataRequest(t *testing.T) {
	client := setupClient(t)
	ctx := context.Background()

	t.Run("GET example", func(t *testing.T) {
		var result exampleResult
		require.NoError(t, client.GetData(ctx, &opac.DataRequest{
			Path: "example",
		}, &result))
		assert.False(t, result.Allowed)
		assert.Equal(t, "orange", result.MyData)
	})

	t.Run("GET example/mydata", func(t *testing.T) {
		var result string
		require.NoError(t, client.GetData(ctx, &opac.DataRequest{
			Path: "example/mydata",
		}, &result))
		assert.Equal(t, "orange", result)
	})

	t.Run("POST example/mydata", func(t *testing.T) {
		var result exampleResult
		req := &opac.DataRequest{
			Input: exampleRequest{
				User: "blue",
			},
			Path: "example",
		}
		require.NoError(t, client.GetData(ctx, req, &result))
		assert.True(t, result.Allowed)
	})
}
