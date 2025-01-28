/*
Package `awss3` provides functionality for interacting with Amazon S3 (Simple Storage Service) in Go applications.
It offers a simple and efficient way to read from and write to S3 buckets using streaming operations.

	import (
		"context"
		"log"

		"github.com/aws/aws-sdk-go-v2/service/s3"
		"github.com/nisimpson/piper/aws/awss3"
	)

	func main() {
	    // Initialize S3 client (aws-sdk-go-v2)
	    s3Client := s3.New(s3.Options{})

	    // Create awss3 client
	    client := awss3.New("s3://src-bucket/path", s3Client)

	    // Example 1: Read from S3 Object
	    ctx := context.Background()
	    source := client.FromObject(ctx)
	    for data := range source.Out() {
	        // Process data from S3 object
	        log.Printf("Received data: %v\n", data)
	    }

	    // Example 2: Write to S3 Object
	    sink := client.ToObject(ctx)
	    go func() {
	        sink.In() <- []byte("Hello, S3!")
	        close(sink.In())
	    }()
	    outputs, err := sink.Out()
	    if err != nil {
	        log.Fatal(err)
	    }
	    log.Printf("Upload complete: %v\n", outputs)

	    // Example 3: List Objects in Directory
	    listing := client.FromDirectory(ctx)
	    for obj := range listing.Out() {
	        if s3Obj, ok := obj.(*s3.Object); ok {
	            log.Printf("Found object: %s\n", *s3Obj.Key)
	        }
	    }
	}

The package supports customization through functional options:

	// Customizing GetObject operation
	client.GetObject(ctx, func(input *s3.GetObjectInput) {
	    input.Range = aws.String("bytes=0-1024") // First 1KB only
	})

	// Customizing PutObject operation
	client.ToObject(ctx, func(input *s3.PutObjectInput) {
	    input.ContentType = aws.String("application/json")
	})

	// Customizing ListObjects operation
	client.FromDirectory(ctx, func(input *s3.ListObjectsV2Input) {
	    input.Prefix = aws.String("subfolder/")
	})
*/
package awss3
