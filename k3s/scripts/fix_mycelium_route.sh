#!/bin/bash
set -e

# ============================================
# Fix IPv6 routing for Kubernetes nodes
# --------------------------------------------
# On nodes with multiple "mycelium" IPv6 networks, this script ensures:
# - All traffic matching 400::/7 is routed through the proper interface.
# - Traffic from the VM (i.e. source IP of the ethX interface) is routed directly
#   through the matching interface, avoiding being forwarded via the mycelium interface.
# - This avoids packets being dropped due to mismatched source subnets.
# ============================================

TARGET_PREFIX="400::/7"
CUSTOM_TABLE_NAME="myctable"
CUSTOM_TABLE_ID=100

# Ensure custom routing table exists in /etc/iproute2/rt_tables
if ! grep -q "$CUSTOM_TABLE_ID $CUSTOM_TABLE_NAME" /etc/iproute2/rt_tables; then
  echo "$CUSTOM_TABLE_ID $CUSTOM_TABLE_NAME" | tee -a /etc/iproute2/rt_tables > /dev/null
fi

# Find all IPv6 routes matching the target prefix and using ethX interfaces
ip -6 route | grep "^$TARGET_PREFIX" | grep -E "dev eth[0-9]+" | while read -r line; do
  echo "🔍 Matched route: $line"

  # Extract gateway and interface name
  GATEWAY=$(echo "$line" | awk '{for (i=1; i<=NF; i++) if ($i=="via") print $(i+1)}')
  DEV=$(echo "$line" | awk '{for (i=1; i<=NF; i++) if ($i=="dev") print $(i+1)}')

  echo "🌐 Gateway: ${GATEWAY:-<none>}"
  echo "🔧 Device: $DEV"

  # Get the first global (non-link-local) IPv6 address of the device
  DEVIP=$(ip -6 addr show dev "$DEV" | \
    awk '/inet6/ && $2 ~ /\// && $2 !~ /^fe80/ { sub(/\/.*/, "", $2); print $2 }' | head -n 1)

  echo "📤 Using source IP: $DEVIP"

  # Add rule only if not already present
  if ! ip -6 rule show | grep -q "from $DEVIP.*table $CUSTOM_TABLE_NAME"; then
    ip -6 rule add from "$DEVIP" table "$CUSTOM_TABLE_NAME" priority "$CUSTOM_TABLE_ID"
    echo "✅ Added rule for $DEVIP"
  fi

  # Add route only if not already present
  if ! ip -6 route show table "$CUSTOM_TABLE_NAME" | grep -q "^$TARGET_PREFIX.*via $GATEWAY.*dev $DEV"; then
    ip -6 route add "$TARGET_PREFIX" via "$GATEWAY" dev "$DEV" table "$CUSTOM_TABLE_NAME"
    echo "✅ Added route for $TARGET_PREFIX via $GATEWAY dev $DEV"
  fi

done
