package pipeline

import "github.com/nisimpson/piper"

// Chunk creates a [piper.Pipe] that splits an input slice into smaller
// batches of the specified size. Chunk panics if the size is
// less than 1.
//
// Example:
//
//	// Create a pipeline that chunks input into groups of 3
//	pipe := pipeline.Chunk[[]int](3)
//
//	in := pipe.In()
//	go func() {
//		numbers := []int{1, 2, 3, 4, 5, 6, 7, 8}
//		in <- numbers
//		close(in)
//	}()
//
//	for chunk := range pipe.Out() {
//		// Results in: [[1, 2, 3], [4, 5, 6], [7]]
//	}
func Chunk[In []T, T any](size int) piper.Pipe {
	if size < 1 {
		panic("chunk size must be greater than 0")
	}
	return Join(
		Flatten[In](),
		BatchN[T](size),
	)
}
