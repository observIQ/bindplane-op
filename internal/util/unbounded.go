package util

import (
	"sync"
	"time"
)

// UnboundedChan is a channel with an unbounded buffer
type UnboundedChan[T any] interface {
	In() chan<- T
	Out() <-chan T
	Close()
}

// unboundedChan is the standard implementation of UnboundedChan
type unboundedChan[T any] struct {
	in       chan T
	out      chan T
	buffer   []T
	mux      sync.Mutex
	interval time.Duration
	closed   bool
}

// In returns the inbound channel
func (u *unboundedChan[T]) In() chan<- T {
	return u.in
}

// Out returns the outbound channel
func (u *unboundedChan[T]) Out() <-chan T {
	return u.out
}

// Close closes the unbounded channel
func (u *unboundedChan[T]) Close() {
	close(u.in)
}

// receive will receive items from the inbound channel
func (u *unboundedChan[T]) receive() {
	for {
		item, ok := <-u.in
		if !ok {
			u.mux.Lock()
			defer u.mux.Unlock()
			u.closed = true
			return
		}

		u.push(item)
	}
}

// send will send items to the outbound channel
func (u *unboundedChan[T]) send() {
	for {
		if u.done() {
			close(u.out)
			return
		}

		if u.empty() {
			time.Sleep(u.interval)
			continue
		}

		item := u.pop()
		u.out <- item
	}
}

// done returns true if the channel is closed and the buffer is drained
func (u *unboundedChan[T]) done() bool {
	u.mux.Lock()
	defer u.mux.Unlock()
	return u.closed && len(u.buffer) == 0
}

// empty returns true if the buffer is empty
func (u *unboundedChan[T]) empty() bool {
	u.mux.Lock()
	defer u.mux.Unlock()
	return len(u.buffer) == 0
}

// pop returns the first item in the buffer
func (u *unboundedChan[T]) pop() T {
	u.mux.Lock()
	defer u.mux.Unlock()
	item := u.buffer[0]
	u.buffer = u.buffer[1:]
	return item
}

// push pushes an item into the buffer
func (u *unboundedChan[T]) push(item T) {
	u.mux.Lock()
	defer u.mux.Unlock()
	u.buffer = append(u.buffer, item)
}

// NewUnboundedChan creates a new unbounded channel
func NewUnboundedChan[T any](interval time.Duration) UnboundedChan[T] {
	unbounded := &unboundedChan[T]{
		in:       make(chan T),
		out:      make(chan T),
		interval: interval,
	}
	go unbounded.send()
	go unbounded.receive()
	return unbounded
}
