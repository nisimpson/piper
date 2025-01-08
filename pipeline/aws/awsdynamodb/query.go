package awsdynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nisimpson/piper"
	"github.com/nisimpson/piper/pipeline"
)

type Querier interface {
	Query(ctx context.Context, input *dynamodb.QueryInput, opts ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
}

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

type QueryMapFunction[In any] func(In) *dynamodb.QueryInput

func mapToQueryInput[In any](mapfn QueryMapFunction[In]) piper.Pipe {
	return pipeline.Map(func(input In) *dynamodb.QueryInput {
		return mapfn(input)
	})
}

func Query[In any](q Querier, ctx context.Context, mapfn QueryMapFunction[In], opts ...func(*Options)) piper.Pipe {
	options := newClientOptions().apply(opts)
	return piper.Join(
		mapToQueryInput(mapfn),     // convert input data into query request
		sendQuery(q, ctx, options), // send dynamodb query
		dropIfNil(),                // if result is nil, do not pass it along
	)
}
