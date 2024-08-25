package main

import (
	"context"
	"encoding/binary"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"

	"github.com/ksysoev/oneway/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// RevProxy will be executed on server side  to publish service to the exchange service

var serviceRegistry = make(map[string]string)

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer cancel()

	parseServices()
	if len(serviceRegistry) == 0 {
		slog.Error("no services provided")
		return
	}

	exchangeManageAPI := os.Getenv("EXCHANGE_MANAGE_API")
	exchangeConnectionAPI := os.Getenv("EXCHANGE_CONNECTION_API")

	if exchangeManageAPI == "" || exchangeConnectionAPI == "" {
		slog.Error("exchange manage or connection api not provided")
		return
	}

	nameSpace := os.Getenv("NAMESPACE")

	if nameSpace == "" {
		slog.Error("namespace not provided")
		return
	}

	conn, err := grpc.NewClient(exchangeManageAPI, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("failed to dial exchange", slog.Any("error", err))
		return
	}

	exchangeService := api.NewExchangeServiceClient(conn)

	serviceNames := make([]string, 0, len(serviceRegistry))
	for name := range serviceRegistry {
		serviceNames = append(serviceNames, name)
	}

	sub, err := exchangeService.RegisterService(ctx, &api.RegisterRequest{
		NameSpace:   nameSpace,
		ServiceName: serviceNames,
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

		if cmd.NameSpace != nameSpace {
			slog.Error("invalid namespace")
			continue
		}

		address, ok := serviceRegistry[cmd.ServiceName]
		if !ok {
			slog.Error("service not found")
			continue
		}
		go handleRequest(ctx, exchangeConnectionAPI, cmd, address)
	}

}

func handleRequest(ctx context.Context, connAPI string, cmd *api.ConnectCommand, dest string) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	connDest, err := net.Dial("tcp", dest)

	if err != nil {
		slog.Error("failed to dial", slog.Any("error", err))
	}

	defer connDest.Close()

	revConn, err := net.Dial("tcp", connAPI)
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

func parseServices() {
	type Service struct {
		Name    string
		Address string
	}

	serviceMap := make(map[string]Service)

	for _, env := range os.Environ() {

		splits := strings.Split(env, "=")
		if len(splits) != 2 {
			continue
		}
		name, value := splits[0], splits[1]

		if strings.HasPrefix(name, "SERVICE_NAME") {
			serviceID := strings.TrimPrefix(name, "SERVICE_NAME")
			svc, ok := serviceMap[serviceID]
			if !ok {
				svc = Service{}
			}
			svc.Name = value
			serviceMap[serviceID] = svc
		}

		if strings.HasPrefix(name, "SERVICE_ADDRESS") {
			serviceID := strings.TrimPrefix(name, "SERVICE_ADDRESS")
			svc, ok := serviceMap[serviceID]
			if !ok {
				svc = Service{}
			}
			svc.Address = value
			serviceMap[serviceID] = svc
		}
	}

	for _, svc := range serviceMap {
		serviceRegistry[svc.Name] = svc.Address
	}
}
