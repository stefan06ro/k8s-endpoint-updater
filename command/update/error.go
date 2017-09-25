package update

import "github.com/giantswarm/microerror"

var cancelledError = microerror.New("cancelled")

// IsCancelled asserts cancelledError.
func IsCancelled(err error) bool {
	return microerror.Cause(err) == cancelledError
}

var executionFailedError = microerror.New("execution failed")

// IsExecutionFailed asserts executionFailedError.
func IsExecutionFailed(err error) bool {
	return microerror.Cause(err) == executionFailedError
}

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}
