package repo

import (
	"fmt"
	"sync"

	"github.com/ksysoev/oneway/pkg/core/exchange"
)

type RevProxyRegistry struct {
	store map[string]*exchange.RevProxy
	mu    sync.RWMutex
}

// NewRevProxyRegistry creates a new instance of RevProxyRegistry.
// It initializes the store with an empty map.
// Returns a pointer to the newly created RevProxyRegistry.
func NewRevProxyRegistry() *RevProxyRegistry {
	return &RevProxyRegistry{
		store: make(map[string]*exchange.RevProxy),
	}
}

// AddRevProxy adds a RevProxy to the RevProxyRegistry.
// It takes a pointer to a RevProxy as the parameter.
// The RevProxy is added to the registry's store using the proxy's NameSpace as the key.
func (r *RevProxyRegistry) Register(proxy *exchange.RevProxy) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.store[proxy.NameSpace] = proxy
}

// Unregister removes the specified proxy from the RevProxyRegistry.
// It takes a pointer to a RevProxy as the argument.
// The proxy is removed from the registry by deleting its namespace from the store.
func (r *RevProxyRegistry) Unregister(proxy *exchange.RevProxy) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.store, proxy.NameSpace)
}

// Find searches for a reverse proxy in the registry based on the given namespace.
// It returns a pointer to the found reverse proxy and an error if the reverse proxy is not found.
func (r *RevProxyRegistry) Find(nameSpace string) (*exchange.RevProxy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	proxy, ok := r.store[nameSpace]

	if !ok {
		return nil, fmt.Errorf("revproxy not found")
	}

	return proxy, nil
}
