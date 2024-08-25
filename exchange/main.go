package main

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/ksysoev/oneway/api"
	"google.golang.org/grpc"
	"tailscale.com/net/socks5"
)

type ExchangeService struct {
	ctx     context.Context
	lock    sync.Mutex
	srvs    map[string]*Service
	pooller *ConnectionPooler
	api.UnimplementedExchangeServiceServer
}

func (s *ExchangeService) RegisterService(req *api.RegisterRequest, stream grpc.ServerStreamingServer[api.ConnectCommand]) error {
	commandChan := make(chan ConnectCommand)
	s.lock.Lock()
	for _, name := range req.ServiceName {
		srv := NewService(stream.Context(), req.NameSpace, name, commandChan)
		slog.Info("Registering service", slog.String("namespace", req.NameSpace), slog.String("service", name))
		s.srvs[name] = srv
	}
	s.lock.Unlock()

	for {
		select {
		case <-s.ctx.Done():
			return nil
		case <-stream.Context().Done():
			return nil
		case cmd, ok := <-commandChan:
			if !ok {
				return nil
			}
			if _, ok := s.srvs[cmd.Name]; !ok {
				slog.Error("invalid service name", slog.String("name", cmd.Name))

				cmd.RespChan <- ConnectCommandResponse{Err: fmt.Errorf("service not found")}
				continue
			}

			s.pooller.WaitForConn(&cmd)

			err := stream.Send(&api.ConnectCommand{
				NameSpace:   cmd.NameSpace,
				ServiceName: cmd.Name,
				Id:          cmd.ID,
			})

			if err != nil {
				return fmt.Errorf("failed to send command: %w", err)
			}
		}
	}
}

func (s *ExchangeService) GetService(ctx context.Context, address string) (*Service, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	srv, ok := s.srvs[address]
	if !ok {
		for k, v := range s.srvs {
			slog.Info("Service", slog.String("key", k), slog.Any("value", v))
		}

		return nil, fmt.Errorf("service not found")
	}

	return srv, nil
}

// Exchange will be between the client and the server and will be routing and multiplexing client connections to the correct service
func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer cancel()

	errs := make(chan error, 1)

	go func() { errs <- startAPI(ctx) }()

	if err := <-errs; err != nil {
		slog.Error("Failed to start exchange service", slog.Any("error", err))
	}
}

func startAPI(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	connAPI := os.Getenv("CONNECTION_API")
	manageAPI := os.Getenv("MANAGE_API")
	proxyServer := os.Getenv("PROXY_SERVER")

	if connAPI == "" || manageAPI == "" || proxyServer == "" {
		return fmt.Errorf("connection, manage or proxy server api not provided")
	}

	pooler := NewConnectionPooler(connAPI)

	exchange := &ExchangeService{
		srvs:    make(map[string]*Service),
		pooller: pooler,
		ctx:     ctx,
	}

	grpcServer := grpc.NewServer()
	api.RegisterExchangeServiceServer(grpcServer, exchange)

	socks5Server := socks5.Server{
		Dialer: func(ctx context.Context, network, address string) (net.Conn, error) {
			slog.Info("Dialing", slog.String("address", address))

			address, _, err := net.SplitHostPort(address)

			slog.Info("GetService", slog.String("address", address))

			if err != nil {
				return nil, fmt.Errorf("failed to split address: %w", err)
			}

			srv, err := exchange.GetService(ctx, address)
			if err != nil {
				slog.Error("Failed to get service", slog.Any("error", err))
				return nil, fmt.Errorf("failed to get service: %w", err)
			}

			conn, err := srv.RequestConn(ctx, address)
			if err != nil {
				slog.Error("Failed to request connection", slog.Any("error", err))
				return nil, fmt.Errorf("failed to request connection: %w", err)
			}

			slog.Info("Connection established", slog.String("address", address))
			return conn, nil
		},
	}

	lis, err := net.Listen("tcp", manageAPI)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()

	go func() {
		defer cancel()
		err := pooler.Run(ctx)
		if err != nil {
			slog.Error("Failed to run connection pooler", slog.Any("error", err))
		}
	}()

	go func() {
		defer cancel()

		lis, err := net.Listen("tcp", proxyServer)
		if err != nil {
			slog.Error("Failed to listen", slog.Any("error", err))
			return
		}

		slog.Info("SOCKS5 server started", slog.Int64("port", 1080))

		go func() {
			<-ctx.Done()
			lis.Close()
		}()

		err = socks5Server.Serve(lis)
		if errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.EPIPE) {
			return
		}

		if err != nil {
			slog.Error("Failed to serve", slog.Any("error", err))
		}
	}()

	slog.Info("Exchange service started", slog.Int64("port", 9090))
	err = grpcServer.Serve(lis)
	if err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

type connection struct {
	conn net.Conn
	id   uint64
}

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

	slog.Info("Connection pooler started", slog.String("address", p.listen))

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
