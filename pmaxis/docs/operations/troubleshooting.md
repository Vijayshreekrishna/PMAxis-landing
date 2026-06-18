# PMAxis — Troubleshooting Guide

> Common errors encountered in development and production, with root causes and exact fixes.
> All issues here are real incidents from this project — not hypothetical.

---

## 1. Hub services crash-loop on startup — Kafka not ready

**Symptom:**
```
hub-discovery exited with code 1
hub-ingestion exited with code 1
...repeated restart loop
```

Container logs show:
```
failed to connect to kafka: dial tcp: connection refused
```

**Root cause:** Docker starts containers in the order defined in `docker-compose.yml`, but "started" does not mean "ready". Kafka takes 10–20 seconds after its container starts before it accepts connections. Hub services that try to connect immediately fail and exit.

**Fix:** All hub services use `condition: service_healthy` on the Kafka dependency in `docker-compose.yml`:
```yaml
depends_on:
  kafka:
    condition: service_healthy
```

Kafka's healthcheck runs `kafka-topics --list` every 15 seconds and must pass before any hub is allowed to start. This eliminates the crash loop entirely.

**If you see this anyway:** Check that Kafka's healthcheck is actually passing — `docker ps` will show `(health: starting)` or `(unhealthy)` on the `package-kafka-1` row. Wait for it to show `(healthy)` before expecting hubs to start.

---

## 2. `docker save >` produces a corrupted tar file

**Symptom:** After running `docker save pmaxis-api:latest > api.tar` in PowerShell, loading on VPS fails:
```
docker load -i api.tar
open api.tar: invalid argument
```
or the file loads silently but images are missing.

**Root cause:** PowerShell's `>` redirect operator writes files in **UTF-16 LE with a BOM**. Docker `.tar` files are binary. The UTF-16 encoding corrupts the binary content — the BOM prepended to the file makes it unreadable as a tar archive.

**Fix:** Always use the `-o` flag instead of `>`:
```powershell
# WRONG — corrupts binary
docker save pmaxis-api:latest > api.tar

# CORRECT — writes raw binary
docker save pmaxis-api:latest -o api.tar
```

The `build-and-ship.bat` script in `pmaxis/` already uses `-o` for all 5 images. Only use `>` if you are writing text output.

---

## 3. SSH key rejected — `Permission denied (publickey,password)`

**Symptom:** Running `scp` or `ssh` with a specific key name fails:
```
Permission denied (publickey,password).
```

**Root cause:** The key name passed with `-i` does not match any key that the VPS accepts. You may have multiple SSH keys in `~/.ssh/` and the correct one has a different filename than expected.

**Fix:** List your available SSH keys first:
```powershell
ls $env:USERPROFILE\.ssh\
```

Identify the correct key (typically `id_ed25519` or `id_rsa`) and use that exact filename:
```powershell
# Check which key works
ssh -i $env:USERPROFILE\.ssh\id_ed25519 user@YOUR_VPS_IP "echo ok"
```

Once you identify the working key, use it consistently across all `scp` and `ssh` commands. The VPS only accepts keys that have been added to `~/.ssh/authorized_keys` on the server side.

---

## 4. ClickHouse queries fail — "database pmaxis does not exist"

**Symptom:** Manual ClickHouse queries or debugging tools fail with:
```
DB::Exception: Database pmaxis doesn't exist
```

**Root cause:** The ClickHouse schema is initialized by `services/storage/main.go:ensureSchemas()` using `CREATE TABLE IF NOT EXISTS default.market_metadata ...`. All tables live in the **`default`** database, not a `pmaxis` database. The `deployments/clickhouse/init.sql` file references `pmaxis` but is a stale copy and is not mounted by Docker — it has no effect.

**Fix:** Always use `default` as the database in manual queries:
```sql
-- WRONG
SELECT * FROM pmaxis.market_metadata LIMIT 5;

-- CORRECT
SELECT * FROM default.market_metadata LIMIT 5;
```

When connecting via `clickhouse-client` inside the container:
```bash
docker exec -it package-clickhouse-1 clickhouse-client --database default
```

**Table list:**
```sql
SHOW TABLES FROM default;
```

---

## 5. Containers show `package-*` prefix instead of `pmaxis-*`

**Symptom:** After running `docker compose up`, containers are named:
```
package-hub-api-1
package-hub-discovery-1
package-kafka-1
...
```
instead of the expected `pmaxis-hub-api-1` or `pmaxis-kafka-1`.

**Root cause:** Docker Compose derives the project name from the **directory name** of the compose file, not from any field in the compose file itself. The compose files live in `~/pmaxis/package/`, so Docker uses `package` as the project name and prefixes all container and volume names with `package-`.

**This is correct and expected behaviour.** All commands in the documentation use `package-*` names. Do not try to rename the directory — it will break volume references.

When referencing containers manually:
```bash
docker exec -it package-clickhouse-1 clickhouse-client
docker logs -f package-hub-ingestion-1
docker compose logs hub-api    # compose command — no prefix needed
```

---

## 6. `/v1/markets/{id}/price` returns `"price not found"` for all markets

**Symptom:** Every call to `/v1/markets/{id}/price` returns:
```json
"price not found"
```
even for actively traded markets.

**Root cause:** Redis key mismatch between the writer and the reader.
- **Processor hub** writes live prices to Redis key: `live:price:<marketID>`
- **API `GetPrice` handler** was reading from: `pmaxis:price:<marketID>`

These two keys never match, so the API always returns 404.

**Fix:** In `services/api/rest.go`, the Redis GET key was changed from `pmaxis:price:` to `live:price:` to match the processor's write key.

**How to verify the key is present:**
```bash
docker exec -it package-redis-1 redis-cli KEYS "live:price:*" | head -5
docker exec -it package-redis-1 redis-cli GET "live:price:<your-market-id>"
```

If the key exists in Redis but the API still returns 404, pull the latest `hub-api` image — the fix may not be deployed yet.

---

## 7. `docker compose restart` does not pick up a new image

**Symptom:** After loading a new `.tar` file and running `docker compose restart hub-api`, the container still runs the old image.

**Root cause:** `docker compose restart` sends `SIGTERM` + `SIGSTART` to the existing container. It does **not** recreate the container from the image. The container continues to use whatever image layer it was originally created from, even if a newer image with the same tag is now loaded.

**Fix:** Use `up -d --no-deps` to recreate the container from the new image:
```bash
docker load -i api.tar
docker compose up -d --no-deps hub-api
```

`--no-deps` ensures only `hub-api` is recreated and no other dependent services are restarted unnecessarily.

---

## 8. `event_id` is empty — `/related` returns `[]` for all markets

**Symptom:** After boot, most markets have `event_id: ""` in ClickHouse. `/v1/markets/{id}/related` always returns `[]`.

**Root cause:** Two compounding issues:
1. The Gamma `/markets` endpoint omits `eventId` on many markets.
2. `discoverEvents` was running *after* `discoverMarkets` finished — which took 30–60+ minutes depending on the number of active markets.

**Fix:** Both discovery loops now run concurrently with `sync.WaitGroup`. `discoverEvents` paginates the Gamma `/events` endpoint, which guarantees `event_id` on all nested markets by backfilling from the parent event.

**How to verify event_id is being populated:**
```bash
docker exec -it package-clickhouse-1 clickhouse-client --query \
  "SELECT market_id, event_id FROM default.market_metadata WHERE event_id != '' LIMIT 5"
```

If this returns rows within ~60 seconds of boot, event discovery is working. If `event_id` is still empty after 2+ minutes, check hub-discovery logs:
```bash
docker compose logs hub-discovery --tail=50
```

Look for lines like:
```json
{"msg":"fetched events from gamma","count":100,"offset":0}
{"msg":"published event market to kafka","id":"558935","event_id":"30615"}
```

---

## 9. `docker compose up` fails — image not found

**Symptom:**
```
Error response from daemon: No such image: pmaxis-api:latest
```

**Root cause:** The `.tar` files were not loaded before running `docker compose up`. The compose file references `pmaxis-api:latest` but that image doesn't exist in the local Docker daemon on the VPS.

**Fix:** Always load images before starting the stack:
```bash
ls *.tar | xargs -I {} docker load -i {}
docker compose up -d
```

Verify images are loaded:
```bash
docker images | grep pmaxis
```

All 5 images (`pmaxis-api`, `pmaxis-discovery`, `pmaxis-ingestion`, `pmaxis-processor`, `pmaxis-storage`) must appear before running `docker compose up`.

---

## 11. ClickHouse OOM on startup — orphaned store directories

**Symptom:** ClickHouse container restarts repeatedly. Logs show:
```
Memory limit (total) exceeded: ... maximum: 1.50 GiB
```
Even after restart, ClickHouse cannot accept connections. `docker stats` shows RSS near the container's memory limit before any queries run.

**Root cause:** Every time a ClickHouse table is dropped and recreated (during development or schema migrations), the old data directory — named by a UUID — is left on disk under `/var/lib/clickhouse/store/`. ClickHouse does not automatically delete these. On startup, ClickHouse loads metadata for all directories in `store/`, including orphaned ones, consuming significant memory before any queries are run.

**Diagnosis:**
```bash
# Enter the volume with Alpine (while clickhouse is stopped)
docker stop package-clickhouse-1
docker run --rm -it -v package_clickhouse_data:/data alpine sh

# Size of store directories
du -sh /data/store/*/*/ 2>/dev/null | sort -rh | head -20

# Live table UUIDs (from metadata)
ls /data/metadata/default/

# Compare — any UUID in /data/store/ not referenced by a .sql file in /data/metadata/default/ is orphaned
```

**Fix:**
```bash
# Inside the Alpine container
rm -rf /data/store/<orphaned-uuid>/
# Repeat for each orphaned directory
exit

docker start package-clickhouse-1
```

After cleanup, restart ClickHouse. RSS will be a fraction of the container limit and it will start accepting connections normally.

**Prevention:** Run the store size query periodically (see [clickhouse-queries.md](clickhouse-queries.md)). If `store/` is significantly larger than the sum of live table data, orphaned dirs have accumulated.

---

## 12. `/v1/categories` and `/v1/tags` return `[]`

**Symptom:** Both endpoints return an empty array even though `market_metadata` has 100k+ rows.

**Verification:**
```bash
# Check if category is actually populated in any rows
docker exec -it package-clickhouse-1 clickhouse-client \
  --query "SELECT count() FROM default.market_metadata WHERE category != ''"
# Returns 0
```

**Root cause:** Polymarket Gamma API returns `"tags": null` on individual market objects (`/markets` endpoint). The discovery service used `m.Tags` (market-level) for `extractTagSlugs()` and `extractCategory()`, which always returned nil/empty. Tags exist at the **parent event level** in the Gamma `/events` response, but `PolymarketEvent` struct was missing a `Tags` field — so event-level tags were never captured.

**Fix:** In `services/discovery/main.go`:
1. Added `Tags []PolymarketTag \`json:"tags"\`` to `PolymarketEvent` struct
2. In `discoverEvents`, fall back to `ev.Tags` when the market's own `m.Tags` is empty

After deploying the fixed image and waiting one poll cycle (~60s), categories and tags begin populating.

**Note:** `series` is unaffected — it comes from `groupItemTitle` which Polymarket does populate. If tags are still empty after a fix deploy, confirm the fix is deployed: `docker compose logs hub-discovery --tail=20 | grep "event market"`.

---

## 13. Debug endpoints return "debug endpoints disabled"

**Symptom:**
```
GET /debug/markets
→ 403 debug endpoints disabled
```

**Root cause:** The `/debug/*` routes and `/metrics` are gated by `middleware.DebugAuth`, which checks for `DEBUG_TOKEN` env var. If `DEBUG_TOKEN` is not set, the middleware rejects all requests with 403.

**Fix:** Add `DEBUG_TOKEN` to the `.env` file and restart `hub-api`:
```bash
echo "DEBUG_TOKEN=your-secret-token" >> ~/pmaxis/package/.env
cd ~/pmaxis/package && docker compose up -d --no-deps hub-api
```

**Usage:** Pass the token in requests:
```bash
# As a header
curl -H "X-Debug-Token: your-secret-token" http://YOUR_VPS_IP:8088/debug/markets

# As a query param
curl "http://YOUR_VPS_IP:8088/debug/markets?debug_token=your-secret-token"
```

**Security note:** Do not use a guessable token. The debug endpoints expose raw ClickHouse data and orderbook state — they should not be accessible without authentication.

---

## 14. Disk fills up steadily — ClickHouse system log tables

**Symptom:**
Disk usage climbs 1–2% per hour even though your actual market/trade data is small.
`docker system df` shows Local Volumes at 10–15 GB. ClickHouse table query reveals the
culprit:

```bash
docker exec -it package-clickhouse-1 clickhouse-client --query \
  "SELECT table, formatReadableSize(sum(bytes)) AS size FROM system.parts WHERE active GROUP BY table ORDER BY sum(bytes) DESC"
```

Output shows `text_log`, `query_log`, `trace_log` etc. at several GB while your actual
tables (`trades`, `onchain_trades`, `market_metadata`) are under 200 MB total.

**Root cause:** ClickHouse logs every internal operation into its own system tables by
default with **no expiry**. On a large server this is negligible; on a 4 GB VPS it
eventually fills the disk.

**Immediate fix — truncate now:**
```bash
docker exec -it package-clickhouse-1 clickhouse-client --query "
  TRUNCATE TABLE system.text_log;
  TRUNCATE TABLE system.trace_log;
  TRUNCATE TABLE system.part_log;
  TRUNCATE TABLE system.query_log;
  TRUNCATE TABLE system.processors_profile_log;
  TRUNCATE TABLE system.metric_log;
  TRUNCATE TABLE system.asynchronous_insert_log;
  TRUNCATE TABLE system.query_views_log;
  TRUNCATE TABLE system.asynchronous_metric_log;
  TRUNCATE TABLE system.error_log;
  TRUNCATE TABLE system.background_schedule_pool_log;
  TRUNCATE TABLE system.query_metric_log;
"
```

**Permanent fix:** `deployments/clickhouse-logging.xml` is mounted into the ClickHouse
container and sets short TTLs on all system tables (1 day for high-volume tables like
`trace_log` and `processors_profile_log`, 2 days for `query_log`/`part_log`, 3 days
for the rest) plus reduces `text_log` to `warning` level. This is already wired into
`docker-compose.yml`. If deploying fresh, the truncation above is still needed once to
clear any bloat that accumulated before the config was applied.

See [vps-deploy.md — ClickHouse System Log Auto-Deletion](../deployment/vps-deploy.md) for full details.

---

## 10. Hub shows `(unhealthy)` in `docker ps`

**Symptom:**
```
package-hub-storage-1   Up 5 minutes (unhealthy)
```

**Root cause (most common):** The `METRICS_PORT` env var for that hub does not match the port the healthcheck is probing. `BaseService` starts an HTTP `/metrics` server on `METRICS_PORT` (default `8080`). If the compose file's healthcheck calls a different port (e.g. `8083`), every check returns connection refused.

**Fix:** Ensure each hub in `docker-compose.yml` sets a unique `METRICS_PORT` that matches its healthcheck:
```yaml
environment:
  - METRICS_PORT=8083
healthcheck:
  test: ["CMD", "wget", "-qO-", "http://localhost:8083/health"]
```

**Diagnosis:** Check the hub's log for the metrics server startup line:
```bash
docker compose logs hub-storage --tail=30 | grep "metrics"
```

It should show which port the metrics server started on.
