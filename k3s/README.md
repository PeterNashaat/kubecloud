# K3s Image

[![K3s Version](https://img.shields.io/badge/K3s-v1.26.0-blue)](https://github.com/k3s-io/k3s/releases/tag/v1.26.0)
[![Docker Image](https://img.shields.io/badge/Docker-threefoldtech%2Fk3s-green)](https://hub.docker.com/r/threefoldtech/k3s)

This image provides a lightweight K3s Kubernetes distribution with enhanced support for:

- High availability (HA) cluster configuration
- Dual stack networking (IPv4/IPv6)
- Easy deployment on ThreeFold Grid

## Features

- **Lightweight**: Minimal resource requirements compared to full Kubernetes
- **High Availability**: Support for multi-master setup
- **Dual Stack**: Full IPv4/IPv6 networking support
- **Simplified Setup**: Easy configuration through environment variables
- **ThreeFold Integration**: Ready for deployment on ThreeFold Grid

## Prerequisites

- Docker installed (for local development)
- Network connectivity between nodes
- Sufficient privileges for container execution

## Building

To build the K3s image locally:

```bash
# Navigate to the k3s directory
cd k3s

# Build the Docker image
docker build -t threefoldtech/k3s:latest .
```

## Running

### Leader/Master Node

```bash
docker run -it --name master \
  -e K3S_URL="" \
  -e K3S_TOKEN="<YOUR_CLUSTER_TOKEN>" \
  -e MASTER="true" \
  --privileged \
  threefoldtech/k3s:latest
```

### Worker Node

```bash
docker run -it --name worker \
  -e K3S_URL="https://<MASTER_IP>:6443" \
  -e K3S_TOKEN="<YOUR_CLUSTER_TOKEN>" \
  --privileged \
  threefoldtech/k3s:latest
```

### High Availability Setup

```bash
# First master (leader)
docker run -it --name master1 \
  -e K3S_URL="" \
  -e K3S_TOKEN="<YOUR_CLUSTER_TOKEN>" \
  -e MASTER="true" \
  -e HA="true" \
  --privileged \
  threefoldtech/k3s:latest

# Additional masters
docker run -it --name master2 \
  -e K3S_URL="https://<LEADER_IP>:6443" \
  -e K3S_TOKEN="<YOUR_CLUSTER_TOKEN>" \
  -e MASTER="true" \
  --privileged \
  threefoldtech/k3s:latest
```

## ThreeFold Deployment

### Flist

The K3s image is available as a Flist for ThreeFold Grid deployment:

[https://hub.grid.tf/samehabouelsaad.3bot/abouelsaad-k3s_1.26.0-latest.flist](https://hub.grid.tf/samehabouelsaad.3bot/abouelsaad-k3s_1.26.0-latest.flist)

### Entrypoint

The default entrypoint for the container is:

```bash
zinit init
```

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `K3S_URL` | URL of the leader node. Empty for leader, `https://<LEADER_IP>:6443` for workers/additional masters | - | Yes |
| `K3S_TOKEN` | Authentication token for the cluster (must be identical across all nodes) | - | Yes |
| `K3S_DATA_DIR` | Data directory for Kubernetes | `/var/lib/rancher/k3s/` | No |
| `K3S_FLANNEL_IFACE` | Network interface used by Flannel | `eth0` (or `mycelium-br` for dual stack) | No |
| `K3S_DATASTORE_ENDPOINT` | External datastore endpoint (etcd, sqlite, postgres, mysql) | - | No |
| `K3S_NODE_NAME` | Custom node name | Hostname | No |
| `DUAL_STACK` | Enable dual stack (IPv4/IPv6) networking | `false` | No |
| `MASTER` | Configure node as a master | `false` | No |
| `HA` | Enable high availability mode on leader node | `false` | No |

## Persistent Storage

For production deployments, it's recommended to mount persistent storage to an external location and set K3S_DATA_DIR to point to this location:

```bash
docker run -it --name master \
  -e K3S_URL="" \
  -e K3S_TOKEN="<YOUR_CLUSTER_TOKEN>" \
  -e MASTER="true" \
  -e K3S_DATA_DIR="/mnt/data" \
  -v /path/to/storage:/mnt/data \
  --privileged \
  threefoldtech/k3s:latest
```

This approach ensures your Kubernetes data is stored on the mounted volume rather than inside the container.

## Troubleshooting

### Common Issues

- **Nodes not joining the cluster**: Verify network connectivity and that the correct K3S_TOKEN is being used
- **Dual stack not working**: Ensure the correct network interface is specified with K3S_FLANNEL_IFACE
- **Container fails to start**: Check for sufficient privileges (--privileged flag)

### Logs

To view container logs:

```bash
docker logs <container_name>
```

## Contributing

Contributions to improve this image are welcome. Please follow the standard GitHub workflow:

1. Fork the repository
2. Create a feature branch
3. Submit a pull request

## License

This project is licensed under the Apache License 2.0 - see the LICENSE file for details.
