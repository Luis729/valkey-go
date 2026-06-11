package conformance

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
)

func TestBroaderRuntimeBehavior(t *testing.T) {
	ctx := context.Background()

	t.Run("ScanFamily", func(t *testing.T) {
		rdb, mock := redismock.NewClientMock()
		defer rdb.Close()

		mock.ExpectScanType(0, "k*", 10, "string").SetVal([]string{"k1", "k2"}, 11)
		keys, cursor, err := rdb.ScanType(ctx, 0, "k*", 10, "string").Result()
		assertResult(t, keys, []string{"k1", "k2"}, err)
		assertResult(t, cursor, uint64(11), nil)

		mock.ExpectSScan("s", 0, "m*", 5).SetVal([]string{"m1"}, 0)
		keys, cursor, err = rdb.SScan(ctx, "s", 0, "m*", 5).Result()
		assertResult(t, keys, []string{"m1"}, err)
		assertResult(t, cursor, uint64(0), nil)

		mock.ExpectHScan("h", 0, "f*", 5).SetVal([]string{"f1", "v1"}, 3)
		keys, cursor, err = rdb.HScan(ctx, "h", 0, "f*", 5).Result()
		assertResult(t, keys, []string{"f1", "v1"}, err)
		assertResult(t, cursor, uint64(3), nil)

		mock.ExpectZScan("z", 0, "m*", 5).SetVal([]string{"m1", "1.5"}, 0)
		keys, cursor, err = rdb.ZScan(ctx, "z", 0, "m*", 5).Result()
		assertResult(t, keys, []string{"m1", "1.5"}, err)
		assertResult(t, cursor, uint64(0), nil)

		assertExpectations(t, mock)
	})

	t.Run("EvalAndFunctionCalls", func(t *testing.T) {
		rdb, mock := redismock.NewClientMock()
		defer rdb.Close()

		mock.ExpectEval("return {KEYS[1],ARGV[1]}", []string{"k"}, "a").SetVal([]any{"k", "a"})
		got, err := rdb.Eval(ctx, "return {KEYS[1],ARGV[1]}", []string{"k"}, "a").Result()
		assertResult(t, got, []any{"k", "a"}, err)

		mock.ExpectEvalSha("sha1", []string{"k"}, "a").SetVal("sha-ok")
		got, err = rdb.EvalSha(ctx, "sha1", []string{"k"}, "a").Result()
		assertResult(t, got, "sha-ok", err)

		mock.ExpectFCall("fn", []string{"k1"}, "a", "b").SetVal(int64(2))
		got, err = rdb.FCall(ctx, "fn", []string{"k1"}, "a", "b").Result()
		assertResult(t, got, int64(2), err)

		mock.ExpectFCallRo("readonly", []string{"k1"}, "a").SetVal([]any{"v"})
		got, err = rdb.FCallRo(ctx, "readonly", []string{"k1"}, "a").Result()
		assertResult(t, got, []any{"v"}, err)

		mock.ExpectEval("return nil", []string{}).RedisNil()
		if err := rdb.Eval(ctx, "return nil", []string{}).Err(); !errors.Is(err, redis.Nil) {
			t.Fatalf("expected redis.Nil, got %v", err)
		}

		mock.ExpectFCall("missing", []string{}).SetErr(errors.New("boom"))
		if err := rdb.FCall(ctx, "missing", []string{}).Err(); err == nil || err.Error() != "boom" {
			t.Fatalf("expected boom, got %v", err)
		}

		assertExpectations(t, mock)
	})

	t.Run("ServerAndCommandMetadata", func(t *testing.T) {
		rdb, mock := redismock.NewClientMock()
		defer rdb.Close()

		now := time.Unix(10, 0)
		mock.ExpectTime().SetVal(now)
		gotTime, err := rdb.Time(ctx).Result()
		assertResult(t, gotTime, now, err)

		slowLogs := []redis.SlowLog{{
			ID:         7,
			Time:       time.Unix(20, 0),
			Duration:   2 * time.Millisecond,
			Args:       []string{"GET", "k"},
			ClientAddr: "127.0.0.1:6379",
			ClientName: "client",
		}}
		mock.ExpectSlowLogGet(1).SetVal(slowLogs)
		gotSlowLogs, err := rdb.SlowLogGet(ctx, 1).Result()
		assertResult(t, gotSlowLogs, slowLogs, err)

		commandInfos := []*redis.CommandInfo{{
			Name:        "get",
			Arity:       2,
			Flags:       []string{"readonly", "fast"},
			FirstKeyPos: 1,
			LastKeyPos:  1,
			StepCount:   1,
			ReadOnly:    true,
		}}
		mock.ExpectCommand().SetVal(commandInfos)
		gotCommands, err := rdb.Command(ctx).Result()
		assertResult(t, gotCommands, map[string]*redis.CommandInfo{"get": commandInfos[0]}, err)

		filter := &redis.FilterBy{Module: "bf", ACLCat: "read", Pattern: "cms.*"}
		mock.ExpectCommandList(filter).SetVal([]string{"cms.info"})
		gotCommandList, err := rdb.CommandList(ctx, filter).Result()
		assertResult(t, gotCommandList, []string{"cms.info"}, err)

		mock.ExpectCommandGetKeys("EVAL", "return redis.call('GET', KEYS[1])", 1, "k").SetVal([]string{"k"})
		gotKeys, err := rdb.CommandGetKeys(ctx, "EVAL", "return redis.call('GET', KEYS[1])", 1, "k").Result()
		assertResult(t, gotKeys, []string{"k"}, err)

		keyFlags := []redis.KeyFlags{{Key: "k", Flags: []string{"RO", "access"}}}
		mock.ExpectCommandGetKeysAndFlags("GET", "k").SetVal(keyFlags)
		gotKeyFlags, err := rdb.CommandGetKeysAndFlags(ctx, "GET", "k").Result()
		assertResult(t, gotKeyFlags, keyFlags, err)

		assertExpectations(t, mock)
	})

	t.Run("SMIsMember", func(t *testing.T) {
		rdb, mock := redismock.NewClientMock()
		defer rdb.Close()

		want := []bool{true, false, true}
		mock.ExpectSMIsMember("s", "a", "b", "c").SetVal(want)
		got, err := rdb.SMIsMember(ctx, "s", "a", "b", "c").Result()
		assertResult(t, got, want, err)
		assertExpectations(t, mock)
	})

	t.Run("Geo", func(t *testing.T) {
		rdb, mock := redismock.NewClientMock()
		defer rdb.Close()

		location := &redis.GeoLocation{Name: "Palermo", Longitude: 13.361389, Latitude: 38.115556}
		mock.ExpectGeoAdd("geo", location).SetVal(1)
		added, err := rdb.GeoAdd(ctx, "geo", location).Result()
		assertResult(t, added, int64(1), err)

		positions := []*redis.GeoPos{{Longitude: 13.361389, Latitude: 38.115556}}
		mock.ExpectGeoPos("geo", "Palermo").SetVal(positions)
		gotPositions, err := rdb.GeoPos(ctx, "geo", "Palermo").Result()
		assertResult(t, gotPositions, positions, err)

		mock.ExpectGeoDist("geo", "Palermo", "Catania", "km").SetVal(166.2742)
		dist, err := rdb.GeoDist(ctx, "geo", "Palermo", "Catania", "km").Result()
		assertResult(t, dist, 166.2742, err)

		mock.ExpectGeoHash("geo", "Palermo").SetVal([]string{"sqc8b49rny0"})
		hashes, err := rdb.GeoHash(ctx, "geo", "Palermo").Result()
		assertResult(t, hashes, []string{"sqc8b49rny0"}, err)

		assertExpectations(t, mock)
	})

	t.Run("SortedSets", func(t *testing.T) {
		rdb, mock := redismock.NewClientMock()
		defer rdb.Close()

		zs := []redis.Z{{Member: "a", Score: 1.5}, {Member: "b", Score: 2.5}}
		mock.ExpectZRange("z", 0, -1).SetVal([]string{"a", "b"})
		members, err := rdb.ZRange(ctx, "z", 0, -1).Result()
		assertResult(t, members, []string{"a", "b"}, err)

		mock.ExpectZRangeWithScores("z", 0, -1).SetVal(zs)
		gotZS, err := rdb.ZRangeWithScores(ctx, "z", 0, -1).Result()
		assertResult(t, gotZS, zs, err)

		mock.ExpectZPopMax("z", 2).SetVal(zs)
		popped, err := rdb.ZPopMax(ctx, "z", 2).Result()
		assertResult(t, popped, zs, err)

		mock.ExpectZMScore("z", "a", "b").SetVal([]float64{1.5, 2.5})
		scores, err := rdb.ZMScore(ctx, "z", "a", "b").Result()
		assertResult(t, scores, []float64{1.5, 2.5}, err)

		assertExpectations(t, mock)
	})

	t.Run("Streams", func(t *testing.T) {
		rdb, mock := redismock.NewClientMock()
		defer rdb.Close()

		xAddArgs := &redis.XAddArgs{Stream: "s", ID: "*", Values: map[string]any{"f": "v"}}
		xMsg := redis.XMessage{ID: "1-0", Values: map[string]any{"f": "v"}}
		xStream := redis.XStream{Stream: "s", Messages: []redis.XMessage{xMsg}}

		mock.ExpectXAdd(xAddArgs).SetVal("1-0")
		id, err := rdb.XAdd(ctx, xAddArgs).Result()
		assertResult(t, id, "1-0", err)

		mock.ExpectXRange("s", "-", "+").SetVal([]redis.XMessage{xMsg})
		messages, err := rdb.XRange(ctx, "s", "-", "+").Result()
		assertResult(t, messages, []redis.XMessage{xMsg}, err)

		xReadArgs := &redis.XReadArgs{Streams: []string{"s", "0"}}
		mock.ExpectXRead(xReadArgs).SetVal([]redis.XStream{xStream})
		streams, err := rdb.XRead(ctx, xReadArgs).Result()
		assertResult(t, streams, []redis.XStream{xStream}, err)

		pending := &redis.XPending{Count: 1, Lower: "1-0", Higher: "1-0", Consumers: map[string]int64{"c": 1}}
		mock.ExpectXPending("s", "g").SetVal(pending)
		gotPending, err := rdb.XPending(ctx, "s", "g").Result()
		assertResult(t, gotPending, pending, err)

		xAutoClaimArgs := &redis.XAutoClaimArgs{Stream: "s", Group: "g", Consumer: "c", Start: "0-0"}
		mock.ExpectXAutoClaim(xAutoClaimArgs).SetVal([]redis.XMessage{xMsg}, "1-0")
		claimed, start, err := rdb.XAutoClaim(ctx, xAutoClaimArgs).Result()
		assertResult(t, claimed, []redis.XMessage{xMsg}, err)
		assertResult(t, start, "1-0", nil)

		assertExpectations(t, mock)
	})

	t.Run("SetErrForBroaderCommands", func(t *testing.T) {
		rdb, mock := redismock.NewClientMock()
		defer rdb.Close()

		errBoom := errors.New("boom")

		mock.ExpectCommand().SetErr(errBoom)
		assertErrString(t, rdb.Command(ctx).Err(), "boom")

		mock.ExpectSMIsMember("s", "a").SetErr(errBoom)
		assertErrString(t, rdb.SMIsMember(ctx, "s", "a").Err(), "boom")

		mock.ExpectTime().SetErr(errBoom)
		assertErrString(t, rdb.Time(ctx).Err(), "boom")

		mock.ExpectGeoPos("geo", "m").SetErr(errBoom)
		assertErrString(t, rdb.GeoPos(ctx, "geo", "m").Err(), "boom")

		mock.ExpectXRange("s", "-", "+").SetErr(errBoom)
		assertErrString(t, rdb.XRange(ctx, "s", "-", "+").Err(), "boom")

		mock.ExpectZPopMax("z").SetErr(errBoom)
		assertErrString(t, rdb.ZPopMax(ctx, "z").Err(), "boom")

		assertExpectations(t, mock)
	})
}

func assertErrString(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil || err.Error() != want {
		t.Fatalf("expected %s, got %v", want, err)
	}
}
