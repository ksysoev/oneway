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
		connAPI:  connection.NewClient(cfg.ConnAPI),
	}
}

func (s *RCPService) NameSpace() string {
	return s.config.NameSpace
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

	revConn, err := s.connAPI.Connect(id)
	if err != nil {
		slog.Error("failed to dial exchange", slog.Any("error", err))
	}

	defer revConn.Close()

	if err != nil {
		slog.Error("failed to handle exchange proto", slog.Any("error", err))
	}

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

func (s *RCPService) ServiceNames() []string {
	serviceNames := make([]string, 0, len(s.config.Services))
	for _, service := range s.config.Services {
		serviceNames = append(serviceNames, service.Name)
	}
	return serviceNames
}
