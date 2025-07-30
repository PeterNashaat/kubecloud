#!/usr/bin/env bash
set -e

CONFIG_URL="https://raw.githubusercontent.com/threefoldtech/zos-config/main/production.json"

DEFAULT_PEERS=(
  tcp://188.40.132.242:9651
  tcp://136.243.47.186:9651
  tcp://185.69.166.7:9651
  tcp://185.69.166.8:9651
  tcp://65.21.231.58:9651
  tcp://65.109.18.113:9651
  tcp://209.159.146.190:9651
  tcp://5.78.122.16:9651
  tcp://5.223.43.251:9651
  tcp://142.93.217.194:9651
)

echo "🔄 Downloading production.json..."
if curl -fsSL "$CONFIG_URL" -o /tmp/production.json; then
  echo "✅ Download succeeded — parsing .mycelium.peers..."
  PEERS=( $(jq -r '.mycelium.peers[]?' /tmp/production.json) )
  if [[ ${#PEERS[@]} -eq 0 ]]; then
    echo "⚠️  No peers found in .mycelium.peers — using default."
    PEERS=("${DEFAULT_PEERS[@]}")
  else
    echo "✅ Found ${#PEERS[@]} peers from JSON."
  fi
else
  echo "❌ Download failed — using default peers."
  PEERS=("${DEFAULT_PEERS[@]}")
fi

# Build and execute the command
cmd=(mycelium --key-file /etc/netseed --peers)
cmd+=("${PEERS[@]}")

echo "🚀 Starting: ${cmd[*]}"
exec "${cmd[@]}"