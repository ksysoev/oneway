package main

import (
	"context"
	"net"
	"sync/atomic"
)

type ConnectCommand struct {
	NameSpace string
	Name      string
	ID        uint64
	RespChan  chan<- ConnectCommandResponse
}

type ConnectCommandResponse struct {
	Conn net.Conn
	Err  error
}

type Service struct {
	NameSpace string
	Name      string
	ctx       context.Context
	currentID atomic.Uint64
	cmdChan   chan<- ConnectCommand
}

func NewService(ctx context.Context, nameSpace, name string, cmdChan chan<- ConnectCommand) *Service {
	return &Service{
		ctx:       ctx,
		NameSpace: nameSpace,
		Name:      name,
		currentID: atomic.Uint64{},
		cmdChan:   cmdChan,
	}
}

func (s *Service) RequestConn(ctx context.Context, name string) (net.Conn, error) {
	id := s.currentID.Add(1)
	respChan := make(chan ConnectCommandResponse, 1)

	cmd := ConnectCommand{
		NameSpace: s.NameSpace,
		Name:      name,
		ID:        id,
		RespChan:  respChan,
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case s.cmdChan <- cmd:
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case resp := <-respChan:
		if resp.Err != nil {
			return nil, resp.Err
		}

		return resp.Conn, nil
	}
}
