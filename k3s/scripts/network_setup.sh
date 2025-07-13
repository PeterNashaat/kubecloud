#!/bin/bash
# ============================================================================
# Script: network_setup.sh
# 
# Purpose: Sets up the mycelium-br bridge interface for IPv6 networking in K3s.
#          This script creates a bridge interface that will be used by the
#          mycelium CNI plugin to assign IPv6 addresses to Kubernetes pods.
#          The bridge provides IPv6 connectivity from the mycelium overlay network
#          to the containerized workloads running in the cluster.
# ============================================================================

# Wait for the mycelium interface to be available
# The script blocks until the mycelium interface is detected by the system
# Mycelium is a peer-to-peer IPv6 overlay network used for ThreeFold Grid connectivity
/scripts/wait_for_interface.sh mycelium

# Create a new bridge interface named mycelium-br
# This bridge will be used by the mycelium CNI to provide IPv6 connectivity to pods
ip link add name mycelium-br type bridge

# Calculate the IPv6 address for the bridge
# Takes the first 4 segments of the mycelium IPv6 address and adds ::1/64 to create a bridge address
# Example: if mycelium has 2001:db8:1:2::3/64, the bridge gets 2001:db8:1:2::1/64
# This ensures the bridge has an address in the same subnet as the mycelium interface
BRIDGE_IP=$(ip -6 addr show dev mycelium | awk '/inet6/ && /scope global/ {print $2}' | cut -d/ -f1 | cut -d: -f1-4 | awk '{print $0 "::1/64"}')

# Assign the calculated IPv6 address to the bridge interface
# This allows the bridge to participate in IPv6 networking
ip addr add $BRIDGE_IP dev mycelium-br 

# Activate the bridge interface
# This brings the interface up so it can start passing traffic
ip link set dev mycelium-br up

# Enable IPv6 forwarding globally
# This is necessary for routing IPv6 traffic between the bridge and the pod network
# Without this, pods would not be able to communicate over IPv6
sysctl -w net.ipv6.conf.all.forwarding=1

# Enable Neighbor Discovery Protocol (NDP) proxying on br0 and all interfaces
# NDP proxying allows the bridge to answer Neighbor Solicitation messages for pods
# This is critical for proper IPv6 address resolution for pod IPs
echo "net.ipv6.conf.br0.proxy_ndp=1" | tee -a /etc/sysctl.conf
echo "net.ipv6.conf.all.proxy_ndp=1" | tee -a /etc/sysctl.conf

# Apply the sysctl settings immediately
sysctl -p

# Get the full IPv6 address of the mycelium interface
# This will be used to set up proxying for the mycelium interface
MYCELIUM_IP=$(ip -6 addr show dev mycelium | awk '/inet6/ && /scope global/ {print $2}' | cut -d/ -f1)

# Set up IPv6 neighbor proxy for the mycelium interface
# This allows the bridge to answer Neighbor Solicitation messages for the mycelium interface
# Essential for proper IPv6 connectivity between pods and the mycelium network
ip -6 neigh add proxy $MYCELIUM_IP dev mycelium-br 
