package repo

import (
	"sync"

	"github.com/ksysoev/oneway/pkg/core/exchange"
)

type ConnectionQueue struct {
	l         sync.Mutex
	currentId uint64
	store     map[uint64]chan exchange.ConnResult
}

func NewConnectionQueue() *ConnectionQueue {
	return &ConnectionQueue{
		store: make(map[uint64]chan exchange.ConnResult),
	}
}

func (q *ConnectionQueue) AddRequest(connChan chan exchange.ConnResult) uint64 {
	q.l.Lock()
	defer q.l.Unlock()

	q.currentId++
	q.store[q.currentId] = connChan

	return q.currentId
}

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
