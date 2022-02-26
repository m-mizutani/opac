package opac

import "github.com/m-mizutani/goerr"

var (
	ErrNoEvalResult = goerr.New("no evaluation result")
	ErrReadRegoDir  = goerr.New("fail to read rego directory")
	ErrReadRegoFile = goerr.New("fail to read rego file")
)
