package awsdynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nisimpson/piper"
	"github.com/nisimpson/piper/pipeline"
)

// Putter defines an interface for DynamoDB put operations.
type Putter interface {
	PutItem(context.Context, *dynamodb.PutItemInput, ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
}

// MapPutFunction maps a generic input type
// to a DynamoDB PutItemInput. This allows for flexible transformation
// of various input types into DynamoDB put requests.
type MapPutFunction[In any] func(In) *dynamodb.PutItemInput

// mapToPut creates a pipeline transformation that converts input data
// of type In to a DynamoDB PutItemInput using the provided mapping function.
func mapToPut[In any](mapfn MapPutFunction[In]) piper.Pipe {
	return pipeline.Map(func(in In) *dynamodb.PutItemInput {
		return mapfn(in)
	})
}

// sendPut creates a pipeline transformation that executes the put operation
// against DynamoDB using the provided [Putter] interface. It handles any errors
// through the options error handler and returns the operation output.
func sendPut(p Putter, ctx context.Context, opts *Options) piper.Pipe {
	return pipeline.Map(func(in *dynamodb.PutItemInput) *dynamodb.PutItemOutput {
		out, err := p.PutItem(ctx, in, opts.DynamoDBOptions...)
		if err != nil {
			opts.HandleError(err)
			return nil
		}
		return out
	})
}

// Put creates a [piper.Pipe] for putting upstream items into DynamoDB.
// It combines three operations:
//  1. Converting input data to a put request via [MapPutFunction]
//  2. Sending the put request to DynamoDB
//  3. Filtering out any error results
//  4. Sending successful [dynamodb.PutItemOutput] results downstream
func Put[In any](p Putter, ctx context.Context, mapfn MapPutFunction[In], opts ...func(*Options)) piper.Pipe {
	options := newClientOptions().apply(opts)
	return pipeline.Join(
		mapToPut(mapfn),                     // convert the incoming data to a put item request
		sendPut(p, ctx, options),            // send the put item request
		dropIfNil[dynamodb.PutItemOutput](), // don't pass along any errors
	)
}
