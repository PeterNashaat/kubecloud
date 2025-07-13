package kubedeployer

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"slices"
	"sync"

	"github.com/pkg/errors"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
)

var (
	usedIPsTracker = make(map[string]map[uint32][]byte)
	usedIPsMutex   sync.RWMutex
)

func getRandomMyceliumNetSeed() (string, error) {
	key := make([]byte, MYC_NET_SEED_LEN)
	_, err := rand.Read(key)
	return hex.EncodeToString(key), err
}

func getIpForVm(ctx context.Context, tfPluginClient deployer.TFPluginClient, networkName string, nodeID uint32) (string, error) {
	network := tfPluginClient.State.Networks.GetNetwork(networkName)
	ipRange := network.GetNodeSubnet(nodeID)

	ip, ipRangeCIDR, err := net.ParseCIDR(ipRange)
	if err != nil {
		return "", errors.Wrapf(err, "invalid IP range %s for node %d", ipRange, nodeID)
	}

	usedHostIDs, err := getUsedHostIDs(ctx, tfPluginClient, nodeID, networkName, ipRangeCIDR)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get used IPs for node %d", nodeID)
	}

	usedIPsMutex.Lock()
	defer usedIPsMutex.Unlock()

	if usedIPsTracker[networkName] == nil {
		usedIPsTracker[networkName] = make(map[uint32][]byte)
	}
	trackedIPs := usedIPsTracker[networkName][nodeID]
	usedHostIDs = append(usedHostIDs, trackedIPs...)

	for hostID := byte(2); hostID < 255; hostID++ {
		if !slices.Contains(usedHostIDs, hostID) {
			usedIPsTracker[networkName][nodeID] = append(usedIPsTracker[networkName][nodeID], hostID)
			vmIP := ip.To4()
			vmIP[3] = hostID
			return vmIP.String(), nil
		}
	}

	return "", fmt.Errorf("all IPs are exhausted for network %s on node %d", networkName, nodeID)
}

func getUsedHostIDs(ctx context.Context, tfPluginClient deployer.TFPluginClient, nodeID uint32, networkName string, ipRangeCIDR *net.IPNet) ([]byte, error) {
	nodeClient, err := tfPluginClient.NcPool.GetNodeClient(tfPluginClient.SubstrateConn, nodeID)
	if err != nil {
		return nil, errors.Wrapf(err, "could not get node client for node %d", nodeID)
	}

	privateIPs, err := nodeClient.NetworkListPrivateIPs(ctx, networkName)
	if err != nil {
		return nil, errors.Wrapf(err, "could not list private IPs from node %d", nodeID)
	}

	var usedHostIDs []byte
	for _, privateIP := range privateIPs {
		parsedIP := net.ParseIP(privateIP).To4()
		if parsedIP != nil && ipRangeCIDR.Contains(parsedIP) {
			usedHostIDs = append(usedHostIDs, parsedIP[3])
		}
	}

	return usedHostIDs, nil
}
