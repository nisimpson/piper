package awss3

import (
	"bytes"
	"context"
	"io"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nisimpson/piper/internal/must"
)

// ObjectListerV2 provides an interface for listing objects in an S3 bucket using the V2 API.
// This interface is implemented by the AWS SDK's S3 client.
type ObjectListerV2 interface {
	ListObjectsV2(context.Context, *s3.ListObjectsV2Input, ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}

// ObjectGetter provides an interface for retrieving objects from an S3 bucket.
// This interface is implemented by the AWS SDK's S3 client.
type ObjectGetter interface {
	GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

// ObjectPutter provides an interface for uploading objects to an S3 bucket.
// This interface is implemented by the AWS SDK's S3 client.
type ObjectPutter interface {
	PutObject(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

// Client provides high-level operations for interacting with Amazon S3.
// It wraps the AWS SDK S3 client and provides additional functionality for
// listing, getting, and putting objects.
type Client struct {
	// s3URL is the parsed URL pointing to the S3 bucket and optional key prefix.
	s3URL *url.URL
	// lister implements the ObjectListerV2 interface for listing S3 objects.
	lister ObjectListerV2
	// getter implements the ObjectGetter interface for retrieving S3 objects.
	getter ObjectGetter
	// putter implements the ObjectPutter interface for uploading S3 objects.
	putter ObjectPutter
	// HandleError is called when an error occurs during S3 operations.
	HandleError func(err error)
	// MarshalFunc converts objects to bytes before uploading to S3.
	MarshalFunc func(any) ([]byte, error)
	// AutoPaginateWhenTruncated determines if list operations should automatically handle pagination.
	AutoPaginateWhenTruncated bool
}

// New creates a new S3 Client with the specified S3 URL and AWS SDK S3 client.
// Additional options can be provided to customize the client's behavior.
func New(s3URL string, s3client *s3.Client, opts ...func(*Client)) Client {
	c := Client{
		s3URL:                     must.Return(url.Parse(s3URL)),
		lister:                    s3client,
		getter:                    s3client,
		putter:                    s3client,
		HandleError:               func(err error) {},
		MarshalFunc:               func(a any) ([]byte, error) { return a.([]byte), nil },
		AutoPaginateWhenTruncated: true,
	}

	for _, opt := range opts {
		opt(&c)
	}

	return c
}

// extractPath splits the s3URL into bucket and key components.
func (s Client) extractPath() (bucket, key string) {
	bucket = s.s3URL.Host
	if s.s3URL.Path != "" {
		key = s.s3URL.Path[1:]
	}
	return
}

// getObject retrieves an object from S3 and returns its contents as bytes.
func (c Client) getObject(ctx context.Context, bucket, key string, opts []func(*s3.GetObjectInput)) []byte {
	input := &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}

	for _, opt := range opts {
		opt(input)
	}

	out, err := c.getter.GetObject(ctx, input)
	if err != nil {
		c.HandleError(err)
		return []byte{}
	}

	defer out.Body.Close()
	data := must.Return(io.ReadAll(out.Body))
	return data
}

// listObjects lists objects in an S3 bucket with the given prefix and handles pagination.
func (c Client) listObjects(
	ctx context.Context,
	bucket, prefix string,
	opts []func(*s3.ListObjectsV2Input),
	h func(*s3.ListObjectsV2Output) (next bool),
) {
	input := &s3.ListObjectsV2Input{
		Bucket: &bucket,
		Prefix: &prefix,
	}

	for _, opt := range opts {
		opt(input)
	}

	for {
		out, err := c.lister.ListObjectsV2(ctx, input)
		if err != nil {
			c.HandleError(err)
			return
		}

		next := h(out)
		input.ContinuationToken = out.NextContinuationToken
		if !next || !c.AutoPaginateWhenTruncated {
			break
		}
	}
}

// putObject handles the upload of a single object to S3.
// It marshals the input value to bytes and performs the upload operation.
func (c Client) putObject(ctx context.Context, input *s3.PutObjectInput, v any) (*s3.PutObjectOutput, error) {
	data, err := c.MarshalFunc(v)
	if err != nil {
		c.HandleError(err)
		return nil, err
	}

	body := bytes.NewBuffer(data)
	input.Body = body

	out, err := c.putter.PutObject(ctx, input)
	if err != nil {
		c.HandleError(err)
		return nil, err
	}
	return out, nil
}
