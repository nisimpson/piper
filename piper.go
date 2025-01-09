/*
Package piper provides a data pipeline framework for Go applications.
It enables the construction of flexible data processing pipelines with support
for operations like mapping, filtering, fan-out, and fan-in patterns.

Build data processing workflows by connecting various components:
  - A [Source] to generate or provide input data
  - A [Pipe] to transform, filter, or process data
  - A [Sink] to consume or store the final results

Key features include:
  - Type-safe(ish) pipeline operations using Go generics
  - Concurrent processing using Go channels
  - Composable pipeline components that can be chained together
  - Built-in support for common operations like map, filter, reduce
  - Fan-out/fan-in capabilities for splitting and merging data streams

[Pipeline] construction follows a fluent builder pattern:
 1. Start with the [PipelineFrom] constructor to specify a data source
 2. Add processing steps with [Pipeline.Thru] to transform data
 3. Complete with [Pipeline.To] to specify where results should be delivered

Piper leverages Go's concurrency primitives to ensure efficient parallel processing while maintaining
data ordering guarantees.
*/
package piper

// Inlet represents a component that can receive data through a channel.
// Any component that needs to receive data from another component in a [Pipeline] must implement this interface.
type Inlet interface {
	// In returns a write-only channel that can be used to send data into this component
	In() chan<- any
}

// Outlet represents a component that can send data through a channel.
// Any component that needs to output data to another component in a [Pipeline] must implement this interface.
type Outlet interface {
	// Out returns a read-only channel that can be used to receive data from this component
	Out() <-chan any
}

// Pipe represents a component that can both receive and send data.
// It combines both [Inlet] and [Outlet] interfaces, making it suitable for middleware components
// that process data and pass it along in a pipeline.
type Pipe interface {
	Inlet
	Outlet
}

// Source represents the starting point of a [Pipeline].
// It only implements [Outlet] since it only produces data and doesn't receive any input.
type Source interface {
	Outlet
}

// Sink represents the endpoint of a [Pipeline].
// It only implements [Inlet] since it only receives data and doesn't produce any output.
type Sink interface {
	Inlet
}
