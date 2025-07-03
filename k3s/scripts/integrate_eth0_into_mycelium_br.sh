#!/bin/bash
set -e

bridge="mycelium-br"
eth_iface="eth0"

echo "[*] Migrating IP configuration from $eth_iface to $bridge..."

# Step 1: Detect IPv4 and IPv6 addresses
ipv4=$(ip -4 addr show dev "$eth_iface" | awk '$1 == "inet" {print $2}')
ipv4_gw=$(ip route show | awk '$1 == "default" && $5 == "'"$eth_iface"'" {print $3; exit}')
ipv6_addrs=$(ip -6 addr show dev "$eth_iface" | awk '$1 == "inet6" && $2 !~ /^fe80::/ {print $2}')
ipv6_gw=$(ip -6 route show default | awk '$0 ~ / dev '"$eth_iface"' / {print $3; exit}')
has_ipv6_default=$(ip -6 route show default | grep -q "dev $eth_iface" && echo yes || echo no)

# Step 2: Capture all non-default IPv4 routes on eth0
mapfile -t old_ipv4_routes < <(ip route show | awk '$1 != "default" && $5 == "'"$eth_iface"'"')

# Step 3: Prepare interfaces
ip link set "$eth_iface" nomaster || true
ip link set "$eth_iface" down
ip addr flush dev "$eth_iface"


ip link set "$bridge" up
ip link set "$eth_iface" master "$bridge"
ip link set "$eth_iface" up

# Step 4: Move IPs to bridge
[[ -n "$ipv4" ]] && ip addr del "$ipv4" dev "$eth_iface" && ip addr add "$ipv4" dev "$bridge"

while read -r ipv6; do
  [[ -n "$ipv6" ]] && ip -6 addr del "$ipv6" dev "$eth_iface" && ip -6 addr add "$ipv6" dev "$bridge"
done <<< "$ipv6_addrs"

# Step 5: Replace default routes
[[ -n "$ipv4_gw" ]] && ip route del default dev "$eth_iface" 2>/dev/null || true
[[ -n "$ipv4_gw" ]] && ip route replace default via "$ipv4_gw" dev "$bridge"

if [[ "$has_ipv6_default" == "yes" && -n "$ipv6_gw" ]]; then
  ip -6 route del default dev "$eth_iface" 2>/dev/null || true
  ip -6 route replace default via "$ipv6_gw" dev "$bridge"
fi

# Step 6: Reapply non-default IPv4 routes that used to be on eth0
echo "[*] Re-applying non-default IPv4 routes originally on $eth_iface..."
for route in "${old_ipv4_routes[@]}"; do
  # Replace ' dev eth0' with ' dev mycelium-br'
  new_route=$(echo "$route" | sed "s/ dev $eth_iface/ dev $bridge/")
  echo "    ➤ $new_route"
  ip route replace $new_route
done


echo "[*] Enabling forwarding and proxy features..."
sysctl -w net.ipv4.ip_forward=1
sysctl -w net.ipv6.conf.mycelium-br.forwarding=1
sysctl -w net.ipv4.conf.mycelium-br.proxy_arp=1
sysctl -w net.ipv6.conf.mycelium-br.proxy_ndp=1

