<p align="center">
  <img src="https://img.shields.io/badge/Language-Go%201.23-00ADD8?logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/Streaming-Kafka-231F20?logo=apachekafka" alt="Kafka">
  <img src="https://img.shields.io/badge/State-Redis-DC382D?logo=redis" alt="Redis">
  <img src="https://img.shields.io/badge/OLAP-ClickHouse-FFCC01?logo=clickhouse" alt="ClickHouse">
</p>

# PMAxis — Real-Time Prediction Market Data Engine

**PMAxis** is a production-grade, event-driven data platform built in Go. It ingests live market data and on-chain trade events from Polymarket, normalizes them, maintains real-time orderbook/price state, and exposes everything through a unified REST API and WebSocket gateway.

> The project was originally scaffolded as ~13 fine-grained microservices and has since been **consolidated into 5 high-performance "Core Hub" services** to keep resource usage lean on small VPS deployments (4GB RAM). The standalone React frontend (`PMAxis-Stat`) has been retired — this repo is now backend-only.

---

## Features

- **Live Ingestion**: Connects to Polymarket WebSockets and on-chain Polygon RPC (with automatic multi-endpoint failover/rotation — see [`docs/infrastructure/rpc-failover-design.md`](docs/infrastructure/rpc-failover-design.md)) with robust connection pooling and pacing.
- **Consolidated Hub Architecture**: 5 specialized Go services (`discovery`, `ingestion`, `processor`, `storage`, `api`) decoupled by Apache Kafka event streams.
- **In-Memory Orderbooks**: Maintains high-performance L2 orderbooks and live price state directly in Go, persisting snapshots to Redis.
- **Wallet Activity Tracking**: Watch any wallet address to track on-chain trades via the `wallet_activity` ClickHouse table; query raw settlement history for any wallet with no pre-watching needed via `/v1/wallets/{address}/onchain`.
- **Unified API & Real-Time Gateway**: REST endpoints (`/v1/*`) with Redis-backed response caching and rate limiting, plus a WebSocket hub (`/stream`) pushing live trades, orderbook updates, prices, and signals — see [`docs/api/api-reference.md`](docs/api/api-reference.md).
- **Interactive API Docs**: Scalar-powered docs page at `/docs` with theme switcher and live try-it-out.
- **High-Throughput Storage**: Batch-inserts trades, candles, and metrics into ClickHouse, with Redis for hot state and PostgreSQL for relational/offset data.

## Architecture

The system is decoupled via Kafka to allow horizontal scaling of any individual hub.

```text
Polymarket (WebSocket + On-chain RPC)
      │
      ▼
[ Discovery Hub ]  ──► market & token discovery   ──► pmaxis.markets.*
      │
      ▼
[ Ingestion Hub ]  ──► live trades & on-chain logs ──► pmaxis.trades / pmaxis.onchain.*
      │
      ▼
[ Processor Hub ]  ──► orderbooks, prices, signals ──► pmaxis.orderbook / pmaxis.prices / pmaxis.signals
      │
      ▼
[ Storage Hub ]    ──► batch writes ──► ClickHouse · Redis · PostgreSQL
      │
      ▼
[ API Hub ]        ──► REST (/v1/*) & WebSocket Gateway (port 8088)
```

## Project Structure

- `libs/`: Shared production-grade libraries (Kafka clients, ClickHouse, Redis, Postgres, WebSocket pools, schemas, retry/config/logging).
- `services/`: The 5 Core Hub services — `discovery`, `ingestion`, `processor`, `storage`, `api`.
- `cmd/pmaxis/`: A built-in developer CLI tool for testing data streams directly from the terminal.
- `deployments/`: Docker Compose configurations that orchestrate the full container stack.
- `docs/`: Documentation organized by topic — see [`docs/README.md`](docs/README.md) for the full index.

---

## Getting Started (Docker)

The easiest way to start the entire data pipeline (Kafka, Redis, ClickHouse, PostgreSQL, and all 5 Core Hub services) is via the provided Windows batch scripts (run from the repo root, one level up):

```bash
# 1. Build and start the full stack in the background
start-docker.bat

# 2. To completely stop and tear down the environment
stop-docker.bat
```

> **Note:** The first `start` may take a few minutes as Go downloads module dependencies inside the containers. Subsequent boots are heavily cached and very fast.

### Health Check

```bash
curl http://localhost:8088/health
```

### API Docs

Open `http://localhost:8088/docs` in a browser for the interactive Scalar API reference with live try-it-out.

---

## Documentation

| Topic | File |
|-------|------|
| API endpoint reference | [`docs/api/api-reference.md`](docs/api/api-reference.md) |
| VPS deploy guide | [`docs/deployment/vps-deploy.md`](docs/deployment/vps-deploy.md) |
| Environment variables | [`docs/deployment/env-vars.md`](docs/deployment/env-vars.md) |
| Rollback procedures | [`docs/deployment/rollback.md`](docs/deployment/rollback.md) |
| ClickHouse query guide | [`docs/data/clickhouse-queries.md`](docs/data/clickhouse-queries.md) |
| Monitoring & alerting | [`docs/operations/monitoring.md`](docs/operations/monitoring.md) |
| Troubleshooting | [`docs/operations/troubleshooting.md`](docs/operations/troubleshooting.md) |
| Port security | [`docs/infrastructure/port-security.md`](docs/infrastructure/port-security.md) |
| Storage projections | [`docs/infrastructure/storage-projections.md`](docs/infrastructure/storage-projections.md) |
| RPC failover design | [`docs/infrastructure/rpc-failover-design.md`](docs/infrastructure/rpc-failover-design.md) |
