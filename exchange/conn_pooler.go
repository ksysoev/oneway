package main

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"syscall"
)

type ConnectionPooler struct {
	lock   sync.Mutex
	conns  map[uint64]*ConnectCommand
	listen string
}

func NewConnectionPooler(listen string) *ConnectionPooler {
	return &ConnectionPooler{
		listen: listen,
		conns:  make(map[uint64]*ConnectCommand),
	}
}

func (p *ConnectionPooler) WaitForConn(cmd *ConnectCommand) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.conns[cmd.ID] = cmd
}

func (p *ConnectionPooler) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ln, err := net.Listen("tcp", p.listen)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	go func() {
		<-ctx.Done()
		ln.Close()
	}()

	wg := sync.WaitGroup{}
	for ctx.Err() == nil {
		conn, err := ln.Accept()
		if errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.EPIPE) {
			break
		}

		if err != nil {
			return fmt.Errorf("failed to accept connection: %w", err)
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := p.handleConn(conn)
			if err != nil {
				slog.Debug("failed to handle connection", slog.Any("error", err))
			}
		}()
	}

	wg.Wait()
	return nil
}

func (p *ConnectionPooler) handleConn(conn net.Conn) error {
	// read protocol version and authetication method from connenction

	buf := make([]byte, 2)
	n, err := conn.Read(buf)

	if err != nil {
		conn.Close()
		return err
	}

	if n != 2 {
		conn.Close()
		return fmt.Errorf("invalid protocol version and authentication method")
	}

	ver := buf[0]
	authMethod := buf[1]

	if ver != 1 || authMethod != 0 {
		conn.Close()
		return fmt.Errorf("unsupported protocol version and authentication method")
	}

	buf = make([]byte, 8)

	n, err = conn.Read(buf)

	if err != nil {
		conn.Close()
		return err
	}

	if n != 8 {
		conn.Close()
		return fmt.Errorf("invalid command")
	}

	connId := binary.BigEndian.Uint64(buf)

	p.lock.Lock()
	cmd, ok := p.conns[connId]
	p.lock.Unlock()

	if !ok {
		conn.Close()
		return fmt.Errorf("invalid connection id")
	}

	cmd.RespChan <- ConnectCommandResponse{
		Conn: conn,
	}

	return nil
}
