package opaclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/open-policy-agent/opa/server/types"
)

type policyOutput struct {
	Result []types.PolicyV1 `json:"result"`
}

func (x *Client) Policy(ctx context.Context, req *DataRequest, dst interface{}) error {
	url := x.baseURL + "/v1/data/" + req.Path
	method := "GET"
	var data io.Reader
	if req.Input != nil {
		input := dataInput{
			Input: req.Input,
		}

		method = "POST"
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
