package kubernetes

import (
	"github.com/giantswarm/k8s-endpoint-updater/command/update/flag/kubernetes/cluster"
	"github.com/giantswarm/k8s-endpoint-updater/command/update/flag/kubernetes/pod"
	"github.com/giantswarm/k8s-endpoint-updater/command/update/flag/kubernetes/tls"
)

type Kubernetes struct {
	Address   string
	Cluster   cluster.Cluster
	InCluster bool
	Pod       pod.Pod
	TLS       tls.TLS
}
