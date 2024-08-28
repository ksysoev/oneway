package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/net/proxy"
)

func main() {
	// SOCKS5 proxy address
	proxyAddr := "127.0.0.1:1080"

	// Create a SOCKS5 dialer
	dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
	if err != nil {
		fmt.Printf("Failed to create SOCKS5 dialer: %v\n", err)
		return
	}

	// Create an HTTP transport with the SOCKS5 dialer
	transport := &http.Transport{
		Dial: dialer.Dial,
	}

	// Create an HTTP client with the transport
	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	// Make the HTTP request
	resp, err := client.Get("http://restapi.example/health")
	if err != nil {
		fmt.Printf("Failed to make HTTP request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Read and print the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response body: %v\n", err)
		return
	}

	fmt.Printf("Response: %s\n", body)
}
