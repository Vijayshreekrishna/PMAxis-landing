# PMAxis — Data Retention & Disk Projections

## What Gets Saved in ClickHouse

Every table in the system either has a TTL (auto-deleted after N days) or is permanent.

| Table | What's stored | TTL |
|---|---|---|
| `market_metadata` | Market registry — title, slug, condition ID, status, outcomes, tags, category | **Permanent** |
| `trades` | Normalized live trade events (price, size, side, market) | 7 days |
| `events_raw` | Raw WebSocket payloads from Polymarket before normalization | 1 day |
| `orderbook_history` | Full L2 orderbook snapshots (bids + asks JSON) | 2 days |
| `market_metrics` | Aggregated per-market stats (volume, trade count, avg price) | 2 days |
| `onchain_trades` | Raw on-chain settlements from Polygon RPC | 14 days |
| `price_history_1m` | 1-minute OHLCV candles (built from trades via materialized view) | 90 days |
| `price_history_1h` | 1-hour OHLCV candles (built from trades via materialized view) | 365 days |
| `wallet_activity` | Trades for explicitly watched wallets | 90 days |

> TTL means ClickHouse automatically deletes rows older than the window. No manual cleanup needed.

---

## Current System — Steady State Disk Usage

At steady state (all TTLs filled), the disk usage is bounded and never grows past this:

| Table | Steady state size |
|---|---|
| `market_metadata` | ~5–25 MB |
| `trades` (7 days) | ~150–600 MB |
| `events_raw` (1 day) | ~200–400 MB |
| `orderbook_history` (2 days) | ~1.0–1.6 GB |
| `market_metrics` (2 days) | ~100–300 MB |
| `onchain_trades` (14 days) | ~100–500 MB |
| `price_history_1m` (90 days) | ~500 MB–2.7 GB |
| `price_history_1h` (365 days) | ~40–180 MB |
| `wallet_activity` (90 days) | ~50–200 MB |
| **Total ClickHouse** | **~2–6.5 GB** |

**Full stack on a 40 GB VPS:**

| Layer | Usage |
|---|---|
| OS + Docker images | ~5 GB |
| Kafka + Zookeeper | ~1–2 GB |
| PostgreSQL | ~0.5 GB |
| Docker logs | ~0.6 GB |
| ClickHouse | ~2–6.5 GB |
| **Total** | **~9–15 GB** |

**Free space remaining: ~25–31 GB out of 40 GB.**

---

## 1-Year Retention Projection (Hypothetical)

If all TTLs were extended to 1 year, disk requirements would be:

| Table | Daily write (compressed) | 1 year total |
|---|---|---|
| `events_raw` | ~200–400 MB/day | ~73–146 GB |
| `orderbook_history` | ~500–800 MB/day | ~182–292 GB |
| `trades` | ~10–30 MB/day | ~4–11 GB |
| `market_metrics` | ~50–100 MB/day | ~18–37 GB |
| `onchain_trades` | ~7–34 MB/day | ~3–12 GB |
| `price_history_1m` | ~6–30 MB/day | ~2–11 GB |
| `price_history_1h` | ~0.1–0.5 MB/day | ~0.2 GB |
| `wallet_activity` | varies | ~5–20 GB |
| **Total** | | **~280–530 GB** |

The two tables that make 1-year retention impractical are `events_raw` and `orderbook_history`. They account for **~250–440 GB** of the total on their own.

**If only the useful analytical tables were kept for 1 year** (trades, candles, metrics, onchain, wallet — no raw events or orderbook history):

| | Size |
|---|---|
| Selective 1-year retention | ~35–90 GB |
| Recommended disk for this | **100 GB VPS** |

---

## Verdict — Current System Is Stable

The current TTL configuration is well-suited to the 40 GB VPS:

- Disk usage is **bounded** — no table grows indefinitely.
- At steady state the system uses **~9–15 GB**, leaving 25+ GB of free headroom.
- The only permanent table (`market_metadata`) grows slowly and will never exceed ~25 MB even across Polymarket's full market lifetime.
- `price_history_1m` and `orderbook_history` are the two largest contributors and are already capped at 90 days and 2 days respectively.

No disk expansion is needed under the current configuration.
