#!/bin/bash

IFACE="$1"
INTERVAL=1  # seconds

if [ -z "$IFACE" ]; then
  echo "Usage: $0 <interface-name>"
  exit 1
fi

echo "⏳ Waiting for network interface '$IFACE'..."

while [ ! -d "/sys/class/net/$IFACE" ]; do
  sleep "$INTERVAL"
done

echo "✅ Interface '$IFACE' is now available."
