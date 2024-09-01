package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"

	"github.com/ksysoev/oneway/example/api"
	"google.golang.org/grpc"
)

type EchoService struct {
	api.UnimplementedEchoServiceServer
}

func (s *EchoService) Echo(_ context.Context, req *api.StringMessage) (*api.StringMessage, error) {
	slog.Info("Echo", slog.String("message", req.Value))

	return &api.StringMessage{Value: req.Value}, nil
}

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill)

	defer cancel()

	grpcServer := grpc.NewServer()
	api.RegisterEchoServiceServer(grpcServer, &EchoService{})

	grpcAddr := os.Getenv("GRPC_LISTEN")
	if grpcAddr == "" {
		slog.Error("GRPC_PORT not provided")
		return
	}

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		slog.Error("Failed to listen", slog.Any("error", err))
		return
	}

	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()

	slog.Info("Server started", slog.String("address", lis.Addr().String()))

	if err := grpcServer.Serve(lis); err != nil {
		slog.Error("Failed to serve", slog.Any("error", err))
	}
}
