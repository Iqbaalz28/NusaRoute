// Package broker is a tiny in-memory pub/sub used to push new OPEN jobs to
// connected couriers over Server-Sent Events. Single-instance only — fine for
// this deployment (dispatch-service runs one replica). A multi-replica setup
// would back this with Redis pub/sub instead.
package broker

import "sync"

type Broker struct {
	mu   sync.RWMutex
	subs map[chan []byte]struct{}
}

func New() *Broker {
	return &Broker{subs: make(map[chan []byte]struct{})}
}

// Subscribe returns a new buffered channel that receives broadcast messages.
func (b *Broker) Subscribe() chan []byte {
	ch := make(chan []byte, 8)
	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *Broker) Unsubscribe(ch chan []byte) {
	b.mu.Lock()
	if _, ok := b.subs[ch]; ok {
		delete(b.subs, ch)
		close(ch)
	}
	b.mu.Unlock()
}

// Broadcast sends msg to every subscriber, dropping it for any slow consumer
// rather than blocking the publisher.
func (b *Broker) Broadcast(msg []byte) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.subs {
		select {
		case ch <- msg:
		default: // subscriber buffer full — skip rather than block
		}
	}
}
