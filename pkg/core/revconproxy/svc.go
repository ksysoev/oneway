package revconproxy

import (
	"context"
	"fmt"
	"io"
)

type BridgeProvider interface {
	CreateConnection(ctx context.Context, id uint64, addr string) (src, dest io.ReadWriteCloser, err error)
}

type Config struct {
	NameSpace string           `yaml:"namespace"`
	CtrlAPI   string           `mapstructure:"ctrl_api"`
	Services  []ServiceCongfig `yaml:"services"`
}

type ServiceCongfig struct {
	Name    string `yaml:"name"`
	Address string `yaml:"address"`
}

type RCPService struct {
	config     *Config
	srvcIndx   map[string]string
	bridgeProv BridgeProvider
}

func New(cfg *Config, bridgeProv BridgeProvider) *RCPService {
	srvcIndx := make(map[string]string)
	for _, service := range cfg.Services {
		srvcIndx[service.Name] = service.Address
	}

	return &RCPService{
		config:     cfg,
		srvcIndx:   srvcIndx,
		bridgeProv: bridgeProv,
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

	connSrc, connDest, err := s.bridgeProv.CreateConnection(ctx, id, dest)
	if err != nil {
		return fmt.Errorf("failed to create connection: %w", err)
	}

	go func() {
		defer cancel()
		_, _ = io.Copy(connDest, connSrc)
	}()

	go func() {
		defer cancel()
		_, _ = io.Copy(connSrc, connDest)
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
