package flag

import (
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/k8s-endpoint-updater/command/update/flag/kubernetes"
	"github.com/giantswarm/k8s-endpoint-updater/command/update/flag/provider"
)

type Flag struct {
	Kubernetes kubernetes.Kubernetes
	Provider   provider.Provider
}

func (f *Flag) Validate() error {
	if f.Kubernetes.Cluster.Namespace == "" {
		return microerror.Maskf(invalidFlagsError, "guest cluster namespace must not be empty")
	}
	if f.Kubernetes.Cluster.Service == "" {
		return microerror.Maskf(invalidFlagsError, "guest cluster service must not be empty")
	}

	if f.Provider.Kind == "env" && f.Provider.Env.Prefix == "" {
		return microerror.Maskf(invalidFlagsError, "env prefix must not be empty")
	}
	if f.Provider.Kind == "" {
		return microerror.Maskf(invalidFlagsError, "provider kind must not be empty")
	}

	return nil
}
