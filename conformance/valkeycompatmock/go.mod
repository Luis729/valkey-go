module github.com/valkey-io/valkey-go/conformance/valkeycompatmock

go 1.25.0

replace github.com/valkey-io/valkey-go => ../..

replace github.com/valkey-io/valkey-go/mock => ../../mock

replace github.com/valkey-io/valkey-go/valkeycompat => ../../valkeycompat

replace github.com/valkey-io/valkey-go/valkeycompatmock => ../../valkeycompatmock

require (
	github.com/go-redis/redismock/v9 v9.2.0
	github.com/redis/go-redis/v9 v9.2.0
	github.com/valkey-io/valkey-go/valkeycompat v1.0.75
	github.com/valkey-io/valkey-go/valkeycompatmock v1.0.75
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/valkey-io/valkey-go v1.0.75 // indirect
	github.com/valkey-io/valkey-go/mock v1.0.75 // indirect
	go.uber.org/mock v0.6.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
)
