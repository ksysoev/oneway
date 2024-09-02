package bridge

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/ksysoev/oneway/api/revconn"
	"github.com/ksysoev/oneway/pkg/core/network"
)

type Config struct {
	Address string
}

type Connector interface {
	Connect(ctx context.Context, id uint64) (net.Conn, error)
}

type Bridge struct {
	apiClient Connector
	dialer    net.Dialer
}

func New(cfg *Config) *Bridge {
	apiClient := revconn.NewClient(cfg.Address)

	return &Bridge{
		apiClient: apiClient,
		dialer:    net.Dialer{},
	}
}

func (r *Bridge) CreateConnection(ctx context.Context, id uint64, addr string) (*network.Bridge, error) {
	src, err := r.createBackConnection(ctx, id)
	if err != nil {
		return nil, err
	}

	dest, err := r.createDestConnection(ctx, addr)
	if err != nil {
		src.Close()
		return nil, err
	}

	return network.NewBridge(src, dest), nil
}

func (r *Bridge) createBackConnection(ctx context.Context, id uint64) (io.ReadWriteCloser, error) {
	conn, err := r.apiClient.Connect(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to connect with for id %d: %w", id, err)
	}

	return conn, nil
}

func (r *Bridge) createDestConnection(ctx context.Context, dest string) (io.ReadWriteCloser, error) {
	connDest, err := r.dialer.DialContext(ctx, "tcp", dest)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", dest, err)
	}

	return connDest, nil
}
