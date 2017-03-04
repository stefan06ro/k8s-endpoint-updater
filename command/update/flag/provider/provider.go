package provider

import (
	"github.com/giantswarm/k8s-endpoint-updater/command/update/flag/provider/env"
)

type Provider struct {
	Env  env.Env
	Kind string
}
