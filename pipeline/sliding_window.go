package pipeline

import (
	"time"

	"github.com/nisimpson/piper"
)

// SlidingWindowOptions configures how items are batched using a sliding window.
type SlidingWindowOptions struct {
	// WindowSize is the size of the sliding window
	WindowSize int
	// StepSize is how many items to slide forward for each new window
	StepSize int
	// Interval is the maximum time to wait before sliding the window
	// To disable time-based sliding, set Interval to zero
	Interval time.Duration
}

// slidingWindow implements a pipeline component that groups items using a sliding window approach
type slidingWindow[In any] struct {
	in      chan any
	out     chan any
	options SlidingWindowOptions
}

// SlidingWindow creates a new pipeline component that groups items using a sliding window.
// The window moves forward by StepSize items after each batch is emitted.
func SlidingWindow[In any](opts ...func(*SlidingWindowOptions)) piper.Pipe {
	options := SlidingWindowOptions{
		WindowSize: 2,
		StepSize:   1,
	}
	for _, opt := range opts {
		opt(&options)
	}

	if options.StepSize <= 0 {
		options.StepSize = 1
	}
	if options.WindowSize < options.StepSize {
		options.WindowSize = options.StepSize
	}

	pipe := slidingWindow[In]{
		in:      make(chan any),
		out:     make(chan any),
		options: options,
	}

	go pipe.start()
	return pipe
}

func (sw slidingWindow[In]) In() chan<- any  { return sw.in }
func (sw slidingWindow[In]) Out() <-chan any { return sw.out }

func (sw slidingWindow[In]) start() {
	defer close(sw.out)

	// buffer stores all items that might be needed for future windows
	buffer := make([]In, 0, sw.options.WindowSize)

	if sw.options.Interval == 0 {
		for {
			input, ok := <-sw.in
			if !ok {
				sw.emitRemainingWindows(buffer)
				return
			}

			buffer = append(buffer, input.(In))
			if len(buffer) >= sw.options.WindowSize {
				// Create and send the current window
				window := make([]In, sw.options.WindowSize)
				copy(window, buffer)
				sw.out <- window

				// Slide the window forward
				buffer = buffer[sw.options.StepSize:]
			}
		}
	}

	for {
		select {
		case input, ok := <-sw.in:
			if !ok {
				sw.emitRemainingWindows(buffer)
				return
			}

			buffer = append(buffer, input.(In))
			if len(buffer) >= sw.options.WindowSize {
				window := make([]In, sw.options.WindowSize)
				copy(window, buffer)
				sw.out <- window
				buffer = buffer[sw.options.StepSize:]
			}

		case <-time.After(sw.options.Interval):
			if len(buffer) >= sw.options.WindowSize {
				window := make([]In, sw.options.WindowSize)
				copy(window, buffer)
				sw.out <- window
				buffer = buffer[sw.options.StepSize:]
			}
		}
	}
}

func (sw slidingWindow[In]) emitRemainingWindows(buffer []In) {
	// Emit any remaining complete windows
	for len(buffer) >= sw.options.WindowSize {
		window := make([]In, sw.options.WindowSize)
		copy(window, buffer)
		sw.out <- window
		buffer = buffer[sw.options.StepSize:]
	}
}
