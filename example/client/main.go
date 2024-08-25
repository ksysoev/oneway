package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"

	"github.com/ksysoev/oneway/example/api"
	"golang.org/x/net/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer cancel()

	dialer, err := proxy.SOCKS5("tcp", "localhost:1080", nil, nil)
	ctxDialer, ok := dialer.(proxy.ContextDialer)
	if !ok {
		panic("dialer does not implement proxy.ContextDialer")
	}

	conn, err := grpc.NewClient("echoservice:0",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(
			func(ctx context.Context, addr string) (net.Conn, error) {
				return ctxDialer.DialContext(ctx, "tcp", addr)
			},
		))

	if err != nil {
		slog.Error("failed to dial exchange", slog.Any("error", err))
		return
	}

	exchangeService := api.NewEchoServiceClient(conn)

	resp, err := exchangeService.Echo(ctx, &api.StringMessage{Value: "Hello, world!"})
	if err != nil {
		slog.Error("failed to send message", slog.Any("error", err))
		return
	}

	slog.Info("response received", slog.Any("response", resp))
}
