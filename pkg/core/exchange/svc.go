package exchange

import (
	"context"
	"fmt"
	"net"
)

var ErrConnReqNotFound = fmt.Errorf("connection request not found")

type Service struct {
	revProxyRepo RevProxyRepo
}

type ConnResult struct {
	Conn net.Conn
	Err  error
}

type RevProxyRepo interface {
	AddRevConProxy(proxy *RevConProxy)
	GetRevConProxy(nameSpace string) (*RevConProxy, error)
}

func New(revProxyRepo RevProxyRepo) *Service {
	return &Service{
		revProxyRepo: revProxyRepo,
	}
}

func (s *Service) NewConnection(ctx context.Context, address string) (net.Conn, error) {

	return nil, nil
}

func (s *Service) RegisterRevConProxy(ctx context.Context, nameSpace string, services []string) (*RevConProxy, error) {
	proxy, err := NewRevConProxy(nameSpace, services)
	if err != nil {
		return nil, fmt.Errorf("failed to create reverse connection proxy: %w", err)
	}

	s.revProxyRepo.AddRevConProxy(proxy)

	return proxy, nil
}

func (s *Service) AddConnection(id uint64, conn net.Conn) error {
	return nil
}
