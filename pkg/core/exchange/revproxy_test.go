package exchange

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRevProxy(t *testing.T) {
	// Test case 1: Valid name space and services
	nameSpace := "example"
	services := []string{"service1", "service2"}

	revProxy, err := NewRevProxy(nameSpace, services)

	assert.NoError(t, err)
	assert.NotNil(t, revProxy)
	assert.Equal(t, nameSpace, revProxy.NameSpace)
	assert.Equal(t, services, revProxy.Services)

	// Test case 2: Empty name space
	nameSpace = ""
	services = []string{"service1", "service2"}

	_, err = NewRevProxy(nameSpace, services)

	assert.ErrorIs(t, err, ErrNameSpaceEmpty)

	// Test case 3: Empty services list
	nameSpace = "example"
	services = []string{}

	_, err = NewRevProxy(nameSpace, services)

	assert.ErrorIs(t, err, ErrServicesEmpty)

	// Test case 4: Duplicate service names
	nameSpace = "example"
	services = []string{"service1", "service1"}

	_, err = NewRevProxy(nameSpace, services)

	assert.ErrorIs(t, err, ErrDuplicateService)

	// Test case 5: Empty service name
	nameSpace = "example"
	services = []string{"service1", ""}

	_, err = NewRevProxy(nameSpace, services)

	assert.ErrorIs(t, err, ErrServiceNameEmpty)
}
func TestRevProxy_CommandStream(t *testing.T) {
	// Create a new RevProxy
	nameSpace := "example"
	services := []string{"service1", "service2"}
	revProxy, err := NewRevProxy(nameSpace, services)

	assert.NoError(t, err)

	// Get the command stream
	cmdStream := revProxy.CommandStream()

	assert.NotNil(t, cmdStream)
}
func TestRevProxy_Start(t *testing.T) {
	// Test case 1: Start RevProxy successfully
	nameSpace := "example"
	services := []string{"service1", "service2"}
	revProxy, err := NewRevProxy(nameSpace, services)

	assert.NoError(t, err)

	ctx := context.Background()
	err = revProxy.Start(ctx)

	assert.NoError(t, err)

	// Test case 2: Start RevProxy when it is already running
	err = revProxy.Start(ctx)

	assert.ErrorIs(t, err, ErrRevProxyStarted)
}

func TestRevProxy_Stop(t *testing.T) {
	// Test case 1: Stop RevProxy successfully
	nameSpace := "example"
	services := []string{"service1", "service2"}
	revProxy, err := NewRevProxy(nameSpace, services)

	assert.NoError(t, err)

	ctx := context.Background()
	err = revProxy.Start(ctx)

	assert.NoError(t, err)

	revProxy.Stop()

	assert.Nil(t, revProxy.ctx)

	select {
	case _, ok := <-revProxy.cmdStream:
		assert.False(t, ok)
	default:
		t.Error("Expected command stream to be closed after stopping RevProxy")
	}
}

func TestRevProxy_RequestConnection_Success(t *testing.T) {
	nameSpace := "example"
	services := []string{"service1", "service2"}
	revProxy, err := NewRevProxy(nameSpace, services)
	assert.NoError(t, err)

	ctx := context.Background()
	err = revProxy.Start(ctx)
	assert.NoError(t, err)

	connID := uint64(1)
	serviceName := "service1"

	done := make(chan struct{})
	go func() {
		err = revProxy.RequestConnection(ctx, connID, serviceName)
		assert.NoError(t, err)
		close(done)
	}()

	select {
	case _, ok := <-revProxy.CommandStream():
		assert.True(t, ok)
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected command to be sent to RevProxy")
	}

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected connection request to be successful")
	}
}
func TestRevProxy_RequestConnection_CancelContext(t *testing.T) {
	nameSpace := "example"
	services := []string{"service1", "service2"}
	revProxy, err := NewRevProxy(nameSpace, services)
	assert.NoError(t, err)

	ctx := context.Background()
	err = revProxy.Start(ctx)
	assert.NoError(t, err)

	connID := uint64(1)
	serviceName := "service1"

	cancelCtx, cancel := context.WithCancel(ctx)

	cancel()

	err = revProxy.RequestConnection(cancelCtx, connID, serviceName)
	assert.Equal(t, context.Canceled, err)
}

func TestRevProxy_RequestConnection_ClosedRevProxy(t *testing.T) {
	nameSpace := "example"
	services := []string{"service1", "service2"}
	revProxy, err := NewRevProxy(nameSpace, services)
	assert.NoError(t, err)

	ctx := context.Background()
	err = revProxy.Start(ctx)
	assert.NoError(t, err)

	connID := uint64(1)
	serviceName := "service1"

	revProxy.Stop()
	err = revProxy.RequestConnection(ctx, connID, serviceName)
	assert.Equal(t, ErrRevProxyStopped, err)
}
