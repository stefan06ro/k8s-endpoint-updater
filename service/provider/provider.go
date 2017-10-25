package provider

import (
	"net"
)

type Provider interface {
	Lookup() (net.IP, error)
}
