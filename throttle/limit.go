package throttle

import (
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

	// pipe is the next stage in the pipeline that will receive
	// the rate-limited items. It represents the downstream
	// processing stage.
	pipe piper.Pipe
}

// Limit creates a new [piper.Pipe] that rate limits the flow of items
// using the provided [rate.Limiter].
func Limit(limit *rate.Limiter, pipe piper.Pipe) piper.Pipe {
	limiter := limiter{
		limit: limit,
		in:    make(chan any),
		pipe:  pipe,
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
func (l limiter) Out() <-chan any { return l.pipe.Out() }

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
	defer close(l.pipe.In())
	for {
		// Wait for rate limit allowance
		if !l.limit.Allow() {
			continue
		}

		// Process next item if available
		if data, ok := <-l.in; ok {
			l.pipe.In() <- data
		} else {
			return
		}
	}
}
