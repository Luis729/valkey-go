package conformance

import (
	"context"
	"testing"

	"github.com/go-redis/redismock/v9"
)

func TestExpectedSetValBehavior(t *testing.T) {
	ctx := context.Background()

	t.Run("CmdAny", func(t *testing.T) {
		rdb, mock := redismock.NewClientMock()
		defer rdb.Close()

		mock.ExpectDo("PING").SetVal(int64(7))
		got, err := rdb.Do(ctx, "PING").Result()
		assertResult(t, got, int64(7), err)
		assertExpectations(t, mock)
	})

	t.Run("ScriptExists", func(t *testing.T) {
		rdb, mock := redismock.NewClientMock()
		defer rdb.Close()

		want := []bool{true, false}
		mock.ExpectScriptExists("a", "b").SetVal(want)
		got, err := rdb.ScriptExists(ctx, "a", "b").Result()
		assertResult(t, got, want, err)
		assertExpectations(t, mock)
	})
}
