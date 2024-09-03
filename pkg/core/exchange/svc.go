package exchange

import (
	"context"
	"fmt"
	"net"

	"github.com/ksysoev/oneway/pkg/core/network"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var meter = otel.GetMeterProvider().Meter("oneway")
var tracer = otel.Tracer("github.com/ksysoev/oneway/pkg/core/exchange")

var ErrConnReqNotFound = fmt.Errorf("connection request not found")

type Service struct {
	revProxyRepo RevProxyRepo
	connQueue    ConnectionQueue
}

type ConnResult struct {
	Conn net.Conn
	Err  error
}

type RevProxyRepo interface {
	Register(proxy *RevProxy)
	Find(nameSpace string) (*RevProxy, error)
	Unregister(proxy *RevProxy)
}

type ConnectionQueue interface {
	AddRequest(connChan chan ConnResult) uint64
	AddConnection(id uint64, conn ConnResult) error
}

func New(revProxyRepo RevProxyRepo, connQueue ConnectionQueue) *Service {
	return &Service{
		revProxyRepo: revProxyRepo,
		connQueue:    connQueue,
	}
}

func (s *Service) NewConnection(ctx context.Context, addr *network.Address) (net.Conn, error) {
	ctx, span := tracer.Start(ctx, "Exchange.NewConnection")
	defer span.End()

	counter, _ := meter.Int64Counter("connection")
	counter.Add(ctx, 1, metric.WithAttributes(attribute.String("address", addr.String())))

	connChan := make(chan ConnResult, 1)

	id := s.connQueue.AddRequest(connChan)

	span.AddEvent("Request added")

	proxy, err := s.revProxyRepo.Find(addr.NameSpace)
	if err != nil {
		return nil, fmt.Errorf("failed to get reverse connection proxy: %w", err)
	}

	if err = proxy.RequestConnection(ctx, id, addr.Service); err != nil {
		return nil, fmt.Errorf("failed to request connection: %w", err)
	}

	span.AddEvent("Request sent")

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res, ok := <-connChan:
		if !ok {
			return nil, fmt.Errorf("failed to get connection")
		}

		return res.Conn, res.Err
	}
}

func (s *Service) RegisterRevProxy(_ context.Context, nameSpace string, services []string) (*RevProxy, error) {
	proxy, err := NewRevProxy(nameSpace, services)
	if err != nil {
		return nil, fmt.Errorf("failed to create reverse connection proxy: %w", err)
	}

	s.revProxyRepo.Register(proxy)

	return proxy, nil
}

func (s *Service) UnregisterRevProxy(proxy *RevProxy) {
	s.revProxyRepo.Unregister(proxy)
}

func (s *Service) AddConnection(id uint64, conn net.Conn) error {
	return s.connQueue.AddConnection(id, ConnResult{
		Conn: conn,
	})
}
