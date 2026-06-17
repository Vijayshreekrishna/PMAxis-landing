# PredSX — Uptime Kuma Status Dashboard

> Uptime Kuma is the visual monitoring and status-page system for PredSX. It runs as a Docker container inside the same network as all other services, pings each one every 60 seconds, keeps history, and alerts on Discord when anything goes down.

**Status: Live as of June 17, 2026 — all 12 monitors green.**

---

## Overview

| Property | Value |
|---|---|
| Image | `louislam/uptime-kuma:1` |
| Container name | `package-uptime-kuma-1` |
| Dashboard URL | `http://167.233.97.217:3001` |
| Public status page | `http://167.233.97.217:3001/status/predsx` |
| Internal port | `3001` |
| Storage | Named Docker volume `package_uptime_kuma_data` (~50–100 MB/year) |
| RAM usage | ~60–80 MB steady state |
| Check interval | 60 seconds per monitor |
| VPS compose path | `/home/vijay/predsx/package/docker-compose.yml` |

Uptime Kuma can reach all PredSX services directly by their Docker service names (same internal network), so monitors point to internal hostnames like `hub-api:8088` rather than the public IP — faster, more reliable, no firewall concerns.

---

## Part 1 — Deploy to VPS

### Step 1 — Push the updated docker-compose

On your local machine, ship the updated `docker-compose.yml` to VPS:

```bash
scp -i ~/.ssh/id_vsk predsx/deployments/package/docker-compose.yml vijay@167.233.97.217:/home/vijay/predsx/package/docker-compose.yml
```

### Step 2 — Pull and start Uptime Kuma

SSH into VPS, then:

```bash
ssh -i ~/.ssh/id_vsk vijay@167.233.97.217
cd /home/vijay/predsx/package
docker compose pull uptime-kuma
docker compose up -d uptime-kuma
```

### Step 3 — Verify it started

```bash
docker compose ps uptime-kuma
docker compose logs uptime-kuma --tail=20
```

Expected: container is `Up`, logs show `Listening on 3001`.

### Step 4 — Open the dashboard

Navigate to `http://167.233.97.217:3001` in your browser.

On first visit, Uptime Kuma prompts you to create an admin username and password. Set these once — they are stored in the `uptime_kuma_data` volume and persist across restarts.

---

## Part 2 — Configure Monitors

After logging in, click **Add New Monitor** for each entry below.

### Infrastructure Monitors (TCP)

These check that the port is open and accepting connections.

| Name | Type | Host | Port |
|---|---|---|---|
| ClickHouse | TCP Port | `clickhouse` | `9000` |
| Redis | TCP Port | `redis` | `6379` |
| Kafka | TCP Port | `kafka` | `9092` |
| PostgreSQL | TCP Port | `postgres` | `5432` |

**How to add a TCP monitor:**
1. Click **Add New Monitor**
2. Monitor Type → **TCP Port**
3. Friendly Name → e.g. `ClickHouse`
4. Hostname → e.g. `clickhouse`
5. Port → e.g. `9000`
6. Heartbeat Interval → `60` seconds
7. Click **Save**

> **Note:** Hostnames are case-sensitive. Use exactly `postgres` (all lowercase) — `postgreSQL` or `Postgres` will fail with `ENOTFOUND`.

---

### Service Health Monitors (HTTP)

These call each service's `/health` endpoint and expect HTTP 200.

| Name | Type | URL |
|---|---|---|
| Hub API | HTTP | `http://hub-api:8088/health` |
| Hub Discovery | HTTP | `http://hub-discovery:8081/health` |
| Hub Ingestion | HTTP | `http://hub-ingestion:8082/health` |
| Hub Processor | HTTP | `http://hub-processor:8083/health` |
| Hub Storage | HTTP | `http://hub-storage:8084/health` |

**How to add an HTTP monitor:**
1. Click **Add New Monitor**
2. Monitor Type → **HTTP(s)**
3. Friendly Name → e.g. `Hub API`
4. URL → e.g. `http://hub-api:8088/health`
5. Heartbeat Interval → `60` seconds
6. Expected Status Code → `200`
7. Click **Save**

---

### API Endpoint Monitors (HTTP)

These verify that key public endpoints return valid data, not just that the server is up.

| Name | Type | URL |
|---|---|---|
| API — Markets | HTTP | `http://hub-api:8088/v1/markets` |
| API — Categories | HTTP | `http://hub-api:8088/v1/categories` |
| API — Tags | HTTP | `http://hub-api:8088/v1/tags` |

Set up the same way as service health monitors above.

---

### Total: 12 monitors

| Group | Count |
|---|---|
| Infrastructure (TCP) | 4 |
| Service health (/health) | 5 |
| API endpoints | 3 |
| **Total** | **12** |

---

## Part 3 — Configure Discord Alerts

Uptime Kuma can send alerts to the same Discord webhook used by the existing cron-based health script.

### Step 1 — Add Discord notification

1. In Uptime Kuma, click your **profile icon (top right)** → **Notifications**
2. Click **Add Notification**
3. Notification Type → **Discord**
4. Friendly Name → `PredSX Discord`
5. Webhook URL → paste your Discord webhook URL (same one from `predsx-health.sh`)
6. Click **Test** — you should receive a test message in your Discord channel
7. Click **Save**

### Step 2 — Attach to all monitors

When adding or editing each monitor:
- Scroll to **Notifications** section
- Toggle on `PredSX Discord`

Now any monitor going down sends an immediate Discord alert, and another when it recovers.

---

## Part 4 — Status Page (Public)

Uptime Kuma can generate a public-facing status page showing current status and incident history, without exposing the admin UI.

### Create a status page

1. Click **Status Page** in the left sidebar
2. Click **New Status Page**
3. Title → `PredSX Status`
4. Content → `Real-time status of all PredSX services`
5. Slug → `predsx`
6. Click **Create**
7. On the editor, click **Add Group** → name it `Services`
8. Click **+** inside the group → add all 12 monitors
9. Click **Save**

Public URL (no login required):
```
http://167.233.97.217:3001/status/predsx
```

Share this URL with anyone who needs to check service health without admin access.

---

## Part 5 — Maintenance

All commands must be run from `/home/vijay/predsx/package` on the VPS.

### View logs

```bash
docker compose logs uptime-kuma --tail=50
```

### Restart Uptime Kuma only

```bash
docker compose restart uptime-kuma
```

### Upgrade Uptime Kuma

```bash
docker compose pull uptime-kuma
docker compose up -d --no-deps uptime-kuma
```

Data in `package_uptime_kuma_data` volume is preserved across upgrades.

### Check volume size

```bash
docker system df -v | grep uptime_kuma_data
```

### Backup Uptime Kuma data

```bash
docker run --rm -v package_uptime_kuma_data:/data -v $(pwd):/backup alpine \
  tar czf /backup/uptime-kuma-backup.tar.gz -C /data .
```

---

## Part 6 — Resource Usage

| Resource | Expected |
|---|---|
| RAM | 60–80 MB steady state |
| CPU | < 1% (pings are tiny HTTP/TCP checks) |
| Disk (SQLite) | ~50–100 MB per year at 12 monitors × 60s intervals |

No impact on API, Kafka, or ClickHouse performance.

---

## Relationship with Existing Monitoring

PredSX runs two complementary monitoring systems:

| System | What it does | Where alerts go |
|---|---|---|
| **Uptime Kuma** | Monitors each service individually, tracks response-time history, shows visual status page, alerts instantly on downtime | Discord |
| **Cron script** (`predsx-health.sh`) | Checks OS-level metrics: disk %, RAM %, swap %, stopped containers — things Uptime Kuma cannot see | Discord |

They are independent. Both should be running.

---

## Related Docs

- [monitoring.md](monitoring.md) — Cron-based Discord alert system (OS-level metrics)
- [vps-deploy.md](vps-deploy.md) — Full VPS deployment guide
- [port-security.md](port-security.md) — Port bindings and UFW rules
- [troubleshooting.md](troubleshooting.md) — Common errors and fixes
