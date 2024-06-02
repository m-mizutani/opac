package opac_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/opac"
)

type httpMock struct {
	do func(req *http.Request) (*http.Response, error)
}

func (x *httpMock) Do(req *http.Request) (*http.Response, error) {
	return x.do(req)
}

func TestRemote(t *testing.T) {
	type testCase struct {
		url    string
		query  string
		input  map[string]any
		do     func(req *http.Request) (*http.Response, error)
		expect bool
		isErr  bool
	}

	doTest := func(tc testCase) func(t *testing.T) {
		return func(t *testing.T) {
			mock := &httpMock{
				do: tc.do,
			}
			client := gt.R1(opac.New(
				opac.Remote(tc.url, opac.WithHTTPClient(mock)),
			)).NoError(t)
			ctx := context.Background()
			input := map[string]interface{}{
				"user": "admin",
			}
			var output struct {
				Allow bool `json:"allow"`
			}

			err := client.Query(ctx, "data.system.authz", input, &output)
			if tc.isErr {
				gt.Error(t, err)
			} else {
				gt.Equal(t, output.Allow, tc.expect)
			}
		}
	}

	t.Run("success", doTest(testCase{
		url:   "https://example.com/v1",
		query: "data.system.authz",
		input: map[string]any{
			"user": "admin",
		},
		do: func(req *http.Request) (*http.Response, error) {
			var body struct {
				Input map[string]any `json:"input"`
			}
			gt.NoError(t, json.NewDecoder(req.Body).Decode(&body))
			gt.Equal(t, body.Input, map[string]any{"user": "admin"})

			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`{"result": {"allow": true}}`)),
			}, nil
		},
		expect: true,
	}))

	t.Run("client error", doTest(testCase{
		url:   "https://example.com/v1",
		query: "data.system.authz",
		input: map[string]any{
			"user": "admin",
		},
		do: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("some error")
		},
		isErr: true,
	}))

	t.Run("server error", doTest(testCase{
		url:   "https://example.com/v1",
		query: "data.system.authz",
		input: map[string]any{
			"user": "admin",
		},
		do: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 500,
				Body:       io.NopCloser(strings.NewReader(`{"error": "some error"}`)),
			}, nil
		},
		isErr: true,
	}))

	t.Run("invalid response", doTest(testCase{
		url:   "https://example.com/v1",
		query: "data.system.authz",
		input: map[string]any{
			"user": "admin",
		},
		do: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`invalid json`)),
			}, nil
		},
		isErr: true,
	}))

	t.Run("invalid response body", doTest(testCase{
		url:   "https://example.com/v1",
		query: "data.system.authz",
		input: map[string]any{
			"user": "admin",
		},
		do: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`{"result": {"allow": 1}}`)),
			}, nil
		},
		isErr: true,
	}))

	t.Run("no result, but valid response", doTest(testCase{
		url:   "https://example.com/v1",
		query: "data.system.authz",
		input: map[string]any{
			"user": "admin",
		},
		do: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`{"result": {}}`)),
			}, nil
		},
		isErr:  false,
		expect: false,
	}))

	t.Run("nil result and get error", doTest(testCase{
		url:   "https://example.com/v1",
		query: "data.system.authz",
		input: map[string]any{
			"user": "admin",
		},
		do: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`{}`)),
			}, nil
		},
		isErr: true,
	}))
}

func loadEnvVar(t *testing.T, key string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		t.Skipf("missing environment variable: %s", key)
	}

	return v
}

func TestRemoteWithOPACommand(t *testing.T) {
	opaAddr := "localhost:18181"
	// Start OPA server
	opaPath := loadEnvVar(t, "TEST_OPA_PATH")
	cmd := exec.Command(opaPath, "run", "--server", "-a", opaAddr, "testdata/server")
	cmd.Start()
	gt.NotEqual(t, cmd.Process, nil)
	defer cmd.Process.Kill()

	time.Sleep(1 * time.Second)

	client := gt.R1(opac.New(opac.Remote("http://" + opaAddr + "/v1"))).NoError(t)
	ctx := context.Background()
	input := map[string]interface{}{
		"user": "admin",
	}
	var output struct {
		Allow bool `json:"allow"`
	}

	gt.NoError(t, client.Query(ctx, "data.system.authz", input, &output))
	gt.True(t, output.Allow)
}
