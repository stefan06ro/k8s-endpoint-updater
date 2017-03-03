package update

import (
	microerror "github.com/giantswarm/microkit/error"
)

type Flags struct {
	Foo string
}

func (f *Flags) Validate() error {
	if f.Foo == "" {
		return microerror.MaskAnyf(invalidFlagsError, "foo must not be empty")
	}

	return nil
}
