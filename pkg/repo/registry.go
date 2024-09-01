package repo

import (
	"fmt"
	"sync"

	"github.com/ksysoev/oneway/pkg/core/exchange"
)

type RevProxyRegistry struct {
	store map[string]*exchange.RevConProxy
	rwl   sync.RWMutex
}

func NewRevProxyRegistry() *RevProxyRegistry {
	return &RevProxyRegistry{
		store: make(map[string]*exchange.RevConProxy),
	}
}

func (r *RevProxyRegistry) AddRevConProxy(proxy *exchange.RevConProxy) {
	r.rwl.Lock()
	defer r.rwl.Unlock()

	r.store[proxy.NameSpace] = proxy
}

func (r *RevProxyRegistry) GetRevConProxy(nameSpace string) (*exchange.RevConProxy, error) {
	r.rwl.RLock()
	defer r.rwl.RUnlock()

	proxy, ok := r.store[nameSpace]

	if !ok {
		return nil, fmt.Errorf("rev con proxy not found")
	}

	return proxy, nil
}
