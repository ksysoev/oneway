package exchange

import (
	"context"
	"fmt"
)

type RevProxy struct {
	cmdStream chan RevProxyCommand
	NameSpace string
	Services  []string
}

type RevProxyCommand struct {
	NameSpace string
	Name      string
	ConnID    uint64
}

func NewRevProxy(nameSpace string, services []string) (*RevProxy, error) {
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

	return &RevProxy{
		NameSpace: nameSpace,
		Services:  services,
		cmdStream: make(chan RevProxyCommand),
	}, nil
}

func (r *RevProxy) CommandStream() <-chan RevProxyCommand {
	return r.cmdStream
}

func (r *RevProxy) RequestConnection(ctx context.Context, id uint64, name string) error {
	cmd := RevProxyCommand{
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
