package kubernetes

import (
	"github.com/giantswarm/k8s-endpoint-updater/command/update/flag/kubernetes/tls"
)

type Kubernetes struct {
	Address   string
	InCluster bool
	TLS       tls.TLS
}
