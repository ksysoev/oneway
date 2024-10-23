package client

import (
	"context"
	"log/slog"
	"net"

	"golang.org/x/net/proxy"
	"google.golang.org/grpc"
)

// NewGRPCClient creates a new gRPC client connection.
// It establishes a connection to the specified service address using the provided proxy address and options.
// The proxy address should be in the format "host:port".
// The service address should be in the format "serviceName.nameSpace".
// Additional options can be passed as variadic arguments of type grpc.DialOption.
// The function returns a *grpc.ClientConn and an error.
func NewGRPCClient(proxyAddr, serviceAddr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, nil)
	if err != nil {
		return nil, err
	}

	ctxDialer, ok := dialer.(proxy.ContextDialer)
	if !ok {
		panic("dialer does not implement proxy.ContextDialer")
	}

	opts = append(opts, grpc.WithContextDialer(
		func(ctx context.Context, addr string) (net.Conn, error) {
			slog.Info("dialing", slog.String("address", addr))

			conn, err := ctxDialer.DialContext(ctx, "tcp", serviceAddr+":1")
			if err != nil {
				slog.Error("failed to dial", slog.Any("error", err))
				return nil, err
			}

			slog.Info("connection established", slog.Any("address", addr))

			return conn, nil
		},
	))

	return grpc.Dial("passthrough://", opts...)
}
