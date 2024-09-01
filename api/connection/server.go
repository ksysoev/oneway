package connection

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"syscall"
)

const (
	maxConcurency = 25
)

type token struct{}

type OnConnectCB func(id uint64, conn net.Conn)

type Server struct {
	onConnect OnConnectCB
	sem       chan struct{}
	listen    string
}

func NewServer(cb OnConnectCB) *Server {
	return &Server{
		onConnect: cb,
		sem:       make(chan struct{}, maxConcurency),
	}
}

func (s *Server) Serve(lis net.Listener) error {
	wg := sync.WaitGroup{}
	defer wg.Wait()

	for {
		s.sem <- token{}
		conn, err := lis.Accept()

		if err != nil {
			if errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.EPIPE) {
				return nil
			}

			return fmt.Errorf("failed to accept connection: %w", err)
		}

		wg.Add(1)
		go func() {
			defer func() {
				<-s.sem
				wg.Done()
			}()

			s.handleConn(conn)
		}()
	}
}

func (s *Server) handleConn(conn net.Conn) {
	err := s.initialize(conn)
	if err != nil {
		slog.Error("failed to initialize connection", slog.Any("error", err))
		conn.Close()

		return
	}

	id, err := s.getConnectionID(conn)
	if err != nil {
		slog.Error("failed to get connection id", slog.Any("error", err))
		conn.Close()

		return
	}

	s.onConnect(id, conn)
}

func (s *Server) initialize(conn net.Conn) error {
	buf := make([]byte, 2)
	n, err := conn.Read(buf)

	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to read protocol version and authentication method: %w", err)
	}

	if n != 2 {
		conn.Close()
		return fmt.Errorf("invalid protocol version and authentication method")
	}

	ver := Version(buf[0])
	authMethod := AuthMethod(buf[1])

	if ver != V1 {
		conn.Close()
		return fmt.Errorf("unsupported protocol version")
	}

	if authMethod != NoAuth {
		conn.Close()
		return fmt.Errorf("unsupported authentication method")
	}

	return nil
}

func (s *Server) getConnectionID(conn net.Conn) (uint64, error) {
	buf := make([]byte, 8)

	n, err := conn.Read(buf)

	if err != nil {
		conn.Close()
		return 0, fmt.Errorf("failed to read connection id: %w", err)
	}

	if n != 8 {
		conn.Close()
		return 0, fmt.Errorf("invalid connection id")
	}

	return binary.BigEndian.Uint64(buf), nil
}
