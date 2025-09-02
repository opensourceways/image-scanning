package primitive

import "errors"

const (
	registryDocker = "docker.io"
	registryQuay   = "quay.io"
	registryOepky  = "hub.oepkgs.net"
)

type Registry interface {
	String() string
}

func NewRegistry(r string) (Registry, error) {
	if r != registryDocker && r != registryQuay && r != registryOepky {
		return nil, errors.New("unsupported registry")
	}

	return registry(r), nil
}

func CreateRegistry(r string) Registry {
	return registry(r)
}

type registry string

func (r registry) String() string {
	return string(r)
}
