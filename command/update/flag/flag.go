package flag

import (
	microerror "github.com/giantswarm/microkit/error"

	"github.com/giantswarm/k8s-endpoint-updater/command/update/flag/pod"
	"github.com/giantswarm/k8s-endpoint-updater/command/update/flag/provider"
)

type Flag struct {
	Pod      pod.Pod
	Provider provider.Provider
}

func (f *Flag) Validate() error {
	if len(f.Pod.Names) == 0 {
		return microerror.MaskAnyf(invalidFlagsError, "pod names must not be empty")
	}
	if f.Provider.Kind == "env" && f.Provider.Env.Prefix == "" {
		return microerror.MaskAnyf(invalidFlagsError, "env prefix must not be empty")
	}
	if f.Provider.Kind != "env" {
		return microerror.MaskAnyf(invalidFlagsError, "provider kind must be 'env'")
	}

	return nil
}
