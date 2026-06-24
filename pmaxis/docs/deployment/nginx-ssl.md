# nginx + SSL Setup for api.pmaxis.trade

> Sets up nginx as a reverse proxy in front of `hub-api` (port 8088), with automatic HTTPS via Let's Encrypt.
> After this is done, the API is reachable at `https://api.pmaxis.trade` — no port number in the URL.

---

## Prerequisites

- VPS running at `167.233.97.217`
- `hub-api` already running on `127.0.0.1:8088` (verify: `curl http://localhost:8088/health`)
- Domain `pmaxis.trade` with DNS access (Namecheap, Cloudflare, etc.)

---

## Step 1 — Add DNS Record

In your domain registrar's DNS panel, add:

| Type | Host | Value | TTL |
|------|------|-------|-----|
| A | `api` | `167.233.97.217` | Automatic |

This creates `api.pmaxis.trade → 167.233.97.217`.

Wait for propagation (usually instant on Cloudflare, up to 30 min elsewhere). Verify:

```bash
nslookup api.pmaxis.trade
# Should return 167.233.97.217
```

---

## Step 2 — Install nginx and certbot

SSH into the VPS:

```bash
ssh vijay@167.233.97.217
```

Install:

```bash
sudo apt update
sudo apt install -y nginx certbot python3-certbot-nginx
```

---

## Step 3 — Create nginx site config

```bash
sudo nano /etc/nginx/sites-available/api.pmaxis.trade
```

Paste this config:

```nginx
server {
    listen 80;
    server_name api.pmaxis.trade;

    location / {
        proxy_pass         http://127.0.0.1:8088;
        proxy_http_version 1.1;
        proxy_set_header   Upgrade $http_upgrade;
        proxy_set_header   Connection "upgrade";
        proxy_set_header   Host $host;
        proxy_set_header   X-Real-IP $remote_addr;
        proxy_set_header   X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
        proxy_read_timeout 86400;
    }
}
```

> `proxy_read_timeout 86400` keeps WebSocket connections alive for up to 24 hours.

---

## Step 4 — Enable the site

```bash
sudo ln -s /etc/nginx/sites-available/api.pmaxis.trade /etc/nginx/sites-enabled/

# Test config syntax
sudo nginx -t

# Reload nginx
sudo systemctl reload nginx
```

Verify HTTP is working before adding SSL:

```bash
curl http://api.pmaxis.trade/health
```

Should return the same JSON as `curl http://localhost:8088/health`.

---

## Step 5 — Get SSL certificate

```bash
sudo certbot --nginx -d api.pmaxis.trade
```

Certbot will:
1. Prompt for your email address (for renewal notices)
2. Ask about Terms of Service → type `Y`
3. Ask about sharing email with EFF → your choice
4. Automatically edit the nginx config to add HTTPS and redirect HTTP→HTTPS
5. Test renewal

When done, verify HTTPS:

```bash
curl https://api.pmaxis.trade/health
```

---

## Step 6 — Open firewall for HTTPS

```bash
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw status
```

Expected:

```
80/tcp    ALLOW
443/tcp   ALLOW
22/tcp    ALLOW
8088/tcp  ALLOW
```

> Port 8088 remains open for direct VPS access during development. You can close it later if you want to force all traffic through nginx: `sudo ufw delete allow 8088/tcp`

---

## Step 7 — Update CORS

Once the domain is live, update `ALLOWED_ORIGINS` in `~/pmaxis/package/.env`:

```env
ALLOWED_ORIGINS=https://api.pmaxis.trade,https://pmaxis.trade
```

Then restart hub-api to pick up the new env:

```bash
cd ~/pmaxis/package
COMPOSE_PROJECT_NAME=package docker compose up -d --no-deps --force-recreate hub-api
```

---

## Ongoing maintenance

### Auto-renewal

Certbot registers a systemd timer for auto-renewal. Check it's active:

```bash
sudo systemctl status certbot.timer
```

Certificates renew automatically when they're within 30 days of expiry. Let's Encrypt certs are valid for 90 days.

To test renewal manually:

```bash
sudo certbot renew --dry-run
```

### Viewing nginx access logs

```bash
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log
```

### Restarting nginx

```bash
sudo systemctl restart nginx
```

### Checking the certificate expiry

```bash
sudo certbot certificates
```

---

## WebSocket connections

WebSocket connections to `wss://api.pmaxis.trade/stream` work through nginx because of the `Upgrade` / `Connection` headers in the config above. The `proxy_read_timeout 86400` ensures connections are not dropped after the default 60-second idle timeout.

Client usage:

```js
// Old (direct IP)
new WebSocket('ws://167.233.97.217:8088/stream?api_key=pmx_live_...')

// New (domain, secure)
new WebSocket('wss://api.pmaxis.trade/stream?api_key=pmx_live_...')
```

---

## Related Docs

- [vps-deploy.md](vps-deploy.md) — Full VPS deployment guide
- [../frontend/pages.md](../frontend/pages.md) — All embedded pages and their URLs
- [../api/key-management.md](../api/key-management.md) — Key management and admin panel
