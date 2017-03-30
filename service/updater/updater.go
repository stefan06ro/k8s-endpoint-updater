package updater

import (
	"net"

	microerror "github.com/giantswarm/microkit/error"
	micrologger "github.com/giantswarm/microkit/logger"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"

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

func (p *Updater) Update(namespace, service string, podInfos []provider.PodInfo) error {
	for _, pi := range podInfos {
		s, err := p.kubernetesClient.Services(namespace).Get(service, metav1.GetOptions{})
		if err != nil {
			return microerror.MaskAny(err)
		}

		endpoint := &apiv1.Endpoints{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: service,
			},
			Subsets: []apiv1.EndpointSubset{
				{
					Addresses: []apiv1.EndpointAddress{
						{
							IP: pi.IP.String(),
						},
					},
					Ports: serviceToPorts(s),
				},
			},
		}

		_, err = p.kubernetesClient.Endpoints(namespace).Create(endpoint)
		if errors.IsAlreadyExists(err) {
			// In case the endpoint we tried to create does already exist, we only
			// need to append the IPs we got in podInfos.
			err := p.appendIPs(namespace, service, podInfos)
			if err != nil {
				return microerror.MaskAny(err)
			}
		} else if err != nil {
			return microerror.MaskAny(err)
		}
	}

	return nil
}

func (p *Updater) appendIPs(namespace, service string, podInfos []provider.PodInfo) error {
	endpoints, err := p.kubernetesClient.Endpoints(namespace).List(metav1.ListOptions{})
	if err != nil {
		return microerror.MaskAny(err)
	}

	for i, e := range endpoints.Items {
		if e.Name != service {
			// In case the service is set to "worker" and the endpoint name is
			// "master" we skip until we find the right endpoint.
			continue
		}

		for j, _ := range e.Subsets {
			for _, pi := range podInfos {
				addresses := endpoints.Items[i].Subsets[j].Addresses
				if ipInAddresses(addresses, pi.IP) {
					continue
				}

				address := apiv1.EndpointAddress{
					IP: pi.IP.String(),
				}

				addresses = append(addresses, address)
				endpoints.Items[i].Subsets[j].Addresses = addresses
			}
		}

		_, err = p.kubernetesClient.Endpoints(namespace).Update(&endpoints.Items[i])
		if err != nil {
			return microerror.MaskAny(err)
		}
	}

	return nil
}

func ipInAddresses(addresses []apiv1.EndpointAddress, IP net.IP) bool {
	for _, a := range addresses {
		if a.IP == IP.String() {
			return true
		}
	}

	return false
}

func podInfoByName(podInfos []provider.PodInfo, name string) (provider.PodInfo, error) {
	for _, pi := range podInfos {
		if pi.Name == name {
			return pi, nil
		}
	}

	return provider.PodInfo{}, microerror.MaskAnyf(executionFailedError, "pod info for name '%s' not found", name)
}

func serviceToPorts(s *apiv1.Service) []apiv1.EndpointPort {
	var ports []apiv1.EndpointPort

	for _, p := range s.Spec.Ports {
		port := apiv1.EndpointPort{
			Name: p.Name,
			Port: p.Port,
		}

		ports = append(ports, port)
	}

	return ports
}
