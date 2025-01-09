package awsdynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nisimpson/piper"
	"github.com/nisimpson/piper/pipeline"
)

// Updater defines an interface for DynamoDB update operations.
type Updater interface {
	UpdateItem(ctx context.Context, input *dynamodb.UpdateItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
}

// sendUpdate creates a pipeline transformation that executes the update operation
// against DynamoDB using the provided Updater interface. It handles any errors
// through the options error handler and returns the operation output.
func sendUpdate(u Updater, ctx context.Context, opts *Options) piper.Pipe {
	return pipeline.Map(func(input *dynamodb.UpdateItemInput) *dynamodb.UpdateItemOutput {
		output, err := u.UpdateItem(ctx, input, opts.DynamoDBOptions...)
		if err != nil {
			opts.HandleError(err)
			return nil
		}
		return output
	})
}

// FromUpdate creates a [piper.Pipeline] that processes a single DynamoDB Update operation.
// It sets up a pipeline that:
//  1. Takes a single UpdateItemInput
//  2. Executes the update operation
//  3. Filters out any error results
//  4. Sends [dynamodb.UpdateItemOutput] items downstream
func FromUpdate(u Updater, ctx context.Context, input *dynamodb.UpdateItemInput, opts ...func(*Options)) piper.Pipeline {
	var (
		options = newClientOptions().apply(opts)
		in      = make(chan any, 1)

		p = pipeline.FromChannel(in).Thru(
			sendUpdate(u, ctx, options), // send dynamodb update
			dropIfNil(),                 // if result is nil, do not pass it along
		)
	)

	in <- input // pass in input
	close(in)   // close the input channel
	return p    // return the pipeline
}

// MapUpdateFunction maps a generic input type
// to a DynamoDB UpdateItemInput. This allows for flexible transformation
// of various input types into DynamoDB update requests.
type MapUpdateFunction[In any] func(In) *dynamodb.UpdateItemInput

// mapToUpdateInput creates a pipeline transformation that converts input data
// of type In to a DynamoDB UpdateItemInput using the provided mapping function.
func mapToUpdateItemInput[In any](mapfn MapUpdateFunction[In]) piper.Pipe {
	return pipeline.Map(func(input In) *dynamodb.UpdateItemInput {
		return mapfn(input)
	})
}

// Update creates a [piper.Pipe] for using upstream items as context for updating DynamoDB items.
// It combines three operations:
//  1. Converting input data to an update request
//  2. Sending the update request to DynamoDB
//  3. Filtering out any error results
//  4. Sending [dynamodb.UpdateItemOutput] results downstream
func Update[In any](u Updater, ctx context.Context, mapfn MapUpdateFunction[In], opts ...func(*Options)) piper.Pipe {
	options := newClientOptions().apply(opts)
	return pipeline.Join(
		mapToUpdateItemInput(mapfn), // convert input data into update request
		sendUpdate(u, ctx, options), // send dynamodb update
		dropIfNil(),                 // if result is nil, do not pass it along
	)
}
