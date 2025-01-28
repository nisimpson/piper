package awss3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Mock implementations for AWS S3 interfaces
type mockS3Client struct {
	ObjectListerV2
	ObjectGetter
	ObjectPutter
}

type mockObjectLister struct {
	listObjectsFunc func(context.Context, *s3.ListObjectsV2Input, ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}

func (m mockObjectLister) ListObjectsV2(ctx context.Context, input *s3.ListObjectsV2Input, opts ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	return m.listObjectsFunc(ctx, input, opts...)
}

type mockObjectGetter struct {
	getObjectFunc func(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

func (m mockObjectGetter) GetObject(ctx context.Context, input *s3.GetObjectInput, opts ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return m.getObjectFunc(ctx, input, opts...)
}

type mockObjectPutter struct {
	putObjectFunc func(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

func (m mockObjectPutter) PutObject(ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return m.putObjectFunc(ctx, input, opts...)
}