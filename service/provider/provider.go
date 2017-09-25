package provider

import (
	"net"
)

type PodInfo struct {
	// IP is the pod IP found associated with the given pod name.
	IP net.IP
}

type Provider interface {
	Lookup() ([]PodInfo, error)
}
