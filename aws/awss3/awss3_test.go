package awss3

import (
	"context"
	"net/url"
	"testing"
	"time"

	"bytes"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		s3URL   string
		options []func(*Client)
		wantErr bool
	}{
		{
			name:    "valid s3 URL",
			s3URL:   "s3://bucket/key",
			wantErr: false,
		},
		{
			name:    "invalid URL",
			s3URL:   "://invalid",
			wantErr: true,
		},
		{
			name:  "custom options",
			s3URL: "s3://bucket/key",
			options: []func(*Client){
				func(c *Client) { c.AutoPaginateWhenTruncated = false },
				func(c *Client) { c.HandleError = func(err error) { panic(err) } },
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &s3.Client{}

			if tt.wantErr {
				assert.Panics(t, func() {
					_ = New(tt.s3URL, mockClient, tt.options...)
				})
				return
			}

			var client Client
			assert.NotPanics(t, func() {
				client = New(tt.s3URL, mockClient, tt.options...)
			})

			parsedURL, _ := url.Parse(tt.s3URL)
			assert.Equal(t, parsedURL, client.s3URL)
			assert.NotNil(t, client.lister)
			assert.NotNil(t, client.getter)
			assert.NotNil(t, client.putter)

			if tt.options != nil {
				assert.False(t, client.AutoPaginateWhenTruncated)
			} else {
				assert.True(t, client.AutoPaginateWhenTruncated)
			}
		})
	}
}

func TestToObject(t *testing.T) {
	tests := []struct {
		name         string
		input        []byte
		putResponse  *s3.PutObjectOutput
		putError     error
		expectError  bool
		expectCancel bool
	}{
		{
			name:        "successful put",
			input:       []byte("test content"),
			putResponse: &s3.PutObjectOutput{},
			expectError: false,
		},
		{
			name:        "error case",
			input:       []byte("test content"),
			putError:    fmt.Errorf("put error"),
			expectError: true,
		},
		{
			name:         "context canceled",
			input:        []byte("test content"),
			expectError:  true,
			expectCancel: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			defer cancel()

			mockPutter := mockObjectPutter{
				putObjectFunc: func(_ context.Context, _ *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
					return tt.putResponse, tt.putError
				},
			}

			client := New("s3://test-bucket/test-key", &s3.Client{})
			client.putter = mockPutter
			sink := client.ToObject(ctx, func(poi *s3.PutObjectInput) {})
			done := make(chan struct{})

			go func() {
				if tt.expectCancel {
					cancel()
				} else {
					sink.In() <- tt.input
					close(sink.In())
				}
				time.Sleep(time.Millisecond)
				done <- struct{}{}
			}()

			// Wait for results
			<-done
			outputs, err := sink.Out()

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, outputs)
			assert.Len(t, outputs, 1)
		})
	}
}

func TestExtractPath(t *testing.T) {
	type testcase struct {
		name       string
		uri        string
		wantBucket string
		wantKey    string
	}

	for _, tc := range []testcase{
		{
			name:       "bucket only",
			uri:        "s3://bucket-name",
			wantBucket: "bucket-name",
			wantKey:    "",
		},
		{
			name:       "bucket with slash",
			uri:        "s3://bucket-name/",
			wantBucket: "bucket-name",
			wantKey:    "",
		},
		{
			name:       "bucket and key",
			uri:        "s3://bucket-name/and/a/path",
			wantBucket: "bucket-name",
			wantKey:    "and/a/path",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := New(tc.uri, &s3.Client{})
			bucket, key := c.extractPath()
			if bucket != tc.wantBucket {
				t.Fatalf("Expected %s, got %s", tc.wantBucket, bucket)
			}

			if key != tc.wantKey {
				t.Fatalf("Expected %s, got %s", tc.wantKey, key)
			}
		})
	}
}

func TestGetObject(t *testing.T) {
	tests := []struct {
		name            string
		getResponse     *s3.GetObjectOutput
		getError        error
		expectedContent []byte
		expectError     bool
	}{
		{
			name: "successful get",
			getResponse: &s3.GetObjectOutput{
				Body: io.NopCloser(bytes.NewReader([]byte("test content"))),
			},
			expectedContent: []byte("test content"),
			expectError:     false,
		},
		{
			name:        "error case",
			getError:    fmt.Errorf("get error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockGetter := mockObjectGetter{
				getObjectFunc: func(_ context.Context, _ *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
					return tt.getResponse, tt.getError
				},
			}

			client := New("s3://test-bucket", &s3.Client{})
			client.getter = mockGetter

			pipe := client.GetObject(ctx, func(goi *s3.GetObjectInput) {})

			go func() {
				input := "test-key"
				pipe.In() <- input
				close(pipe.In())
			}()

			content := <-pipe.Out()

			if tt.expectError {
				assert.Nil(t, content)
				return
			}

			assert.Equal(t, tt.expectedContent, content)
		})
	}
}

func TestFromObject(t *testing.T) {
	tests := []struct {
		name            string
		getResponse     *s3.GetObjectOutput
		getError        error
		expectedContent []byte
		expectError     bool
	}{
		{
			name: "successful get",
			getResponse: &s3.GetObjectOutput{
				Body: io.NopCloser(bytes.NewReader([]byte("test content"))),
			},
			expectedContent: []byte("test content"),
			expectError:     false,
		},
		{
			name:            "error case",
			getError:        fmt.Errorf("get error"),
			expectedContent: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockGetter := mockObjectGetter{
				getObjectFunc: func(_ context.Context, _ *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
					return tt.getResponse, tt.getError
				},
			}

			client := New("s3://test-bucket", &s3.Client{})
			client.getter = mockGetter

			source := client.FromObject(ctx)

			var content []byte
			var err error

			done := make(chan bool)
			go func() {
				for data := range source.Out() {
					content = data.([]byte)
				}
				done <- true
			}()

			<-done

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedContent, content)
		})
	}
}

func TestFromDirectory(t *testing.T) {
	tests := []struct {
		name         string
		listResponse *s3.ListObjectsV2Output
		listError    error
		expectedKeys []string
		expectError  bool
		cancelCtx    bool
	}{
		{
			name: "successful listing",
			listResponse: &s3.ListObjectsV2Output{
				Contents: []types.Object{
					{Key: aws.String("file1.txt")},
					{Key: aws.String("file2.txt")},
				},
			},
			expectedKeys: []string{"file1.txt", "file2.txt"},
			expectError:  false,
		},
		{
			name: "empty listing",
			listResponse: &s3.ListObjectsV2Output{
				Contents: []types.Object{},
			},
			expectedKeys: []string{},
			expectError:  false,
		},
		{
			name:        "error case",
			listError:   fmt.Errorf("listing error"),
			expectError: true,
		},
		{
			name: "context canceled",
			listResponse: &s3.ListObjectsV2Output{
				Contents: []types.Object{
					{Key: aws.String("file1.txt")},
					{Key: aws.String("file2.txt")},
				},
			},
			expectError: true,
			cancelCtx:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			if tt.cancelCtx {
				cancel()
			}

			defer cancel()

			mockLister := mockObjectLister{
				listObjectsFunc: func(_ context.Context, _ *s3.ListObjectsV2Input, _ ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
					return tt.listResponse, tt.listError
				},
			}

			client := New("s3://test-bucket", &s3.Client{})
			client.lister = mockLister

			client.HandleError = func(err error) {
				if tt.expectError {
					assert.Error(t, err)
					return
				}
				assert.NoError(t, err)
			}

			source := client.FromDirectory(ctx, func(lovi *s3.ListObjectsV2Input) {})

			var keys []string

			done := make(chan bool)
			go func() {
				for key := range source.Out() {
					keys = append(keys, *key.(types.Object).Key)
				}
				done <- true
			}()

			<-done

			if tt.expectError {
				return
			}
			assert.ElementsMatch(t, tt.expectedKeys, keys)
		})
	}
}
