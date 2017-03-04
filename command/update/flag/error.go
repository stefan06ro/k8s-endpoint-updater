package flag

import (
	"github.com/juju/errgo"
)

var invalidFlagsError = errgo.New("invalid flags")

// IsInvalidFlags asserts invalidFlagsError.
func IsInvalidFlags(err error) bool {
	return errgo.Cause(err) == invalidFlagsError
}
