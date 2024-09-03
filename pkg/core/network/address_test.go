package network

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAddress(t *testing.T) {
	tests := []struct {
		addr        string
		expected    *Address
		err         error
		expectedStr string
	}{
		{
			addr:        "service.namespace:0",
			expected:    NewAddress("service", "namespace"),
			err:         nil,
			expectedStr: "service.namespace",
		},
		{
			addr:     "invalidaddress",
			expected: nil,
			err:      ErrInvalidAddress,
		},
		{
			addr:     "invalidaddress:2",
			expected: nil,
			err:      ErrInvalidAddress,
		},
	}

	for _, test := range tests {
		actual, err := ParseAddress(test.addr)

		assert.Equal(t, test.err, err)
		assert.Equal(t, test.expected, actual)

		if test.expected != nil {
			assert.Equal(t, test.expected.String(), test.expectedStr)
		}
	}
}
