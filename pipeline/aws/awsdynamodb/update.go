package awsdynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nisimpson/piper"
	"github.com/nisimpson/piper/pipeline"
)

type Updater interface {
	UpdateItem(ctx context.Context, input *dynamodb.UpdateItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
}

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

type UpdateMapFunction[In any] func(In) *dynamodb.UpdateItemInput

func mapToUpdateItemInput[In any](mapfn UpdateMapFunction[In]) piper.Pipe {
	return pipeline.Map(func(input In) *dynamodb.UpdateItemInput {
		return mapfn(input)
	})
}

func Update[In any](u Updater, ctx context.Context, mapfn UpdateMapFunction[In], opts ...func(*Options)) piper.Pipe {
	options := newClientOptions().apply(opts)
	return piper.Join(
		mapToUpdateItemInput(mapfn), // convert input data into update request
		sendUpdate(u, ctx, options), // send dynamodb update
		dropIfNil(),                 // if result is nil, do not pass it along
	)
}
