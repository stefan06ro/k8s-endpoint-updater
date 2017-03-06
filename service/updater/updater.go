package updater

import (
	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/k8s-endpoint-updater/service/provider"
)

// Config represents the configuration used to create a new updater.
type Config struct {
	// Dependencies.
	KubernetesClient *kubernetes.Clientset
	Logger           micrologger.Logger
}

// DefaultConfig provides a default configuration to create a new updater
// by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		KubernetesClient: nil,
		Logger:           nil,
	}
}

// New creates a new updater.
func New(config Config) (*Updater, error) {
	// Dependencies.
	if config.KubernetesClient == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "kubernetes client must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "logger must not be empty")
	}

	newUpdater := &Updater{
		// Dependencies.
		kubernetesClient: config.KubernetesClient,
		logger:           config.Logger,
	}

	return newUpdater, nil
}

type Updater struct {
	// Dependencies.
	kubernetesClient *kubernetes.Clientset
	logger           micrologger.Logger
}

func (p *Updater) Update(podInfos []provider.PodInfo) error {
	// TODO
	return nil
}
