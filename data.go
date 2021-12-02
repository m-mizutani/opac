package opaclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

type DataRequest struct {
	Input interface{}
	Path  string
}

type dataInput struct {
	Input interface{} `json:"input"`
}

type dataOutput struct {
	Result interface{} `json:"result"`
}

func (x *Client) GetData(ctx context.Context, req *DataRequest, dst interface{}) error {
	url := x.baseURL + "/v1/data/" + req.Path
	method := http.MethodGet

	var data io.Reader
	if req.Input != nil {
		input := dataInput{
			Input: req.Input,
		}

		method = http.MethodPost
		raw, err := json.Marshal(input)
		if err != nil {
			return ErrInvalidInput.Wrap(err)
		}
		data = bytes.NewReader(raw)
	}

	var resp dataOutput
	if err := x.request(ctx, method, url, data, &resp); err != nil {
		return err
	}
	if resp.Result == nil {
		return nil
	}
	raw, err := json.Marshal(resp.Result)
	if err != nil {
		return ErrUnexpectedResp.Wrap(err).With("result", resp.Result)
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		return ErrUnexpectedResp.Wrap(err).With("result data", string(raw))
	}

	return nil
}
