package opaclient_test

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	opaclient "github.com/m-mizutani/opa-go-client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPRequest(t *testing.T) {
	var called int
	req := func(ctx context.Context, method, url string, data io.Reader) (*http.Response, error) {
		called++
		assert.Equal(t, "https://example.com/v1/data", url)
		assert.Equal(t, "POST", method)
		raw, err := ioutil.ReadAll(data)
		require.NoError(t, err)
		assert.Equal(t, `{"input":{"user":"blue"}}`, string(raw))

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBuffer([]byte(`{"result":{}}`))),
		}, nil
	}

	client, err := opaclient.New("https://example.com", opaclient.OptHTTPRequest(req))
	require.NoError(t, err)
	var out interface{}
	require.NoError(t, client.GetData(context.Background(), &opaclient.DataRequest{
		Input: map[string]string{
			"user": "blue",
		},
	}, &out))
	assert.Equal(t, 1, called)
	t.Log("output =>", out)
}
