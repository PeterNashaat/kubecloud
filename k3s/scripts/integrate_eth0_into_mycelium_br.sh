#!/bin/bash
set -e


if [ -z "${DUAL_STACK}" ]; then
  echo "❌ Not a dual stack setup"
  exit 1
fi

bridge="mycelium-br"
eth_iface="eth0"

echo "[*] Migrating IPv4 configuration from $eth_iface to $bridge..."
# Step 1: Detect IPv4 address and default gateway
ipv4=$(ip -4 addr show dev "$eth_iface" | awk '$1 == "inet" {print $2}')
ipv4_gw=$(ip route show | awk '$1 == "default" && $5 == "'"$eth_iface"'" {print $3; exit}')

# Step 2: Capture all non-default IPv4 routes on eth0
mapfile -t old_ipv4_routes < <(ip route show | awk '$1 != "default" && $5 == "'"$eth_iface"'"')

# Step 3: Clean up eth0
ip addr del $ipv4 dev $eth_iface

ip link set "$bridge" up
ip link set "$eth_iface" master "$bridge"
ip link set "$eth_iface" up

# Step 5: Reassign IPv4 address
if [[ -n "$ipv4" ]]; then
  ip addr add "$ipv4" dev "$bridge"
fi

# Step 6: Reapply default IPv4 route
if [[ -n "$ipv4_gw" ]]; then
  ip route del default dev "$eth_iface" 2>/dev/null || true
  ip route add default via "$ipv4_gw" dev "$bridge"
fi

# Step 7: Remove IPv6 default route via eth0, if present
if ip -6 route show default | grep -q "dev $eth_iface"; then
  echo "[*] Removing default IPv6 route on $eth_iface..."
  ip -6 route del default dev "$eth_iface" || true
fi

# Step 8: Reapply non-default IPv4 routes
echo "[*] Re-applying non-default IPv4 routes previously on $eth_iface..."
for route in "${old_ipv4_routes[@]}"; do
  new_route=$(echo "$route" | sed "s/ dev $eth_iface/ dev $bridge/")
  echo "    ➤ $new_route"
  ip route replace $new_route
done

echo "[*] Enabling forwarding and proxy features..."
sysctl -w net.ipv4.ip_forward=1
sysctl -w net.ipv6.conf."$bridge".forwarding=1
sysctl -w net.ipv4.conf."$bridge".proxy_arp=1
sysctl -w net.ipv6.conf."$bridge".proxy_ndp=1

