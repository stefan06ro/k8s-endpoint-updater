package etcd

import (
	"context"
	"net"

	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"
	microstorage "github.com/giantswarm/microkit/storage"

	"github.com/giantswarm/k8s-endpoint-updater/service/provider"
)

const (
	Kind = "etcd"
)

// Config represents the configuration used to create a new provider.
type Config struct {
	// Dependencies.
	Logger  micrologger.Logger
	Storage microstorage.Service

	// Settings.
	PodNames []string
}

// DefaultConfig provides a default configuration to create a new provider
// by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger:  nil,
		Storage: nil,

		// Settings.
		PodNames: nil,
	}
}

// New creates a new provider.
func New(config Config) (*Provider, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "logger must not be empty")
	}
	if config.Storage == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "storage must not be empty")
	}

	// Settings.
	if len(config.PodNames) == 0 {
		return nil, microerror.MaskAnyf(invalidConfigError, "pod names must not be empty")
	}

	newProvider := &Provider{
		// Dependencies.
		logger:  config.Logger,
		storage: config.Storage,

		// Settings.
		podNames: config.PodNames,
	}

	return newProvider, nil
}

type Provider struct {
	// Dependencies.
	logger  micrologger.Logger
	storage microstorage.Service

	// Settings.
	podNames []string
}

func (p *Provider) Lookup() ([]provider.PodInfo, error) {
	var podInfos []provider.PodInfo

	for _, pn := range p.podNames {
		IP, err := p.storage.Search(context.TODO(), pn)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}

		podInfo := provider.PodInfo{
			IP: net.ParseIP(IP),
		}

		podInfos = append(podInfos, podInfo)
	}

	return podInfos, nil
}
