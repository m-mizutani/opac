package opac

import "github.com/m-mizutani/goerr"

var (
	ErrNoEvalResult       = goerr.New("no evaluation result")
	ErrReadRegoDir        = goerr.New("fail to read rego directory")
	ErrReadRegoFile       = goerr.New("fail to read rego file")
	ErrNoPolicyData       = goerr.New("no policy data, one ore more file or policy data are required")
	ErrInvalidQueryOption = goerr.New("invalid query option")
)
