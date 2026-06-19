// Package broker is an in-memory per-user pub/sub used to push notifications to a
// user's connected browser tabs in real time over SSE. Single-replica only (state
// is in-process), matching the dispatch job-board broker.
package broker

import "sync"

type Broker struct {
	mu   sync.RWMutex
	subs map[string]map[chan []byte]struct{} // userID -> set of subscriber channels
}

func New() *Broker {
	return &Broker{subs: make(map[string]map[chan []byte]struct{})}
}

// Subscribe registers a new subscriber channel for a user and returns it.
func (b *Broker) Subscribe(userID string) chan []byte {
	ch := make(chan []byte, 8)
	b.mu.Lock()
	if b.subs[userID] == nil {
		b.subs[userID] = make(map[chan []byte]struct{})
	}
	b.subs[userID][ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

// Unsubscribe removes and closes a subscriber channel.
func (b *Broker) Unsubscribe(userID string, ch chan []byte) {
	b.mu.Lock()
	if set, ok := b.subs[userID]; ok {
		if _, ok := set[ch]; ok {
			delete(set, ch)
			close(ch)
		}
		if len(set) == 0 {
			delete(b.subs, userID)
		}
	}
	b.mu.Unlock()
}

// Publish delivers a message to all of a user's subscribers (non-blocking).
func (b *Broker) Publish(userID string, msg []byte) {
	if userID == "" {
		return
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.subs[userID] {
		select {
		case ch <- msg:
		default: // drop if the subscriber is slow; the REST list is the source of truth
		}
	}
}
