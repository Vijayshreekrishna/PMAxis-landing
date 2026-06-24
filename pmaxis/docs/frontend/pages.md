# PMAxis — Embedded Pages

All pages are compiled into the `hub-api` binary via Go `//go:embed` and served directly from the API server. No separate frontend server needed.

---

## Pages at a Glance

| URL | Who uses it | Auth | Description |
|-----|-------------|------|-------------|
| `/status` | Public | None | Real-time system status with 90-day uptime history |
| `/register` | Developers | None | Self-serve API key registration |
| `/viz` | Public | None | Live data explorer (markets, candles, orderbooks) |
| `/docs` | Public | None | Interactive API reference (Scalar) |
| `/admin` | Operator | `DEBUG_TOKEN` | API key management panel |

---

## Design System

All pages share a consistent PMAxis brand:

| Token | Value | Usage |
|-------|-------|-------|
| Deep Black | `#0A0A0A` | Page background (dark mode) |
| Signal Green | `#00E676` | Primary accent, CTAs, status indicators |
| Intelligence Purple | `#8B5CF6` | Secondary accent, tags |
| Soft Grey | `#A1A1AA` | Muted text, secondary info |
| Pure White | `#FFFFFF` | Page background (light mode), dark-mode text |
| Font | Space Grotesk 300–700 | All pages except `/docs` |
| Nav height | 68px | All pages |
| Logo size | 38px × 38px inline SVG | All pages |

### Light / Dark Mode

Pages `/status`, `/register`, and `/viz` support light/dark toggle:
- Toggle button (sun/moon) in every nav bar
- Preference stored in `localStorage` as `pmaxis-theme`
- **Shared across pages** — toggling on one page persists when navigating to another
- Auto-detects `prefers-color-scheme` on first visit
- Logo paths switch between white (dark) and `#0A0A0A` (light); green centre always stays `#00E676`

The `/docs` page is permanently dark (Scalar `deepSpace` theme) with its own built-in toggle.
The `/admin` page is permanently dark (internal tool, always dark).

---

## /status — Public Status Page

**File:** `pmaxis/services/api/status.html`

ChatGPT-style system status page with real uptime history.

### Features
- **Overall status banner** — pulsing green dot when all systems operational
- **90-day timeline bars** per component — green ≥99%, yellow 90–99%, red <90%, grey = no data
- **Hover tooltips** on bars showing date and uptime %
- **Live component cards** — fetches `/health` for current status
- **Auto-refreshes** every 30 seconds
- **Resource links** — Docs, Explorer, Get API Key

### Components tracked

| Name | Slug | How checked |
|------|------|-------------|
| API | `api` | Always operational (if this endpoint responds, API is up) |
| Redis Cache | `redis` | `PING` command |
| ClickHouse | `clickhouse` | `SELECT 1` |
| Postgres | `postgres` | `Ping()` |
| Data Pipeline | `pipeline` | `SELECT max(timestamp) FROM trades` — operational if < 10 min old, degraded 10–60 min, outage > 60 min |

### Uptime time-series (Redis)

The background goroutine in `hub-api` records a snapshot every 5 minutes:
- Key: `pmaxis:uptime:{slug}` (sorted set)
- Member: `{unix_ts}:{0|1}` (0 = down, 1 = up)
- Score: unix timestamp
- Retention: 90 days (old entries pruned automatically)

Endpoint: `GET /viz/data/uptime` — returns daily aggregated uptime % per component for the last 90 days.

### Data sources
- `GET /health` — current component status and latency
- `GET /viz/data/uptime` — 90-day daily history from Redis

---

## /register — API Key Registration

**File:** `pmaxis/services/api/register.html`

Two-column layout for developer onboarding.

### Features
- Left side: eyebrow tag, headline, feature list with icons
- Right side: sticky form card (App Name, Email, Use Case)
- On success: shows the full key with copy button (key shown **once only**)
- Green CTA button (`#00E676` bg, `#0A0A0A` text)
- Key rotation section: existing users can enter their current key to rotate

### Behaviour
- Calls `POST /v1/keys/register` — creates a free-tier key in Postgres + Redis
- Key is never shown again after the session — user must copy it immediately
- For rotation: calls `POST /v1/keys/rotate` with the current key in `X-API-Key` header

---

## /viz — Data Explorer

**File:** `pmaxis/services/api/viz.html`

Live market data dashboard.

### Sections
1. **Platform Stats** — total markets, trades, volume, signals (from `/viz/data/stats`)
2. **Top Markets** — sortable table of highest-volume markets (from `/viz/data/markets/top`)
3. **Candle Chart** — TradingView lightweight-charts OHLCV candles per market
4. **Orderbook** — live bid/ask depth from Redis (separate picker from candles)
5. **Recent Trades** — scrolling feed of latest trades

### Orderbook vs Candle pickers
These are **deliberately separate** because the market IDs that have live orderbook snapshots in Redis may differ from the IDs returned by the top-markets query. The orderbook picker is populated from `GET /viz/data/orderbooks/available` which scans actual Redis keys (`pmaxis:orderbook:*`) — so it only shows markets with live data.

### Auto-refresh
Live toggle in nav bar — when on, refreshes all sections every 30 seconds.

---

## /docs — API Reference

**Served from:** `main.go` (inline HTML string, not a separate file)

Interactive API documentation powered by [Scalar API Reference](https://scalar.com).

### Configuration
- Theme: `deepSpace` (permanent dark)
- Layout: `modern`
- Spec: served from `GET /openapi.json` (embedded from `openapi.json`)
- Nav bar: PMAxis branding (logo, links to Explorer / Status / Get API Key)
- No custom CSS injection — Scalar renders in its own style

### Why inline in main.go
The docs page has a strict CSP (Content-Security-Policy) that differs from other pages — it needs to allow `cdn.jsdelivr.net` for the Scalar bundle but block most other origins. Keeping it inline makes the CSP easy to maintain alongside the route handler.

### Updating the OpenAPI spec
Edit `pmaxis/services/api/openapi.json`, then rebuild and redeploy `hub-api`. The spec is embedded at compile time via `//go:embed openapi.json`.

---

## /admin — API Key Management

**File:** `pmaxis/services/api/admin.html`

Internal operator panel for managing API keys.

### Auth
Protected by `DEBUG_TOKEN` middleware — all `/admin/*` routes require `X-Debug-Token: <token>` header. The page stores the token in `localStorage` after first login so you don't re-enter it.

Access with token in URL (first time):
```
/admin?debug_token=your-token-here
```

### Features

**Dashboard stats:**
- Total keys, Active keys, Free tier count, Pro tier count

**Keys table:**
- All keys sorted by creation date (newest first)
- Columns: App Name, Email, Tier, Created, Status, Usage bar

**Detail panel (click any row):**
- Full masked key, metadata, 7-day usage breakdown
- Actions:

| Button | What it does |
|--------|-------------|
| Save Changes | Update tier / custom rate limit |
| Reset Usage | Clear current rate-limit window counter |
| Revoke Key | Sets `active=false` instantly in Redis — key stops working immediately |
| Activate Key | Re-enables a revoked key |
| Delete Permanently | **Only on revoked keys.** Removes from Postgres + cleans Redis. Cannot be undone. |

**Create Key modal:**
- App name, email, tier (free / pro / enterprise), optional custom rate limit
- Use this for internal/monitoring keys or manually provisioning enterprise customers

### Admin API endpoints

All require `X-Debug-Token: <DEBUG_TOKEN>` header.

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/admin/keys` | List all keys |
| `POST` | `/admin/keys` | Create a key (any tier) |
| `GET` | `/admin/keys/{key}` | Get key details |
| `PUT` | `/admin/keys/{key}` | Update tier / rate_limit / active |
| `DELETE` | `/admin/keys/{key}` | Permanently delete a revoked key |
| `POST` | `/admin/keys/{key}/revoke` | Revoke key immediately |
| `POST` | `/admin/keys/{key}/activate` | Re-activate a revoked key |
| `POST` | `/admin/keys/{key}/reset` | Clear rate-limit counter |
| `GET` | `/admin/keys/{key}/usage` | Usage stats (current window + 7-day daily) |
| `GET` | `/admin/keys/stats` | Aggregate dashboard stats |

---

## Embedding New Pages

To add a new embedded page:

1. Create `pmaxis/services/api/yourpage.html`
2. Add `//go:embed yourpage.html` and `var yourpageHTML []byte` in `main.go`
3. Register the route with appropriate CSP headers
4. Rebuild and redeploy `hub-api`

Use the existing pages as templates — all share the same CSS variables, logo SVG, nav structure, and theme toggle pattern.

---

## Related Docs

- [../api/key-management.md](../api/key-management.md) — Full key management guide
- [../deployment/vps-deploy.md](../deployment/vps-deploy.md) — How to rebuild and deploy hub-api
- [../operations/monitoring.md](../operations/monitoring.md) — Monitoring and alerting setup
