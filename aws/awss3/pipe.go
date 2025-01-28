package awss3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nisimpson/piper"
	"github.com/nisimpson/piper/pipeline"
)

// GetObject creates a pipeline that transforms input strings into S3 object content.
// Each input string is treated as an object key, and the corresponding object's content
// is retrieved from S3. Empty objects are dropped from the pipeline.
func (c Client) GetObject(ctx context.Context, opts ...func(*s3.GetObjectInput)) piper.Pipe {
	// use the upstream input as the key to an object within an s3 bucket
	mapper := pipeline.Map(func(key string) []byte {
		bucket, _ := c.extractPath()
		return c.getObject(ctx, bucket, key, opts)
	})

	// do not pass the result along if the length of the body is empty
	dropper := pipeline.DropIf(func(b []byte) bool {
		return len(b) == 0
	})

	return pipeline.Join(
		mapper,
		dropper,
	)
}
