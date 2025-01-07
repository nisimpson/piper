package awsdynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nisimpson/piper"
	"github.com/nisimpson/piper/pipeline"
)

type Scanner interface {
	Scan(ctx context.Context, input *dynamodb.ScanInput, opts ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
}

func sendScan(s Scanner, ctx context.Context, opts *Options) piper.Pipe {
	return pipeline.Map(func(input *dynamodb.ScanInput) *dynamodb.ScanOutput {
		output, err := s.Scan(ctx, input, opts.DynamoDBOptions...)
		if err != nil {
			opts.HandleError(err)
			return nil
		}
		return output
	})
}

func FromScan(s Scanner, ctx context.Context, input *dynamodb.ScanInput, opts ...func(*Options)) piper.Pipeline {
	var (
		options = newClientOptions().apply(opts)
		in      = make(chan any, 1)

		p = pipeline.FromChannel(in).Thru(
			sendScan(s, ctx, options), // send dynamodb scan
			dropIfNil(),               // if result is nil, do not pass it along
		)
	)

	in <- input // pass in input
	close(in)   // close the input channel
	return p    // return the pipeline
}

type ScanMapFunction[In any] func(In) *dynamodb.ScanInput

func mapToScanInput[In any](mapfn ScanMapFunction[In]) piper.Pipe {
	return pipeline.Map(func(input In) *dynamodb.ScanInput {
		return mapfn(input)
	})
}

func Scan[In any](s Scanner, ctx context.Context, mapfn ScanMapFunction[In], opts ...func(*Options)) piper.Pipe {
	options := newClientOptions().apply(opts)
	return piper.Join(
		mapToScanInput(mapfn),
		sendScan(s, ctx, options),
	)
}
