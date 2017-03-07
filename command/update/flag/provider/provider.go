package provider

import (
	"github.com/giantswarm/k8s-endpoint-updater/command/update/flag/provider/env"
	"github.com/giantswarm/k8s-endpoint-updater/command/update/flag/provider/etcd"
)

type Provider struct {
	Env  env.Env
	Etcd etcd.Etcd
	Kind string
}
