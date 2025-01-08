package awsdynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nisimpson/piper"
	"github.com/nisimpson/piper/pipeline"
)

type Getter interface {
	GetItem(ctx context.Context, input *dynamodb.GetItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
}

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

func FromGet(g Getter, ctx context.Context, input *dynamodb.GetItemInput, opts ...func(*Options)) piper.Pipeline {
	var (
		options = newClientOptions().apply(opts)
		in      = make(chan any, 1)

		p = pipeline.FromChannel(in).Thru(
			sendGet(g, ctx, options), // send dynamodb get
			dropIfNil(),              // if result is nil, do not pass it along
		)
	)

	in <- input // pass in input
	close(in)   // close the input channel
	return p    // return the pipeline
}

type MapGetFunction[In any] func(In) *dynamodb.GetItemInput

func mapToGetItemInput[In any](mapfn MapGetFunction[In]) piper.Pipe {
	return pipeline.Map(func(input In) *dynamodb.GetItemInput {
		return mapfn(input)
	})
}

func Get[In any](q Getter, ctx context.Context, mapfn MapGetFunction[In], opts ...func(*Options)) piper.Pipe {
	options := newClientOptions().apply(opts)
	return piper.Join(
		mapToGetItemInput(mapfn), // convert input data into get request
		sendGet(q, ctx, options), // send dynamodb get
		dropIfNil(),              // if result is nil, do not pass it along
	)
}
