package socks5

import (
	"fmt"
	"net"
)

type Client struct {
	conn net.Conn
}

func NewClient(conn net.Conn) *Client {
	return &Client{
		conn: conn,
	}
}
func (c *Client) Connect(net.Addr) (net.Conn, error) {
	return nil, fmt.Errorf("Associate not implemented")
}

func (c *Client) Bind(net.Addr) (net.Conn, error) {
	return nil, fmt.Errorf("Associate not implemented")
}

func (c *Client) Associate(net.Addr) (net.Conn, error) {
	return nil, fmt.Errorf("Associate not implemented")
}
