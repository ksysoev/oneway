package revconproxy

import (
	"context"
	"fmt"
	"sync"

	"github.com/ksysoev/oneway/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type rcpService interface {
	CreateConnection(ctx context.Context, nameSpace string, serviceName string, id uint64) error
}

type ServiceCongfig struct {
	Name    string
	Address string
}

type Config struct {
	NameSpace string
	CtrlAPI   string
	ConnAPI   string
	Services  []ServiceCongfig
}

type Proxy struct {
	config  *Config
	rcpServ rcpService
}

func New(cfg *Config, rcpServ rcpService) *Proxy {
	return &Proxy{
		config:  cfg,
		rcpServ: rcpServ,
	}
}

func (s *Proxy) Run(ctx context.Context) error {
	conn, err := grpc.NewClient(s.config.CtrlAPI, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to dial control api: %w", err)
	}

	exchangeService := api.NewExchangeServiceClient(conn)

	serviceNames := make([]string, 0, len(s.config.Services))
	for _, service := range s.config.Services {
		serviceNames = append(serviceNames, service.Name)
	}

	sub, err := exchangeService.RegisterService(ctx, &api.RegisterRequest{
		NameSpace:   s.config.NameSpace,
		ServiceName: serviceNames,
	})

	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	wg := sync.WaitGroup{}
	defer wg.Wait()

	for {
		cmd, err := sub.Recv()
		if err != nil {
			return fmt.Errorf("failed to receive command: %w", err)
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			s.ConnectCommandHandler(ctx, cmd)
		}()
	}
}
