package piper

// Inlet represents a component that can receive data through a channel.
// Any component that needs to receive data from another component in the pipeline must implement this interface.
type Inlet interface {
	// In returns a write-only channel that can be used to send data into this component
	In() chan<- any
}

// Outlet represents a component that can send data through a channel.
// Any component that needs to output data to another component in the pipeline must implement this interface.
type Outlet interface {
	// Out returns a read-only channel that can be used to receive data from this component
	Out() <-chan any
}

// Pipe represents a component that can both receive and send data.
// It combines both Inlet and Outlet interfaces, making it suitable for middleware components
// that process data and pass it along in a pipeline.
type Pipe interface {
	Inlet
	Outlet
}

// Source represents the starting point of a pipeline.
// It only implements Outlet since it only produces data and doesn't receive any input.
type Source interface {
	Outlet
}

// Sink represents the endpoint of a pipeline.
// It only implements Inlet since it only receives data and doesn't produce any output.
type Sink interface {
	Inlet
}
