package awss3

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// putObjectSink implements a pipeline Sink that uploads data to S3.
// It buffers outputs and errors for retrieval after the upload completes.
type putObjectSink struct {
	// client is the S3 client used for upload operations
	client Client
	// in is the channel receiving data to be uploaded
	in chan any
	// out collects the responses from successful S3 put operations
	out []*s3.PutObjectOutput
	// err collects any errors that occur during upload operations
	err []error
	// done is closed when the upload is complete
	done chan struct{}
}

// ToObject creates a pipeline Sink that uploads data to S3.
// Each input value is marshaled to bytes and uploaded as a separate object.
func (c Client) ToObject(ctx context.Context, opts ...func(*s3.PutObjectInput)) *putObjectSink {
	sink := &putObjectSink{
		client: c,
		in:     make(chan any),
		done:   make(chan struct{}),
	}
	go sink.start(ctx, opts)
	return sink
}

// In implements the piper.Sink interface.
func (s *putObjectSink) In() chan<- any { return s.in }

// Out returns the results of all upload operations and any errors that occurred.
// This method blocks until all uploads are complete.
func (s putObjectSink) Out() ([]*s3.PutObjectOutput, error) {
	<-s.done
	return s.out, errors.Join(s.err...)
}

// start begins processing input data and uploading it to S3.
// This method runs in its own goroutine and coordinates upload operations.
func (s *putObjectSink) start(ctx context.Context, opts []func(*s3.PutObjectInput)) {
	defer close(s.done)

	var (
		bucket, key = s.client.extractPath()
		input       = s3.PutObjectInput{
			Bucket: &bucket,
			Key:    &key,
		}
	)

	for _, opt := range opts {
		opt(&input)
	}

	for {
		select {
		case <-ctx.Done():
			s.err = append(s.err, ctx.Err())
			return
		case v, ok := <-s.in:
			if !ok {
				return
			} else if out, err := s.client.putObject(ctx, &input, v); err != nil {
				s.err = append(s.err, err)
			} else {
				s.out = append(s.out, out)
			}
		}
	}
}
