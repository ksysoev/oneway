package main

import (
	"context"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"

	"github.com/ksysoev/oneway/api"
	"google.golang.org/grpc"
)

type ExchangeService struct{}

func (s *ExchangeService) RegisterService(ctx context.Context, in *api.RegisterRequest, opts ...grpc.CallOption) (api.ExchangeService_RegisterServiceClient, error) {
	return nil, nil
}

// Exchange will be between the client and the server and will be routing and multiplexing client connections to the correct service
func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer cancel()

	service := api.NewExchangeServiceService(&ExchangeService{})

	grpcServer := grpc.NewServer()
	api.RegisterExchangeServiceService(grpcServer, service)

	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()

	slog.Info("Exchange service started", slog.Int64("port", 9090))
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
