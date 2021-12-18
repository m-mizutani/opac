package opac

import "github.com/m-mizutani/goerr"

var (
	ErrInvalidInput   = goerr.New("invalid input")
	ErrRequestFailed  = goerr.New("request to OPA server failed")
	ErrUnexpectedResp = goerr.New("unexpected response from OPA server")

	// Internal errors
	ErrInvalidConfiguration = goerr.New("invalid configuration")
	ErrExitWithNonZero      = goerr.New("exit with non-zero")
)
