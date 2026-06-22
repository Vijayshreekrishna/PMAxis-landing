# Graph Report - .  (2026-06-22)

## Corpus Check
- 84 files · ~62,757 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 704 nodes · 1259 edges · 85 communities (41 shown, 44 thin omitted)
- Extraction: 92% EXTRACTED · 8% INFERRED · 0% AMBIGUOUS · INFERRED: 95 edges (avg confidence: 0.79)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- [[_COMMUNITY_API Handler Layer|API Handler Layer]]
- [[_COMMUNITY_Kafka Client & Messaging|Kafka Client & Messaging]]
- [[_COMMUNITY_Event Schemas & Normalization|Event Schemas & Normalization]]
- [[_COMMUNITY_Ingestion & On-Chain Indexer|Ingestion & On-Chain Indexer]]
- [[_COMMUNITY_Schema Definitions & Events|Schema Definitions & Events]]
- [[_COMMUNITY_Market Discovery|Market Discovery]]
- [[_COMMUNITY_Market Data Endpoints|Market Data Endpoints]]
- [[_COMMUNITY_API Routing & Config|API Routing & Config]]
- [[_COMMUNITY_Deployment & Caddy|Deployment & Caddy]]
- [[_COMMUNITY_Events & Taxonomy API|Events & Taxonomy API]]
- [[_COMMUNITY_Infrastructure Services|Infrastructure Services]]
- [[_COMMUNITY_Redis Client|Redis Client]]
- [[_COMMUNITY_Monitoring & Alerting|Monitoring & Alerting]]
- [[_COMMUNITY_Shared Libraries|Shared Libraries]]
- [[_COMMUNITY_Processor & Signal Pipeline|Processor & Signal Pipeline]]
- [[_COMMUNITY_Community 15|Community 15]]
- [[_COMMUNITY_Community 16|Community 16]]
- [[_COMMUNITY_Community 17|Community 17]]
- [[_COMMUNITY_Community 18|Community 18]]
- [[_COMMUNITY_Community 19|Community 19]]
- [[_COMMUNITY_Community 20|Community 20]]
- [[_COMMUNITY_Community 21|Community 21]]
- [[_COMMUNITY_Community 22|Community 22]]
- [[_COMMUNITY_Community 23|Community 23]]
- [[_COMMUNITY_Community 24|Community 24]]
- [[_COMMUNITY_Community 25|Community 25]]
- [[_COMMUNITY_Community 26|Community 26]]
- [[_COMMUNITY_Community 27|Community 27]]
- [[_COMMUNITY_Community 28|Community 28]]
- [[_COMMUNITY_Community 29|Community 29]]
- [[_COMMUNITY_Community 30|Community 30]]
- [[_COMMUNITY_Community 31|Community 31]]
- [[_COMMUNITY_Community 32|Community 32]]
- [[_COMMUNITY_Community 33|Community 33]]
- [[_COMMUNITY_Community 34|Community 34]]
- [[_COMMUNITY_Community 35|Community 35]]
- [[_COMMUNITY_Community 36|Community 36]]
- [[_COMMUNITY_Community 37|Community 37]]
- [[_COMMUNITY_Community 38|Community 38]]
- [[_COMMUNITY_Community 39|Community 39]]
- [[_COMMUNITY_Community 40|Community 40]]
- [[_COMMUNITY_Community 41|Community 41]]
- [[_COMMUNITY_Community 42|Community 42]]
- [[_COMMUNITY_Community 43|Community 43]]
- [[_COMMUNITY_Community 44|Community 44]]
- [[_COMMUNITY_Community 45|Community 45]]
- [[_COMMUNITY_Community 46|Community 46]]
- [[_COMMUNITY_Community 47|Community 47]]
- [[_COMMUNITY_Community 48|Community 48]]
- [[_COMMUNITY_Community 49|Community 49]]
- [[_COMMUNITY_Community 50|Community 50]]
- [[_COMMUNITY_Community 51|Community 51]]
- [[_COMMUNITY_Community 52|Community 52]]
- [[_COMMUNITY_Community 53|Community 53]]
- [[_COMMUNITY_Community 54|Community 54]]
- [[_COMMUNITY_Community 55|Community 55]]
- [[_COMMUNITY_Community 56|Community 56]]
- [[_COMMUNITY_Community 57|Community 57]]
- [[_COMMUNITY_Community 58|Community 58]]
- [[_COMMUNITY_Community 59|Community 59]]
- [[_COMMUNITY_Community 60|Community 60]]
- [[_COMMUNITY_Community 61|Community 61]]
- [[_COMMUNITY_Community 62|Community 62]]
- [[_COMMUNITY_Community 63|Community 63]]
- [[_COMMUNITY_Community 64|Community 64]]
- [[_COMMUNITY_Community 65|Community 65]]
- [[_COMMUNITY_Community 66|Community 66]]
- [[_COMMUNITY_Community 67|Community 67]]
- [[_COMMUNITY_Community 68|Community 68]]
- [[_COMMUNITY_Community 69|Community 69]]
- [[_COMMUNITY_Community 70|Community 70]]
- [[_COMMUNITY_Community 71|Community 71]]
- [[_COMMUNITY_Community 72|Community 72]]
- [[_COMMUNITY_Community 73|Community 73]]
- [[_COMMUNITY_Community 74|Community 74]]
- [[_COMMUNITY_Community 75|Community 75]]
- [[_COMMUNITY_Community 76|Community 76]]
- [[_COMMUNITY_Community 77|Community 77]]
- [[_COMMUNITY_Community 78|Community 78]]
- [[_COMMUNITY_Community 79|Community 79]]
- [[_COMMUNITY_Community 80|Community 80]]
- [[_COMMUNITY_Community 81|Community 81]]
- [[_COMMUNITY_Community 82|Community 82]]
- [[_COMMUNITY_Community 83|Community 83]]
- [[_COMMUNITY_Community 84|Community 84]]

## God Nodes (most connected - your core abstractions)
1. `APIHandler` - 51 edges
2. `Context` - 44 edges
3. `Request` - 36 edges
4. `ResponseWriter` - 35 edges
5. `GetEnv()` - 17 edges
6. `getQuery()` - 17 edges
7. `marshal()` - 17 edges
8. `discoverMarkets()` - 17 edges
9. `discoverEvents()` - 17 edges
10. `Hub` - 15 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `NewLogger()`  [INFERRED]
  pmaxis/libs/kafka-client/examples/main.go → pmaxis/libs/logger/logger.go
- `main()` --calls--> `NewLogger()`  [INFERRED]
  pmaxis/libs/clickhouse-client/examples/main.go → pmaxis/libs/logger/logger.go
- `TestConfig()` --calls--> `New()`  [INFERRED]
  pmaxis/libs/config/config_test.go → pmaxis/libs/config/config.go
- `processLog()` --calls--> `New()`  [INFERRED]
  pmaxis/services/ingestion/main.go → pmaxis/libs/config/config.go
- `runIndexerCycle()` --calls--> `New()`  [INFERRED]
  pmaxis/services/ingestion/main.go → pmaxis/libs/config/config.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **** — hub_discovery, hub_ingestion, hub_processor, hub_storage, hub_api [EXTRACTED]
- **** — apache_kafka, redis, clickhouse, postgresql, zookeeper [EXTRACTED]
- **** — table_trades, table_events_raw, table_orderbook_history, table_onchain_trades, table_market_metrics, table_price_history_1m, table_price_history_1h, table_wallet_activity, table_market_metadata [EXTRACTED]
- **** — api_tier_free, api_tier_pro, api_tier_enterprise [EXTRACTED]
- **** — hub_discovery, hub_ingestion, hub_processor, hub_storage, hub_api [EXTRACTED]
- **** — kafka, redis, clickhouse, postgres, zookeeper [EXTRACTED]
- **** — events_raw_table, trades_table, orderbook_history_table, market_metrics_table, onchain_trades_table, price_history_1m_table, price_history_1h_table, market_metadata_table [EXTRACTED]
- **** — uptime_kuma_dashboard, cron_job, discord_webhook [EXTRACTED]
- **** — security_headers_middleware, cors_middleware, ratelimit_middleware, debug_token_var [EXTRACTED]

## Communities (85 total, 44 thin omitted)

### Community 0 - "API Handler Layer"
Cohesion: 0.12
Nodes (18): APIHandler, decodeCursor(), encodeCursor(), fallback(), getQuery(), parseLimit(), parseOffset(), parseOutcomes() (+10 more)

### Community 1 - "Kafka Client & Messaging"
Cohesion: 0.09
Nodes (42): Client, ClientInterface, ConsumerInterface, NewClient(), NewTypedConsumer(), NewTypedProducer(), ProducerInterface, TypedConsumer (+34 more)

### Community 2 - "Event Schemas & Normalization"
Cohesion: 0.11
Nodes (33): GenerateEventID(), HistoricalEvent, NormalizedEvent, OrderbookUpdate, Context, Duration, T, BaseService (+25 more)

### Community 3 - "Ingestion & On-Chain Indexer"
Cohesion: 0.12
Nodes (28): Hash, extractEventInfo(), flushMetrics(), newRPCRotator(), processLog(), runIndexerCycle(), startAggregatorComponent(), startIndexerComponent() (+20 more)

### Community 4 - "Schema Definitions & Events"
Cohesion: 0.08
Nodes (20): PriceLevel, Time, T, GetVersion(), MarshalEvent(), TestEventSerialization(), TestEventValidation(), UnmarshalEvent() (+12 more)

### Community 5 - "Market Discovery"
Cohesion: 0.17
Nodes (26): GammaResponse, discoverEvents(), discoverMarkets(), extractCategory(), extractTagSlugs(), parseClobTokenIDs(), parseTime(), processMarket() (+18 more)

### Community 6 - "Market Data Endpoints"
Cohesion: 0.10
Nodes (25): GET /v1/markets/{id}/candles, GET /v1/markets/{id}/positions, GET /v1/markets/{id}/price-history, GET /v1/markets/{id}/trades, GET /v1/positions, GET /v1/positions/closed, GET /v1/wallets/{address}/activity, GET /v1/wallets/{address}/onchain (+17 more)

### Community 7 - "API Routing & Config"
Cohesion: 0.09
Nodes (23): ALLOWED_ORIGINS env var, API /docs (Scalar UI) endpoint, API_PORT env var, API /v1/events endpoint, API /v1/positions endpoint, API /v1/markets/{id}/signals endpoint, CORS Allowlist Middleware, Debug Endpoints /debug/* (+15 more)

### Community 8 - "Deployment & Caddy"
Cohesion: 0.11
Nodes (20): build-and-ship.bat script, Caddy (TLS Reverse Proxy), Caddy Reverse Proxy, Caddyfile, Docker Compose, docker-compose.caddy.yml overlay, Docker Healthcheck, Docker Image TAR files (+12 more)

### Community 9 - "Events & Taxonomy API"
Cohesion: 0.12
Nodes (18): GET /v1/events, API /v1/categories endpoint, API /v1/tags endpoint, DerivedStatus() method, Env: GAMMA_EVENTS_URL, Polymarket Gamma API, GAMMA_API_URL env var, GAMMA_EVENTS_URL env var (+10 more)

### Community 10 - "Infrastructure Services"
Cohesion: 0.12
Nodes (17): Apache Kafka, CLICKHOUSE_ADDR env var, Env: POLYGON_RPC_URLS, FNV-1a Consistent Hashing (WS distribution), Ingestion Hub, KAFKA_BROKERS env var, Kafka Topic: pmaxis.onchain.*, Kafka Topic: pmaxis.tokens.extracted (+9 more)

### Community 11 - "Redis Client"
Cohesion: 0.21
Nodes (9): BoolCmd, IntCmd, Context, Duration, Client, Interface, Options, NewClient() (+1 more)

### Community 12 - "Monitoring & Alerting"
Cohesion: 0.14
Nodes (16): Hourly Cron Job, Discord Health Monitoring Cron, Discord Webhook, Go 1.23, Apache Kafka, Kafka Healthcheck (kafka-topics --list), Kafka Retention Configuration, Microservices Architecture (+8 more)

### Community 13 - "Shared Libraries"
Cohesion: 0.39
Nodes (16): github.com/pmaxis/pmaxis/libs/clickhouse-client, github.com/pmaxis/pmaxis/libs/config, github.com/pmaxis/pmaxis/libs/crypto, github.com/pmaxis/pmaxis/libs/kafka-client, github.com/pmaxis/pmaxis/libs/logger, github.com/pmaxis/pmaxis/libs/postgres-client, github.com/pmaxis/pmaxis/libs/redis-client, github.com/pmaxis/pmaxis/libs/retry-utils (+8 more)

### Community 14 - "Processor & Signal Pipeline"
Cohesion: 0.13
Nodes (15): POST /v1/wallets/watch, ClickHouse Orphaned Store Directories, Processor Hub, Kafka Topic: pmaxis.orderbook, Kafka Topic: pmaxis.orderbook.updates, Kafka Topic: pmaxis.prices, Kafka Topic: pmaxis.signals, Kafka Topic: pmaxis.trades.live (+7 more)

### Community 15 - "Community 15"
Cohesion: 0.19
Nodes (10): Conn, Request, ResponseWriter, Upgrader, Client, Interface, Server, ServerInterface (+2 more)

### Community 16 - "Community 16"
Cohesion: 0.19
Nodes (9): Client, Context, Interface, broadcastMsg, Hub, NewHub(), subscribeCmd, walletBroadcastMsg (+1 more)

### Community 17 - "Community 17"
Cohesion: 0.14
Nodes (14): API Growth Plan, API Key Authentication (Planned), API Tier: Enterprise, API Tier: Free, API Tier: Pro, Developer CLI (cmd/pmaxis/), Key Management Endpoints (/v1/keys), PMAxis (+6 more)

### Community 18 - "Community 18"
Cohesion: 0.22
Nodes (13): ClickHouse Table TTL Configuration, events_raw ClickHouse Table, Storage Hub, market_metadata ClickHouse Table, market_metrics ClickHouse Table, onchain_trades ClickHouse Table, orderbook_history ClickHouse Table, Port 8084 (Storage Metrics) (+5 more)

### Community 19 - "Community 19"
Cohesion: 0.18
Nodes (8): NewLogger(), TestLogger(), TestWith(), main(), T, main(), main(), main()

### Community 20 - "Community 20"
Cohesion: 0.24
Nodes (11): Env: TRUSTED_PROXIES, IPNet, libs/redis-client/redis.go, extractClientIP(), middleware/ratelimit.go, isTrusted(), parseTrustedProxies(), RateLimit() (+3 more)

### Community 21 - "Community 21"
Cohesion: 0.24
Nodes (4): Interface, Logger, NewNoOp(), NoOpLogger

### Community 22 - "Community 22"
Cohesion: 0.27
Nodes (9): main(), GetEnv(), main(), main(), EnsureTopics(), DebugAuth(), Context, Interface (+1 more)

### Community 23 - "Community 23"
Cohesion: 0.27
Nodes (7): NewClient(), Client, Interface, Options, Conn, Context, Rows

### Community 24 - "Community 24"
Cohesion: 0.22
Nodes (10): GET /v1/markets/{id}/stats, GET /v1/markets/{id}/summary, GET /v1/markets, GET /v1/markets/search, GET /v1/markets/top, ClickHouse, clickhouse-logging.xml, ClickHouse System Log TTLs (+2 more)

### Community 25 - "Community 25"
Cohesion: 0.22
Nodes (10): API /v1/markets endpoint, API /v1/markets/{id}/orderbook endpoint, API /v1/markets/{id}/price endpoint, API /v1/markets/{id}/trades endpoint, CLI markets subcommand, CLI orderbook subcommand, CLI price subcommand, CLI trades subcommand (+2 more)

### Community 26 - "Community 26"
Cohesion: 0.39
Nodes (5): Config, GetEnvBool(), GetEnvInt(), New(), Interface

### Community 27 - "Community 27"
Cohesion: 0.25
Nodes (7): GET /metrics, Env: DEBUG_TOKEN, Auth(), middleware/auth.go, Handler, Interface, Prometheus Metrics

### Community 28 - "Community 28"
Cohesion: 0.29
Nodes (3): NewClient(), PMAxisClient, Client

### Community 29 - "Community 29"
Cohesion: 0.39
Nodes (6): Conn, Hub, Interface, Client, NewClient(), clientMsg

### Community 30 - "Community 30"
Cohesion: 0.29
Nodes (7): Alpha Leaderboard /v1/leaderboard (planned), Automated Execution Engine (planned), Backtesting Service (planned), ML Signal Refinement (planned), Virtual Paper Trading (planned), PMAxis Roadmap (recipe.md), Whale Tracking (planned)

### Community 31 - "Community 31"
Cohesion: 0.29
Nodes (7): Polymarket CTF Contract, go-ethereum ethclient, On-chain Indexer (startIndexerComponent), Polygon Blockchain, defaultRPCPool (Polygon endpoints), RPC Failover & Rotation Design, rpcRotator (round-robin scheduler)

### Community 32 - "Community 32"
Cohesion: 0.48
Nodes (5): Context, Pool, Client, Interface, NewClient()

### Community 33 - "Community 33"
Cohesion: 0.33
Nodes (5): Context, Interface, Server, BaseService, Config

### Community 35 - "Community 35"
Cohesion: 0.33
Nodes (5): Hub, Interface, Request, ResponseWriter, ServeWS()

### Community 36 - "Community 36"
Cohesion: 0.40
Nodes (4): Env: ALLOWED_ORIGINS, CORS(), middleware/cors.go, Handler

### Community 37 - "Community 37"
Cohesion: 0.40
Nodes (5): K8s ConfigMap (pmaxis-config), K8s Deployment: api, K8s Deployment: market-discovery, K8s Service: api-service (LoadBalancer), Kubernetes

### Community 38 - "Community 38"
Cohesion: 0.50
Nodes (3): MyEvent, main(), Time

### Community 39 - "Community 39"
Cohesion: 0.50
Nodes (3): Time, MarketInfo, TokenInfo

### Community 42 - "Community 42"
Cohesion: 0.67
Nodes (3): Package Docker Compose (deployments/package/), VPS Docker Compose (deployments/docker-compose.vps.yml), VPS Deployment (4 GB RAM)

### Community 43 - "Community 43"
Cohesion: 0.67
Nodes (3): Storage & Ingestion Projections, VPS Disk Budget (30 GB), VPS RAM Budget (4 GB)

## Knowledge Gaps
- **72 isolated node(s):** `Client`, `github.com/pmaxis/pmaxis/cmd/pmaxis`, `Conn`, `Rows`, `T` (+67 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **44 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `GetEnv()` connect `Community 22` to `Kafka Client & Messaging`, `Event Schemas & Normalization`, `Community 36`, `Market Discovery`, `Community 20`, `Community 26`?**
  _High betweenness centrality (0.190) - this node is a cross-community bridge._
- **Why does `marshal()` connect `Market Discovery` to `API Handler Layer`, `Kafka Client & Messaging`, `Event Schemas & Normalization`, `Ingestion & On-Chain Indexer`, `Schema Definitions & Events`?**
  _High betweenness centrality (0.159) - this node is a cross-community bridge._
- **Why does `RateLimit()` connect `Community 20` to `Community 22`?**
  _High betweenness centrality (0.129) - this node is a cross-community bridge._
- **What connects `Client`, `github.com/pmaxis/pmaxis/cmd/pmaxis`, `Conn` to the rest of the system?**
  _72 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `API Handler Layer` be split into smaller, more focused modules?**
  _Cohesion score 0.11808604038630377 - nodes in this community are weakly interconnected._
- **Should `Kafka Client & Messaging` be split into smaller, more focused modules?**
  _Cohesion score 0.08599290780141844 - nodes in this community are weakly interconnected._
- **Should `Event Schemas & Normalization` be split into smaller, more focused modules?**
  _Cohesion score 0.10960960960960961 - nodes in this community are weakly interconnected._