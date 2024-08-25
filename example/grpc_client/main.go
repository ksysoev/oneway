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

	conn, err := grpc.NewClient("passthrough://",
		grpc.WithContextDialer(
			func(ctx context.Context, addr string) (net.Conn, error) {
				slog.Info("dialing", slog.String("address", addr))

				conn, err := ctxDialer.DialContext(ctx, "tcp", "echoserver:9095")
				if err != nil {
					slog.Error("failed to dial", slog.Any("error", err))
					return nil, err
				}

				slog.Info("connection established", slog.Any("address", addr))
				return conn, nil
			},
		),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

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
