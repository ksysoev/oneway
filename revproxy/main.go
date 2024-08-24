package main

import (
	"context"
	"encoding/binary"
	"log/slog"
	"net"
	"os"
	"os/signal"

	"github.com/ksysoev/oneway/api"
	"google.golang.org/grpc"
)

// RevProxy will be executed on server side  to publish service to the exchange service

var serviceRegistry = make(map[string]string)

const NameSpace = "oneway"

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer cancel()

	serviceRegistry["service1"] = "localhost:8080"

	conn, err := grpc.NewClient("localhost:9090")

	exchangeService := api.NewExchangeServiceClient(conn)

	sub, err := exchangeService.RegisterService(ctx, &api.RegisterRequest{
		NameSpace:   NameSpace,
		ServiceName: "service1",
	})

	if err != nil {
		panic(err)
	}

	for {
		cmd, err := sub.Recv()
		if err != nil {
			panic(err)
		}

		if cmd.NameSpace != NameSpace {
			panic("invalid namespace")
		}

		address, ok := serviceRegistry[cmd.ServiceName]
		if !ok {
			panic("service not found")
		}
		go handleRequest(cmd, address)
	}

}

func handleRequest(cmd *api.ConnectCommand, dest string) {
	connDest, err := net.Dial("tcp", dest)

	if err != nil {
		slog.Error("failed to dial", slog.Any("error", err))
	}

	revConn, err := net.Dial("tcp", "localhost:9091")
	if err != nil {
		slog.Error("failed to dial exchange", slog.Any("error", err))
	}

	err = handleExchangeProto(cmd.Id, revConn)
	if err != nil {
		slog.Error("failed to handle exchange proto", slog.Any("error", err))
	}

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
