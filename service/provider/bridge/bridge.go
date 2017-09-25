package bridge

import (
	"errors"
	"net"

	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"

	"github.com/giantswarm/k8s-endpoint-updater/service/provider"
)

const (
	Kind = "bridge"
)

// Config represents the configuration used to create a new provider.
type Config struct {
	// Dependencies.
	Logger micrologger.Logger

	// Settings.

	// BridgeName is the bridge name of the underlying host used to lookup the endpoint
	// IP.
	BridgeName string
}

// DefaultConfig provides a default configuration to create a new provider
// by best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		Logger: nil,

		// Settings.
		BridgeName: "",
	}
}

// New creates a new provider.
func New(config Config) (*Provider, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.MaskAnyf(invalidConfigError, "config.Logger must not be empty")
	}

	// Settings.
	if config.BridgeName == "" {
		return nil, microerror.MaskAnyf(invalidConfigError, "config.BridgeName must not be empty")
	}

	newProvider := &Provider{
		// Dependencies.
		logger: config.Logger,

		// Settings.
		bridgeName: config.BridgeName,
	}

	return newProvider, nil
}

type Provider struct {
	// Dependencies.
	logger micrologger.Logger

	// Settings.
	bridgeName string
}

func (p *Provider) Lookup() ([]provider.PodInfo, error) {
	// We fetch the interface first because it holds all IP addresses associated
	// with it.
	netInterface, err := net.InterfaceByName(p.bridgeName)
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	// The interface addresses have to be parsed to find the actual IPV4 we are
	// interested in.
	ip, err := ipv4FromInterface(netInterface)
	if err != nil {
		return nil, microerror.MaskAny(err)
	}

	// The bridge provider lookup assumes some aspects of our setup. The following
	// explains why we need to increment the bridge IP.
	//
	//     - We use Flannel.
	//     - Flannel creates IP addresses in a deterministic way.
	//     - The IP address after the IP address of the Flannel bridge is the IP
	//       address of the guest cluster VM.
	//
	next := incrIPV4(ip)

	// The bridge provider can only lookup one IP for one pod at a time, because
	// it is running inside a pod and inspects the host network the container is
	// running on.
	podInfos := []provider.PodInfo{
		{
			IP: next,
		},
	}

	return podInfos, nil
}

func incrIPV4(ip net.IP) net.IP {
	c := net.ParseIP(ip.String())

	for j := len(c) - 1; j >= 0; j-- {
		c[j]++
		if c[j] > 0 {
			break
		}
	}

	return c
}

func ipv4FromInterface(netInterface *net.Interface) (net.IP, error) {
	addrs, err := netInterface.Addrs()
	if err != nil {
		return nil, microerror.MaskAny(err)
	}
	for _, addr := range addrs {
		var ip net.IP

		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}

		if ip == nil {
			continue
		}

		ipv4 := ip.To4()
		if ipv4 == nil {
			// Not an ipv4 address.
			continue
		}

		return ipv4, nil
	}

	return nil, errors.New("IPV4 not found")
}
