package opaclient

import "github.com/m-mizutani/goerr"

var (
	ErrInvalidInput   = goerr.New("invalid input")
	ErrRequestFailed  = goerr.New("request to OPA server failed")
	ErrUnexpectedResp = goerr.New("unexpected response from OPA server")
)
