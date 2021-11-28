package opaclient

import (
	"context"

	"github.com/open-policy-agent/opa/server/types"
)

type listPolicyOutput struct {
	Result []*types.PolicyV1 `json:"result"`
}

func (x *Client) ListPolicy(ctx context.Context) ([]*types.PolicyV1, error) {
	url := x.baseURL + "/v1/policies/"

	var resp listPolicyOutput
	if err := x.request(ctx, "GET", url, nil, &resp); err != nil {
		return nil, err
	}
	if resp.Result == nil {
		return nil, nil
	}

	return resp.Result, nil
}
