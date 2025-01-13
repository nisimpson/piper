package awsddb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nisimpson/piper"
	"github.com/nisimpson/piper/pipeline"
)

// Deleter defines an interface for DynamoDB delete operations.
type Deleter interface {
	DeleteItem(context.Context, *dynamodb.DeleteItemInput, ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
}

// MapDeleteFunction defines a function signature for mapping
// a generic input type to a DynamoDB DeleteItemInput.
type MapDeleteFunction[In any] func(In) *dynamodb.DeleteItemInput

// mapToDeleteInput creates a pipeline transformation that converts input data
// of type In to a DynamoDB DeleteItemInput using the provided mapping function.
// This function is part of the delete operation pipeline.
func mapToDeleteInput[In any](mapfn MapDeleteFunction[In]) piper.Pipe {
	return pipeline.Map(func(in In) *dynamodb.DeleteItemInput {
		return mapfn(in)
	})
}

// sendDelete creates a pipeline transformation that executes the delete operation
// against DynamoDB using the provided Deleter interface. It handles any errors
// through the options error handler and returns the operation output.
func sendDelete(d Deleter, ctx context.Context, opts *Options) piper.Pipe {
	return pipeline.Map(func(in *dynamodb.DeleteItemInput) *dynamodb.DeleteItemOutput {
		out, err := d.DeleteItem(ctx, in, opts.DynamoDBOptions...)
		if err != nil {
			opts.HandleError(err)
			return nil
		}
		return out
	})
}

// Delete creates a [piper.Pipe] for using upstream input as context for deleting items from DynamoDB.
// It combines three operations:
//  1. Converting input data to a delete request
//  2. Sending the delete request to DynamoDB
//  3. Filtering out any error results
func Delete[In any](d Deleter, ctx context.Context, mapfn MapDeleteFunction[In], opts ...func(*Options)) piper.Pipe {
	options := newClientOptions().apply(opts)
	return pipeline.Join(
		mapToDeleteInput(mapfn),                // convert the incoming data to a delete item request
		sendDelete(d, ctx, options),            // send the put item request
		dropIfNil[dynamodb.DeleteItemOutput](), // don't pass along any errors
	)
}
