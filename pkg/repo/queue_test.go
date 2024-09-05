package repo

import (
	"testing"
	"time"

	"github.com/ksysoev/oneway/pkg/core/exchange"
	"github.com/stretchr/testify/assert"
)

func TestNewConnectionQueue(t *testing.T) {
	q := NewConnectionQueue()

	assert.NotNil(t, q)
	assert.NotNil(t, q.store)
	assert.Equal(t, 0, len(q.store))
}

func TestAddRequest(t *testing.T) {
	q := NewConnectionQueue()

	connChan := make(chan exchange.ConnResult)
	id := q.AddRequest(connChan)

	assert.NotEqual(t, 0, id)
	assert.Equal(t, 1, len(q.store))
	assert.Equal(t, connChan, q.store[id])
}

func TestAddConnection(t *testing.T) {
	q := NewConnectionQueue()

	connChan := make(chan exchange.ConnResult)
	id := q.AddRequest(connChan)

	conn := exchange.ConnResult{
		Conn: nil,
		Err:  nil,
	}

	go func() {
		err := q.AddConnection(id, conn)

		assert.NoError(t, err)
		assert.Equal(t, 0, len(q.store))
	}()

	select {
	case result := <-connChan:
		assert.Equal(t, conn, result)
	case <-time.After(1 * time.Second):
		t.Error("AddConnection should send the connection result to the request channel")
	}
}
