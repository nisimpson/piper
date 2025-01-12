package awsddb

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nisimpson/piper"
	"github.com/nisimpson/piper/internal/must"
	"github.com/nisimpson/piper/pipeline"
)

// Options defines the configuration settings for DynamoDB pipeline operations.
// It provides customization for error handling and DynamoDB client options.
type Options struct {
	// HandleError is a function that processes errors encountered during
	// DynamoDB operations. It allows custom error handling strategies.
	HandleError func(error)

	// DynamoDBOptions is a slice of option functions that modify the
	// configuration of DynamoDB operations. These options are passed
	// directly to the DynamoDB client calls.
	DynamoDBOptions []func(*dynamodb.Options)
}

// newClientOptions creates and returns a new Options instance with default settings.
// By default, it sets the HandleError function to ignore errors using [must.IgnoreError].
func newClientOptions() *Options {
	return &Options{
		HandleError: must.IgnoreError,
	}
}

// apply takes a slice of option modifier functions and applies them in sequence
// to the current Options instance. This allows for flexible configuration of
// the Options struct through functional options pattern.
//
// Parameters:
//   - opts: A slice of functions that modify the Options struct
//
// Returns the modified Options instance
func (o *Options) apply(opts []func(*Options)) *Options {
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// dropIfNil creates a pipeline transformation that filters out nil values
// from the pipeline. This is commonly used to remove error results or
// empty responses from the processing chain.
//
// Returns a pipe that drops nil values from the pipeline
func dropIfNil[In any]() piper.Pipe {
	return pipeline.DropIf(func(in *In) bool {
		return in == nil
	})
}
