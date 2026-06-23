#!/usr/bin/env bash
# Deploy only hub-api to VPS.
# Run from the pmaxis/ root: bash deployments/deploy-api.sh
set -e

SSH_KEY=~/.ssh/id_ed25519
VPS=vijay@167.233.97.217
REMOTE=~/pmaxis/package

echo "==> Building pmaxis-api image..."
docker build -t pmaxis-api:latest -f services/api/Dockerfile .

echo "==> Saving image to tar (~400MB, takes ~30s)..."
docker save pmaxis-api:latest -o deployments/package/api.tar

echo "==> Uploading image + compose to VPS..."
scp -i "$SSH_KEY" \
  deployments/package/api.tar \
  deployments/package/docker-compose.yml \
  "$VPS:$REMOTE/"

echo "==> Loading image and restarting hub-api on VPS..."
ssh -i "$SSH_KEY" "$VPS" "
  cd $REMOTE &&
  docker load < api.tar &&
  COMPOSE_PROJECT_NAME=package docker compose up -d --no-deps --force-recreate hub-api &&
  echo 'Waiting for health...' &&
  sleep 5 &&
  docker compose -p package ps hub-api
"

echo "==> Done. Hub-api is live."
