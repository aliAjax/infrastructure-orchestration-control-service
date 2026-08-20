#!/usr/bin/env sh
set -eu

docker compose up -d

for i in $(seq 1 30); do
  if docker compose exec -T postgres pg_isready -U infra -d infra_controlplane >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

go build ./...
exec go run ./cmd/server -config=configs/config.yaml
