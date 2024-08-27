package repo

import (
	"sync"

	"github.com/ksysoev/oneway/pkg/core/exchange"
)

type RevProxyRegistry struct {
	rwl   sync.RWMutex
	store map[string]*exchange.RevConProxy
}

func NewRevProxyRegistry() *RevProxyRegistry {
	return &RevProxyRegistry{
		store: make(map[string]*exchange.RevConProxy),
	}
}

func (r *RevProxyRegistry) RegisterRevConProxy(proxy *exchange.RevConProxy) {
	r.rwl.Lock()
	defer r.rwl.Unlock()

	r.store[proxy.NameSpace] = proxy
}

func (r *RevProxyRegistry) GetRevConProxy(nameSpace string) *exchange.RevConProxy {
	r.rwl.RLock()
	defer r.rwl.RUnlock()

	return r.store[nameSpace]
}
