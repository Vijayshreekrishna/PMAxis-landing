# PMAxis — VPS Monitoring & Alerting Guide

> Covers how to set up, manage, and maintain the Discord alert system and cron job that monitors the PMAxis VPS health.

---

## Overview

An hourly cron job runs on the VPS and checks:
- Disk usage
- Container health (unhealthy or exited containers)
- Memory usage
- Swap usage

If anything crosses the threshold, a Discord message is sent instantly. If everything is healthy, no message is sent.

---

## Part 1 — Discord Webhook Setup

### Step 1 — Create a Discord Server & Channel

1. Open Discord → click **+** to create a new server (or use an existing one)
2. Create a new text channel e.g. `#vps-alerts`

### Step 2 — Create a Webhook

1. Click the **gear icon** next to the `#vps-alerts` channel → **Integrations** → **Webhooks**
2. Click **New Webhook**
3. Give it a name: `PMAxis Monitor`
4. Click **Copy Webhook URL** — it looks like:
   ```
   https://discord.com/api/webhooks/1234567890/xxxxxxxxxxxxxxxxxxxx
   ```
5. Save this URL — you will need it in the next step

---

## Part 2 — Creating the Monitoring Script

SSH into your VPS, then run this command to create the script:

```bash
python3 << 'PYEOF'
script = r"""#!/bin/bash

WEBHOOK="YOUR_DISCORD_WEBHOOK_URL"
THRESHOLD_DISK=80
THRESHOLD_MEM=85
ALERTS=""

DISK=$(df / | awk 'NR==2 {print $5}' | tr -d '%')
if [ "$DISK" -gt "$THRESHOLD_DISK" ]; then
  ALERTS="${ALERTS}[DISK] Usage is ${DISK}%  "
fi

UNHEALTHY=$(docker ps --filter "health=unhealthy" --format "{{.Names}}" | tr '\n' ', ')
EXITED=$(docker ps -a --filter "status=exited" --format "{{.Names}}" | tr '\n' ', ')
if [ -n "$UNHEALTHY" ]; then
  ALERTS="${ALERTS}[UNHEALTHY] ${UNHEALTHY}  "
fi
if [ -n "$EXITED" ]; then
  ALERTS="${ALERTS}[EXITED] ${EXITED}  "
fi

MEM=$(free | awk '/^Mem:/ {printf "%.0f", $3/$2*100}')
SWAP=$(free | awk '/^Swap:/ {if ($2>0) printf "%.0f", $3/$2*100; else print 0}')
if [ "$MEM" -gt "$THRESHOLD_MEM" ]; then
  ALERTS="${ALERTS}[MEMORY] ${MEM}%  "
fi
if [ "$SWAP" -gt 50 ]; then
  ALERTS="${ALERTS}[SWAP] ${SWAP}%  "
fi

if [ -n "$ALERTS" ]; then
  MSG="PMAxis Alert $(date '+%Y-%m-%d %H:%M:%S'): ${ALERTS}"
  curl -s -X POST "$WEBHOOK" \
    -H "Content-Type: application/json" \
    -d "{\"content\": \"${MSG}\"}"
fi
"""
with open('/usr/local/bin/pmaxis-health.sh', 'w') as f:
    f.write(script)
print("Done")
PYEOF
```

Then make it executable:

```bash
sudo chmod +x /usr/local/bin/pmaxis-health.sh
```

Then update the webhook URL inside the script:

```bash
sudo sed -i 's|YOUR_DISCORD_WEBHOOK_URL|https://discord.com/api/webhooks/YOUR_ID/YOUR_TOKEN|' /usr/local/bin/pmaxis-health.sh
```

---

## Part 3 — Testing the Script

### Test the webhook directly (confirm Discord works)

```bash
curl -s -X POST "YOUR_DISCORD_WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -d '{"content": "PMAxis test alert - Discord webhook is working!"}'
```

You should receive a message in your Discord channel immediately.

### Test the full script (force a disk alert)

Temporarily lower the disk threshold to 1% so it always triggers:

```bash
sudo sed -i 's/THRESHOLD_DISK=80/THRESHOLD_DISK=1/' /usr/local/bin/pmaxis-health.sh
sudo bash /usr/local/bin/pmaxis-health.sh
```

You should receive a Discord alert like:
```
PMAxis Alert 2026-06-16 17:08:09: [DISK] Usage is 57%
```

Reset the threshold back to 80 after testing:

```bash
sudo sed -i 's/THRESHOLD_DISK=1/THRESHOLD_DISK=80/' /usr/local/bin/pmaxis-health.sh
```

---

## Part 4 — Setting Up the Cron Job

Add the cron job to run the script every hour:

```bash
(crontab -l 2>/dev/null; echo "0 * * * * /usr/local/bin/pmaxis-health.sh") | crontab -
```

Verify it was added:

```bash
crontab -l
```

Expected output:
```
0 * * * * /usr/local/bin/pmaxis-health.sh
```

### Cron schedule explained

```
0 * * * *
│ │ │ │ │
│ │ │ │ └─ day of week (0-7)
│ │ │ └─── month (1-12)
│ │ └───── day of month (1-31)
│ └─────── hour (0-23)
└───────── minute (0-59)
```

`0 * * * *` = runs at minute 0 of every hour (hourly).

---

## Part 5 — Alert Thresholds

| Variable | Default | Meaning |
|---|---|---|
| `THRESHOLD_DISK` | 80 | Alert if disk usage exceeds 80% |
| `THRESHOLD_MEM` | 85 | Alert if RAM usage exceeds 85% |
| Swap | 50 (hardcoded) | Alert if swap usage exceeds 50% |

To change thresholds:

```bash
sudo nano /usr/local/bin/pmaxis-health.sh
```

Edit the values at the top of the file:
```bash
THRESHOLD_DISK=80
THRESHOLD_MEM=85
```

---

## Part 6 — Maintenance

### Update the Discord webhook URL

If you regenerate the webhook in Discord:

1. Go to Discord → channel gear → Integrations → Webhooks → Regenerate
2. Copy the new URL
3. Update the script:

```bash
sudo sed -i 's|https://discord.com/api/webhooks/OLD_ID/OLD_TOKEN|https://discord.com/api/webhooks/NEW_ID/NEW_TOKEN|' /usr/local/bin/pmaxis-health.sh
```

Or open manually:
```bash
sudo nano /usr/local/bin/pmaxis-health.sh
```

### View the current script

```bash
cat /usr/local/bin/pmaxis-health.sh
```

### Run the script manually at any time

```bash
sudo bash /usr/local/bin/pmaxis-health.sh
```

### View all cron jobs

```bash
crontab -l
```

### Remove the cron job

```bash
crontab -e
```

Delete the line with `pmaxis-health.sh`, save and exit.

### Change how often it runs

Edit the cron job:

```bash
crontab -e
```

Replace `0 * * * *` with one of these:

| Schedule | Cron expression |
|---|---|
| Every 30 minutes | `*/30 * * * *` |
| Every hour | `0 * * * *` |
| Every 6 hours | `0 */6 * * *` |
| Every day at 9am | `0 9 * * *` |

---

## Part 7 — Alert Reference

| Alert prefix | Meaning | Action |
|---|---|---|
| `[DISK]` | Disk usage above 80% | Run `sudo find /var/lib/docker/containers -name "*.log" -exec truncate -s 0 {} \;` to free logs |
| `[UNHEALTHY]` | A container failed its healthcheck | Run `docker compose logs <container-name> --tail=50` to diagnose |
| `[EXITED]` | A container stopped unexpectedly | Run `docker compose up -d` to restart it |
| `[MEMORY]` | RAM usage above 85% | Run `docker stats --no-stream` to find which container is using the most |
| `[SWAP]` | Swap above 50% — memory pressure building | Check `docker stats`, may need to restart a heavy container |

---

---

## Part 8 — Uptime Kuma (Visual Status Dashboard)

In addition to the cron script, PMAxis runs **Uptime Kuma** — a self-hosted visual monitoring dashboard at `http://167.233.97.217:3001`.

| Feature | Cron Script | Uptime Kuma |
|---|---|---|
| OS metrics (disk, RAM, swap) | ✅ | ❌ |
| Per-service uptime monitoring | ❌ | ✅ |
| Response-time graphs & history | ❌ | ✅ |
| Public status page | ❌ | ✅ |
| Discord alerts | ✅ | ✅ |

Both systems run independently and complement each other.

Full setup guide: [uptime-kuma.md](uptime-kuma.md)

---

## Related Docs

- [uptime-kuma.md](uptime-kuma.md) — Uptime Kuma setup, monitors, and status page
- [vps-deploy.md](vps-deploy.md) — Full VPS deployment guide including stability fixes
- [troubleshooting.md](troubleshooting.md) — Common errors and how to fix them
