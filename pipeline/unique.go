package pipeline

import (
	"fmt"

	"github.com/nisimpson/piper"
)

// UniqueOptions defines configuration options for the [Unique] pipe.
// It allows customization of how uniqueness is determined for elements.
type UniqueOptions[T any] struct {
	// KeyFunc is a function that generates a string key for a given element.
	// This key is used to determine uniqueness. If two elements generate the
	// same key, they are considered duplicates.
	KeyFunc func(T) string
}

// uniquePipe implements a [piper.Pipe] that filters out duplicate elements
// based on a key function. It maintains internal channels for communication
// and configuration options for customization.
type uniquePipe[In any] struct {
	// options holds the configuration for determining element uniqueness
	options UniqueOptions[In]
	// in is the channel for receiving input elements
	in chan any
	// out is the channel for sending unique elements
	out chan any
}

// Unique creates a new [piper.Pipe] that filters out duplicate elements from the stream.
// It accepts optional configuration functions to customize the pipe's behavior.
// By default, it uses the string representation of elements as the uniqueness key.
func Unique[In any](opts ...func(*UniqueOptions[In])) piper.Pipe {
	options := UniqueOptions[In]{
		KeyFunc: func(in In) string { return fmt.Sprintf("%v", in) },
	}
	for _, opt := range opts {
		opt(&options)
	}
	pipe := uniquePipe[In]{
		options: options,
		in:      make(chan any),
		out:     make(chan any),
	}
	go pipe.start()
	return pipe
}

// In returns the input channel for the pipe.
// Elements sent to this channel will be processed for uniqueness.
func (u uniquePipe[In]) In() chan<- any { return u.in }

// Out returns the output channel for the pipe.
// Only unique elements will be sent to this channel.
func (u uniquePipe[In]) Out() <-chan any { return u.out }

// start begins the unique filtering process.
// It reads elements from the input channel, applies the key function,
// and forwards only unique elements to the output channel.
// The function runs until the input channel is closed, then closes
// the output channel.
func (u uniquePipe[In]) start() {
	defer close(u.out)
	unique := make(map[string]struct{})
	for item := range u.in {
		var (
			input = item.(In)
			key   = u.options.KeyFunc(input)
		)
		if _, ok := unique[key]; !ok {
			unique[key] = struct{}{}
			u.out <- item
		}
	}
}
