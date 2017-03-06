package provider

import (
	"net"
)

type PodInfo struct {
	// IP is the pod IP found associated with the given pod name.
	IP net.IP
	// Name is the pod name found associated with the given pod IP.
	Name string
}

type Provider interface {
	Lookup() ([]PodInfo, error)
}
