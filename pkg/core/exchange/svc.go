package exchange

import (
	"context"
	"fmt"
	"net"
	"strings"
)

var ErrConnReqNotFound = fmt.Errorf("connection request not found")

type Service struct {
	revProxyRepo RevProxyRepo
	connQueue    ConnectionQueue
}

type ConnResult struct {
	Conn net.Conn
	Err  error
}

type RevProxyRepo interface {
	AddRevConProxy(proxy *RevConProxy)
	GetRevConProxy(nameSpace string) (*RevConProxy, error)
}

type ConnectionQueue interface {
	AddRequest(connChan chan ConnResult) uint64
	AddConnection(id uint64, conn ConnResult) error
}

func New(revProxyRepo RevProxyRepo, connQueue ConnectionQueue) *Service {
	return &Service{
		revProxyRepo: revProxyRepo,
		connQueue:    connQueue,
	}
}

func (s *Service) NewConnection(ctx context.Context, address string) (net.Conn, error) {
	connChan := make(chan ConnResult, 1)
	id := s.connQueue.AddRequest(connChan)

	splits := strings.Split(address, ".")

	if len(splits) != 2 {
		return nil, fmt.Errorf("invalid address")
	}

	service, nameSpace := splits[0], splits[1]

	proxy, err := s.revProxyRepo.GetRevConProxy(nameSpace)
	if err != nil {
		return nil, fmt.Errorf("failed to get reverse connection proxy: %w", err)
	}

	err = proxy.RequestConnection(ctx, id, service)
	if err != nil {
		return nil, fmt.Errorf("failed to request connection: %w", err)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res, ok := <-connChan:
		if !ok {
			return nil, fmt.Errorf("failed to get connection")
		}

		return res.Conn, res.Err
	}
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
	return s.connQueue.AddConnection(id, ConnResult{
		Conn: conn,
	})
}
