package provider

import (
	"net"
)

type PodInfo struct {
	// IP is the pod IP found associated with the given pod name.
	IP   net.IP
	Name string
}

type Provider interface {
	Lookup() (PodInfo, error)
}
