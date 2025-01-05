package pipeline_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/nisimpson/piper/internal/must"
	"github.com/nisimpson/piper/pipeline"
)

func TestFromHTTP(t *testing.T) {
	t.Parallel()

	var (
		handlers = map[string]http.Handler{
			"ok": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
			"echo": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// echo request body
				data, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}
				w.WriteHeader(http.StatusOK)
				w.Write(data)
			}),
			"error": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "http error", http.StatusInternalServerError)
			}),
		}
	)

	t.Run("get http with defaults", func(t *testing.T) {
		var (
			server = httptest.NewServer(handlers["ok"])
			source = pipeline.FromHTTP(http.MethodPost, server.URL, nil)
			got    = Consume[*http.Response](source)
		)

		defer server.Close()

		if len(got) != 1 {
			t.Errorf("expected 1 response, got %d", len(got))
			return
		}
	})

	t.Run("post http with defaults", func(t *testing.T) {
		var (
			server = httptest.NewServer(handlers["echo"])
			body   = bytes.NewBufferString("hello, world!")
			source = pipeline.FromHTTP(http.MethodPost, server.URL, body)
			got    = Consume[*http.Response](source)
		)

		defer server.Close()

		if len(got) != 1 {
			t.Errorf("expected 1 response, got %d", len(got))
			return
		}

		var (
			res     = got[0]
			payload = must.Return(io.ReadAll(res.Body))
		)

		if res.StatusCode != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, res.StatusCode)
			return
		}

		if string(payload) != "hello, world!" {
			t.Errorf("expected payload %q, got %q", "hello, world!", string(payload))
			return
		}
	})

	t.Run("handles request error", func(t *testing.T) {
		var (
			handled = []bool{false}
			body    = bytes.NewBufferString("hello, world!")
			source  = pipeline.FromHTTP(http.MethodPost, "", body, // request will fail
				func(hpo *pipeline.HttpPipeOptions) {
					hpo.HandleError = func(error) { handled[0] = true }
				},
			)
			want = []*http.Response{}
			got  = Consume[*http.Response](source)
		)

		if !reflect.DeepEqual(want, got) {
			t.Errorf("wanted %#v, got %#v", want, got)
			return
		}

		if handled[0] == false {
			t.Errorf("error was not handled")
			return
		}
	})

	t.Run("overrides response", func(t *testing.T) {
		var (
			server = httptest.NewServer(handlers["echo"])
			body   = bytes.NewBufferString("hello, world!")
			source = pipeline.FromHTTP(http.MethodPost, server.URL, body,
				func(hpo *pipeline.HttpPipeOptions) {
					hpo.HandleResponse = func(r *http.Response) (any, error) {
						data, err := io.ReadAll(r.Body)
						if err != nil {
							t.Fatal(err)
						}
						return strings.ToUpper(string(data)), nil
					}
				},
			)
			want = []string{"HELLO, WORLD!"}
			got  = Consume[string](source)
		)

		defer server.Close()

		if !reflect.DeepEqual(want, got) {
			t.Errorf("wanted %#v, got %#v", want, got)
		}
	})

	t.Run("drops response error", func(t *testing.T) {
		var (
			handled = []bool{false}
			server  = httptest.NewServer(handlers["error"])
			body    = bytes.NewBufferString("hello, world!")
			source  = pipeline.FromHTTP(http.MethodPost, server.URL, body,
				func(hpo *pipeline.HttpPipeOptions) {
					hpo.HandleError = func(err error) {
						handled[0] = true
					}
					hpo.HandleResponse = func(r *http.Response) (any, error) {
						if r.StatusCode != http.StatusOK {
							return nil, errors.New("response error")
						}
						return "ok", nil
					}
				},
			)
			want = []string{}
			got  = Consume[string](source)
		)

		defer server.Close()

		if !reflect.DeepEqual(want, got) {
			t.Errorf("wanted %#v, got %#v", want, got)
			return
		}

		if handled[0] == false {
			t.Error("error was not handled")
		}
	})
}

func TestSendHTTP(t *testing.T) {
	type payload struct {
		Value int `json:"value"`
	}

	type response struct {
		Value int `json:"value"`
	}

	var (
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			data := must.Return(io.ReadAll(r.Body))

			p := payload{}
			err := json.Unmarshal(data, &p)
			if err != nil {
				t.Fatal(err)
			}

			data = must.Return(json.Marshal(response{Value: p.Value * 2}))
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		})
		server = httptest.NewServer(handler)
		source = pipeline.FromSlice(payload{Value: 2})
		action = pipeline.SendHTTP(http.MethodPost, server.URL,
			func(hpo *pipeline.HttpPipeOptions) {
				hpo.HandleResponse = func(r *http.Response) (any, error) {
					data := must.Return(io.ReadAll(r.Body))
					res := response{}
					err := json.Unmarshal(data, &res)
					return res, err
				}
			},
		)
	)

	defer server.Close()
	source = source.Thru(action)

	var (
		want = []response{{Value: 4}}
		got  = Consume[response](source)
	)

	if !reflect.DeepEqual(want, got) {
		t.Errorf("wanted %#v, got %#v", want, got)
		return
	}
}
