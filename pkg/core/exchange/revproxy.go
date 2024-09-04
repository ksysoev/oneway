package exchange

import (
	"context"
	"fmt"
	"sync"
)

var (
	ErrNameSpaceEmpty   = fmt.Errorf("name space is empty")
	ErrServicesEmpty    = fmt.Errorf("services list is empty")
	ErrDuplicateService = fmt.Errorf("duplicate service name")
	ErrRevProxyStopped  = fmt.Errorf("revproxy is stopped")
	ErrServiceNameEmpty = fmt.Errorf("service name is empty")
	ErrRevProxyStarted  = fmt.Errorf("revproxy is already started")
)

type RevProxy struct {
	ctx       context.Context
	cancel    context.CancelFunc
	cmdStream chan RevProxyCommand
	NameSpace string
	Services  []string
	mu        sync.RWMutex
	wg        sync.WaitGroup
}

type RevProxyCommand struct {
	NameSpace string
	Name      string
	ConnID    uint64
}

// NewRevProxy creates a new RevProxy with the specified name space and services.
// It returns an error if the name space is empty or if the services list is empty.
// It also returns an error if the services list contains duplicate service names.
// The RevProxy is created with an empty command stream.
func NewRevProxy(nameSpace string, services []string) (*RevProxy, error) {
	if nameSpace == "" {
		return nil, ErrNameSpaceEmpty
	}

	if len(services) == 0 {
		return nil, ErrServicesEmpty
	}

	uniqIndex := make(map[string]struct{})

	for _, service := range services {
		if service == "" {
			return nil, ErrServiceNameEmpty
		}

		if _, ok := uniqIndex[service]; ok {
			return nil, ErrDuplicateService
		}

		uniqIndex[service] = struct{}{}
	}

	return &RevProxy{
		NameSpace: nameSpace,
		Services:  services,
		cmdStream: make(chan RevProxyCommand),
	}, nil
}

// Start starts the RevProxy and returns an error if the RevProxy is already running.
// The RevProxy is started with the specified context.
func (r *RevProxy) Start(ctx context.Context) error {
	r.mu.Lock()
	if r.ctx != nil {
		r.mu.Unlock()
		return ErrRevProxyStarted
	}

	r.ctx, r.cancel = context.WithCancel(ctx)
	ctx = r.ctx

	r.wg.Add(1)
	r.mu.Unlock()

	go func() {
		defer r.wg.Done()

		<-ctx.Done()
		close(r.cmdStream)
	}()

	return nil
}

// Stop stops the RevProxy.
// It closes the command stream and sets the context to nil.
func (r *RevProxy) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cancel()

	r.ctx = nil
	r.wg.Wait()
}

// CommandStream returns a read-only channel of RevProxyCommand.
// This channel is used to receive commands for the RevProxy.
// Any commands sent through this channel will be processed by the RevProxy.
// The returned channel is read-only, meaning it can only be used for receiving commands.
func (r *RevProxy) CommandStream() <-chan RevProxyCommand {
	return r.cmdStream
}

// RequestConnection sends a request to establish a connection with the specified ID and name.
// It returns an error if the context is canceled or if the command cannot be sent to the command stream.
func (r *RevProxy) RequestConnection(ctx context.Context, id uint64, name string) error {
	r.mu.RLock()
	proxyCtx := r.ctx
	r.mu.RUnlock()

	if proxyCtx == nil {
		return ErrRevProxyStopped
	}

	cmd := RevProxyCommand{
		NameSpace: r.NameSpace,
		Name:      name,
		ConnID:    id,
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-proxyCtx.Done():
		return ErrRevProxyStopped
	case r.cmdStream <- cmd:
		return nil
	}
}
