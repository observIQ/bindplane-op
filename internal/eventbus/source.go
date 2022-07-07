// Copyright  observIQ, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package eventbus

import (
	"context"
	"sync"
)

const subscriberChannelBufferSize = 10

// UnsubscribeFunc is a function that allows a subscriber to unsubscribe
type UnsubscribeFunc func()

// Subscriber can be notified of events of type T. Instead of using this interface directly, use one of the eventbus.Subscribe
// functions to receive a channel of events.
type Subscriber[T any] interface {
	// Receive will be called when an event is available
	Receive(event T)

	// Close will be called when the subscriber is unsubscribed
	Close()
}

// Source is a source of events.
type Source[T any] interface {
	// Send the event to all of the subscribers
	Send(event T)

	// SubscribeUntilDone adds a subscriber to the source and automatically unsubscribes when the context is done. If the
	// context is nil, the unsubscribe function must be called to unsubscribe. Instead of using this method to subscribe
	// to a source, use one of the eventbus.Subscribe functions to receive a channel of events.
	SubscribeUntilDone(context.Context, Subscriber[T]) UnsubscribeFunc

	// Subscribers returns the current number of subscribers
	Subscribers() int
}

// SubscriptionFilter can filter on events and map from an event to another type. It can also ignore events. If accept
// is false, the result is ignored and not sent to subscribers.
type SubscriptionFilter[T, R any] func(event T) (result R, accept bool)

var exists = struct{}{}

// implementation of EventBus
type source[T any] struct {
	// subscribers is the set of current subscribers, implemented as a map to an empty struct
	subscribers map[Subscriber[T]]struct{}
	mtx         sync.RWMutex
}

var _ Source[any] = (*source[any])(nil)

type subscription[T any] struct {
	channel chan T
}

func newSubscription[T any](options []SubscriptionOption[T]) subscription[T] {
	opts := makeSubscriptionOptions(options)
	channel := opts.channel
	if channel == nil {
		channel = make(chan T, subscriberChannelBufferSize)
	}
	return subscription[T]{
		channel: channel,
	}
}

func (s *subscription[T]) Receive(event T) {
	s.channel <- event
}

func (s *subscription[T]) Close() {
	close(s.channel)
}

var _ Subscriber[int] = (*subscription[int])(nil)

type filterSubscription[T, R any] struct {
	subscription[R]
	filter SubscriptionFilter[T, R]
}

func newFilterSubscription[T, R any](filter SubscriptionFilter[T, R], options []SubscriptionOption[R]) filterSubscription[T, R] {
	return filterSubscription[T, R]{
		subscription: newSubscription(options),
		filter:       filter,
	}
}

func (s *filterSubscription[T, R]) Receive(event T) {
	filtered, accept := s.filter(event)
	if accept {
		s.channel <- filtered
	}
}

func (s *filterSubscription[T, R]) Close() {
	close(s.channel)
}

var _ Subscriber[int] = (*filterSubscription[int, int])(nil)

// NewSource returns a new Source implementation for the specified event type T
func NewSource[T any]() Source[T] {
	return &source[T]{
		subscribers: make(map[Subscriber[T]]struct{}),
	}
}

// Subscribe subscribes to events on the bus and returns a channel to receive events and an unsubscribe function.
func Subscribe[T any](bus Source[T], options ...SubscriptionOption[T]) (<-chan T, UnsubscribeFunc) {
	return SubscribeUntilDone(nil, bus, options...)
}

// SubscribeUntilDone subscribes to events on the bus and returns a channel to receive events and an unsubscribe
// function. It automatically unsubscribes when the context is done.
func SubscribeUntilDone[T any](ctx context.Context, bus Source[T], options ...SubscriptionOption[T]) (<-chan T, UnsubscribeFunc) {
	subscription := newSubscription(options)
	unsubscribe := bus.SubscribeUntilDone(ctx, &subscription)
	return subscription.channel, unsubscribe
}

// SubscribeWithFilter TODO
func SubscribeWithFilter[T, R any](bus Source[T], filter SubscriptionFilter[T, R], options ...SubscriptionOption[R]) (<-chan R, UnsubscribeFunc) {
	return SubscribeWithFilterUntilDone(nil, bus, filter, options...)
}

// SubscribeWithFilterUntilDone TODO
func SubscribeWithFilterUntilDone[T, R any](ctx context.Context, source Source[T], filter SubscriptionFilter[T, R], options ...SubscriptionOption[R]) (<-chan R, UnsubscribeFunc) {
	subscription := newFilterSubscription(filter, options)
	unsubscribe := source.SubscribeUntilDone(ctx, &subscription)
	return subscription.channel, unsubscribe
}

// SubscribeUntilDone adds the subscriber and returns a cancel function.
func (s *source[T]) SubscribeUntilDone(ctx context.Context, subscriber Subscriber[T]) UnsubscribeFunc {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	// create a new subscription and add it
	s.subscribers[subscriber] = exists

	// create the unsubscribe function
	unsubscribe := func() {
		s.mtx.Lock()
		defer s.mtx.Unlock()
		delete(s.subscribers, subscriber)
		subscriber.Close()
	}

	// wait for context done and auto-unsubscribe
	if ctx != nil {
		go func() {
			<-ctx.Done()
			unsubscribe()
		}()
	}

	return unsubscribe
}

// Send the event to all of the subscribers
func (s *source[T]) Send(event T) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	for sub := range s.subscribers {
		sub.Receive(event)
	}
}

// Subscribers returns the current number of subscribers
func (s *source[T]) Subscribers() int {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return len(s.subscribers)
}

// Relay will relay from source to destination. It runs a separate goroutine, consuming events from the source and
// sending events to the destination. When the supplied context is Done, the relay is automatically unsubscribed from
// the source and the destination will no longer receive events.
func Relay[T any](ctx context.Context, source Source[T], destination Source[T]) {
	channel, unsubscribe := Subscribe(source)
	go relay(ctx, channel, unsubscribe, destination)
}

// RelayWithFilter will relay from source to destination with the specified filter. It runs a separate goroutine,
// consuming events from the source, running the filter, and sending events to the destination. When the supplied
// context is Done, the relay is automatically unsubscribed from the source and the destination will no longer receive
// events.
func RelayWithFilter[T, R any](ctx context.Context, source Source[T], filter SubscriptionFilter[T, R], destination Source[R]) {
	channel, unsubscribe := SubscribeWithFilter(source, filter)
	go relay(ctx, channel, unsubscribe, destination)
}

func relay[T any](ctx context.Context, channel <-chan T, unsubscribe UnsubscribeFunc, destination Source[T]) {
	defer unsubscribe()
	for {
		select {
		case event, ok := <-channel:
			if ok {
				destination.Send(event)
			} else {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
