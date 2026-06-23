# API Key Management Guide

This covers everything about API keys — how users get them, how you manage them as admin, and how to maintain the system over time.

---

## Pages at a Glance

| URL | Who uses it | Auth required |
|-----|------------|--------------|
| `/register` | Developers signing up | None (public) |
| `/docs` | Anyone reading API docs | None (public) |
| `/admin` | You (operator) | `DEBUG_TOKEN` |

---

## User Flow: Getting a Key

1. Developer visits `http://167.233.97.217:8088/register`
2. Fills in **App Name** + **Email** (use_case optional)
3. Clicks submit — receives a `pmx_live_...` key on screen
4. **The key is shown once.** They must save it immediately.

From then on they pass it in every API request:
```
X-API-Key: pmx_live_their_key_here
```

Or for WebSocket:
```
ws://167.233.97.217:8088/stream?api_key=pmx_live_their_key_here
```

### Key rotation (user-initiated)
If a user loses their key or suspects it's leaked, they can rotate it themselves:

```bash
curl -X POST http://167.233.97.217:8088/v1/keys/rotate \
  -H "X-API-Key: pmx_live_their_current_key"
```

The old key dies immediately. The response contains the new key (shown once).  
They can also rotate via the `/register` page's "Rotate Key" button.

---

## Admin Panel: `/admin`

Access at `http://167.233.97.217:8088/admin`

### First-time login
The page asks for your **Admin Token** — this is the `DEBUG_TOKEN` value set in the VPS `.env`.

The token is saved in `localStorage` so you only type it once per browser. To log out, clear localStorage or open the URL with `?debug_token=<token>` to override.

### Dashboard stats
Four top-line numbers shown on load:
- **Total Keys** — all keys ever created (including inactive)
- **Active Keys** — currently usable keys
- **Free Tier** — count of free-tier keys
- **Pro Tier** — count of pro-tier keys

### Keys table
Shows all keys sorted by creation date (newest first). Columns:
- App Name / Email
- Tier badge (Free / Pro / Enterprise)
- Created date
- Status badge (Active / Revoked)
- Usage progress bar (current minute window / rate limit)

Click any row to open the **detail panel**.

### Detail panel (click a key row)
Shows full key info:
- Masked key (`pmx_live_abc***xyz4`)
- App name, email, tier, created date, active status
- **Lifetime & Daily Usage block** — total requests ever + 7-day bar chart

Actions available in the panel:
| Button | What it does |
|--------|-------------|
| Change Tier | Upgrade/downgrade tier; takes effect on next request |
| Set Custom Rate Limit | Override the tier default (0 = use tier default) |
| Revoke | Sets `active=false` instantly in Redis — key stops working immediately |
| Activate | Re-enables a revoked key |
| Reset Usage | Clears the current rate-limit window counter (does NOT reset daily/total stats) |

### Creating a key manually
Click **+ New Key** in the top-right. Fill in app name, email, tier. The full key is returned in the API response — copy it from the browser console or the toast notification.

Use this for:
- Creating monitoring keys (for Uptime Kuma)
- Creating your own internal keys
- Provisioning pro/enterprise keys for paying customers

---

## Admin API Endpoints

All require `X-Debug-Token: pmaxis-debug-2026` header.

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/admin/keys` | List all keys |
| `POST` | `/admin/keys` | Create a key (any tier) |
| `GET` | `/admin/keys/{key}` | Get key details |
| `PATCH` | `/admin/keys/{key}` | Update tier / rate_limit / active |
| `POST` | `/admin/keys/{key}/revoke` | Revoke key |
| `POST` | `/admin/keys/{key}/activate` | Re-activate key |
| `GET` | `/admin/keys/{key}/usage` | Usage stats (current window + 7-day daily) |
| `POST` | `/admin/keys/{key}/reset-usage` | Clear rate-limit counter |
| `GET` | `/admin/stats` | Aggregate dashboard stats |

### Upgrade a key to Pro

```bash
curl -X PATCH http://167.233.97.217:8088/admin/keys/pmx_live_... \
  -H "X-Debug-Token: pmaxis-debug-2026" \
  -H "Content-Type: application/json" \
  -d '{"tier": "pro"}'
```

### Revoke a key immediately

```bash
curl -X POST http://167.233.97.217:8088/admin/keys/pmx_live_.../revoke \
  -H "X-Debug-Token: pmaxis-debug-2026"
```

---

## Uptime Kuma Monitoring

The monitoring key must be added as a header in Kuma's HTTP monitor settings.  
Create a dedicated key via admin, then add to each monitor:

**Headers field (must be valid JSON):**
```json
{"X-API-Key": "pmx_live_your_monitoring_key"}
```

Endpoints to monitor:
- `GET /health` — no key needed
- `GET /v1/markets?limit=1` — needs key
- `GET /v1/categories` — needs key
- `GET /v1/tags` — needs key

---

## Ongoing Maintenance

### What to do regularly

**Weekly:**
- Open `/admin` and scan for keys with suspicious usage patterns (e.g. hitting 100% rate limit every minute all day)
- Check total key count growth

**When a user reports their key isn't working:**
1. Go to `/admin`, find their key by email
2. Check status badge — may have been auto-revoked (not currently implemented) or manually revoked
3. Click Activate if it was accidentally revoked
4. If lost/leaked, tell them to call `POST /v1/keys/rotate` with their current key, or use "Reset Usage" if they just got rate-limited

**When you want to add a Pro user:**
1. `/admin` → find their key → click row → Change Tier → Pro → Save
2. Change takes effect on the next API request (Redis is updated immediately)

### Redis key structure
```
apikey:{key}        → JSON: {tier, active, rate_limit}   (permanent)
rl:{key}            → integer counter, TTL ~60s           (rate limit window)
usage:total:{key}   → integer (lifetime request count)    (permanent)
usage:daily:{key}:YYYYMMDD → integer (daily count)       (30-day TTL)
```

If Redis is wiped or a key is missing, the auth middleware returns 401 until the key is re-synced. To re-sync all keys from Postgres into Redis:

```bash
# SSH into VPS then exec into hub-api
ssh -i ~/.ssh/id_ed25519 vijay@167.233.97.217
docker exec -it package-hub-api-1 sh

# Or call the admin endpoint for each key to trigger a re-sync
# (re-sync happens automatically on PATCH /admin/keys/{key})
```

In practice, restarting hub-api is the fastest fix — `MigrateAPIKeys` runs on boot but does not re-seed Redis. The simplest recovery is:

```bash
# On VPS
cd ~/pmaxis/package
COMPOSE_PROJECT_NAME=package docker compose restart hub-api
# Then re-activate each key via /admin panel which writes back to Redis
```

### Rotating the DEBUG_TOKEN (admin token)
If you need to change the admin token:
1. Edit `.env` on the VPS: `DEBUG_TOKEN=new_value_here`
2. Restart hub-api: `COMPOSE_PROJECT_NAME=package docker compose up -d --no-deps --force-recreate hub-api`
3. Update Uptime Kuma monitors that use the debug token
4. Update your browser localStorage: clear `pmx_admin_token` and re-login at `/admin`

---

## What's Not Built Yet (Future)

- Email on registration (removed for now — can be added back)
- Pro upgrade self-serve flow (currently manual via admin panel)
- Key expiry / auto-revocation after N days of inactivity
- Usage webhooks / alerts when a key approaches its rate limit
