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

import "time"

type subscriptionOptions[T any] struct {
	channel           chan T
	unbounded         bool
	unboundedInterval time.Duration
}

func makeSubscriptionOptions[T any](options []SubscriptionOption[T]) subscriptionOptions[T] {
	opts := subscriptionOptions[T]{}
	for _, opt := range options {
		opt(&opts)
	}
	return opts
}

// SubscriptionOption is used to provide options when creating a new subscription to a Source.
type SubscriptionOption[T any] func(*subscriptionOptions[T])

// WithChannel allows a subscriber to specify the channel that will receive events. This allows the subscriber to
// control the size.
func WithChannel[T any](channel chan T) SubscriptionOption[T] {
	return func(opts *subscriptionOptions[T]) {
		opts.channel = channel
	}
}

// WithUnboundedChannel specifies that util.UnboundedChan should be used for the channel. This will allow an unbounded
// number of events to be received before being dispatched to subscribers.
func WithUnboundedChannel[T any](interval time.Duration) SubscriptionOption[T] {
	return func(opts *subscriptionOptions[T]) {
		opts.unbounded = true
		opts.unboundedInterval = interval
	}
}
