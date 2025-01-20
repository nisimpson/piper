package pipeline_test

import (
	"fmt"
	"strings"
	"time"

	"github.com/nisimpson/piper/pipeline"
)

func ExampleFlatMap() {
	// Create a pipeline that splits words into characters
	pipe := pipeline.FlatMap(func(word string) []string {
		chars := make([]string, len(word))
		for i, c := range word {
			chars[i] = string(c)
		}
		return chars
	})

	// Send input
	in := pipe.In()
	go func() {
		in <- "hello"
		close(in)
	}()

	// Receive output
	for char := range pipe.Out() {
		fmt.Println(char)
	}

	// Output:
	// h
	// e
	// l
	// l
	// o
}

// ExampleBatch demonstrates custom batch configuration
func ExampleBatch() {
	// Create a batcher that groups up to 2 items with a 100ms timeout
	pipe := pipeline.Batch[int](func(bo *pipeline.BatcherOptions) {
		bo.MaxSize = 2
		bo.Interval = 100 * time.Millisecond
	})

	// Send input
	in := pipe.In()
	go func() {
		in <- 1
		in <- 2
		in <- 3
		close(in)
	}()

	// Receive batches
	for batch := range pipe.Out() {
		fmt.Printf("%v\n", batch.([]int))
	}
	// Output:
	// [1 2]
	// [3]
}

// ExampleBatchN demonstrates fixed-size batching
func ExampleBatchN() {
	// Create a batcher that groups exactly 3 items
	pipe := pipeline.BatchN[string](3)

	// Send input
	in := pipe.In()
	go func() {
		in <- "a"
		in <- "b"
		in <- "c"
		in <- "d"
		in <- "e"
		in <- "f"
		close(in)
	}()

	// Receive batches
	for batch := range pipe.Out() {
		fmt.Printf("%v\n", batch.([]string))
	}
	// Output:
	// [a b c]
	// [d e f]
}

// ExampleBatchEvery demonstrates time-based batching
func ExampleBatchEvery() {
	// Create a batcher that sends items every 100ms
	pipe := pipeline.BatchEvery[int](100 * time.Millisecond)

	// Send input with delays
	in := pipe.In()
	go func() {
		in <- 1
		in <- 2
		time.Sleep(150 * time.Millisecond)
		in <- 3
		in <- 4
		close(in)
	}()

	// Receive batches
	for batch := range pipe.Out() {
		fmt.Printf("%v\n", batch.([]int))
	}
	// Output:
	// [1 2]
	// [3 4]
}

// ExampleBatch_empty demonstrates behavior with no items
func ExampleBatch_empty() {
	pipe := pipeline.BatchN[int](2)

	// Close immediately with no input
	in := pipe.In()
	close(in)

	// Receive batches
	for batch := range pipe.Out() {
		fmt.Printf("%v\n", batch.([]int))
	}
	// Output:
}

// ExampleBatch_partial demonstrates handling incomplete batches
func ExampleBatch_partial() {
	pipe := pipeline.BatchN[int](3)

	// Send fewer items than batch size
	in := pipe.In()
	go func() {
		in <- 1
		in <- 2
		close(in)
	}()

	// Receive batches
	for batch := range pipe.Out() {
		fmt.Printf("%v\n", batch.([]int))
	}
	// Output:
	// [1 2]
}

// ExampleSlidingWindow demonstrates basic sliding window functionality
func ExampleSlidingWindow() {
	// Create a sliding window with size 3 and step 1
	pipe := pipeline.SlidingWindow[int](func(o *pipeline.SlidingWindowOptions) {
		o.WindowSize = 3
		o.StepSize = 1
	})

	// Send input
	in := pipe.In()
	go func() {
		for i := 1; i <= 5; i++ {
			in <- i
		}
		close(in)
	}()

	// Receive windows
	for window := range pipe.Out() {
		fmt.Printf("%v\n", window.([]int))
	}
	// Output:
	// [1 2 3]
	// [2 3 4]
	// [3 4 5]
}

// ExampleSlidingWindow_withStep demonstrates sliding with larger step size
func ExampleSlidingWindow_withStep() {
	// Create a sliding window with size 4 and step 2
	pipe := pipeline.SlidingWindow[int](func(o *pipeline.SlidingWindowOptions) {
		o.WindowSize = 4
		o.StepSize = 2
	})

	// Send input
	in := pipe.In()
	go func() {
		for i := 1; i <= 6; i++ {
			in <- i
		}
		close(in)
	}()

	// Receive windows
	for window := range pipe.Out() {
		fmt.Printf("%v\n", window.([]int))
	}
	// Output:
	// [1 2 3 4]
	// [3 4 5 6]
}

// ExampleSlidingWindow_withInterval demonstrates time-based sliding
func ExampleSlidingWindow_withInterval() {
	// Create a sliding window with interval
	pipe := pipeline.SlidingWindow[int](func(o *pipeline.SlidingWindowOptions) {
		o.WindowSize = 2
		o.StepSize = 1
		o.Interval = 100 * time.Millisecond
	})

	// Send input with delays
	in := pipe.In()
	go func() {
		in <- 1
		in <- 2
		time.Sleep(150 * time.Millisecond)
		in <- 3
		close(in)
	}()

	// Receive windows
	for window := range pipe.Out() {
		fmt.Printf("%v\n", window.([]int))
	}
	// Output:
	// [1 2]
	// [2 3]
}

// ExampleFilter demonstrates basic filtering of integers
func ExampleFilter() {
	// Create a filter for even numbers
	pipe := pipeline.Filter(func(n int) bool {
		return n%2 == 0
	})

	// Send input
	in := pipe.In()
	go func() {
		for i := 1; i <= 5; i++ {
			in <- i
		}
		close(in)
	}()

	// Receive filtered items
	for item := range pipe.Out() {
		fmt.Println(item)
	}
	// Output:
	// 2
	// 4
}

// ExampleFilter_strings demonstrates filtering strings
func ExampleFilter_strings() {
	// Create a filter for non-empty strings
	pipe := pipeline.Filter(func(s string) bool {
		return len(strings.TrimSpace(s)) > 0
	})

	// Send input
	in := pipe.In()
	go func() {
		inputs := []string{"hello", "", "  ", "world", "\n"}
		for _, s := range inputs {
			in <- s
		}
		close(in)
	}()

	// Receive filtered items
	for item := range pipe.Out() {
		fmt.Printf("%q\n", item)
	}
	// Output:
	// "hello"
	// "world"
}

// ExampleKeepIf demonstrates using the KeepIf alias
func ExampleKeepIf() {
	// Keep numbers greater than 3
	pipe := pipeline.KeepIf(func(n int) bool {
		return n > 3
	})

	// Send input
	in := pipe.In()
	go func() {
		for i := 1; i <= 5; i++ {
			in <- i
		}
		close(in)
	}()

	// Receive kept items
	for item := range pipe.Out() {
		fmt.Println(item)
	}
	// Output:
	// 4
	// 5
}

// ExampleDropIf demonstrates dropping items that match a condition
func ExampleDropIf() {
	// Drop negative numbers
	pipe := pipeline.DropIf(func(n int) bool {
		return n < 0
	})

	// Send input
	in := pipe.In()
	go func() {
		nums := []int{-2, -1, 0, 1, 2}
		for _, n := range nums {
			in <- n
		}
		close(in)
	}()

	// Receive non-dropped items
	for item := range pipe.Out() {
		fmt.Println(item)
	}
	// Output:
	// 0
	// 1
	// 2
}

// ExampleFilter_struct demonstrates filtering custom structs
func ExampleFilter_struct() {
	type Person struct {
		Name string
		Age  int
	}

	// Filter for adults (age >= 18)
	pipe := pipeline.Filter(func(p Person) bool {
		return p.Age >= 18
	})

	// Send input
	in := pipe.In()
	go func() {
		people := []Person{
			{"Alice", 25},
			{"Bob", 17},
			{"Charlie", 30},
			{"David", 15},
		}
		for _, person := range people {
			in <- person
		}
		close(in)
	}()

	// Receive filtered items
	for item := range pipe.Out() {
		person := item.(Person)
		fmt.Printf("%s: %d\n", person.Name, person.Age)
	}
	// Output:
	// Alice: 25
	// Charlie: 30
}

// ExampleFilter_empty demonstrates behavior with no matching items
func ExampleFilter_empty() {
	// Filter that matches nothing
	pipe := pipeline.Filter(func(n int) bool {
		return false
	})

	// Send input
	in := pipe.In()
	go func() {
		for i := 1; i <= 3; i++ {
			in <- i
		}
		close(in)
	}()

	// Receive filtered items (should be none)
	for item := range pipe.Out() {
		fmt.Println(item)
	}
	// Output:
}

// ExampleFilter_all demonstrates behavior when all items match
func ExampleFilter_all() {
	// Filter that matches everything
	pipe := pipeline.Filter(func(n int) bool {
		return true
	})

	// Send input
	in := pipe.In()
	go func() {
		for i := 1; i <= 3; i++ {
			in <- i
		}
		close(in)
	}()

	// Receive all items
	for item := range pipe.Out() {
		fmt.Println(item)
	}
	// Output:
	// 1
	// 2
	// 3
}

// ExampleJoin demonstrates basic pipe joining
func ExampleJoin() {
	// Create two simple transforming pipes
	doubler := pipeline.Map(func(x int) int {
		return x * 2
	})

	adder := pipeline.Map(func(x int) int {
		return x + 1
	})

	// Join the pipes
	pipe := pipeline.Join(doubler, adder)

	// Send input
	in := pipe.In()
	go func() {
		in <- 5
		close(in)
	}()

	// Receive output
	for item := range pipe.Out() {
		fmt.Println(item)
	}
	// Output:
	// 11
}

// ExampleJoin_multiple demonstrates joining multiple pipes
func ExampleJoin_multiple() {
	// Create three pipes for string manipulation
	upper := pipeline.Map(func(s string) string {
		return strings.ToUpper(s)
	})

	trim := pipeline.Map(func(s string) string {
		return strings.TrimSpace(s)
	})

	exclaim := pipeline.Map(func(s string) string {
		return s + "!"
	})

	// Join all pipes
	pipe := pipeline.Join(upper, trim, exclaim)

	// Send input
	in := pipe.In()
	go func() {
		in <- " hello world "
		close(in)
	}()

	// Receive output
	for item := range pipe.Out() {
		fmt.Println(item)
	}
	// Output:
	// HELLO WORLD!
}

// ExampleMap demonstrates basic integer transformation
func ExampleMap() {
	// Create a mapper that doubles numbers
	pipe := pipeline.Map(func(x int) int {
		return x * 2
	})

	// Send input
	in := pipe.In()
	go func() {
		for i := 1; i <= 3; i++ {
			in <- i
		}
		close(in)
	}()

	// Receive transformed items
	for item := range pipe.Out() {
		fmt.Println(item)
	}
	// Output:
	// 2
	// 4
	// 6
}

// ExampleMap_typeConversion demonstrates converting between types
func ExampleMap_typeConversion() {
	// Create a mapper that converts integers to strings
	pipe := pipeline.Map(func(x int) string {
		return fmt.Sprintf("value: %d", x)
	})

	// Send input
	in := pipe.In()
	go func() {
		for i := 1; i <= 2; i++ {
			in <- i
		}
		close(in)
	}()

	// Receive transformed items
	for item := range pipe.Out() {
		fmt.Println(item)
	}
	// Output:
	// value: 1
	// value: 2
}

// ExampleMap_strings demonstrates string transformation
func ExampleMap_strings() {
	// Create a mapper that uppercases strings
	pipe := pipeline.Map(func(s string) string {
		return strings.ToUpper(s)
	})

	// Send input
	in := pipe.In()
	go func() {
		words := []string{"hello", "world"}
		for _, word := range words {
			in <- word
		}
		close(in)
	}()

	// Receive transformed items
	for item := range pipe.Out() {
		fmt.Println(item)
	}
	// Output:
	// HELLO
	// WORLD
}

// ExampleMap_struct demonstrates transforming between struct types
func ExampleMap_struct() {
	type Person struct {
		Name string
		Age  int
	}

	type Greeting struct {
		Message string
	}

	// Create a mapper that generates greetings
	pipe := pipeline.Map(func(p Person) Greeting {
		return Greeting{
			Message: fmt.Sprintf("Hello, %s! You are %d years old.",
				p.Name,
				p.Age,
			),
		}
	})

	// Send input
	in := pipe.In()

	go func() {
		people := []Person{
			{"Alice", 25},
			{"Bob", 30},
			{"Charlie", 35},
		}
		for _, person := range people {
			in <- person
		}
		close(in)
	}()

	for item := range pipe.Out() {
		greeting := item.(Greeting)
		fmt.Println(greeting.Message)
	}
	// Output:
	// Hello, Alice! You are 25 years old.
	// Hello, Bob! You are 30 years old.
	// Hello, Charlie! You are 35 years old.
}

// ExampleReduce demonstrates basic number reduction
func ExampleReduce() {
	// Create a reducer that sums numbers
	pipe := pipeline.Reduce(func(acc int, item int) int {
		return acc + item
	})

	// Send input
	in := pipe.In()
	go func() {
		for i := 1; i <= 5; i++ {
			in <- i
		}
		close(in)
	}()

	// Receive accumulated values
	for sum := range pipe.Out() {
		fmt.Println(sum)
	}
	// Output:
	// 1
	// 3
	// 6
	// 10
	// 15
}

// ExampleReduce_strings demonstrates string concatenation
func ExampleReduce_strings() {
	// Create a reducer that joins strings
	pipe := pipeline.Reduce(func(acc string, item string) string {
		return acc + " " + item
	})

	// Send input
	in := pipe.In()
	go func() {
		words := []string{"Hello", "world", "from", "pipeline"}
		for _, word := range words {
			in <- word
		}
		close(in)
	}()

	// Receive accumulated values
	for result := range pipe.Out() {
		fmt.Printf("%q\n", result)
	}
	// Output:
	// "Hello"
	// "Hello world"
	// "Hello world from"
	// "Hello world from pipeline"
}

// ExampleReduce_struct demonstrates reducing custom structs
func ExampleReduce_struct() {
	type Counter struct {
		Even int
		Odd  int
		Cur  int
	}

	// Create a reducer that counts even and odd numbers
	pipe := pipeline.Reduce(func(acc Counter, item Counter) Counter {
		if item.Cur%2 == 0 {
			acc.Even++
		} else {
			acc.Odd++
		}
		return acc
	})

	// Send input
	in := pipe.In()
	go func() {
		numbers := []Counter{
			{0, 0, 1},
			{0, 0, 2},
			{0, 0, 3},
			{0, 0, 4},
			{0, 0, 5},
		}
		// Initialize with first item
		in <- Counter{0, 0, 0}
		for _, n := range numbers {
			in <- n
		}
		close(in)
	}()

	// Receive accumulated values
	for counts := range pipe.Out() {
		c := counts.(Counter)
		fmt.Printf("Even: %d, Odd: %d\n", c.Even, c.Odd)
	}
	// Output:
	// Even: 0, Odd: 0
	// Even: 0, Odd: 1
	// Even: 1, Odd: 1
	// Even: 1, Odd: 2
	// Even: 2, Odd: 2
	// Even: 2, Odd: 3
}

// ExampleReduce_empty demonstrates behavior with no input
func ExampleReduce_empty() {
	pipe := pipeline.Reduce(func(acc int, item int) int {
		return acc + item
	})

	// Close immediately with no input
	in := pipe.In()
	close(in)

	// Receive accumulated values (should be none)
	for sum := range pipe.Out() {
		fmt.Println(sum)
	}
	// Output:
}

// ExampleFromChannel demonstrates basic integer channel pipeline conversion
func ExampleFromChannel() {
	// Create a source channel of integers
	source := make(chan int)

	// Create pipeline from the source channel
	pipe := pipeline.FromChannel(source)

	// Send values to source channel
	go func() {
		for i := 1; i <= 3; i++ {
			source <- i
		}
		close(source)
	}()

	// Read from pipeline
	for val := range pipe.Out() {
		fmt.Println(val)
	}
	// Output:
	// 1
	// 2
	// 3
}

// ExampleFromChannel_strings demonstrates string channel conversion
func ExampleFromChannel_strings() {
	// Create a source channel of strings
	source := make(chan string)

	// Create pipeline from the source channel
	pipe := pipeline.FromChannel(source)

	// Send values to source channel
	go func() {
		messages := []string{"Hello", "Pipeline", "World"}
		for _, msg := range messages {
			source <- msg
		}
		close(source)
	}()

	// Read from pipeline
	for val := range pipe.Out() {
		fmt.Printf("%s\n", val)
	}
	// Output:
	// Hello
	// Pipeline
	// World
}

// ExampleFromChannel_struct demonstrates custom struct channel conversion
func ExampleFromChannel_struct() {
	type User struct {
		ID   int
		Name string
	}

	// Create a source channel of User structs
	source := make(chan User)

	// Create pipeline from the source channel
	pipe := pipeline.FromChannel(source)

	// Send values to source channel
	go func() {
		users := []User{
			{1, "Alice"},
			{2, "Bob"},
			{3, "Charlie"},
		}
		for _, user := range users {
			source <- user
		}
		close(source)
	}()

	// Read from pipeline
	for val := range pipe.Out() {
		user := val.(User)
		fmt.Printf("User %d: %s\n", user.ID, user.Name)
	}
	// Output:
	// User 1: Alice
	// User 2: Bob
	// User 3: Charlie
}

// ExampleFromChannel_empty demonstrates behavior with empty channel
func ExampleFromChannel_empty() {
	source := make(chan int)
	pipe := pipeline.FromChannel(source)

	// Close immediately with no values
	close(source)

	// Read from pipeline (should be empty)
	for range pipe.Out() {
		fmt.Println("This won't execute")
	}
	// Output:
}

// ExampleFromChannel_composition demonstrates composing with other pipeline operations
func ExampleFromChannel_composition() {
	source := make(chan int)

	// Create pipeline that doubles numbers from the source channel
	pipe := pipeline.
		FromChannel(source).
		Thru(pipeline.Map(func(x any) any {
			return x.(int) * 2
		}))

	// Send values to source channel
	go func() {
		for i := 1; i <= 3; i++ {
			source <- i
		}
		close(source)
	}()

	// Read transformed values from pipeline
	for val := range pipe.Out() {
		fmt.Println(val)
	}
	// Output:
	// 2
	// 4
	// 6
}

// ExampleFromCmd demonstrates executing a command once without input
func ExampleFromCmd() {
	// Create a command that returns current timestamp
	timeCmd := pipeline.CommandFunc(func(input string) (string, int, error) {
		return time.Now().Format(time.RFC3339), 0, nil
	})

	// Create pipeline from command
	flow := pipeline.FromCmd(timeCmd)

	// Read the single output
	for output := range flow.Out() {
		fmt.Println("Timestamp:", output)
	}
	// Output example:
	// Timestamp: 2024-01-01T12:00:00Z
}

// ExampleFromCmd_withError demonstrates error handling
func ExampleFromCmd_withError() {
	// Create a command that always fails
	failCmd := pipeline.CommandFunc(func(input string) (string, int, error) {
		return "", 1, fmt.Errorf("command failed")
	})

	// Create error handler
	var lastError error
	handleError := func(err error) {
		lastError = err
	}

	// Create pipeline with error handling
	flow := pipeline.FromCmd(failCmd, func(opts *pipeline.CommandPipeOptions[string]) {
		opts.HandleError = handleError
	})

	// Read output (will be empty due to error)
	for range flow.Out() {
	}

	fmt.Println("Error:", lastError)
	// Output:
	// Error: command failed
}

// ExampleTakeN demonstrates basic usage of TakeN to limit output
func ExampleTakeN() {
	// Create a pipeline that takes only the first 3 items
	pipe := pipeline.TakeN(3)

	// Send input
	in := pipe.In()
	go func() {
		for i := 1; i <= 5; i++ {
			in <- i
		}
		close(in)
	}()

	// Receive limited output
	for val := range pipe.Out() {
		fmt.Println(val)
	}
	// Output:
	// 1
	// 2
	// 3
}

// ExampleTakeN_zero demonstrates behavior when taking zero items
func ExampleTakeN_zero() {
	// Create a pipeline that takes no items
	pipe := pipeline.TakeN(0)

	// Send input
	in := pipe.In()
	go func() {
		for i := 1; i <= 3; i++ {
			in <- i
		}
		close(in)
	}()

	// Receive output (should be empty)
	for val := range pipe.Out() {
		fmt.Println(val)
	}
	// Output:
}

// ExampleTakeN_negative demonstrates behavior with negative count
func ExampleTakeN_negative() {
	// Create a pipeline that takes all items (negative count)
	pipe := pipeline.TakeN(-1)

	// Send input
	in := pipe.In()
	go func() {
		for i := 1; i <= 3; i++ {
			in <- i
		}
		close(in)
	}()

	// Receive all output
	for val := range pipe.Out() {
		fmt.Println(val)
	}
	// Output:
	// 1
	// 2
	// 3
}

// ExampleTakeN_fewerItems demonstrates behavior when input has fewer items than count
func ExampleTakeN_fewerItems() {
	// Create a pipeline that takes up to 5 items
	pipe := pipeline.TakeN(5)

	// Send fewer items than the limit
	in := pipe.In()
	go func() {
		for i := 1; i <= 3; i++ {
			in <- i
		}
		close(in)
	}()

	// Receive all available items
	for val := range pipe.Out() {
		fmt.Println(val)
	}
	// Output:
	// 1
	// 2
	// 3
}

// ExampleTakeN_composition demonstrates composing TakeN with other pipeline operations
func ExampleTakeN_composition() {
	// Create a source of numbers
	source := pipeline.FromSlice(1, 2, 3, 4, 5)

	// Create a pipeline that doubles numbers and takes first 3 results
	result := source.Thru(
		pipeline.Map(func(x any) any {
			return x.(int) * 2
		}),
		pipeline.TakeN(3),
	)

	// Receive limited output
	for val := range result.Out() {
		fmt.Println(val)
	}
	// Output:
	// 2
	// 4
	// 6
}

// ExampleTakeN_strings demonstrates TakeN with string values
func ExampleTakeN_strings() {
	pipe := pipeline.TakeN(2)

	// Send string input
	in := pipe.In()
	go func() {
		words := []string{"apple", "banana", "cherry", "date"}
		for _, word := range words {
			in <- word
		}
		close(in)
	}()

	// Receive limited output
	for val := range pipe.Out() {
		fmt.Printf("%s\n", val)
	}
	// Output:
	// apple
	// banana
}

// ExampleDropN demonstrates basic usage of DropN to skip initial items
func ExampleDropN() {
	// Create a pipeline that drops the first 3 items
	pipe := pipeline.DropN(3)

	// Send input
	in := pipe.In()
	go func() {
		for i := 1; i <= 5; i++ {
			in <- i
		}
		close(in)
	}()

	// Receive remaining output
	for val := range pipe.Out() {
		fmt.Println(val)
	}
	// Output:
	// 4
	// 5
}

// ExampleDropN_zero demonstrates behavior when dropping zero items
func ExampleDropN_zero() {
	// Create a pipeline that drops no items (passthrough)
	pipe := pipeline.DropN(0)

	// Send input
	in := pipe.In()
	go func() {
		for i := 1; i <= 3; i++ {
			in <- i
		}
		close(in)
	}()

	// Receive all output
	for val := range pipe.Out() {
		fmt.Println(val)
	}
	// Output:
	// 1
	// 2
	// 3
}

// ExampleDropN_negative demonstrates behavior with negative count
func ExampleDropN_negative() {
	// Create a pipeline that drops no items (negative count)
	pipe := pipeline.DropN(-1)

	// Send input
	in := pipe.In()
	go func() {
		for i := 1; i <= 3; i++ {
			in <- i
		}
		close(in)
	}()

	// Receive all output
	for val := range pipe.Out() {
		fmt.Println(val)
	}
	// Output:
	// 1
	// 2
	// 3
}

// ExampleDropN_moreItems demonstrates behavior when dropping more items than available
func ExampleDropN_moreItems() {
	// Create a pipeline that drops more items than input has
	pipe := pipeline.DropN(5)

	// Send fewer items than the drop count
	in := pipe.In()
	go func() {
		for i := 1; i <= 3; i++ {
			in <- i
		}
		close(in)
	}()

	// Receive output (should be empty)
	for val := range pipe.Out() {
		fmt.Println(val)
	}
	// Output:
}

// ExampleDropN_strings demonstrates DropN with string values
func ExampleDropN_strings() {
	pipe := pipeline.DropN(2)

	// Send string input
	in := pipe.In()
	go func() {
		words := []string{"apple", "banana", "cherry", "date"}
		for _, word := range words {
			in <- word
		}
		close(in)
	}()

	// Receive remaining output
	for val := range pipe.Out() {
		fmt.Printf("%s\n", val)
	}
	// Output:
	// cherry
	// date
}

// ExampleDropN_composition demonstrates composing DropN with other pipeline operations
func ExampleDropN_composition() {
	// Create a source of numbers
	source := pipeline.FromSlice(1, 2, 3, 4, 5)

	// Create a pipeline that doubles numbers and drops first 2 results
	result := source.
		Thru(pipeline.Map(func(x any) any {
			return x.(int) * 2
		})).
		Thru(pipeline.DropN(2))

	// Receive remaining output
	for val := range result.Out() {
		fmt.Println(val)
	}
	// Output:
	// 6
	// 8
	// 10
}

// ExampleDropN_empty demonstrates behavior with empty input
func ExampleDropN_empty() {
	pipe := pipeline.DropN(3)

	// Send no input
	in := pipe.In()
	close(in)

	// Receive output (should be empty)
	for val := range pipe.Out() {
		fmt.Println(val)
	}
	// Output:
}

// ExampleDropN_mixedTypes demonstrates DropN with mixed value types
func ExampleDropN_mixedTypes() {
	pipe := pipeline.DropN(2)

	// Send mixed type input
	in := pipe.In()
	go func() {
		items := []any{1, "two", 3.0, true, "five"}
		for _, item := range items {
			in <- item
		}
		close(in)
	}()

	// Receive remaining output
	for val := range pipe.Out() {
		fmt.Printf("%v\n", val)
	}
	// Output:
	// 3
	// true
	// five
}

// ExampleChunk demonstrates splitting a slice into smaller chunks
func ExampleChunk() {
	// Create a pipeline that chunks input into groups of 3
	pipe := pipeline.Chunk[[]int](3)

	// Send input
	in := pipe.In()
	go func() {
		// Send a slice of numbers
		numbers := []int{1, 2, 3, 4, 5, 6, 7, 8}
		in <- numbers
		close(in)
	}()

	// Receive and print the chunked output
	for chunk := range pipe.Out() {
		fmt.Printf("%v\n", chunk.([]int))
	}
	// Output:
	// [1 2 3]
	// [4 5 6]
	// [7 8]
}

// ExampleChunk_empty demonstrates behavior with empty input
func ExampleChunk_empty() {
	pipe := pipeline.Chunk[[]int](2)

	// Send empty slice
	in := pipe.In()
	go func() {
		numbers := []int{}
		in <- numbers
		close(in)
	}()

	// Receive chunks
	for chunk := range pipe.Out() {
		fmt.Printf("%v\n", chunk.([]int))
	}
	// Output:
}

// ExampleChunk_strings demonstrates chunking with string slices
func ExampleChunk_strings() {
	pipe := pipeline.Chunk[[]string](2)

	in := pipe.In()
	go func() {
		words := []string{"hello", "world", "how", "are", "you"}
		in <- words
		close(in)
	}()

	for chunk := range pipe.Out() {
		fmt.Printf("%v\n", chunk.([]string))
	}
	// Output:
	// [hello world]
	// [how are]
	// [you]
}

// ExampleUnique demonstrates basic usage of the Unique pipe with integers
func ExampleUnique_basic() {
	// Create a source pipeline with duplicate numbers
	source := pipeline.FromSlice(1, 2, 2, 3, 3, 3, 4)

	// Create a unique pipe and connect it to the source
	result := source.Thru(pipeline.Unique[int]())

	// Collect and print unique values
	for val := range result.Out() {
		fmt.Println(val)
	}

	// Output:
	// 1
	// 2
	// 3
	// 4
}

// ExampleUnique_strings demonstrates using Unique with string values
func ExampleUnique_strings() {
	// Create a source with duplicate strings
	source := pipeline.FromSlice("apple", "banana", "apple", "cherry", "banana")

	// Create a unique pipe for strings
	result := source.Thru(pipeline.Unique[string]())

	// Collect and print unique values
	for val := range result.Out() {
		fmt.Println(val)
	}

	// Output:
	// apple
	// banana
	// cherry
}

// ExampleUnique_customKey demonstrates using Unique with a custom key function
func ExampleUnique_customKey() {
	// Define a simple struct
	type Person struct {
		ID   int
		Name string
		Age  int
	}

	// Create sample data with duplicates (same ID)
	people := []Person{
		{ID: 1, Name: "Alice", Age: 25},
		{ID: 2, Name: "Bob", Age: 30},
		{ID: 1, Name: "Alice", Age: 26}, // Duplicate ID
		{ID: 3, Name: "Charlie", Age: 35},
	}

	// Create a unique pipe with custom key function that uses ID as the unique key
	uniqueByID := func(opts *pipeline.UniqueOptions[Person]) {
		opts.KeyFunc = func(p Person) string {
			return fmt.Sprintf("%d", p.ID)
		}
	}

	// Create the pipeline
	source := pipeline.FromSlice(people...)
	result := source.Thru(pipeline.Unique[Person](uniqueByID))

	// Collect and print unique people
	for val := range result.Out() {
		person := val.(Person)
		fmt.Printf("ID: %d, Name: %s, Age: %d\n", person.ID, person.Name, person.Age)
	}

	// Output:
	// ID: 1, Name: Alice, Age: 25
	// ID: 2, Name: Bob, Age: 30
	// ID: 3, Name: Charlie, Age: 35
}

// ExampleUnique_compositeKey demonstrates using multiple fields as a unique key
func ExampleUnique_compositeKey() {
	type Event struct {
		Date     string
		Category string
		Value    int
	}

	events := []Event{
		{"2024-01-01", "A", 1},
		{"2024-01-01", "A", 2}, // Duplicate date+category
		{"2024-01-01", "B", 3},
		{"2024-01-02", "A", 4},
	}

	// Create unique pipe with composite key function using date and category
	uniqueByDateCategory := func(opts *pipeline.UniqueOptions[Event]) {
		opts.KeyFunc = func(e Event) string {
			return fmt.Sprintf("%s-%s", e.Date, e.Category)
		}
	}

	// Create the pipeline
	source := pipeline.FromSlice(events...)
	result := source.Thru(pipeline.Unique[Event](uniqueByDateCategory))

	// Collect and print unique events
	for val := range result.Out() {
		event := val.(Event)
		fmt.Printf("Date: %s, Category: %s, Value: %d\n",
			event.Date, event.Category, event.Value)
	}

	// Output:
	// Date: 2024-01-01, Category: A, Value: 1
	// Date: 2024-01-01, Category: B, Value: 3
	// Date: 2024-01-02, Category: A, Value: 4
}
