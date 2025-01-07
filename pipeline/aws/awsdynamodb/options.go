package awsdynamodb

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nisimpson/piper"
	"github.com/nisimpson/piper/internal/must"
	"github.com/nisimpson/piper/pipeline"
)

type Options struct {
	HandleError     func(error)
	DynamoDBOptions []func(*dynamodb.Options)
}

func newClientOptions() *Options {
	return &Options{
		HandleError: must.IgnoreError,
	}
}

func (o *Options) apply(opts []func(*Options)) *Options {
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func dropIfNil() piper.Pipe {
	return pipeline.DropIf(func(in any) bool {
		return in == nil
	})
}
