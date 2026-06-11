package conformance

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	redis "github.com/valkey-io/valkey-go/valkeycompat"
	redismock "github.com/valkey-io/valkey-go/valkeycompatmock"
)

func TestConformance(t *testing.T) {
	ctx := context.Background()

	t.Run("Get", func(t *testing.T) {
		rdb, mock := redismock.NewClientMock()
		defer rdb.Close()

		mock.ExpectGet("k").SetVal("v")
		got, err := rdb.Get(ctx, "k").Result()
		assertResult(t, got, "v", err)
		assertExpectations(t, mock)
	})

	t.Run("RedisNil", func(t *testing.T) {
		rdb, mock := redismock.NewClientMock()
		defer rdb.Close()

		mock.ExpectGet("missing").RedisNil()
		if err := rdb.Get(ctx, "missing").Err(); !errors.Is(err, redis.Nil) {
			t.Fatalf("expected redis.Nil, got %v", err)
		}
		assertExpectations(t, mock)
	})

	t.Run("SetErr", func(t *testing.T) {
		rdb, mock := redismock.NewClientMock()
		defer rdb.Close()

		mock.ExpectSet("k", "v", 0).SetErr(errors.New("boom"))
		if err := rdb.Set(ctx, "k", "v", 0).Err(); err == nil || err.Error() != "boom" {
			t.Fatalf("expected boom, got %v", err)
		}
		assertExpectations(t, mock)
	})

	t.Run("SetNXWithTTL", func(t *testing.T) {
		rdb, mock := redismock.NewClientMock()
		defer rdb.Close()

		mock.ExpectSetNX("k", "v", time.Second).SetVal(true)
		got, err := rdb.SetNX(ctx, "k", "v", time.Second).Result()
		assertResult(t, got, true, err)
		assertExpectations(t, mock)
	})

	t.Run("MSetMap", func(t *testing.T) {
		rdb, mock := redismock.NewClientMock()
		defer rdb.Close()

		mock.ExpectMSet(map[string]any{"a": "1", "b": 2}).SetVal("OK")
		if err := rdb.MSet(ctx, map[string]any{"b": 2, "a": "1"}).Err(); err != nil {
			t.Fatalf("unexpected err %v", err)
		}
		assertExpectations(t, mock)
	})

	t.Run("HSetMap", func(t *testing.T) {
		rdb, mock := redismock.NewClientMock()
		defer rdb.Close()

		mock.ExpectHSet("h", map[string]any{"a": "1", "b": 2}).SetVal(2)
		got, err := rdb.HSet(ctx, "h", map[string]any{"b": 2, "a": "1"}).Result()
		assertResult(t, got, int64(2), err)
		assertExpectations(t, mock)
	})

	t.Run("Pipeline", func(t *testing.T) {
		rdb, mock := redismock.NewClientMock()
		defer rdb.Close()

		mock.ExpectGet("k1").SetVal("v1")
		mock.ExpectSet("k2", "v2", 0).SetVal("OK")
		pipe := rdb.Pipeline()
		get := pipe.Get(ctx, "k1")
		set := pipe.Set(ctx, "k2", "v2", 0)
		if _, err := pipe.Exec(ctx); err != nil {
			t.Fatalf("unexpected err %v", err)
		}
		assertResult(t, get.Val(), "v1", get.Err())
		assertResult(t, set.Val(), "OK", set.Err())
		assertExpectations(t, mock)
	})

	t.Run("ZAdd", func(t *testing.T) {
		rdb, mock := redismock.NewClientMock()
		defer rdb.Close()

		member := redis.Z{Score: 1.5, Member: "m"}
		mock.ExpectZAdd("z", member).SetVal(1)
		got, err := rdb.ZAdd(ctx, "z", member).Result()
		assertResult(t, got, int64(1), err)
		assertExpectations(t, mock)
	})

	t.Run("ScanSetVal", func(t *testing.T) {
		rdb, mock := redismock.NewClientMock()
		defer rdb.Close()

		keys := []string{"k1", "k2"}
		mock.ExpectScan(0, "k*", 10).SetVal(keys, 5)
		gotKeys, gotCursor, err := rdb.Scan(ctx, 0, "k*", 10).Result()
		assertResult(t, gotKeys, keys, err)
		if gotCursor != 5 {
			t.Fatalf("cursor got %d, want 5", gotCursor)
		}
		assertExpectations(t, mock)
	})

	t.Run("ACLDryRun", func(t *testing.T) {
		rdb, mock := redismock.NewClientMock()
		defer rdb.Close()

		mock.ExpectACLDryRun("default", "get", "k").SetVal("OK")
		got, err := rdb.ACLDryRun(ctx, "default", "get", "k").Result()
		assertResult(t, got, "OK", err)
		assertExpectations(t, mock)
	})

	t.Run("XReadStreams", func(t *testing.T) {
		rdb, mock := redismock.NewClientMock()
		defer rdb.Close()

		streams := []redis.XStream{{
			Stream: "s",
			Messages: []redis.XMessage{{
				ID:     "1-0",
				Values: map[string]any{"f": "v"},
			}},
		}}
		mock.ExpectXReadStreams("s", "0").SetVal(streams)
		got, err := rdb.XReadStreams(ctx, "s", "0").Result()
		assertResult(t, got, streams, err)
		assertExpectations(t, mock)
	})
}

type expectationVerifier interface {
	ExpectationsWereMet() error
}

func assertExpectations(t *testing.T, mock expectationVerifier) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected expectations err %v", err)
	}
}

func assertResult(t *testing.T, got, want any, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected err %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}
