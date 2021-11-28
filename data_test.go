package opaclient_test

import (
	"context"
	"os"
	"testing"

	opaclient "github.com/m-mizutani/opa-go-client"
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
	url, ok := os.LookupEnv("OPA_BASE_URL")
	if !ok {
		t.Skip("OPA_BASE_URL is not set")
	}

	ctx := context.Background()
	client, err := opaclient.New(url)
	require.NoError(t, err)

	t.Run("GET example", func(t *testing.T) {
		var result exampleResult
		require.NoError(t, client.GetData(ctx, &opaclient.DataRequest{
			Path: "example",
		}, &result))
		assert.False(t, result.Allowed)
		assert.Equal(t, "orange", result.MyData)
	})

	t.Run("GET example/mydata", func(t *testing.T) {
		var result string
		require.NoError(t, client.GetData(ctx, &opaclient.DataRequest{
			Path: "example/mydata",
		}, &result))
		assert.Equal(t, "orange", result)
	})

	t.Run("POST example/mydata", func(t *testing.T) {
		var result exampleResult
		req := &opaclient.DataRequest{
			Input: exampleRequest{
				User: "blue",
			},
			Path: "example",
		}
		require.NoError(t, client.GetData(ctx, req, &result))
		assert.True(t, result.Allowed)
	})
}
