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

type dummyClient struct {
	DoMock func(req *http.Request) (*http.Response, error)
}

func (x *dummyClient) Do(req *http.Request) (*http.Response, error) {
	return x.DoMock(req)
}

func TestHTTPRequest(t *testing.T) {
	var called int
	dummy := &dummyClient{
		DoMock: func(req *http.Request) (*http.Response, error) {
			called++

			assert.Equal(t, "https://example.com/v1/data", req.URL.String())
			assert.Equal(t, "POST", req.Method)
			raw, err := ioutil.ReadAll(req.Body)
			require.NoError(t, err)
			assert.Equal(t, `{"input":{"user":"blue"}}`, string(raw))

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer([]byte(`{"result":{}}`))),
			}, nil
		},
	}

	client, err := opaclient.New("https://example.com", opaclient.WithHTTPClient(dummy))
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
