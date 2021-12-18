package opac

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

// GetData gets a document with/without input data
func (x *Client) GetData(ctx context.Context, req *DataRequest, dst interface{}) error {
	url := x.baseURL + "/v1/data"
	if req.Path != "" {
		url += "/" + req.Path
	}
	method := http.MethodGet

	var reader io.Reader
	if req.Input != nil {
		input := dataInput{
			Input: req.Input,
		}

		method = http.MethodPost
		raw, err := json.Marshal(input)
		if err != nil {
			return ErrInvalidInput.Wrap(err)
		}
		reader = bytes.NewReader(raw)
	}

	var resp dataOutput
	data := &body{
		Reader: reader,
		Type:   contetJSON,
	}

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

// UpdateData creates or overwrite a document.
func (x *Client) UpdateData(ctx context.Context, path string, data interface{}) error {
	url := x.baseURL + "/v1/data/" + path

	raw, err := json.Marshal(data)
	if err != nil {
		return ErrInvalidInput.Wrap(err)
	}
	reader := bytes.NewReader(raw)

	b := &body{
		Reader: reader,
		Type:   contetJSON,
	}

	if err := x.request(ctx, http.MethodPut, url, b, nil); err != nil {
		return err
	}

	return nil
}

// DeleteData deletes a document.
func (x *Client) DeleteData(ctx context.Context, path string) error {
	url := x.baseURL + "/v1/data/" + path

	if err := x.request(ctx, http.MethodDelete, url, nil, nil); err != nil {
		return err
	}

	return nil
}
