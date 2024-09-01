package revconproxy

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ksysoev/oneway/pkg/core/network"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var meter = otel.GetMeterProvider().Meter("oneway")

type BridgeProvider interface {
	CreateConnection(ctx context.Context, id uint64, addr string) (*network.Bridge, error)
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

// TODO Do i need namespace here as argument?
func (s *RCPService) CreateConnection(ctx context.Context, _, serviceName string, id uint64) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	dest, ok := s.srvcIndx[serviceName]
	if !ok {
		return fmt.Errorf("service not found")
	}

	bridge, err := s.bridgeProv.CreateConnection(ctx, id, dest)
	if err != nil {
		return fmt.Errorf("failed to create bridge: %w", err)
	}

	stats, err := bridge.Run(ctx)

	counter, _ := meter.Int64Counter("transmitted_bytes")
	counter.Add(ctx, stats.Sent, metric.WithAttributes(attribute.String("service", serviceName), attribute.String("direction", "sent")))
	counter.Add(ctx, stats.Recv, metric.WithAttributes(attribute.String("service", serviceName), attribute.String("direction", "received")))
	timing, _ := meter.Float64Histogram("connection_duration", metric.WithDescription("Connection duration in milliseconds"), metric.WithUnit("s"))
	timing.Record(ctx, stats.Duration.Seconds(), metric.WithAttributes(attribute.String("service", serviceName)))

	if err != nil {
		slog.Error("failed to run bridge", slog.Any("error", err))
	}

	return err
}

func (s *RCPService) ServiceNames() []string {
	serviceNames := make([]string, 0, len(s.config.Services))
	for _, service := range s.config.Services {
		serviceNames = append(serviceNames, service.Name)
	}

	return serviceNames
}
