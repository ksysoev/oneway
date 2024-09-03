package client

import (
	"net/http"
	"time"

	"golang.org/x/net/proxy"
)

const defaultTimeout = 30 * time.Second

// NewHTTPClient creates a new HTTP client with the specified proxy address.
// The proxyAddr parameter should be in the format "host:port".
// It returns a pointer to an http.Client and an error if any.
func NewHTTPClient(proxyAddr string) (*http.Client, error) {
	dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
	if err != nil {
		return nil, err
	}

	ctxDialer, ok := dialer.(proxy.ContextDialer)
	if !ok {
		panic("dialer does not implement proxy.ContextDialer")
	}

	transport := &http.Transport{
		DialContext: ctxDialer.DialContext,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   defaultTimeout,
	}, nil
}
