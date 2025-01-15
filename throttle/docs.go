/*
Package throttle provides rate limiting and flow control components for data processing pipelines.
It offers flexible mechanisms to control the rate at which items flow through a pipeline,
helping to manage resource utilization and prevent system overload.

Usage:

	// Create a limiter allowing 10 items per second with burst of 1
	limiter := rate.NewLimiter(rate.Every(100*time.Millisecond), 1)

	// Apply rate limiting to a pipeline
	pipeline.
		From(source).
	    Thru(throttle.Limit(ctx, limiter).
	    To(destination)
*/
package throttle
