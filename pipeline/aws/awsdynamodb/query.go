package awsdynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nisimpson/piper"
	"github.com/nisimpson/piper/pipeline"
)

// Querier defines an interface for DynamoDB query operations.
type Querier interface {
	Query(ctx context.Context, input *dynamodb.QueryInput, opts ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
}

// sendQuery creates a pipeline transformation that executes the query operation
// against DynamoDB using the provided Querier interface. It handles any errors
// through the options error handler and returns the operation output.
func sendQuery(q Querier, ctx context.Context, opts *Options) piper.Pipe {
	return pipeline.Map(func(input *dynamodb.QueryInput) *dynamodb.QueryOutput {
		output, err := q.Query(ctx, input, opts.DynamoDBOptions...)
		if err != nil {
			opts.HandleError(err)
			return nil
		}
		return output
	})
}

// FromQuery creates a [piper.Pipeline] that processes a single DynamoDB Query operation.
// It sets up a pipeline that:
//  1. Takes a single QueryInput
//  2. Executes the query operation
//  3. Filters out any error results
//  4. Sends [dynamodb.QueryOutput] items downstream
func FromQuery(q Querier, ctx context.Context, input *dynamodb.QueryInput, opts ...func(*Options)) piper.Pipeline {
	var (
		options = newClientOptions().apply(opts)
		in      = make(chan any, 1)

		p = pipeline.FromChannel(in).Thru(
			sendQuery(q, ctx, options), // send dynamodb query
			dropIfNil(),                // if result is nil, do not pass it along
		)
	)

	in <- input // pass in input
	close(in)   // close the input channel
	return p    // return the pipeline
}

// MapQueryFunction maps a generic input type
// to a DynamoDB QueryInput. This allows for flexible transformation
// of various input types into DynamoDB query requests.
type MapQueryFunction[In any] func(In) *dynamodb.QueryInput

// mapToQueryInput creates a pipeline transformation that converts input data
// of type In to a DynamoDB QueryInput using the provided mapping function.
func mapToQueryInput[In any](mapfn MapQueryFunction[In]) piper.Pipe {
	return pipeline.Map(func(input In) *dynamodb.QueryInput {
		return mapfn(input)
	})
}

// Query creates a [piper.Pipe] for using upstream items as context for querying DynamoDB.
// It combines three operations:
//  1. Converting input data to a query request
//  2. Sending the query request to DynamoDB
//  3. Filtering out any error results
//  4. Sending [dynamodb.QueryOutput] results downstream
func Query[In any](q Querier, ctx context.Context, mapfn MapQueryFunction[In], opts ...func(*Options)) piper.Pipe {
	options := newClientOptions().apply(opts)
	return piper.Join(
		mapToQueryInput(mapfn),     // convert input data into query request
		sendQuery(q, ctx, options), // send dynamodb query
		dropIfNil(),                // if result is nil, do not pass it along
	)
}
