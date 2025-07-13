# K3s Image

This image for k3s with support for high availability and dual stack setup.

## Building

cd to k3s directory
`docker build -t threefoldtech/k3s:latest .`

## Running

for running master node

```bash
docker run -it --name master -e K3S_URL=""  -e K3S_TOKEN="<TOKEN>" --privileged threefoldtech/k3s:latest
```

for running a worker node

```bash
docker run -it --name worker -e K3S_URL="https://<MASTER_IP>:6443" -e K3S_TOKEN="<TOKEN>" --privileged threefoldtech/k3s:latest
```

## Flist

<https://hub.grid.tf/samehabouelsaad.3bot/abouelsaad-k3s_1.26.0-latest.flist>

## Entrypoint

```bash
zinit init 
```

## ENV Vars

- `K3S_URL`: For the leader node this should be empty for worker or other masters nodes should be the leader url for example `https://<LEADER_IP>:6443`
- `K3S_TOKEN`: The token for your cluster should be same for all nodes
- `K3S_DATA_DIR`: Data dir for kubernetes default is `/var/lib/rancher/k3s/` (preferred to mount disk and use its mount path)
- `K3S_FLANNEL_IFACE`: Interface used by flannel. (default is `eth0`, in case of using dual stack, it will be `mycelium-br`)
- `K3S_DATASTORE_ENDPOINT`: For k3s external data store like etcd, sqlite, postgres or mysql ...
- `K3S_NODE_NAME`: Sets node name
- `DUAL_STACK`: Should be added to all nodes to enable the dual stack on the cluster. (default: false)
- `MASTER`: Should be added to make the node master. (default: false)
- `HA`: Should be added to the leader node to enable the high availability. (default: false)
