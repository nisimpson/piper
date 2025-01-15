package throttle_test

import (
	"context"
	"fmt"
	"time"

	"github.com/nisimpson/piper/pipeline"
	"github.com/nisimpson/piper/throttle"
	"golang.org/x/time/rate"
)

// ExampleLimit demonstrates basic rate limiting of pipeline items
func ExampleLimit() {
	// Create a limiter that allows 2 items per second
	limiter := rate.NewLimiter(rate.Every(500*time.Millisecond), 1)

	// Create source pipeline and apply rate limiting
	source := pipeline.FromSlice(1, 2, 3)
	limited := source.Thru(throttle.Limit(limiter, pipeline.Passthrough()))

	start := time.Now()
	for val := range limited.Out() {
		fmt.Printf("Received %v after %v\n", val,
			time.Since(start).Round(100*time.Millisecond))
	}
	// Output:
	// Received 1 after 0s
	// Received 2 after 500ms
	// Received 3 after 1s
}

// ExampleLimit_burst demonstrates rate limiting with burst capacity
func ExampleLimit_burst() {
	// Create a limiter that allows 2 items per second with burst of 2
	limiter := rate.NewLimiter(rate.Every(500*time.Millisecond), 2)

	pipe := throttle.Limit(limiter, pipeline.Passthrough())

	// Send input
	go func() {
		in := pipe.In()
		defer close(in)
		for i := 1; i <= 4; i++ {
			in <- i
		}
	}()

	start := time.Now()
	for val := range pipe.Out() {
		fmt.Printf("Received %v after %v\n", val,
			time.Since(start).Round(100*time.Millisecond))
	}
	// Output:
	// Received 1 after 0s
	// Received 2 after 0s
	// Received 3 after 500ms
	// Received 4 after 1s
}

// ExampleLimitWithContext demonstrates using context for cancellation
func ExampleLimitWithContext() {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Create a limiter that allows 1 item per second
	limiter := rate.NewLimiter(rate.Every(time.Second), 1)

	// Create rate-limited pipeline with context
	pipe := throttle.LimitWithContext(ctx, limiter, pipeline.Passthrough())

	// Send input
	go func() {
		in := pipe.In()
		defer close(in)
		for i := 1; i <= 3; i++ {
			in <- i
		}
	}()

	start := time.Now()
	for val := range pipe.Out() {
		fmt.Printf("Received %v after %v\n", val,
			time.Since(start).Round(100*time.Millisecond))
	}
	// Output:
	// Received 1 after 0s
}

// ExampleLimit_composition demonstrates composing rate limiting with other operations
func ExampleLimit_composition() {
	// Create a limiter that allows 2 items per second
	limiter := rate.NewLimiter(rate.Every(500*time.Millisecond), 1)

	// Create pipeline that doubles numbers and rate limits the output
	source := pipeline.FromSlice(1, 2, 3)
	result := source.
		Thru(pipeline.Map(func(x any) any {
			return x.(int) * 2
		})).
		Thru(throttle.Limit(limiter, pipeline.Passthrough()))

	start := time.Now()
	for val := range result.Out() {
		fmt.Printf("Received %v after %v\n", val,
			time.Since(start).Round(100*time.Millisecond))
	}
	// Output:
	// Received 2 after 0s
	// Received 4 after 500ms
	// Received 6 after 1s
}

// ExampleLimit_strings demonstrates rate limiting with string values
func ExampleLimit_strings() {
	// Create a limiter that allows 2 items per second
	limiter := rate.NewLimiter(rate.Every(500*time.Millisecond), 1)
	pipe := throttle.Limit(limiter, pipeline.Passthrough())

	// Send string input
	go func() {
		in := pipe.In()
		defer close(in)
		words := []string{"fast", "medium", "slow"}
		for _, word := range words {
			in <- word
		}
	}()

	start := time.Now()
	for val := range pipe.Out() {
		fmt.Printf("Received %q after %v\n", val,
			time.Since(start).Round(100*time.Millisecond))
	}
	// Output:
	// Received "fast" after 0s
	// Received "medium" after 500ms
	// Received "slow" after 1s
}

// ExampleLimit_errorHandling demonstrates handling context cancellation
func ExampleLimit_errorHandling() {
	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Create a limiter
	limiter := rate.NewLimiter(rate.Every(time.Second), 1)

	// Create rate-limited pipeline with cancelled context
	pipe := throttle.LimitWithContext(ctx, limiter, pipeline.Passthrough())

	// Send input
	go func() {
		in := pipe.In()
		defer close(in)
		in <- "this won't go through"
	}()

	// Try to receive (should exit immediately)
	count := 0
	for range pipe.Out() {
		count++
	}
	fmt.Printf("Received %d items\n", count)
	// Output:
	// Received 0 items
}

// ExampleLimit_zeroRate demonstrates behavior with zero rate limit
func ExampleLimit_zeroRate() {
	// Create a limiter with zero rate
	limiter := rate.NewLimiter(0, 2)
	pipe := throttle.Limit(limiter, pipeline.Passthrough())

	// Send input
	go func() {
		in := pipe.In()
		defer close(in)
		in <- 1
		in <- 2
	}()

	timeout := time.After(100 * time.Millisecond)

	for {
		select {
		case val := <-pipe.Out():
			fmt.Printf("Received! %v\n", val)
		case <-timeout:
			fmt.Println("Timed out waiting for value")
			return
		}
	}
	// Output:
	// Received! 1
	// Received! 2
	// Timed out waiting for value
}
