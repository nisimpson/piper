package awss3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/nisimpson/piper/pipeline"
)

var _ types.Object

// listObjectsSource implements a pipeline Flow that streams S3 objects from a directory listing.
type listObjectsSource struct {
	Client
	out chan any
}

// FromDirectory creates a [pipeline.Flow] that streams objects from an S3 directory listing.
// The Flow will send individual S3 [types.Object] metadata downstream from the list of results.
func (s Client) FromDirectory(ctx context.Context, opts ...func(*s3.ListObjectsV2Input)) pipeline.Flow {
	source := listObjectsSource{
		Client: s,
		out:    make(chan any),
	}

	go source.start(ctx, opts)
	return pipeline.From(source)
}

// Out returns an output channel for connection to a downstream sink or pipe.
func (l listObjectsSource) Out() <-chan any { return l.out }

// start executes the list objects operation on the target S3 bucket.
func (l listObjectsSource) start(ctx context.Context, opts []func(*s3.ListObjectsV2Input)) {
	defer close(l.out)

	var (
		bucket, prefix = l.extractPath()
	)

	l.listObjects(ctx, bucket, prefix, opts, func(out *s3.ListObjectsV2Output) (next bool) {
		for _, obj := range out.Contents {
			select {
			case <-ctx.Done():
				l.HandleError(ctx.Err())
				return false
			case l.out <- obj:
			}
		}
		if out.IsTruncated == nil {
			out.IsTruncated = aws.Bool(false)
		}
		return *out.IsTruncated
	})
}

// getObjectSource implements a pipeline Flow that streams the contents of a single S3 object.
type getObjectSource struct {
	Client
	out chan any
}

// FromObject creates a [pipeline.Flow] that streams the contents of a single S3 object.
// The flow will emit the object's contents as bytes. Zero-length content or request errors
// will be dropped from the stream.
func (s Client) FromObject(ctx context.Context, opts ...func(*s3.GetObjectInput)) pipeline.Flow {
	source := getObjectSource{
		Client: s,
		out:    make(chan any),
	}

	go source.start(ctx, opts)
	return pipeline.From(source)
}

// Out returns an output channel for connection to a downstream sink or pipe.
func (g getObjectSource) Out() <-chan any { return g.out }

// start performs the retrieval of the object from the target S3 bucket.
func (g getObjectSource) start(ctx context.Context, opts []func(*s3.GetObjectInput)) {
	defer close(g.out)

	var (
		bucket, path = g.extractPath()
		data         = g.getObject(ctx, bucket, path, opts)
	)

	// if we didn't get any data, we're done
	if len(data) == 0 {
		return
	}

	// send the data downstream
	g.out <- data
}
