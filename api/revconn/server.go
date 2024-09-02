package revconn

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"syscall"
)

const (
	maxConcurency = 25
)

type token struct{}

type OnConnectCB func(id uint64, conn net.Conn) error

type Server struct {
	onConnect OnConnectCB
	sem       chan struct{}
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

	if err = s.onConnect(id, conn); err != nil {
		slog.Error("failed to handle connection", slog.Any("error", err))
		conn.Close()

		return
	}
}

func (s *Server) initialize(conn io.ReadWriter) error {
	buf := make([]byte, connectionInitLength)
	n, err := conn.Read(buf)

	if err != nil {
		return fmt.Errorf("failed to read protocol version and authentication method: %w", err)
	}

	if n != connectionInitLength {
		return fmt.Errorf("invalid protocol version and authentication method")
	}

	ver := Version(buf[0])
	authMethod := AuthMethod(buf[1])

	if ver != V1 {
		return fmt.Errorf("unsupported protocol version")
	}

	if authMethod != NoAuth {
		if _, err = conn.Write([]byte{byte(V1), byte(NoAcceptableAuthMethod)}); err != nil {
			return fmt.Errorf("failed to write protocol version and authentication method: %w", err)
		}

		return fmt.Errorf("unsupported authentication method")
	}

	_, err = conn.Write([]byte{byte(V1), byte(NoAuth)})

	if err != nil {
		return fmt.Errorf("failed to write protocol version and authentication method: %w", err)
	}

	return nil
}

func (s *Server) getConnectionID(conn net.Conn) (uint64, error) {
	buf := make([]byte, connectionIDLenght+1)

	n, err := conn.Read(buf)

	if err != nil {
		return 0, fmt.Errorf("failed to read connection id: %w", err)
	}

	if n != connectionIDLenght+1 {
		return 0, fmt.Errorf("invalid connection id")
	}

	ver := Version(buf[0])
	if ver != V1 {
		return 0, fmt.Errorf("unsupported protocol version")
	}

	return binary.BigEndian.Uint64(buf[1:]), nil
}
