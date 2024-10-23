package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/ksysoev/oneway/example/api"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill)

	defer cancel()
	// "localhost:1080", "echoserver.example"
	conn, err := grpc.DialContext(ctx, "localhost:1080", grpc.WithInsecure())

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
