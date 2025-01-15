package pipeline

import (
	"time"

	"github.com/nisimpson/piper"
)

// BatcherOptions configure both how and when items are batched together.
type BatcherOptions struct {
	// MaxSize is the maximum number of items to include in a batch before sending.
	// For an unbounded size, set MaxSize to any value less than 1.
	MaxSize int
	// Interval is the maximum time to wait before sending a batch, even if MaxSize hasn't been reached.
	// To disable, set Interval to a zero duration.
	Interval time.Duration
}

// batcher implements a pipeline component that groups incoming items into batches
// based on size and/or time constraints.
type batcher[In any] struct {
	// in receives individual items to be batched
	in chan any
	// out sends completed batches
	out chan any
	// options controls batching behavior
	options BatcherOptions
}

// Batch creates a new [piper.Pipe] component that groups items into batches.
// The batching behavior can be customized through the provided option functions.
func Batch[In any](opts ...func(*BatcherOptions)) piper.Pipe {
	options := BatcherOptions{
		MaxSize: 1,
	}
	for _, opt := range opts {
		opt(&options)
	}
	pipe := batcher[In]{
		in:      make(chan any),
		out:     make(chan any),
		options: options,
	}

	go pipe.start()
	return pipe
}

// BatchN creates a new batcher that groups exactly N items together before sending them downstream.
// This is a convenience wrapper around [Batch] that sets only the MaxSize option.
func BatchN[In any](size int) piper.Pipe {
	return Batch[In](func(bo *BatcherOptions) {
		bo.MaxSize = size
	})
}

// BatchEvery creates a new batcher that sends batches at regular time intervals.
// Any items received during the interval will be included in the next batch. This is a
// convenience wrapper around [Batch] that leaves the maximum size unbounded.
func BatchEvery[In any](d time.Duration) piper.Pipe {
	return Batch[In](func(bo *BatcherOptions) {
		bo.Interval = d
		bo.MaxSize = -1
	})
}

func (b batcher[In]) In() chan<- any  { return b.in }
func (b batcher[In]) Out() <-chan any { return b.out }

// start begins the batching process, collecting items and sending batches based on the configured options.
// It handles both size-based and time-based batching strategies.
func (b batcher[In]) start() {
	batch := b.newSlice()
	defer close(b.out)

	if b.options.Interval == 0 {
		for {
			input, next := <-b.in
			if !next {
				b.send(batch)
				return
			}
			batch = append(batch, input.(In))
			if len(batch) == b.options.MaxSize {
				batch = b.send(batch)
			}
		}
	}

	for {
		select {
		case input, next := <-b.in:
			if !next {
				b.send(batch)
				return
			}
			batch = append(batch, input.(In))
			if len(batch) == b.options.MaxSize {
				batch = b.send(batch)
			}
		case <-time.After(b.options.Interval):
			batch = b.send(batch)
		}
	}
}

// send emits the current batch downstream and initializes a new empty batch.
// If the current batch is empty, it is returned as-is without sending.
func (b batcher[In]) send(batch []In) []In {
	if len(batch) == 0 {
		return batch
	}
	b.out <- batch
	return b.newSlice()
}

// newSlice creates a new empty slice to hold the next batch of items.
func (b batcher[In]) newSlice() []In {
	if b.options.MaxSize <= 0 {
		return make([]In, 0)
	}
	return make([]In, 0, b.options.MaxSize)
}

// BatchAll returns a [piper.Pipe] that collects all upstream items and sends them downstream as a slice.
// It is the semantic equivalent to
//
//	BatchN[T](-1)
func BatchAll[In any]() piper.Pipe {
	return BatchN[In](-1)
}
