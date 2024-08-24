package main

import (
	"context"
	"encoding/binary"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"

	"github.com/ksysoev/oneway/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// RevProxy will be executed on server side  to publish service to the exchange service

var serviceRegistry = make(map[string]string)

const NameSpace = "oneway"

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer cancel()

	serviceRegistry["service1"] = "localhost:8080"

	conn, err := grpc.NewClient("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("failed to dial exchange", slog.Any("error", err))
		return
	}

	exchangeService := api.NewExchangeServiceClient(conn)

	sub, err := exchangeService.RegisterService(ctx, &api.RegisterRequest{
		NameSpace:   NameSpace,
		ServiceName: "service1",
	})

	if err != nil {
		slog.Error("failed to register service", slog.Any("error", err))
		return
	}

	slog.Info("service registered")

	for {
		cmd, err := sub.Recv()
		if err != nil {
			slog.Error("failed to receive command", slog.Any("error", err))
			break
		}

		if cmd.NameSpace != NameSpace {
			slog.Error("invalid namespace")
			continue
		}

		address, ok := serviceRegistry[cmd.ServiceName]
		if !ok {
			slog.Error("service not found")
			continue
		}
		go handleRequest(ctx, cmd, address)
	}

}

func handleRequest(ctx context.Context, cmd *api.ConnectCommand, dest string) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	connDest, err := net.Dial("tcp", dest)

	if err != nil {
		slog.Error("failed to dial", slog.Any("error", err))
	}

	defer connDest.Close()

	revConn, err := net.Dial("tcp", "localhost:9091")
	if err != nil {
		slog.Error("failed to dial exchange", slog.Any("error", err))
	}

	defer revConn.Close()

	err = handleExchangeProto(cmd.Id, revConn)
	if err != nil {
		slog.Error("failed to handle exchange proto", slog.Any("error", err))
	}

	// TODO: in futre we can use splice or sockmap to avoid copying data in user space
	go func() {
		defer cancel()
		_, _ = io.Copy(connDest, revConn)
	}()

	go func() {
		defer cancel()
		_, _ = io.Copy(revConn, connDest)
	}()

	<-ctx.Done()
}

func handleExchangeProto(id uint64, conn net.Conn) error {
	buf := []byte{1, 0}
	_, err := conn.Write(buf)
	if err != nil {
		return err
	}

	buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, id)

	_, err = conn.Write(buf)

	if err != nil {
		return err
	}

	return nil
}
