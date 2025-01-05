# piper

Create data pipelines with Golang channels.

## Overview

Piper is a Go library that enables the creation of composable data pipelines using Go channels. It provides a set of primitives for building concurrent data processing pipelines with operations like map, filter, reduce, fan-out, and fan-in.

## Key Features

- Type-safe pipeline operations using Go generics
- Concurrent processing with Go channels
- Composable pipeline components
- Built-in support for common operations:
  - Map: Transform data
  - Filter: Include/exclude data
  - Reduce: Aggregate data
  - Fan-out: Split data streams
  - Fan-in: Combine data streams

## Examples

### Basic Pipeline with Map Operation
```go
// Create a pipeline that doubles numbers
source := pipeline.FromSlice(1, 2, 3, 4)
double := func(in int) int { return in * 2 }
result := source.Then(pipeline.Map(double))

// Consume the results
for num := range result.Out() {
    fmt.Println(num) // Prints: 2, 4, 6, 8
}
```

### Filtering Data
```go
// Create a pipeline that keeps only even numbers
source := pipeline.FromSlice(1, 2, 3, 4, 5, 6)
isEven := func(n int) bool { return n%2 == 0 }
result := source.Then(pipeline.Filter(isEven))

// Or using the more expressive KeepIf
result := source.Then(pipeline.KeepIf(isEven))
```

### Complex Pipeline with Multiple Operations
```go
source := pipeline.FromSlice(1, 2, 3, 4, 5)
pipeline := source.
    Then(pipeline.Map(func(n int) int { return n * 2 })).
    Then(pipeline.Filter(func(n int) bool { return n > 5 })).
    Then(pipeline.Map(func(n int) string { return fmt.Sprintf("Value: %d", n) }))
```

### Fan-Out Example
```go
source := pipeline.FromSlice(1, 2, 3, 4)
keyFn := func(n int) string {
    if n%2 == 0 {
        return "even"
    }
    return "odd"
}

generators := map[string]pipeline.FanOutPipelineFunction{
    "even": func(s piper.Source) piper.Pipeline { return piper.PipelineFrom(s) },
    "odd":  func(s piper.Source) piper.Pipeline { return piper.PipelineFrom(s) },
}

fanout := pipeline.ToMultiSource(keyFn, generators)

// Process even and odd numbers separately
evenPipeline := fanout.Sources()[0]
oddPipeline := fanout.Sources()[1]
```

### Fan-In Example
```go
// Create two sources
source1 := pipeline.FromSlice(1, 2, 3)
source2 := pipeline.FromSlice(4, 5, 6)

// Combine the sources into a single pipeline
combined := pipeline.FromMultiSource(source1, source2)
```

### Reduce Example
```go
// Sum all numbers in a pipeline
source := pipeline.FromSlice(1, 2, 3, 4, 5)
sum := func(acc, item int) int { return acc + item }
result := source.Then(pipeline.Reduce(sum))

// The final value will be 15
```

## Installation

```bash
go get github.com/yourusername/piper
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
