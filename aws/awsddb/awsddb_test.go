package awsddb_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/nisimpson/piper/aws/awsddb"
	"github.com/nisimpson/piper/pipeline"
)

type MockDynamoDB struct {
	outerr error
	dynamodb.PutItemInput
	dynamodb.PutItemOutput
	dynamodb.GetItemOutput
	dynamodb.GetItemInput
	dynamodb.ScanInput
	dynamodb.ScanOutput
	dynamodb.QueryInput
	dynamodb.QueryOutput
	dynamodb.UpdateItemInput
	dynamodb.UpdateItemOutput
	dynamodb.DeleteItemInput
	dynamodb.DeleteItemOutput
}

func (m MockDynamoDB) PutItem(_ context.Context, in *dynamodb.PutItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	return &m.PutItemOutput, m.outerr
}

func (m MockDynamoDB) GetItem(_ context.Context, in *dynamodb.GetItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	return &m.GetItemOutput, m.outerr
}

func (m MockDynamoDB) Scan(_ context.Context, in *dynamodb.ScanInput, opts ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
	return &m.ScanOutput, m.outerr
}

func (m MockDynamoDB) Query(_ context.Context, in *dynamodb.QueryInput, opts ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	return &m.QueryOutput, m.outerr
}

func (m MockDynamoDB) UpdateItem(_ context.Context, in *dynamodb.UpdateItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	return &m.UpdateItemOutput, m.outerr
}

func (m MockDynamoDB) DeleteItem(_ context.Context, in *dynamodb.DeleteItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	return &m.DeleteItemOutput, m.outerr
}

func (m MockDynamoDB) mapPut(string) *dynamodb.PutItemInput { return &m.PutItemInput }
func (m MockDynamoDB) mapGet(string) *dynamodb.GetItemInput { return &m.GetItemInput }
func (m MockDynamoDB) mapScan(string) *dynamodb.ScanInput   { return &m.ScanInput }
func (m MockDynamoDB) mapQuery(string) *dynamodb.QueryInput { return &m.QueryInput }
func (m MockDynamoDB) mapUpdate(string) *dynamodb.UpdateItemInput {
	return &m.UpdateItemInput
}
func (m MockDynamoDB) mapDelete(string) *dynamodb.DeleteItemInput {
	return &m.DeleteItemInput
}

func TestScan(t *testing.T) {
	t.Run("scan source", func(t *testing.T) {
		var (
			ddb    = MockDynamoDB{}
			source = awsddb.FromScan(ddb, context.TODO(), &dynamodb.ScanInput{})
			sink   = pipeline.ToSlice[*dynamodb.ScanOutput]()
			want   = []*dynamodb.ScanOutput{{}}
		)

		source.To(sink)
		got := sink.Slice()

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("scans object", func(t *testing.T) {
		var (
			ddb    = MockDynamoDB{}
			source = pipeline.FromSlice("foo")
			pipe   = awsddb.Scan(
				ddb,
				context.TODO(),
				ddb.mapScan,
			)
			sink = pipeline.ToSlice[*dynamodb.ScanOutput]()
			want = []*dynamodb.ScanOutput{{}}
		)

		source.Thru(pipe).To(sink)
		got := sink.Slice()

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("scans object fails", func(t *testing.T) {
		var (
			ddb    = MockDynamoDB{outerr: errors.New("failed")}
			source = pipeline.FromSlice("foo")
			pipe   = awsddb.Scan(
				ddb,
				context.TODO(),
				ddb.mapScan,
			)
			sink = pipeline.ToSlice[*dynamodb.ScanOutput]()
		)

		source.Thru(pipe).To(sink)
		got := sink.Slice()

		if len(got) != 0 {
			t.Errorf("got %v, want %v", got, nil)
		}
	})
}

func TestQuery(t *testing.T) {
	t.Run("query source", func(t *testing.T) {
		var (
			ddb    = MockDynamoDB{}
			source = awsddb.FromQuery(ddb, context.TODO(), &dynamodb.QueryInput{})
			sink   = pipeline.ToSlice[*dynamodb.QueryOutput]()
			want   = []*dynamodb.QueryOutput{{}}
		)

		source.To(sink)
		got := sink.Slice()

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("queries object", func(t *testing.T) {
		var (
			ddb    = MockDynamoDB{}
			source = pipeline.FromSlice("foo")
			pipe   = awsddb.Query(
				ddb,
				context.TODO(),
				ddb.mapQuery,
			)
			sink = pipeline.ToSlice[*dynamodb.QueryOutput]()
			want = []*dynamodb.QueryOutput{{}}
		)

		source.Thru(pipe).To(sink)
		got := sink.Slice()

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("queries object fails", func(t *testing.T) {
		var (
			ddb    = MockDynamoDB{outerr: errors.New("failed")}
			source = pipeline.FromSlice("foo")
			pipe   = awsddb.Query(
				ddb,
				context.TODO(),
				ddb.mapQuery,
			)
			sink = pipeline.ToSlice[*dynamodb.QueryOutput]()
		)

		source.Thru(pipe).To(sink)
		got := sink.Slice()

		if len(got) != 0 {
			t.Errorf("got %v, want %v", got, nil)
		}
	})
}

func TestDelete(t *testing.T) {
	t.Run("deletes object", func(t *testing.T) {
		var (
			ddb    = MockDynamoDB{}
			source = pipeline.FromSlice("foo")
			pipe   = awsddb.Delete(
				ddb,
				context.TODO(),
				ddb.mapDelete,
			)
			sink = pipeline.ToSlice[*dynamodb.DeleteItemOutput]()
			want = []*dynamodb.DeleteItemOutput{{}}
		)

		source.Thru(pipe).To(sink)
		got := sink.Slice()

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("deletes object fails", func(t *testing.T) {
		var (
			ddb    = MockDynamoDB{outerr: errors.New("failed")}
			source = pipeline.FromSlice("foo")
			pipe   = awsddb.Delete(
				ddb,
				context.TODO(),
				ddb.mapDelete,
			)
			sink = pipeline.ToSlice[*dynamodb.DeleteItemOutput]()
		)

		source.Thru(pipe).To(sink)
		got := sink.Slice()

		if len(got) != 0 {
			t.Errorf("got %v, want %v", got, nil)
		}
	})
}

func TestUpdate(t *testing.T) {
	t.Run("updates source", func(t *testing.T) {
		var (
			ddb    = MockDynamoDB{}
			source = awsddb.FromUpdate(ddb, context.TODO(), &dynamodb.UpdateItemInput{})
			sink   = pipeline.ToSlice[*dynamodb.UpdateItemOutput]()
			want   = []*dynamodb.UpdateItemOutput{{}}
		)

		source.To(sink)
		got := sink.Slice()

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("updates object", func(t *testing.T) {
		var (
			ddb    = MockDynamoDB{}
			source = pipeline.FromSlice("foo")
			pipe   = awsddb.Update(
				ddb,
				context.TODO(),
				ddb.mapUpdate,
			)
			sink = pipeline.ToSlice[*dynamodb.UpdateItemOutput]()
			want = []*dynamodb.UpdateItemOutput{{}}
		)

		source.Thru(pipe).To(sink)
		got := sink.Slice()

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("updates object fails", func(t *testing.T) {
		var (
			ddb    = MockDynamoDB{outerr: errors.New("failed")}
			source = pipeline.FromSlice("foo")
			pipe   = awsddb.Update(
				ddb,
				context.TODO(),
				ddb.mapUpdate,
			)
			sink = pipeline.ToSlice[*dynamodb.UpdateItemOutput]()
		)

		source.Thru(pipe).To(sink)
		got := sink.Slice()

		if len(got) != 0 {
			t.Errorf("got %v, want %v", got, nil)
		}
	})
}

func TestGet(t *testing.T) {
	t.Run("gets source", func(t *testing.T) {
		var (
			ddb    = MockDynamoDB{}
			source = awsddb.FromGet(ddb, context.TODO(), &dynamodb.GetItemInput{})
			sink   = pipeline.ToSlice[*dynamodb.GetItemOutput]()
			want   = []*dynamodb.GetItemOutput{{}}
		)

		source.To(sink)
		got := sink.Slice()

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("gets object", func(t *testing.T) {
		var (
			ddb    = MockDynamoDB{}
			source = pipeline.FromSlice("foo")
			pipe   = awsddb.Get(
				ddb,
				context.TODO(),
				ddb.mapGet,
			)
			sink = pipeline.ToSlice[*dynamodb.GetItemOutput]()
			want = []*dynamodb.GetItemOutput{{}}
		)

		source.Thru(pipe).To(sink)
		got := sink.Slice()

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("gets object fails", func(t *testing.T) {
		var (
			ddb    = MockDynamoDB{outerr: errors.New("failed")}
			source = pipeline.FromSlice("foo")
			pipe   = awsddb.Get(
				ddb,
				context.TODO(),
				ddb.mapGet,
			)
			sink = pipeline.ToSlice[*dynamodb.GetItemOutput]()
		)

		source.Thru(pipe).To(sink)
		got := sink.Slice()

		if len(got) != 0 {
			t.Errorf("got %v, want %v", got, nil)
		}
	})
}

func TestPut(t *testing.T) {
	t.Run("puts object", func(t *testing.T) {
		var (
			ddb    = MockDynamoDB{}
			source = pipeline.FromSlice("foo")
			pipe   = awsddb.Put(
				ddb,
				context.TODO(),
				ddb.mapPut,
				func(o *awsddb.Options) {},
			)
			sink = pipeline.ToSlice[*dynamodb.PutItemOutput]()
			want = []*dynamodb.PutItemOutput{{}}
		)

		source.Thru(pipe).To(sink)
		got := sink.Slice()

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("puts object fails", func(t *testing.T) {
		var (
			ddb    = MockDynamoDB{outerr: errors.New("failed")}
			source = pipeline.FromSlice("foo")
			pipe   = awsddb.Put(
				ddb,
				context.TODO(),
				ddb.mapPut,
			)
			sink = pipeline.ToSlice[*dynamodb.PutItemOutput]()
		)

		source.Thru(pipe).To(sink)
		got := sink.Slice()

		if len(got) != 0 {
			t.Errorf("got %v, want %v", got, nil)
		}
	})
}
