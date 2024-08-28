package connapi

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/ksysoev/oneway/api/connection"
)

type ExchangeService interface {
	AddConnection(id uint64, conn net.Conn) error
}

type API struct {
	listen   string
	exchange ExchangeService
}

type Config struct {
	Listen string
}

func New(cfg *Config, exchange ExchangeService) *API {
	return &API{
		listen: cfg.Listen,
	}
}

func (a *API) Run(ctx context.Context) error {
	lis, err := net.Listen("tcp", a.listen)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	go func() {
		<-ctx.Done()
		err := lis.Close()
		if err != nil {
			slog.Error("Failed to close listener", slog.Any("error", err))
		}
	}()

	slog.Info("Connection API started", slog.String("address", lis.Addr().String()))

	connAPI := connection.NewServer(a.ConnectionHandler)

	return connAPI.Serve(lis)
}

func (a *API) ConnectionHandler(id uint64, conn net.Conn) {
	err := a.exchange.AddConnection(id, conn)
	if err != nil {
		slog.Error("Failed to add connection", slog.Any("error", err))
	}
}
