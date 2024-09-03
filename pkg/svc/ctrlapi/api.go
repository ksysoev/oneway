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
	RegisterRevProxy(ctx context.Context, nameSpace string, services []string) (*exchange.RevProxy, error)
	UnregisterRevProxy(proxy *exchange.RevProxy)
}

type API struct {
	api.UnimplementedExchangeServiceServer
	exchange ExchangeService
	listen   string
}

type Config struct {
	Listen string
}

func New(cfg *Config, exchangeSvc ExchangeService) *API {
	return &API{
		exchange: exchangeSvc,
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

	slog.InfoContext(ctx, "Control API started", slog.String("address", lis.Addr().String()))

	return grpcServer.Serve(lis)
}

func (a *API) RegisterService(req *api.RegisterRequest, stream grpc.ServerStreamingServer[api.ConnectCommand]) error {
	rcp, err := a.exchange.RegisterRevProxy(stream.Context(), req.NameSpace, req.ServiceName)
	if err != nil {
		return err
	}

	defer a.exchange.UnregisterRevProxy(rcp)

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
