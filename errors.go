package opac

import "errors"

var (
	// ErrNoPolicyData is returned when no policy data is provided.
	ErrNoPolicyData = errors.New("no policy data, one ore more file or policy data are required")

	// ErrNoPolicySrc is returned when no result of evaluation is provided. If you expect a result, you should check the error. If you don't expect a result, you should ignore the error.
	ErrNoEvalResult = errors.New("no evaluation result")
)
