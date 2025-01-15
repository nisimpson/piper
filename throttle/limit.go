package throttle

import (
	"context"

	"github.com/nisimpson/piper"
	"golang.org/x/time/rate"
)

// limiter represents a pipeline stage that controls the rate at which items
// flow through the pipeline using a rate.Limiter. It implements the piper.Pipe
// interface to integrate with the pipeline processing system.
type limiter struct {
	// limit is the rate limiter that controls the flow of items through
	// the pipeline. It determines how many items can pass through per
	// time unit and handles burst capacity.
	limit *rate.Limiter

	// in is the input channel that receives items from the previous
	// pipeline stage. It's used to accept incoming data that needs
	// to be rate limited.
	in chan any

	// out is the output channel that receives items from the previous
	// pipeline stage. It's used to send outgoing data downstream.
	out chan any

	// ctx is the context that manages cancellation or deadline checks
	// for the limiter.
	ctx context.Context
}

// Limit creates a new [piper.Pipe] that rate limits the flow of items
// using the provided [rate.Limiter].
func Limit(ctx context.Context, limit *rate.Limiter) piper.Pipe {
	limiter := limiter{
		limit: limit,
		in:    make(chan any),
		out:   make(chan any),
		ctx:   ctx,
	}

	go limiter.start()
	return limiter
}

// In returns the input channel for the rate limiter stage.
// This channel is used to receive items that need to be rate limited.
func (l limiter) In() chan<- any { return l.in }

// Out returns the output channel from the next pipeline stage.
// This channel provides access to the rate-limited items after they've
// been processed by the downstream stage.
func (l limiter) Out() <-chan any { return l.out }

// start begins the main processing loop for the rate limiter.
// It continuously reads from the input channel, applies rate limiting,
// and forwards items to the next pipeline stage. The loop continues until
// either:
//  1. The input channel is closed
//  2. The context is cancelled
//  3. A rate limiting error occurs
//
// The function ensures proper cleanup by closing the downstream pipe's
// input channel when processing is complete.
func (l limiter) start() {
	defer close(l.out)
	for {
		// Wait for rate limit allowance
		if err := l.limit.Wait(l.ctx); err != nil {
			return
		}

		// Process next item if available
		if data, ok := <-l.in; ok {
			l.out <- data
		} else {
			return
		}
	}
}
