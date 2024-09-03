package repo

import (
	"testing"

	"github.com/ksysoev/oneway/pkg/core/exchange"
)

func TestRevProxyRegistry(t *testing.T) {
	// Create a new instance of RevProxyRegistry
	registry := NewRevProxyRegistry()

	// Create a mock reverse proxy
	proxy := &exchange.RevProxy{
		NameSpace: "example",
	}

	// Register the mock reverse proxy
	registry.Register(proxy)

	// Test case: Find an existing reverse proxy
	foundProxy, err := registry.Find("example")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if foundProxy != proxy {
		t.Errorf("Expected foundProxy to be equal to proxy")
	}

	// Test case: Find a non-existing reverse proxy
	nonExistingProxy, err := registry.Find("non-existing")
	if err == nil {
		t.Errorf("Expected an error, got nil")
	}
	if nonExistingProxy != nil {
		t.Errorf("Expected nonExistingProxy to be nil")
	}

	// Unregister the mock reverse proxy
	registry.Unregister(proxy)

	// Test case: Find an unregistered reverse proxy
	unregisteredProxy, err := registry.Find("example")
	if err == nil {
		t.Errorf("Expected an error, got nil")
	}
	if unregisteredProxy != nil {
		t.Errorf("Expected unregisteredProxy to be nil")
	}
}
