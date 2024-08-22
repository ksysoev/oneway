package main

import (
	"context"
	"net"

	"golang.org/x/build/revdial/v2"
)

// RevProxy will be executed on server side  to publish service to the exchange service

func main() {
	ctx := context.Background()

}

func revConProxyClient(ctx context.Context) error {
	conn, err := net.Dial("tcp", ":8080")
	if err != nil {
		return err
	}

	lis := revdial.NewListener(conn, 
}

func serviceDialer(ctx context.Context) (net.Conn, error) {
	conn, err := net.Dial("tcp", "google.com:80")
	if err != nil {
		return nil, err
	}

	return conn, nil
}
