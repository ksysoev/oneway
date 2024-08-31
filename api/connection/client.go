package connection

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"

	"golang.org/x/net/context"
)

type Client struct {
	addr   string
	dialer net.Dialer
}

func NewClient(addr string) *Client {
	return &Client{
		addr:   addr,
		dialer: net.Dialer{},
	}
}

func (c *Client) Connect(ctx context.Context, id uint64) (net.Conn, error) {
	conn, err := c.dialer.DialContext(ctx, "tcp", c.addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	err = c.initialize(conn, id)
	if err != nil {
		if errC := conn.Close(); errC != nil {
			err = errors.Join(err, errC)
		}

		return nil, fmt.Errorf("failed to initialize connection: %w", err)
	}

	return conn, nil
}

func (c *Client) initialize(conn net.Conn, id uint64) error {
	buf := []byte{byte(V1), byte(NoAuth)}
	_, err := conn.Write(buf)

	if err != nil {
		return fmt.Errorf("failed to write protocol version and authentication method: %w", err)
	}

	buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, id)

	_, err = conn.Write(buf)

	if err != nil {
		return fmt.Errorf("failed to write connection id: %w", err)
	}

	return nil
}
