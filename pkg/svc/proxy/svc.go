package proxy

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"

	"github.com/ksysoev/oneway/pkg/core/network"
	"go.opentelemetry.io/otel"
	"tailscale.com/net/socks5"
)

type ExchangeService interface {
	NewConnection(ctx context.Context, address *network.Address) (net.Conn, error)
}

type Server interface {
	Serve(net.Listener) error
}

type Config struct {
	Listen string
}

type Service struct {
	srv      Server
	listener net.Listener
	exchange ExchangeService
	addr     string
	l        sync.Mutex
}

func New(cfg *Config, exchange ExchangeService) *Service {
	svc := &Service{
		addr:     cfg.Listen,
		exchange: exchange,
		l:        sync.Mutex{},
	}

	svc.srv = &socks5.Server{
		Logf:   svc.proxyLogf,
		Dialer: svc.dial,
	}

	return svc
}

var tracer = otel.Tracer("github.com/ksysoev/oneway/pkg/svc/proxy")

func (s *Service) dial(ctx context.Context, _, address string) (net.Conn, error) {
	ctx, span := tracer.Start(ctx, "Proxy.Dial")
	defer span.End()

	addr, err := network.ParseAddress(address)
	if err != nil {
		return nil, fmt.Errorf("failed to parse address: %w", err)
	}

	conn, err := s.exchange.NewConnection(ctx, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to get service for %s: %w", address, err)
	}

	return conn, nil
}

func (s *Service) proxyLogf(format string, args ...any) {
	slog.Info(fmt.Sprintf(format, args...))
}

func (s *Service) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.l.Lock()
	s.listener = lis
	s.l.Unlock()

	slog.Info("Proxy Server started", slog.String("address", lis.Addr().String()))

	go func() {
		<-ctx.Done()

		if err := lis.Close(); err != nil {
			slog.Error("Failed to close Proxy listener", slog.Any("error", err))
		}
	}()

	return s.srv.Serve(lis)
}

func (s *Service) Close() error {
	s.l.Lock()
	defer s.l.Unlock()

	return s.listener.Close()
}

func (s *Service) Addr() string {
	s.l.Lock()
	defer s.l.Unlock()

	if s.listener == nil {
		return ""
	}

	return s.listener.Addr().String()
}
