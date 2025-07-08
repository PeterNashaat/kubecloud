#!/bin/bash
set -e

bridge="flannel-br"
eth_iface="eth0"
CUSTOM_TABLE_NAME=myctable

# Step 0: Create bridge
echo "🔧 Creating bridge: $bridge"
ip link add name $bridge type bridge
ip link set $bridge up

echo "[*] Migrating IPv4 configuration from $eth_iface to $bridge..."
# Step 1: Detect IPv4 address and default gateway
ipv4=$(ip -4 addr show dev "$eth_iface" | awk '$1 == "inet" {print $2}')
ipv4_gw=$(ip route show | awk '$1 == "default" && $5 == "'"$eth_iface"'" {print $3; exit}')

# Step 2: Capture all non-default IPv4 routes on eth0
mapfile -t old_ipv4_routes < <(ip route show | awk '$1 != "default" && $5 == "'"$eth_iface"'"')


# Step 3: Find the interface with 400::/7 route among eth1-eth9
for iface in eth{1..9}; do
  if ip -6 route show 400::/7 | grep -qw "dev $iface"; then
    IPV6_IFACE="$iface"
    echo "✅ Found IPv6 interface: $IPV6_IFACE (has 400::/7 route)"
    break
  fi
done

if [[ -z "$IPV6_IFACE" ]]; then
  echo "❌ No interface eth1–eth9 has a route to 400::/7"
  exit 1
fi
ipv6_global=$(ip -6 addr show dev "$IPV6_IFACE" | awk '/inet6/ && !/fe80::/ {print $2}' | head -n1)
ipv6_ll=$(ip -6 addr show dev "$IPV6_IFACE" | awk '/inet6/ && /fe80::/ {print $2}' | head -n1)
ipv6_gw=$(ip -6 route show 400::/7 | grep "dev $IPV6_IFACE" | awk '/via/ {print $3}' | head -n1)


# Step 3: Clean up
ip link set "$eth_iface" nomaster || true
ip link set "$eth_iface" down
ip addr flush dev "$eth_iface"

ip link set "$IPV6_IFACE" nomaster || true
ip link set "$IPV6_IFACE" down
ip addr flush dev "$IPV6_IFACE"

ip link set "$bridge" up
ip link set "$eth_iface" master "$bridge"
ip link set "$IPV6_IFACE" master "$bridge"
ip link set "$eth_iface" up
ip link set "$IPV6_IFACE" up

# Step 5: Reassign IPv4 address
if [[ -n "$ipv4" ]]; then
  ip addr add "$ipv4" dev "$bridge"
fi

if [[ -n "$ipv6_global" ]]; then
  ip addr add "$ipv6_global" dev "$bridge"
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

# Step 7: Clean up and re-add 400::/7 route
echo "🧹 Removing old 400::/7 route via $IPV6_IFACE"
ip -6 route del 400::/7 dev "$IPV6_IFACE" || true

echo "📡 Adding route: 400::/7 via $ipv6_gw on $bridge"
ip -6 route add 400::/7 via "$ipv6_gw" dev "$bridge" table "$CUSTOM_TABLE_NAME"


echo "[*] Enabling forwarding and proxy features..."
sysctl -w net.ipv4.ip_forward=1
sysctl -w net.ipv6.conf."$bridge".forwarding=1
sysctl -w net.ipv4.conf."$bridge".proxy_arp=1
sysctl -w net.ipv6.conf."$bridge".proxy_ndp=1

