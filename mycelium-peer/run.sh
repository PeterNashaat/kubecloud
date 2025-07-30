#!/bin/sh
set -e

CONFIG_URL="https://raw.githubusercontent.com/threefoldtech/zos-config/main/production.json"
DEFAULT_PEERS="tcp://188.40.132.242:9651 tcp://136.243.47.186:9651 tcp://185.69.166.7:9651 tcp://185.69.166.8:9651 tcp://65.21.231.58:9651 tcp://65.109.18.113:9651 tcp://209.159.146.190:9651 tcp://5.78.122.16:9651 tcp://5.223.43.251:9651 tcp://142.93.217.194:9651"
SOURCE_BINARY="/mycelium"
TARGET_BINARY="/usr/local/bin/mycelium"

if [ -f "$SOURCE_BINARY" ]; then
    cp "$SOURCE_BINARY" "$TARGET_BINARY"
    chmod +x "$TARGET_BINARY"
else
    echo "❌ Source mycelium binary not found at $SOURCE_BINARY"
    exit 1
fi

if curl -fsSL "$CONFIG_URL" -o /tmp/production.json; then
  PEERS=$(jq -r '.mycelium.peers[]?' /tmp/production.json 2>/dev/null | tr '\n' ' ' | sed 's/ $//')
  if [ -z "$PEERS" ]; then
    PEERS="$DEFAULT_PEERS"
  fi
else
  echo "❌ Download failed — using default peers."
  PEERS="$DEFAULT_PEERS"
fi

echo "🚀 Starting: $TARGET_BINARY --peers $PEERS"
exec "$TARGET_BINARY" --peers $PEERS
