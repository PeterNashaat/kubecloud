#!/bin/bash
# ============================================================================
# Script: multus_setup.sh
# 
# Purpose: Sets up the environment required for Multus CNI plugin to function
#          properly within K3s. Multus CNI is a container network interface that
#          enables attaching multiple network interfaces to Kubernetes pods.
# ============================================================================

# Make the /mnt/data directory a shared mount
# This is a critical requirement for Multus pods to function correctly
# Shared mounts allow mount events to propagate to slave mounts
mount --make-shared /mnt/data/

# Wait for the K3s CNI configuration directory to be created
# This ensures that K3s has initialized its CNI system before we proceed
# The script will block until this path exists
/scripts/wait_for_path.sh /mnt/data/agent/etc/cni/net.d

# Remove any existing CNI configuration directory to avoid conflicts
# This ensures we start with a clean state
rm -rf /etc/cni

# Create the standard CNI configuration directory expected by CNI plugins
mkdir -p /etc/cni

# Create a symbolic link from the standard CNI location to the K3s CNI location
# Multus expects CNI configurations to be in /etc/cni/net.d, but K3s stores them
# in /mnt/data/agent/etc/cni/net.d, so this symlink bridges that gap
ln -s /mnt/data/agent/etc/cni/net.d /etc/cni
