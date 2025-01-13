/*
Package pipeline provides a comprehensive set of building blocks for constructing data processing pipelines.

At its core, the pipeline package offers reusable components that can be composed to create
complex data transformation workflows. These components include:

  - Basic transformation operations (Map, Filter, Reduce)
  - Channel-based data streaming
  - HTTP request/response handling
  - Command execution pipelines
  - Data multiplexing and demultiplexing

Key features:

  - Generic type support for type-safe data processing
  - Concurrent execution using Go channels
  - Extensible architecture supporting custom pipeline components
  - Built-in error handling and propagation
  - Support for both synchronous and asynchronous processing

Pipeline construction follows a fluent builder pattern:
 1. Start with the [From] constructor to create a new [Flow].
 2. Add processing steps with [Flow.Thru] to transform data.
 3. Complete with [Flow.To] to specify where results should be delivered.
*/
package pipeline
