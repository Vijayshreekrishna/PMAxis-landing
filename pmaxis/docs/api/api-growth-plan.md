# PMAxis — API Growth Plan

## Current State

The API hub (`hub-api`) is live on port 8088 with the following setup:

- All `/v1/*` routes are **public** — auth middleware exists in code but is commented out
- Rate limiting is **per IP** at 60 req/min (Redis-backed sliding window)
- No API key system, no application registration, no usage tracking
- WebSocket gateway at `/stream` is fully open
- Debug routes gated behind `X-Debug-Token` header only

This is fine for internal use and early testing, but not suitable for exposing to external developers or monetizing.

---

## What Needs to Be Built

### Step 1 — API Key Auth

Uncomment and implement the auth middleware at `services/api/middleware/auth.go`.

Each API key needs:

| Field | Description |
|---|---|
| `key` | Random UUID or prefixed token (e.g. `pmx_live_xxxxx`) |
| `app_name` | Name of the application |
| `owner_email` | Developer contact |
| `tier` | `free`, `pro`, `enterprise` |
| `created_at` | Timestamp |
| `active` | Bool — allows instant revocation |

**Storage:** Redis is the right choice for key lookup (sub-millisecond). PostgreSQL as the source of truth for persistence.

Flow on every request:
```
Request hits /v1/*
  → middleware reads X-API-Key header
  → Redis lookup: GET apikey:{key}
  → if miss → 401 Unauthorized
  → if found but inactive → 403 Forbidden
  → if valid → attach tier to request context → proceed
```

---

### Step 2 — Rate Limit Per API Key (Not Per IP)

Change the rate limit key in `services/api/middleware/ratelimit.go` from:
```
rl:{ip}
```
to:
```
rl:{api_key}:{minute_window}
```

Rate limits per tier:

| Tier | Rate limit | WebSocket connections |
|---|---|---|
| Free | 60 req/min | 1 concurrent |
| Pro | 600 req/min | 10 concurrent |
| Enterprise | Custom | Unlimited |

The current Redis sliding window implementation already supports this — only the key prefix needs changing.

---

### Step 3 — Key Management Endpoints

New routes to add under `/v1/keys` (protected by admin token, not API key):

| Method | Route | Description |
|---|---|---|
| `POST` | `/v1/keys` | Create a new API key |
| `GET` | `/v1/keys/{key}` | Get key details and tier |
| `DELETE` | `/v1/keys/{key}` | Revoke a key |
| `GET` | `/v1/keys/{key}/usage` | Request count for current window |

---

## How Many Apps the Current VPS Can Support

### Hardware Ceiling (4 GB RAM)

| Service | RAM allocated |
|---|---|
| ClickHouse | 1 GB |
| Kafka | 512 MB |
| hub-processor | 400 MB |
| hub-api | 256 MB |
| hub-ingestion | 256 MB |
| hub-storage | 256 MB |
| Postgres | 256 MB |
| Redis | 100 MB |
| Zookeeper | 128 MB |
| OS + Docker | ~400 MB |
| **Total** | **~3.8 GB** |

RAM is the tightest constraint. ClickHouse at 1 GB is the hard ceiling for analytical query throughput.

---

### Response Times Per Endpoint

| Endpoint | Data source | Avg response time |
|---|---|---|
| `/v1/markets/{id}/price` | Redis | ~1–3 ms |
| `/v1/markets/top` | Redis cache (15s TTL) | ~1–3 ms |
| `/v1/markets/search` | Redis cache (30s TTL) | ~2–5 ms |
| `/v1/markets/{id}/summary` | Redis cache (10s TTL) | ~2–5 ms |
| `/v1/prices` (batch) | Redis | ~3–8 ms |
| `/v1/markets/{id}/candles` | ClickHouse | ~20–100 ms |
| `/v1/markets/{id}/trades` | ClickHouse | ~20–80 ms |
| `/v1/markets/{id}/price-history` | ClickHouse | ~30–120 ms |
| `/stream` WebSocket | Kafka push | ~10–50 ms latency |

Redis-backed endpoints are essentially free at this scale. ClickHouse is where load accumulates.

---

### App Capacity Projections

**Free tier — 60 req/min per app:**

| Usage pattern | Apps supported comfortably |
|---|---|
| Mostly price / orderbook (Redis) | 300–500 apps |
| Mixed Redis + ClickHouse | 100–200 apps |
| Heavy analytics (candles, history) | 40–80 apps |
| WebSocket only | 500–1,000 concurrent connections |

**Pro tier — 600 req/min per app:**

| Usage pattern | Apps supported |
|---|---|
| Mostly Redis | 50–100 apps |
| Mixed | 20–40 apps |
| Heavy ClickHouse | 10–20 apps |

**Practical ceiling on current VPS:**
- ~100–200 apps at free tier with mixed usage
- ~20–50 apps at pro tier
- ~500–1,000 simultaneous WebSocket connections

ClickHouse handles roughly 50–150 analytical queries/second at 1 GB RAM. That is the hard floor — Redis endpoints have virtually no ceiling at this scale.

---

## Scaling Path (When You Hit the Ceiling)

You do not need to change the architecture — just increase resources:

| Bottleneck | Fix | Effect |
|---|---|---|
| ClickHouse query slowdown | Increase ClickHouse RAM to 2–4 GB | Doubles analytical query capacity |
| Too many WebSocket connections | Increase hub-api RAM to 512 MB | Adds ~2,000 more concurrent WS connections |
| Need more apps at pro tier | Move to 8 GB VPS | Comfortable headroom for 100+ pro apps |
| Enterprise scale | Separate ClickHouse node | Fully decoupled query layer |

The Kafka + service architecture is already horizontally scalable — each hub can be duplicated independently without code changes.

---

## Recommended Rollout Order

1. **Implement API key storage** in Redis + Postgres — 1 table, straightforward
2. **Wire auth middleware** — uncomment + implement key lookup
3. **Switch rate limit key** from IP to API key — one line change in ratelimit.go
4. **Add key management endpoints** — create, revoke, usage
5. **Set up a simple developer registration form** — email + app name → returns key
6. **Add usage tracking** — write request counts to Redis, expose via `/v1/keys/{key}/usage`
7. **Introduce Pro tier** — higher rate limit, more WS connections, longer history access

---

## What Each Tier Should Unlock

| Feature | Free | Pro | Enterprise |
|---|---|---|---|
| Rate limit | 60 req/min | 600 req/min | Custom |
| Price history depth | 7 days | 90 days | 365 days |
| WebSocket connections | 1 | 10 | Unlimited |
| Candle resolution | 1m, 1h | 1m, 5m, 1h | All |
| Batch prices (`/v1/prices`) | 10 ids max | 100 ids max | Unlimited |
| Wallet activity | No | Yes | Yes |
| Debug endpoints | No | No | Yes |
| SLA | None | 99.5% | 99.9% |

---

## Summary

The current API is production-ready in terms of data quality and endpoint coverage. What it lacks is an identity layer — there is no way to know who is calling, enforce per-app limits, or offer tiered access.

The path forward is well-defined, builds on what already exists, and does not require architectural changes. The VPS can comfortably support 100–200 free-tier apps or 20–50 pro-tier apps before any hardware upgrade is needed.
