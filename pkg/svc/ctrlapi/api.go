package ctrlapi

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/ksysoev/oneway/api"
	"github.com/ksysoev/oneway/pkg/core/exchange"
	"google.golang.org/grpc"
)

type ExchangeService interface {
	RegisterRevConProxy(ctx context.Context, nameSpace string, services []string) (*exchange.RevConProxy, error)
}

type API struct {
	api.UnimplementedExchangeServiceServer
	exchange ExchangeService
	listen   string
}

type Config struct {
	Listen string
}

func New(cfg *Config, exchange ExchangeService) *API {
	return &API{
		exchange: exchange,
		listen:   cfg.Listen,
	}
}

func (a *API) Run(ctx context.Context) error {
	srv := grpc.NewServer()
	api.RegisterExchangeServiceServer(srv, a)

	grpcServer := grpc.NewServer()
	api.RegisterExchangeServiceServer(grpcServer, a)

	lis, err := net.Listen("tcp", a.listen)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()

	slog.Info("Control API started", slog.String("address", lis.Addr().String()))

	return grpcServer.Serve(lis)
}

func (a *API) RegisterService(req *api.RegisterRequest, stream grpc.ServerStreamingServer[api.ConnectCommand]) error {
	rcp, err := a.exchange.RegisterRevConProxy(stream.Context(), req.NameSpace, req.ServiceName)
	if err != nil {
		return err
	}

	cmdStream := rcp.CommandStream()

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case cmd, ok := <-cmdStream:
			if !ok {
				return nil
			}

			err := stream.Send(&api.ConnectCommand{
				NameSpace:   cmd.NameSpace,
				ServiceName: cmd.Name,
				Id:          cmd.ConnID,
			})

			if err != nil {
				return fmt.Errorf("failed to send command: %w", err)
			}
		}
	}
}
