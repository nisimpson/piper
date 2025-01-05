package pipeline

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"piper"
	"piper/internal/must"
)

// HttpBodyMarshalFunction represents a function that can serialize pipeline data into bytes for [http.Request] bodies.
type HttpBodyMarshalFunction = func(any) ([]byte, error)

// HttpPipeOptions configures how HTTP requests are made and how their responses are processed.
type HttpPipeOptions struct {
	// Client is the [http.Client] used to make requests
	Client *http.Client
	// Request is the base [http.Request] to be sent (headers, etc. can be configured here).
	Request *http.Request
	// HandleError is called when an HTTP request fails.
	HandleError func(error)
	// HandleResponse processes the HTTP response and converts it to an item to be sent downstream.
	HandleResponse func(*http.Response) (any, error)
	// MarshalFunc converts pipeline items to bytes for the request body.
	MarshalFunc HttpBodyMarshalFunction
}

// httpPipe implements a pipeline component that makes HTTP requests.
// It can be used either as a source (FromHTTP) or as a processing step (SendHTTP).
type httpPipe struct {
	// url is the target URL for HTTP requests
	url string
	// method is the HTTP method to use (GET, POST, etc.)
	method string
	// options configure how requests are made and responses are handled
	options []func(*HttpPipeOptions)
	// in receives items to be sent as request bodies
	in chan any
	// out sends processed responses
	out chan any
}

// FromHTTP creates a new [piper.Pipeline] that starts by making an HTTP request.
// The provided body is used for the initial request, and the response payload becomes the first pipeline item.
// Provide [HttpPipeOptions] to configure the default behavior.
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

// SendHTTP creates a [piper.Pipe] component that sends each upstream item as an HTTP request.
// By default, the response payloads from these requests become the output items in the pipeline.
// Provide [HttpPipeOptions] to configure the default behavior.
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

func (h httpPipe) In() chan<- any  { return h.in }
func (h httpPipe) Out() <-chan any { return h.out }

// apply configures the HttpPipeOptions by running all the provided option functions.
// This is an internal helper method used during pipe initialization.
func (o *HttpPipeOptions) apply(opts ...func(*HttpPipeOptions)) {
	for _, opt := range opts {
		opt(o)
	}
}

// start begins processing HTTP requests/responses.
// For source pipes (FromHTTP) it makes a single request; for processing pipes (SendHTTP)
// it makes a request for each input item.
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

// passResponse is the default handling behavior. It extracts the response payload and sends
// it downstream as a string.
func (httpPipe) passResponse(res *http.Response) (any, error) {
	data, err := io.ReadAll(res.Body)
	return string(data), err
}
