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
	NameSpace() string
	ServiceNames() []string
	CreateConnection(ctx context.Context, nameSpace string, serviceName string, id uint64) error
}

type Proxy struct {
	rcpServ rcpService
	ctrlAPI string
}

func New(rcpServ rcpService, ctrlAPI string) *Proxy {
	return &Proxy{
		ctrlAPI: ctrlAPI,
		rcpServ: rcpServ,
	}
}

func (s *Proxy) Run(ctx context.Context) error {
	conn, err := grpc.NewClient(s.ctrlAPI, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to dial control api: %w", err)
	}

	exchangeService := api.NewExchangeServiceClient(conn)

	sub, err := exchangeService.RegisterService(ctx, &api.RegisterRequest{
		NameSpace:   s.rcpServ.NameSpace(),
		ServiceName: s.rcpServ.ServiceNames(),
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
