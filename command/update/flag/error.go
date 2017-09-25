package flag

import "github.com/giantswarm/microerror"

var invalidFlagsError = microerror.New("invalid flags")

// IsInvalidFlags asserts invalidFlagsError.
func IsInvalidFlags(err error) bool {
	return microerror.Cause(err) == invalidFlagsError
}
