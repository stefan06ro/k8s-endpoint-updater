package flag

import (
	microerror "github.com/giantswarm/microkit/error"

	"github.com/giantswarm/k8s-endpoint-updater/command/update/flag/kubernetes"
	"github.com/giantswarm/k8s-endpoint-updater/command/update/flag/provider"
	"github.com/giantswarm/k8s-endpoint-updater/command/update/flag/updater"
)

type Flag struct {
	Kubernetes kubernetes.Kubernetes
	Provider   provider.Provider
	Updater    updater.Updater
}

func (f *Flag) Validate() error {
	if f.Provider.Kind == "env" && f.Provider.Env.Prefix == "" {
		return microerror.MaskAnyf(invalidFlagsError, "env prefix must not be empty")
	}
	if f.Provider.Kind != "env" {
		return microerror.MaskAnyf(invalidFlagsError, "provider kind must be 'env'")
	}

	if len(f.Updater.Pod.Names) == 0 {
		return microerror.MaskAnyf(invalidFlagsError, "pod names must not be empty")
	}

	return nil
}
