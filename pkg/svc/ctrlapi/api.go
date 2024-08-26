package ctrlapi

import (
	"context"
	"fmt"

	"github.com/ksysoev/oneway/api"
	"github.com/ksysoev/oneway/pkg/core/exchange"
	"google.golang.org/grpc"
)

type RevConProxy interface {
	CommandStream() <-chan exchange.RevConProxyCommand
}

type ExchangeService interface {
	RegisterRevConProxy(ctx context.Context, nameSpace string, services []string) (RevConProxy, error)
}

type API struct {
	exchange ExchangeService
	api.UnimplementedExchangeServiceServer
}

func New(exchange ExchangeService) *API {
	return &API{
		exchange: exchange,
	}
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
