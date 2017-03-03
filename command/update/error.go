package update

import (
	"github.com/juju/errgo"
)

var executionFailedError = errgo.New("execution failed")

// IsExecutionFailed asserts executionFailedError.
func IsExecutionFailed(err error) bool {
	return errgo.Cause(err) == executionFailedError
}

var invalidConfigError = errgo.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return errgo.Cause(err) == invalidConfigError
}

var invalidFlagsError = errgo.New("invalid flags")

// IsInvalidFlags asserts invalidFlagsError.
func IsInvalidFlags(err error) bool {
	return errgo.Cause(err) == invalidFlagsError
}
