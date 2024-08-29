package revconproxy

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"

	"github.com/ksysoev/oneway/api/connection"
)

type Config struct {
	NameSpace string
	CtrlAPI   string
	ConnAPI   string
	Services  []ServiceCongfig
}

type ServiceCongfig struct {
	Name    string
	Address string
}

type RCPService struct {
	config   *Config
	srvcIndx map[string]string
	connAPI  *connection.Client
}

func New(cfg *Config) *RCPService {
	srvcIndx := make(map[string]string)
	for _, service := range cfg.Services {
		srvcIndx[service.Name] = service.Address
	}

	return &RCPService{
		config:   cfg,
		srvcIndx: srvcIndx,
	}
}

func (s *RCPService) CreateConnection(ctx context.Context, nameSpace string, serviceName string, id uint64) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	dest, ok := s.srvcIndx[serviceName]
	if !ok {
		return fmt.Errorf("service not found")
	}

	connDest, err := net.Dial("tcp", dest)

	if err != nil {
		slog.Error("failed to dial", slog.Any("error", err))
	}

	defer connDest.Close()

	revConn, err := net.Dial("tcp", s.config.ConnAPI)
	if err != nil {
		slog.Error("failed to dial exchange", slog.Any("error", err))
	}

	defer revConn.Close()

	api := connection.NewConnectionServiceClient(revConn)
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

	return nil
}
