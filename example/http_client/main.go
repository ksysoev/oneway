package main

import (
	"fmt"
	"io"

	"github.com/ksysoev/oneway/api/client"
)

func main() {
	// Create an HTTP client with the transport
	cl, err := client.NewHTTPClient("127.0.0.1:1080")
	if err != nil {
		fmt.Printf("Failed to create HTTP client: %v\n", err)
		return
	}

	// Make the HTTP request
	resp, err := cl.Get("http://restapi.example/health")
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
