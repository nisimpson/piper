package throttle

import (
	"context"

	"github.com/nisimpson/piper"
	"golang.org/x/time/rate"
)

type limiter struct {
	limit *rate.Limiter
	in    chan any
	pipe  piper.Pipe
	ctx   context.Context
}

func Limit(limit *rate.Limiter, pipe piper.Pipe) piper.Pipe {
	return LimitWithContext(context.Background(), limit, pipe)
}

func LimitWithContext(ctx context.Context, limit *rate.Limiter, pipe piper.Pipe) piper.Pipe {
	limiter := limiter{
		limit: limit,
		in:    make(chan any),
		pipe:  pipe,
		ctx:   ctx,
	}

	go limiter.start()
	return limiter
}

func (l limiter) In() chan<- any  { return l.in }
func (l limiter) Out() <-chan any { return l.pipe.Out() }

func (l limiter) start() {
	defer close(l.pipe.In())
	for {
		if err := l.limit.Wait(l.ctx); err != nil {
			continue
		}
		if data, ok := <-l.in; ok {
			l.pipe.In() <- data
		} else {
			return
		}
	}
}
