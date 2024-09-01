package exchange

import (
	"context"
	"fmt"
)

type RevConProxy struct {
	cmdStream chan RevConProxyCommand
	NameSpace string
	Services  []string
}

type RevConProxyCommand struct {
	NameSpace string
	Name      string
	ConnID    uint64
}

func NewRevConProxy(nameSpace string, services []string) (*RevConProxy, error) {
	if nameSpace == "" {
		return nil, fmt.Errorf("name space is empty")
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("services list is empty")
	}

	uniqIndex := make(map[string]struct{})
	for _, service := range services {
		if service == "" {
			return nil, fmt.Errorf("service name is empty")
		}

		if _, ok := uniqIndex[service]; ok {
			return nil, fmt.Errorf("duplicate service name %s", service)
		}
	}

	return &RevConProxy{
		NameSpace: nameSpace,
		Services:  services,
		cmdStream: make(chan RevConProxyCommand),
	}, nil
}

func (r *RevConProxy) CommandStream() <-chan RevConProxyCommand {
	return r.cmdStream
}

func (r *RevConProxy) RequestConnection(ctx context.Context, id uint64, name string) error {
	cmd := RevConProxyCommand{
		NameSpace: r.NameSpace,
		Name:      name,
		ConnID:    id,
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case r.cmdStream <- cmd:
		return nil
	}
}
