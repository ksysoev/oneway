package ctrlapi

import (
	"github.com/ksysoev/oneway/api"
	"google.golang.org/grpc"
)

type ExchangeService interface{}

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
	return nil
}
