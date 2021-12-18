package opac

import (
	"bytes"
	"context"
	"net/http"

	"github.com/open-policy-agent/opa/server/types"
)

type listPolicyOutput struct {
	Result []*types.PolicyV1 `json:"result"`
}
type getPolicyOutput struct {
	Result *types.PolicyV1 `json:"result"`
}

// ListPolicy retrieves policy modules.
func (x *Client) ListPolicy(ctx context.Context) ([]*types.PolicyV1, error) {
	url := x.baseURL + "/v1/policies"

	var resp listPolicyOutput
	if err := x.request(ctx, "GET", url, nil, &resp); err != nil {
		return nil, err
	}
	if resp.Result == nil {
		return nil, nil
	}

	return resp.Result, nil
}

// GetPolicy retrieves a policy module
func (x *Client) GetPolicy(ctx context.Context, id string) (*types.PolicyV1, error) {
	url := x.baseURL + "/v1/policies/" + id

	var resp getPolicyOutput
	if err := x.request(ctx, http.MethodGet, url, nil, &resp); err != nil {
		return nil, err
	}
	if resp.Result == nil {
		return nil, nil
	}

	return resp.Result, nil
}

// PutPolicy creates or update a policy module
func (x *Client) PutPolicy(ctx context.Context, id string, policy string) error {
	url := x.baseURL + "/v1/policies/" + id
	data := &body{
		Reader: bytes.NewReader([]byte(policy)),
		Type:   contetText,
	}

	if err := x.request(ctx, http.MethodPut, url, data, nil); err != nil {
		return err
	}

	return nil
}

// DeletePolicy deleats a policy module
func (x *Client) DeletePolicy(ctx context.Context, id string) error {
	url := x.baseURL + "/v1/policies/" + id

	if err := x.request(ctx, http.MethodDelete, url, nil, nil); err != nil {
		return err
	}
	return nil
}
