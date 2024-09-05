package repo

import (
	"sync"

	"github.com/ksysoev/oneway/pkg/core/exchange"
)

type ConnectionQueue struct {
	store     map[uint64]chan exchange.ConnResult
	currentID uint64
	l         sync.Mutex
}

// NewConnectionQueue creates a new instance of ConnectionQueue.
// It initializes the store with an empty map.
// Returns a pointer to the newly created ConnectionQueue.
func NewConnectionQueue() *ConnectionQueue {
	return &ConnectionQueue{
		store: make(map[uint64]chan exchange.ConnResult),
	}
}

// AddRequest adds a connection request to the queue.
// It takes a channel of connection results as an argument.
// Returns the ID of the request.
func (q *ConnectionQueue) AddRequest(connChan chan exchange.ConnResult) uint64 {
	q.l.Lock()
	defer q.l.Unlock()

	q.currentID++
	q.store[q.currentID] = connChan

	return q.currentID
}

// AddConnection adds a connection to the queue.
// It takes an ID and a connection result as arguments.
// If the request with the given ID exists in the queue, the connection result is sent to the request channel.
// Returns an error if the request with the given ID is not found.
func (q *ConnectionQueue) AddConnection(id uint64, conn exchange.ConnResult) error {
	q.l.Lock()
	ch, ok := q.store[id]
	q.l.Unlock()

	if ok {
		ch <- conn
		close(ch)
		delete(q.store, id)

		return nil
	}

	return exchange.ErrConnReqNotFound
}
