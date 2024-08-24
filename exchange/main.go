package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"sync"

	"github.com/ksysoev/oneway/api"
	"google.golang.org/grpc"
)

type ExchangeService struct {
	lock sync.Mutex
	srvs map[string]*Service
	api.UnimplementedExchangeServiceServer
}

func (s *ExchangeService) RegisterService(req *api.RegisterRequest, stream grpc.ServerStreamingServer[api.ConnectCommand]) error {
	commandChan := make(chan ConnectCommand)
	srv := NewService(stream.Context(), req.NameSpace, req.ServiceName, commandChan)

	s.lock.Lock()
	s.srvs[req.ServiceName] = srv
	s.lock.Unlock()

	for cmd := range commandChan {
		if cmd.NameSpace != req.NameSpace || cmd.Name != req.ServiceName {
			cmd.RespChan <- ConnectCommandResponse{Err: fmt.Errorf("service not found")}
			continue
		}

		err := stream.Send(&api.ConnectCommand{
			NameSpace:   cmd.NameSpace,
			ServiceName: cmd.Name,
			Id:          cmd.ID,
		})

		if err != nil {
			return fmt.Errorf("failed to send command: %w", err)
		}
	}

	return nil
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
	grpcServer := grpc.NewServer()
	api.RegisterExchangeServiceServer(grpcServer, &ExchangeService{})

	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
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

func startConnectionListener(ctx context.Context) error {
	lis, err := net.Listen("tcp", ":9091")
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	go func() {
		<-ctx.Done()
		lis.Close()
	}()

	for {
		conn, err := lis.Accept()
		if err != nil {
			return fmt.Errorf("failed to accept connection: %w", err)
		}

	}
}
