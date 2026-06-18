# PMAxis — Storage & Ingestion Projections

> Capacity planning reference for a 4 GB RAM / 30 GB disk VPS running the full PMAxis stack.

---

## How Data Flows Into Storage

```
Polymarket WebSocket (4 connections)
         │
         ▼
   Ingestion Hub  ──► Kafka (2-hour transit pipe)
         │
         ▼
   Storage Hub
         │
         ├──► events_raw          (raw WS payloads — 1 day TTL)
         ├──► trades              (normalized trades — 7 day TTL)
         ├──► orderbook_history   (L2 snapshots — 2 day TTL)
         ├──► market_metrics      (per-market aggregates — 2 day TTL)
         ├──► market_metadata     (market registry — permanent)
         ├──► price_history_1m    (1-min OHLCV candles — 90 day TTL)
         └──► price_history_1h    (1-hour OHLCV candles — 365 day TTL)

Polygon RPC (on-chain)
         │
         ▼
   Ingestion Hub ──► onchain_trades (14 day TTL)
```

---

## ClickHouse Table Reference

### TTL-Capped Tables (reach steady state, never grow past it)

#### `events_raw` — 1 day TTL

Raw WebSocket payloads before normalization. Every event type lands here first.

| Metric | Estimate |
|---|---|
| Write rate | ~200 events/sec across all markets |
| Row size (compressed) | ~100–150 bytes |
| Daily ingest | ~1.7 GB raw → ~200–400 MB compressed |
| **Steady state** | **~200 – 400 MB** |

> Highest write frequency table. TTL reduced from 3 days to 1 day on 2026-06-18 to cap disk growth. Raw events are only useful for short-term debugging.

---

#### `orderbook_history` — 2 day TTL

L2 orderbook snapshots captured on every orderbook update. Each row stores the full bids/asks JSON arrays.

| Metric | Estimate |
|---|---|
| Write rate | ~50 snapshots/sec (active markets) |
| Row size (compressed) | ~300–500 bytes (bids + asks JSON) |
| Daily ingest | ~1.3–2.2 GB raw → ~500–800 MB compressed |
| **Steady state** | **~1.0 – 1.6 GB** |

> Second-highest volume table. JSON columns compress poorly compared to numeric columns — this is the main TTL-capped contributor to disk use.

---

#### `trades` — 7 day TTL

Normalized, deduplicated trade events. Source for all price calculations and signals.

| Metric | Estimate |
|---|---|
| Write rate | ~5–20 trades/sec (depends on market activity) |
| Row size (compressed) | ~40–60 bytes |
| Daily ingest | ~25–100 MB raw → ~10–30 MB compressed |
| **Steady state** | **~70 – 210 MB** |

---

#### `market_metrics` — 2 day TTL

Aggregated per-market stats (trade count, volume, avg price) written by the processor hub.

| Metric | Estimate |
|---|---|
| Write rate | low — triggered by trade events |
| Row size (compressed) | ~30 bytes |
| Daily ingest | ~50–100 MB |
| **Steady state** | **~100 – 200 MB** |

---

#### `onchain_trades` — 14 day TTL

On-chain Polygon trade events fetched via RPC. Lower frequency than WebSocket trades.

| Metric | Estimate |
|---|---|
| Write rate | ~1–5 events/sec |
| Row size (compressed) | ~80 bytes |
| **Steady state** | **~50 – 150 MB** |

---

### TTL-Capped Candle Tables

#### `price_history_1m` — 90 day TTL

1-minute OHLCV candles built automatically by a materialized view on `trades`.

| Metric | Estimate |
|---|---|
| Active markets generating candles | ~100–500 at any time |
| Rows per day | 100–500 markets × 1440 min = 144k–720k rows/day |
| Row size (compressed) | ~40 bytes |
| **Daily growth** | **~6 – 30 MB/day** |
| **Steady state (90 days)** | **~540 MB – 2.7 GB** |

> TTL added 2026-06-18 — was previously growing unboundedly. 90 days of minute candles is more than enough for analytics and backtesting. Hourly candles (`price_history_1h`) retain a full year for longer-range queries.

---

#### `price_history_1h` — 365 day TTL

1-hour OHLCV candles. 60× fewer rows than `price_history_1m`.

| Metric | Estimate |
|---|---|
| Rows per day | ~2.4k – 12k rows/day |
| **Daily growth** | **~0.1 – 0.5 MB/day** |
| **Steady state (365 days)** | **~37 – 183 MB** |

> TTL added 2026-06-18 — was previously growing unboundedly.

---

#### `market_metadata`

One row per Polymarket market. Grows only as new markets open.

| Metric | Estimate |
|---|---|
| Total Polymarket markets | ~10,000–50,000 lifetime |
| Row size | ~500 bytes |
| **Total size** | **~5 – 25 MB** (effectively flat) |

---

## Full Stack Disk Budget (30 GB VPS)

```
30 GB total
├──  5 GB   OS + Docker engine + images (5 images × ~200 MB each)
├──  1 GB   Kafka data + Zookeeper (capped by 2-hour retention + 200 MB/partition)
├──  0.5 GB Postgres (backfill offsets — tiny)
├──  0.6 GB Docker container logs (capped at 20 MB × 3 files × 10 services)
│
└── ~23 GB  ClickHouse
    ├──  1–2 GB   TTL-capped short-term tables (events_raw, trades, orderbook, metrics)
    ├──  0.5–3 GB price_history_1m (90-day rolling window)
    └──  0.1–0.2 GB price_history_1h (365-day rolling window)
```

### When does 30 GB fill up?

All ClickHouse tables now have TTLs, so disk reaches a steady state rather than growing indefinitely.

| Table | Steady state |
|---|---|
| `events_raw` (1 day) | ~200–400 MB |
| `trades` (7 days) | ~70–210 MB |
| `orderbook_history` (2 days) | ~1.0–1.6 GB |
| `market_metrics` (2 days) | ~100–200 MB |
| `price_history_1m` (90 days) | ~540 MB–2.7 GB |
| `price_history_1h` (365 days) | ~37–183 MB |
| `market_metadata` (permanent) | ~5–25 MB |
| `onchain_trades` (14 days) | ~50–150 MB |
| **Total ClickHouse** | **~2–5.5 GB steady state** |

**Disk is now bounded. No growth cliff expected.**

---

## RAM Budget (4 GB VPS)

| Service | Memory limit | Notes |
|---|---|---|
| ClickHouse | 1 GB | Largest consumer — query cache + merge buffers |
| Kafka | 512 MB | JVM heap fixed at 256M + overhead |
| Postgres | 256 MB | Mostly idle (backfill offsets only) |
| Redis | 100 MB | 64 MB data cap + overhead |
| Zookeeper | 128 MB | |
| hub-processor | 400 MB | Holds live orderbook state in memory |
| hub-ingestion | 256 MB | 4 WS connections + RPC pool |
| hub-storage | 256 MB | Batch flush buffers |
| hub-discovery | 200 MB | |
| hub-api | 256 MB | Redis-backed response cache |
| OS + Docker | ~400 MB | Kernel, networking, Docker daemon |
| **Total** | **~3.8 GB** | Tight but within limits |

> RAM is the tighter constraint on a 4 GB VPS. If any service hits its limit, Docker kills it and `restart: always` brings it back. Monitor with `docker stats`.

---

## Monitoring Commands

```bash
# ClickHouse table sizes
docker exec -it <clickhouse-container> clickhouse-client --query \
  "SELECT table, formatReadableSize(sum(bytes)) AS size, count() AS parts
   FROM system.parts WHERE active GROUP BY table ORDER BY sum(bytes) DESC"

# Overall disk usage
df -h /var/lib/docker

# Docker volume sizes
du -sh /var/lib/docker/volumes/*

# Live RAM per container
docker stats --no-stream
```

---

## If Disk Starts Running Low

The following TTL changes have already been applied (2026-06-18). If disk pressure returns, these are the remaining levers:

**Option 1 — Shorten `price_history_1m` TTL further (e.g. 30 days)**
```sql
ALTER TABLE price_history_1m MODIFY TTL timestamp + INTERVAL 30 DAY;
```
Saves up to ~2 GB. Use this before anything else — hourly candles still cover older data.

**Option 2 — Shorten `orderbook_history` TTL (from 2 days to 12 hours)**
```sql
ALTER TABLE orderbook_history MODIFY TTL timestamp + INTERVAL 12 HOUR;
```
Saves ~500–800 MB immediately. Real-time signals only need recent orderbook snapshots.

**Option 3 — Shorten `onchain_trades` TTL (from 14 days to 7 days)**
```sql
ALTER TABLE onchain_trades MODIFY TTL timestamp + INTERVAL 7 DAY;
```

### Already applied

| Change | Date | Saved |
|---|---|---|
| `events_raw` TTL: 3 days → 1 day | 2026-06-18 | ~400–800 MB |
| `price_history_1m` TTL: permanent → 90 days | 2026-06-18 | Growing unboundedly → capped |
| `price_history_1h` TTL: permanent → 365 days | 2026-06-18 | Capped |
| Kafka retention bytes: 1 GB → 200 MB/partition | 2026-06-18 | ~4 GB |
| Log rotation added to all hub-* services | 2026-06-18 | Prevents unbounded log growth |

---

## Related Docs

- [info.md](info.md) — Full project reference and development history
- [port-security.md](port-security.md) — Port bindings, networking, and TLS setup
- [api-reference.md](api-reference.md) — API endpoint reference
