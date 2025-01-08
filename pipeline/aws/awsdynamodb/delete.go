package awsdynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nisimpson/piper"
	"github.com/nisimpson/piper/pipeline"
)

type Deleter interface {
	PutItem(context.Context, *dynamodb.PutItemInput, ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
}

type MapDeleteFunction[In any] func(In) *dynamodb.PutItemInput

func mapToDeleteInput[In any](mapfn MapDeleteFunction[In]) piper.Pipe {
	return pipeline.Map(func(in In) *dynamodb.PutItemInput {
		return mapfn(in)
	})
}

func sendDelete(d Deleter, ctx context.Context, opts *Options) piper.Pipe {
	return pipeline.Map(func(in *dynamodb.PutItemInput) *dynamodb.PutItemOutput {
		out, err := d.PutItem(ctx, in, opts.DynamoDBOptions...)
		if err != nil {
			opts.HandleError(err)
			return nil
		}
		return out
	})
}

func Delete[In any](d Deleter, ctx context.Context, mapfn MapDeleteFunction[In], opts ...func(*Options)) piper.Pipe {
	options := newClientOptions().apply(opts)
	return piper.Join(
		mapToDeleteInput(mapfn),     // convert the incoming data to a delete item request
		sendDelete(d, ctx, options), // send the put item request
		dropIfNil(),                 // don't pass along any errors
	)
}
