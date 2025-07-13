#!/bin/bash
# ============================================================================
# Script: integrate_eth0_into_mycelium_br.sh
# 
# Purpose: This script integrates the eth0 interface into the mycelium-br bridge
#          to create a unified interface that supports both IPv4 and IPv6 for
#          dual-stack networking in K3s. This is necessary because eth0 typically
#          has IPv4 connectivity while mycelium-br provides IPv6 connectivity.
#          Flannel in K3s requires a single interface with both IP versions for
#          proper dual-stack operation.
# ============================================================================

# Exit immediately if any command fails
set -e

# Check if this is a dual-stack setup by verifying the DUAL_STACK environment variable
if [ -z "${DUAL_STACK}" ]; then
  echo "❌ Not a dual stack setup"
  exit 1
fi

# Define the bridge and ethernet interface names
bridge="mycelium-br"  # The mycelium bridge interface that has IPv6 connectivity
eth_iface="eth0"      # The standard ethernet interface with IPv4 connectivity

echo "[*] Migrating IPv4 configuration from $eth_iface to $bridge..."

# Step 1: Detect IPv4 address and default gateway on eth0
# Extract the IPv4 address assigned to eth0
ipv4=$(ip -4 addr show dev "$eth_iface" | awk '$1 == "inet" {print $2}')
# Extract the default gateway for IPv4 traffic through eth0
ipv4_gw=$(ip route show | awk '$1 == "default" && $5 == "'"$eth_iface"'" {print $3; exit}')

# Step 2: Capture all non-default IPv4 routes that use eth0
# This preserves all specific routes that will need to be migrated to the bridge
mapfile -t old_ipv4_routes < <(ip route show | awk '$1 != "default" && $5 == "'"$eth_iface"'"')

# Step 3: Remove the IPv4 address from eth0
# This is necessary before adding eth0 to the bridge to avoid IP conflicts
ip addr del $ipv4 dev $eth_iface

# Step 4: Set up the bridge and add eth0 to it
# Ensure the bridge interface is up
ip link set "$bridge" up
# Add eth0 as a port on the bridge (makes eth0 a slave of the bridge)
ip link set "$eth_iface" master "$bridge"
# Ensure eth0 is up
ip link set "$eth_iface" up

# Step 5: Reassign the IPv4 address to the bridge
# This moves the IPv4 connectivity from eth0 to the bridge
if [[ -n "$ipv4" ]]; then
  ip addr add "$ipv4" dev "$bridge"
fi

# Step 6: Reapply the default IPv4 route through the bridge
# This ensures IPv4 traffic is routed through the bridge instead of directly through eth0
if [[ -n "$ipv4_gw" ]]; then
  # Remove any existing default route through eth0 (may fail if already removed)
  ip route del default dev "$eth_iface" 2>/dev/null || true
  # Add the default route via the bridge
  ip route add default via "$ipv4_gw" dev "$bridge"
fi

# Step 7: Remove IPv6 default route via eth0, if present
# This prevents routing conflicts with IPv6 on the bridge
if ip -6 route show default | grep -q "dev $eth_iface"; then
  echo "[*] Removing default IPv6 route on $eth_iface..."
  ip -6 route del default dev "$eth_iface" || true
fi

# Step 8: Reapply all non-default IPv4 routes through the bridge
# This ensures all specific routes continue to work but now through the bridge
echo "[*] Re-applying non-default IPv4 routes previously on $eth_iface..."
for route in "${old_ipv4_routes[@]}"; do
  # Replace eth0 with the bridge name in each route
  new_route=$(echo "$route" | sed "s/ dev $eth_iface/ dev $bridge/")
  echo "    ➤ $new_route"
  ip route replace $new_route
done

# Step 9: Enable IP forwarding and proxy features on the bridge
# These settings are necessary for proper network functionality in a Kubernetes environment
echo "[*] Enabling forwarding and proxy features..."
# Enable IPv4 forwarding globally
sysctl -w net.ipv4.ip_forward=1
# Enable IPv6 forwarding on the bridge
sysctl -w net.ipv6.conf."$bridge".forwarding=1
# Enable proxy ARP on the bridge (helps with IPv4 address resolution)
sysctl -w net.ipv4.conf."$bridge".proxy_arp=1
# Enable proxy NDP on the bridge (helps with IPv6 address resolution)
sysctl -w net.ipv6.conf."$bridge".proxy_ndp=1

