package awsdynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nisimpson/piper"
	"github.com/nisimpson/piper/pipeline"
)

// Scanner defines an interface for DynamoDB scan operations.
type Scanner interface {
	Scan(ctx context.Context, input *dynamodb.ScanInput, opts ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
}

// sendScan creates a pipeline transformation that executes the scan operation
// against DynamoDB using the provided Scanner interface. It handles any errors
// through the options error handler and returns the operation output.
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

// FromScan creates a [piper.Pipeline] that processes a single DynamoDB Scan operation.
// It sets up a pipeline that:
//  1. Takes a single ScanInput
//  2. Executes the scan operation
//  3. Filters out any error results
//  4. Sends [dynamodb.ScanOutput] items downstream
func FromScan(s Scanner, ctx context.Context, input *dynamodb.ScanInput, opts ...func(*Options)) piper.Pipeline {
	var (
		options = newClientOptions().apply(opts)
		in      = make(chan any, 1)

		p = pipeline.FromChannel(in).Thru(
			sendScan(s, ctx, options),        // send dynamodb scan
			dropIfNil[dynamodb.ScanOutput](), // if result is nil, do not pass it along
		)
	)

	in <- input // pass in input
	close(in)   // close the input channel
	return p    // return the pipeline
}

// MapScanFunction maps a generic input type
// to a DynamoDB ScanInput. This allows for flexible transformation
// of various input types into DynamoDB query requests.
type MapScanFunction[In any] func(In) *dynamodb.ScanInput

// mapToScanInput creates a pipeline transformation that converts input data
// of type In to a DynamoDB ScanInput using the provided mapping function.
func mapToScanInput[In any](mapfn MapScanFunction[In]) piper.Pipe {
	return pipeline.Map(func(input In) *dynamodb.ScanInput {
		return mapfn(input)
	})
}

// Scan creates a [piper.Pipe] for using upstream items as context for scanning DynamoDB.
// It combines three operations:
//  1. Converting input data to a scan request
//  2. Sending the query request to DynamoDB
//  3. Filtering out any error results
//  4. Sending [dynamodb.ScanOutput] results downstream
func Scan[In any](s Scanner, ctx context.Context, mapfn MapScanFunction[In], opts ...func(*Options)) piper.Pipe {
	options := newClientOptions().apply(opts)
	return pipeline.Join(
		mapToScanInput(mapfn),            // convert input into scan request
		sendScan(s, ctx, options),        // send scan request
		dropIfNil[dynamodb.ScanOutput](), // do not pass nil results along
	)
}
