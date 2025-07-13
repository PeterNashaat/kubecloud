#!/bin/bash
/scripts/wait_for_path.sh /mnt/data/agent/etc/cni/net.d
mount --make-shared /mnt/data/
rm -rf /etc/cni
mkdir -p /etc/cni
ln -s /mnt/data/agent/etc/cni/net.d /etc/cni
