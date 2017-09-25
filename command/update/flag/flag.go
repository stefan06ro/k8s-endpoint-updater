package flag

import (
	microerror "github.com/giantswarm/microkit/error"

	"github.com/giantswarm/k8s-endpoint-updater/command/update/flag/kubernetes"
	"github.com/giantswarm/k8s-endpoint-updater/command/update/flag/provider"
)

type Flag struct {
	Kubernetes kubernetes.Kubernetes
	Provider   provider.Provider
}

func (f *Flag) Validate() error {
	if f.Kubernetes.Cluster.Namespace == "" {
		return microerror.MaskAnyf(invalidFlagsError, "guest cluster namespace must not be empty")
	}
	if f.Kubernetes.Cluster.Service == "" {
		return microerror.MaskAnyf(invalidFlagsError, "guest cluster service must not be empty")
	}

	if f.Provider.Kind == "env" && f.Provider.Env.Prefix == "" {
		return microerror.MaskAnyf(invalidFlagsError, "env prefix must not be empty")
	}
	if f.Provider.Kind == "" {
		return microerror.MaskAnyf(invalidFlagsError, "provider kind must not be empty")
	}

	return nil
}
