package awsddb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nisimpson/piper"
	"github.com/nisimpson/piper/pipeline"
)

// Getter defines an interface for DynamoDB get operations.
type Getter interface {
	GetItem(ctx context.Context, input *dynamodb.GetItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
}

// sendGet creates a pipeline transformation that executes the get operation
// against DynamoDB using the provided Getter interface. It handles any errors
// through the options error handler and returns the operation output.
func sendGet(g Getter, ctx context.Context, opts *Options) piper.Pipe {
	return pipeline.Map(func(input *dynamodb.GetItemInput) *dynamodb.GetItemOutput {
		output, err := g.GetItem(ctx, input, opts.DynamoDBOptions...)
		if err != nil {
			opts.HandleError(err)
			return nil
		}
		return output
	})
}

// FromGet creates a [pipeline.Flow] that processes a single DynamoDB GetItem operation.
//  1. Takes a single GetItemInput
//  2. Executes the get operation
//  3. Filters out any error results
//  4. Passes successful results downstream
func FromGet(g Getter, ctx context.Context, input *dynamodb.GetItemInput, opts ...func(*Options)) pipeline.Flow {
	var (
		options = newClientOptions().apply(opts)
		in      = make(chan any, 1)

		p = pipeline.FromChannel(in).Thru(
			sendGet(g, ctx, options),            // send dynamodb get
			dropIfNil[dynamodb.GetItemOutput](), // if result is nil, do not pass it along
		)
	)

	in <- input // pass in input
	close(in)   // close the input channel
	return p    // return the pipeline
}

// MapGetFunction maps a generic input type
// to a DynamoDB GetItemInput. This allows for flexible transformation
// of various input types into DynamoDB get requests.
type MapGetFunction[In any] func(In) *dynamodb.GetItemInput

// mapToGetItemInput creates a pipeline transformation that converts input data
// of type In to a DynamoDB GetItemInput using the provided mapping function.
func mapToGetItemInput[In any](mapfn MapGetFunction[In]) piper.Pipe {
	return pipeline.Map(func(input In) *dynamodb.GetItemInput {
		return mapfn(input)
	})
}

// Get creates a [piper.Pipe] for using upstream input as context for retrieving items from DynamoDB.
// It combines three operations:
//  1. Converting input data to a get request
//  2. Sending the get request to DynamoDB
//  3. Filtering out any error results
func Get[In any](g Getter, ctx context.Context, mapfn MapGetFunction[In], opts ...func(*Options)) piper.Pipe {
	options := newClientOptions().apply(opts)
	return pipeline.Join(
		mapToGetItemInput(mapfn),            // convert input data into get request
		sendGet(g, ctx, options),            // send dynamodb get
		dropIfNil[dynamodb.GetItemOutput](), // if result is nil, do not pass it along
	)
}
