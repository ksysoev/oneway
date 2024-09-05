package bridge

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	cfg := &Config{
		Address: "example.com:1234",
	}

	bridge := New(cfg)

	assert.NotNil(t, bridge.apiClient)
	assert.Equal(t, &net.Dialer{}, bridge.dialer)
}

func TestBridge_CreateConnection(t *testing.T) {
	ctx := context.Background()
	expectedProto := "tcp"
	expectedAddr := "example.com:1234"
	expectedID := uint64(1)

	tests := []struct {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiClient := NewMockConnector(t)
			dialer := NewMockContextDialer(t)

			bridgeProv := &Bridge{
				apiClient: apiClient,
				dialer:    dialer,
			}

			srcConn, destConn := net.Pipe()

			apiClient.EXPECT().Connect(ctx, expectedID).Return(srcConn, tt.srcErr)

			if tt.srcErr == nil {
				dialer.EXPECT().DialContext(ctx, expectedProto, expectedAddr).Return(destConn, tt.destErr)
			}

			bridge, err := bridgeProv.CreateConnection(context.Background(), uint64(1), expectedAddr)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, bridge)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, bridge)
			}

			srcConn.Close()
			destConn.Close()
		})
	}
}
