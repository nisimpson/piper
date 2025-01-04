package pipeline

import (
	"piper"
	"time"
)

type BatcherOptions struct {
	MaxSize  int
	Interval time.Duration
}

type batcher[In any] struct {
	in      chan any
	out     chan any
	options BatcherOptions
}

func Batch[In any](opts ...func(*BatcherOptions)) piper.Pipe {
	options := BatcherOptions{
		MaxSize: 1,
	}
	for _, opt := range opts {
		opt(&options)
	}
	pipe := batcher[In]{
		in:      make(chan any),
		out:     make(chan any),
		options: options,
	}

	go pipe.start()
	return pipe
}

func BatchN[In any](size int) piper.Pipe {
	return Batch[In](func(bo *BatcherOptions) {
		bo.MaxSize = size
	})
}

func BatchEvery[In any](d time.Duration) piper.Pipe {
	return Batch[In](func(bo *BatcherOptions) {
		bo.Interval = d
		bo.MaxSize = -1
	})
}

func (b batcher[In]) In() chan<- any  { return b.in }
func (b batcher[In]) Out() <-chan any { return b.out }

func (b batcher[In]) start() {
	batch := b.newSlice()
	defer close(b.out)

	if b.options.Interval == 0 {
		for {
			input, next := <-b.in
			if !next {
				b.send(batch)
				return
			}
			batch = append(batch, input.(In))
			if len(batch) == b.options.MaxSize {
				batch = b.send(batch)
			}
		}
	}

	for {
		select {
		case input, next := <-b.in:
			if !next {
				b.send(batch)
				return
			}
			batch = append(batch, input.(In))
			if len(batch) == b.options.MaxSize {
				batch = b.send(batch)
			}
		case <-time.After(b.options.Interval):
			batch = b.send(batch)
		}
	}
}

func (b batcher[In]) send(batch []In) []In {
	if len(batch) == 0 {
		return batch
	}
	b.out <- batch
	return b.newSlice()
}

func (b batcher[In]) newSlice() []In {
	if b.options.MaxSize <= 0 {
		return make([]In, 0)
	}
	return make([]In, 0, b.options.MaxSize)
}
