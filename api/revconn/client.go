package revconn

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

	if _, err := conn.Write(buf); err != nil {
		return fmt.Errorf("failed to write protocol version and authentication method: %w", err)
	}

	if _, err := conn.Read(buf); err != nil {
		return fmt.Errorf("failed to read protocol version and authentication method: %w", err)
	}

	if ver := Version(buf[0]); ver != V1 {
		return fmt.Errorf("unsupported protocol version")
	}

	if authMethod := AuthMethod(buf[1]); authMethod != NoAuth {
		return fmt.Errorf("unsupported authentication method")
	}

	buf = make([]byte, connectionIDLenght+1)
	buf[0] = byte(V1)
	binary.BigEndian.PutUint64(buf[1:], id)

	if _, err := conn.Write(buf); err != nil {
		return fmt.Errorf("failed to write connection id: %w", err)
	}

	return nil
}
