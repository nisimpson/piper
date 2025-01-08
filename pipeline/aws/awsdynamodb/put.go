package awsdynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nisimpson/piper"
	"github.com/nisimpson/piper/pipeline"
)

type Putter interface {
	PutItem(context.Context, *dynamodb.PutItemInput, ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
}

type MapPutFunction[In any] func(In) *dynamodb.PutItemInput

func mapToPut[In any](mapfn MapPutFunction[In]) piper.Pipe {
	return pipeline.Map(func(in In) *dynamodb.PutItemInput {
		return mapfn(in)
	})
}

func sendPut(ctx context.Context, p Putter, opts *Options) piper.Pipe {
	return pipeline.Map(func(in *dynamodb.PutItemInput) *dynamodb.PutItemOutput {
		out, err := p.PutItem(ctx, in, opts.DynamoDBOptions...)
		if err != nil {
			opts.HandleError(err)
			return nil
		}
		return out
	})
}

func Put[In any](p Putter, ctx context.Context, mapfn MapPutFunction[In], opts ...func(*Options)) piper.Pipe {
	options := newClientOptions().apply(opts)
	return piper.Join(
		mapToPut(mapfn),          // convert the incoming data to a put item request
		sendPut(ctx, p, options), // send the put item request
		dropIfNil(),              // don't pass along any errors
	)
}
