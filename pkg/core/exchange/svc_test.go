package exchange

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/ksysoev/oneway/pkg/core/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNew(t *testing.T) {
	revProxyRepo := NewMockRevProxyRepo(t)
	connQueue := NewMockConnectionQueue(t)

	service := New(revProxyRepo, connQueue)

	assert.Equal(t, revProxyRepo, service.revProxyRepo)
	assert.Equal(t, connQueue, service.connQueue)
}
func TestAddConnection(t *testing.T) {
	tests := []struct {
		err  error
		name string
	}{
		{
			name: "no error",
			err:  nil,
		},
		{
			name: "error",
			err:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			revProxyRepo := NewMockRevProxyRepo(t)
			connQueue := NewMockConnectionQueue(t)

			service := New(revProxyRepo, connQueue)

			mockConn, _ := net.Pipe()
			defer mockConn.Close()

			connQueue.EXPECT().AddConnection(uint64(123), ConnResult{Conn: mockConn}).Return(tt.err)

			// Add the connection to the connection queue
			err := service.AddConnection(123, mockConn)
			assert.ErrorIs(t, err, tt.err)
		})
	}
}

func TestRegisterRevProxy(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		NameSpace string
	}{
		{
			name:      "no error",
			err:       nil,
			NameSpace: "example",
		},
		{
			name:      "error",
			err:       ErrNameSpaceEmpty,
			NameSpace: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			revProxyRepo := NewMockRevProxyRepo(t)
			connQueue := NewMockConnectionQueue(t)

			service := New(revProxyRepo, connQueue)

			nameSpace := tt.NameSpace
			services := []string{"service1", "service2"}

			if tt.err == nil {
				revProxyRepo.EXPECT().Register(mock.Anything).Return()
			}

			result, err := service.RegisterRevProxy(context.Background(), nameSpace, services)

			assert.ErrorIs(t, err, tt.err)

			if err == nil {
				assert.NotNil(t, result)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}
func TestUnregisterRevProxy(t *testing.T) {
	revProxyRepo := NewMockRevProxyRepo(t)
	connQueue := NewMockConnectionQueue(t)

	service := New(revProxyRepo, connQueue)

	proxy := &RevProxy{} // Create a mock RevProxy

	revProxyRepo.EXPECT().Unregister(proxy)

	service.UnregisterRevProxy(proxy)
}

func NewMockConnQ(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockConnQ {
	m := &MockConnQ{
		ready: make(chan struct{}),
	}
	m.Mock.Test(t)

	t.Cleanup(func() { m.AssertExpectations(t) })

	return m
}

type MockConnQ struct {
	ready   chan struct{}
	connRes chan ConnResult
	mock.Mock
}

func (m *MockConnQ) AddRequest(connChan chan ConnResult) uint64 {
	args := m.Called(connChan)
	m.connRes = connChan

	close(m.ready)

	return args.Get(0).(uint64)
}

func (m *MockConnQ) AddConnection(id uint64, conn ConnResult) error {
	args := m.Called(id, conn)
	return args.Error(0)
}

func TestNewConnection(t *testing.T) {
	addr := &network.Address{
		NameSpace: "example",
		Service:   "service1",
	}

	revProxyRepo := NewMockRevProxyRepo(t)
	connQueue := NewMockConnQ(t)

	proxy, err := NewRevProxy(addr.NameSpace, []string{addr.Service})

	assert.NoError(t, err)

	err = proxy.Start(context.Background())

	assert.NoError(t, err)

	defer proxy.Stop()

	service := New(revProxyRepo, connQueue)

	mockConn, _ := net.Pipe()
	defer mockConn.Close()

	connQueue.On("AddRequest", mock.Anything, mock.Anything).Return(uint64(123))
	revProxyRepo.EXPECT().Find(addr.NameSpace).Return(proxy, nil)

	done := make(chan struct{})
	go func() {
		conn, err := service.NewConnection(context.Background(), addr)

		assert.NoError(t, err)
		assert.NotNil(t, conn)

		close(done)
	}()

	select {
	case <-connQueue.ready:
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected AddRequest to be called")
	}

	select {
	case <-proxy.cmdStream:
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected command to be send a message")
	}

	connRes := ConnResult{
		Conn: mockConn,
		Err:  nil,
	}

	select {
	case connQueue.connRes <- connRes:
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected AddConnection to be called")
	}

	select {
	case <-done:
	case <-time.After(1000 * time.Millisecond):
		t.Error("Expected NewConnection to finish")
	}

	time.Sleep(1000 * time.Millisecond)
}

func TestNewConnection_FailToFindNameSpace(t *testing.T) {
	revProxyRepo := NewMockRevProxyRepo(t)
	connQueue := NewMockConnQ(t)

	addr := &network.Address{
		NameSpace: "example",
		Service:   "service1",
	}

	service := New(revProxyRepo, connQueue)

	connQueue.On("AddRequest", mock.Anything, mock.Anything).Return(uint64(123))
	revProxyRepo.EXPECT().Find(addr.NameSpace).Return(nil, assert.AnError)

	conn, err := service.NewConnection(context.Background(), addr)

	assert.ErrorIs(t, err, assert.AnError)
	assert.Nil(t, conn)
}
