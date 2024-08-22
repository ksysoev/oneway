package main

import (
	"context"
	"log"
	"log/slog"
	"net"
	"sync"

	"golang.org/x/build/revdial/v2"
	"tailscale.com/net/socks5"
)

var (
	Dialer *revdial.Dialer
)

// Exchange will be between the client and the server and will be routing and multiplexing client connections to the correct service
func main() {
	ctx := context.Background()
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := revConProxyListener(ctx)

		if err != nil {
			slog.Error(err.Error())
		}
	}()

	wg.Wait()
}

func revConProxyListener(ctx context.Context) error {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	for ctx.Err() == nil {
		conn, err := lis.Accept()
		if err != nil {
			log.Fatalf("failed to accept: %v", err)
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			handleConnection(ctx, conn)

		}()
	}

	return nil
}

func clientConnProxyListener(ctx context.Context) error {
	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		return err
	}

	proxy := socks5.Server{
		Dialer: dialler,
	}

	return proxy.Serve(lis)
}

func dialler(ctx context.Context, _, _ string) (net.Conn, error) {
	return Dialer.Dial(ctx)
}

func handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	Dialer = revdial.NewDialer(conn, "")

	<-ctx.Done()
}
