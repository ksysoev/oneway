package bridge

import (
	"context"
	"fmt"
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

type ContextDialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type Bridge struct {
	apiClient Connector
	dialer    ContextDialer
}

// New creates a new Bridge instance
// with the provided configuration
func New(cfg *Config) *Bridge {
	apiClient := revconn.NewClient(cfg.Address)

	return &Bridge{
		apiClient: apiClient,
		dialer:    &net.Dialer{},
	}
}

// CreateConnection creates a new network bridge connection.
// It takes a context, connection ID, and address as parameters.
// It returns a pointer to a network.Bridge and an error.
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

// createBackConnection creates a connection to the back-end service
// using the provided connection ID.
// It takes a context and connection ID as parameters.
// It returns an io.ReadWriteCloser and an error.
func (r *Bridge) createBackConnection(ctx context.Context, id uint64) (net.Conn, error) {
	conn, err := r.apiClient.Connect(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to connect with for id %d: %w", id, err)
	}

	return conn, nil
}

// createDestConnection creates a connection to the destination service
// using the provided address.
// It takes a context and address as parameters.
// It returns an io.ReadWriteCloser and an error.
func (r *Bridge) createDestConnection(ctx context.Context, dest string) (net.Conn, error) {
	connDest, err := r.dialer.DialContext(ctx, "tcp", dest)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", dest, err)
	}

	return connDest, nil
}
