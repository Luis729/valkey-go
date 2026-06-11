package conformance

import (
	"testing"
	"time"

	redis "github.com/valkey-io/valkey-go/valkeycompat"
	redismock "github.com/valkey-io/valkey-go/valkeycompatmock"
)

func TestExpectedSetValAPISurface(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	defer rdb.Close()
	defer mock.ClearExpect()

	if false {
		mock.ExpectCommand().SetVal([]*redis.CommandInfo{{Name: "get"}})
		mock.ExpectScan(0, "*", 10).SetVal([]string{"k"}, 1)
		mock.ExpectSlowLogGet(1).SetVal([]redis.SlowLog{{ID: 1, Time: time.Unix(1, 0)}})
		mock.ExpectDo("PING").SetVal(int64(1))
		mock.ExpectBZPopMax(time.Second, "z").SetVal(&redis.ZWithKey{Key: "z", Z: redis.Z{Member: "m", Score: 1}})
		mock.ExpectXInfoStream("s").SetVal(&redis.XInfoStream{Length: 1})
		mock.ExpectXInfoStreamFull("s", 1).SetVal(&redis.XInfoStreamFull{Length: 1})
		mock.ExpectXPending("s", "g").SetVal(&redis.XPending{Count: 1})

		var _ *redismock.ExpectedMapStringString = mock.ExpectConfigGet("*")
		var _ *redismock.ExpectedMapStringInt = mock.ExpectPubSubNumSub("ch")
		var _ *redismock.ExpectedGeoSearchLocation = mock.ExpectGeoSearchLocation("g", &redis.GeoSearchLocationQuery{})
		var _ *redismock.ExpectedBoolSlice = mock.ExpectScriptExists("sha")
	}
}
