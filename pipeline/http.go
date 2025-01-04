package pipeline

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"piper"
	"piper/internal/must"
)

type HttpBodyMarshalFunction = func(any) ([]byte, error)

type HttpPipeOptions struct {
	Client         *http.Client
	Request        *http.Request
	HandleError    func(error)
	HandleResponse func(*http.Response) (any, error)
	MarshalFunc    HttpBodyMarshalFunction
}

type httpPipe struct {
	url     string
	method  string
	options []func(*HttpPipeOptions)
	in      chan any
	out     chan any
}

func FromHTTP(method string, url string, body io.Reader, opts ...func(*HttpPipeOptions)) piper.Pipeline {
	source := httpPipe{
		url:     url,
		method:  method,
		options: opts,
		in:      make(chan any, 1),
		out:     make(chan any),
	}

	if body == nil {
		body = bytes.NewBufferString("")
	}

	data := must.Return(io.ReadAll(body))
	source.in <- data
	close(source.in)

	go source.start()
	return piper.PipelineFrom(source)
}

func SendHTTP(method string, url string, opts ...func(*HttpPipeOptions)) piper.Pipe {
	pipe := httpPipe{
		url:     url,
		method:  method,
		options: opts,
		in:      make(chan any),
		out:     make(chan any),
	}

	go pipe.start()
	return pipe
}

func (h httpPipe) In() chan<- any {
	return h.in
}

func (h httpPipe) Out() <-chan any {
	return h.out
}

func (o *HttpPipeOptions) apply(opts ...func(*HttpPipeOptions)) {
	for _, opt := range opts {
		opt(o)
	}
}

func (h httpPipe) start() {
	opts := HttpPipeOptions{
		Request:        must.Return(http.NewRequest(h.method, h.url, nil)),
		Client:         &http.Client{},
		HandleError:    must.IgnoreError,
		HandleResponse: h.passResponse,
		MarshalFunc:    json.Marshal,
	}

	defer close(h.out)
	opts.apply(h.options...)

	for input := range h.in {
		switch item := input.(type) {
		case []byte:
			opts.Request.Body = io.NopCloser(bytes.NewBuffer(item))
		default:
			data := must.Return(opts.MarshalFunc(item))
			opts.Request.Body = io.NopCloser(bytes.NewBuffer(data))
		}

		res, err := opts.Client.Do(opts.Request)
		if err != nil {
			opts.HandleError(err)
			continue
		}
		output, err := opts.HandleResponse(res)
		if err != nil {
			opts.HandleError(err)
			continue
		}
		must.PanicOnError(res.Body.Close())
		h.out <- output
	}
}

func (httpPipe) passResponse(res *http.Response) (any, error) {
	data, err := io.ReadAll(res.Body)
	return string(data), err
}
