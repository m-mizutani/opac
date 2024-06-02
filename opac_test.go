package opac_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/opac"
)

type logWriter struct {
	buf *bytes.Buffer
}

func (x *logWriter) Write(p []byte) (int, error) {
	return x.buf.Write(p)
}

func TestLogger(t *testing.T) {
	w := &logWriter{buf: new(bytes.Buffer)}
	logger := slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	mock := &httpMock{
		do: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"result": {"allow": true}}`)),
			}, nil
		},
	}
	client := gt.R1(opac.New(
		opac.Remote("http://example.com/v1", opac.WithHTTPClient(mock)),
		opac.WithLogger(logger),
	)).NoError(t)

	ctx := context.Background()
	input := map[string]interface{}{
		"user": "admin",
	}
	var output struct {
		Allow bool `json:"allow"`
	}

	gt.NoError(t, client.Query(ctx, "data.system.authz", input, &output))
	var out map[string]any
	decoder := json.NewDecoder(bytes.NewReader(w.buf.Bytes()))

	// 1st log is for sending the request
	gt.NoError(t, decoder.Decode(&out))
	gt.Equal(t, out["url"], "http://example.com/v1/data/system/authz")

	// 2nd log is for receiving the response
	gt.NoError(t, decoder.Decode(&out))
	gt.Equal(t, out["status"].(float64), 200)
}
