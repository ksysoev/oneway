package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/ksysoev/oneway/api/client"
	"github.com/ksysoev/oneway/example/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill)

	defer cancel()

	conn, err := client.NewGRPCClient("localhost:1080", "echoserver.example", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		slog.Error("failed to dial exchange", slog.Any("error", err))
		return
	}

	defer conn.Close()

	exchangeService := api.NewEchoServiceClient(conn)

	resp, err := exchangeService.Echo(ctx, &api.StringMessage{Value: "Hello, world!"})
	if err != nil {
		slog.Error("failed to send message", slog.Any("error", err))
		return
	}

	slog.Info("response received", slog.Any("response", resp))
}
