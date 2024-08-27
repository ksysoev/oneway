package exchange

import (
	"context"
	"fmt"
	"net"
)

var ErrConnReqNotFound = fmt.Errorf("connection request not found")

type RevConProxyCommand struct {
	NameSpace string
	Name      string
	ConnID    uint64
}

type Service struct{}

type RevConProxy struct {
	NameSpace string
}

type ConnResult struct {
	Conn net.Conn
	Err  error
}

func New() *Service {
	return &Service{}
}

func (s *Service) NewConnection(ctx context.Context, address string) (net.Conn, error) {
	return nil, nil
}

func (s *Service) RegisterRevConProxy(ctx context.Context, nameSpace string, services []string) (*RevConProxy, error) {
	return nil, nil
}
