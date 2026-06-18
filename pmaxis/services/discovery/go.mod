module github.com/pmaxis/pmaxis/services/discovery

go 1.24.0

require (
	github.com/pmaxis/pmaxis/libs/config v0.0.0
	github.com/pmaxis/pmaxis/libs/kafka-client v0.0.0
	github.com/pmaxis/pmaxis/libs/redis-client v0.0.0
	github.com/pmaxis/pmaxis/libs/retry-utils v0.0.0
	github.com/pmaxis/pmaxis/libs/schemas v0.0.0
	github.com/pmaxis/pmaxis/libs/service v0.0.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pmaxis/pmaxis/libs/logger v0.0.0 // indirect
	github.com/prometheus/client_golang v1.20.0 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.55.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/redis/go-redis/v9 v9.5.1 // indirect
	github.com/segmentio/kafka-go v0.4.47 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/pmaxis/pmaxis/libs/kafka-client => ../../libs/kafka-client

replace github.com/pmaxis/pmaxis/libs/postgres-client => ../../libs/postgres-client

replace github.com/pmaxis/pmaxis/libs/retry-utils => ../../libs/retry-utils

replace github.com/pmaxis/pmaxis/libs/redis-client => ../../libs/redis-client

replace github.com/pmaxis/pmaxis/libs/schemas => ../../libs/schemas

replace github.com/pmaxis/pmaxis/libs/service => ../../libs/service

replace github.com/pmaxis/pmaxis/libs/logger => ../../libs/logger

replace github.com/pmaxis/pmaxis/libs/config => ../../libs/config

replace github.com/pmaxis/pmaxis/libs/websocket-client => ../../libs/websocket-client

replace github.com/pmaxis/pmaxis/libs/crypto => ../../libs/crypto
