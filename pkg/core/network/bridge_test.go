package network

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewBridge(t *testing.T) {
	src := NewMockConn(t)
	dest := NewMockConn(t)

	bridge := NewBridge(src, dest)

	if bridge.src != src {
		t.Errorf("Expected source to be %v, but got %v", src, bridge.src)
	}

	if bridge.dest != dest {
		t.Errorf("Expected destination to be %v, but got %v", dest, bridge.dest)
	}
}

func TestBridge_Close(t *testing.T) {
	table := []struct {
		srcErr      error
		destErr     error
		expectedErr error
		name        string
	}{
		{
			name:        "no errors",
			srcErr:      nil,
			destErr:     nil,
			expectedErr: nil,
		},
		{
			name:        "source error",
			srcErr:      assert.AnError,
			destErr:     nil,
			expectedErr: assert.AnError,
		},
		{
			name:        "destination error",
			srcErr:      nil,
			destErr:     assert.AnError,
			expectedErr: assert.AnError,
		},
		{
			name:        "both errors",
			srcErr:      assert.AnError,
			destErr:     assert.AnError,
			expectedErr: assert.AnError,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			src := NewMockConn(t)
			dest := NewMockConn(t)

			bridge := NewBridge(src, dest)

			src.EXPECT().Close().Return(tt.srcErr)
			dest.EXPECT().Close().Return(tt.destErr)

			err := bridge.Close()

			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}
func TestStartCopy(t *testing.T) {
	src := strings.NewReader("Hello, World!")
	dest := &bytes.Buffer{}
	sent := int64(0)
	errCh := make(chan error)

	startCopy(src, dest, &sent, errCh)

	err := <-errCh
	assert.NoError(t, err)
	assert.Equal(t, int64(13), sent)
	assert.Equal(t, "Hello, World!", dest.String())
}

func TestStartCopy_Error(t *testing.T) {
	src := NewMockConn(t)
	dest := NewMockConn(t)
	sent := int64(0)
	errCh := make(chan error)

	src.EXPECT().Read(mock.Anything).Return(0, assert.AnError)

	startCopy(src, dest, &sent, errCh)

	err := <-errCh
	assert.ErrorIs(t, err, assert.AnError)
}

func TestBridge_Run(t *testing.T) {
	src := NewMockConn(t)
	dest := NewMockConn(t)

	src.EXPECT().Read(mock.Anything).Return(13, io.EOF)
	src.EXPECT().Write(mock.Anything).Return(13, nil)
	src.EXPECT().Close().Return(nil)
	dest.EXPECT().Read(mock.Anything).Return(13, io.EOF)
	dest.EXPECT().Write(mock.Anything).Return(13, nil)
	dest.EXPECT().Close().Return(nil)

	bridge := NewBridge(src, dest)

	done := make(chan struct{})

	go func() {
		stats, err := bridge.Run(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, int64(13), stats.Sent)
		assert.Equal(t, int64(13), stats.Recv)
		assert.GreaterOrEqual(t, stats.Duration.Milliseconds(), int64(0))

		close(done)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected Run to finish in 1 second")
	}
}

func TestBridge_Run_Error(t *testing.T) {
	src := NewMockConn(t)
	dest := NewMockConn(t)

	src.EXPECT().Read(mock.Anything).Return(0, assert.AnError)
	src.EXPECT().Close().Return(nil)
	dest.EXPECT().Read(mock.Anything).Return(0, io.EOF)
	dest.EXPECT().Close().Return(nil)

	bridge := NewBridge(src, dest)

	done := make(chan struct{})

	go func() {
		stats, err := bridge.Run(context.Background())

		assert.ErrorIs(t, err, assert.AnError)
		assert.Equal(t, int64(0), stats.Sent)
		assert.Equal(t, int64(0), stats.Recv)
		assert.GreaterOrEqual(t, stats.Duration.Milliseconds(), int64(0))

		close(done)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected Run to finish in 1 second")
	}
}
