package eventbus

type subscriptionOptions[T any] struct {
	channel chan T
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
