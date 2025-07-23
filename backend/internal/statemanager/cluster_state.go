package statemanager

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"kubecloud/kubedeployer"

	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/zos"
	"github.com/xmonader/ewf"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// ClusterWrapper wraps the cluster with custom JSON marshaling
type ClusterWrapper struct {
	kubedeployer.Cluster
}

// MarshalJSON implements custom JSON marshaling for ClusterWrapper
func (cw ClusterWrapper) MarshalJSON() ([]byte, error) {
	// Create a copy of the cluster with string-encoded complex fields
	type Alias kubedeployer.Cluster

	// Create a custom network wrapper
	networkWrapper := struct {
		Name             string            `json:"name"`
		Description      string            `json:"description"`
		Nodes            []uint32          `json:"nodes"`
		IPRange          string            `json:"ip_range"`
		AddWGAccess      bool              `json:"add_wg_access"`
		MyceliumKeys     map[string]string `json:"mycelium_keys"` // base64 encoded
		SolutionType     string            `json:"solution_type"`
		AccessWGConfig   string            `json:"access_wg_config"`
		ExternalIP       *string           `json:"external_ip"`
		ExternalSK       string            `json:"external_sk"` // base64 encoded
		PublicNodeID     uint32            `json:"public_node_id"`
		NodesIPRange     map[string]string `json:"nodes_ip_range"`
		NodeDeploymentID map[string]uint64 `json:"node_deployment_id"` // string keys
		WGPort           map[string]int    `json:"wg_port"`
		Keys             map[string]string `json:"keys"` // base64 encoded
	}{
		Name:             cw.Network.Name,
		Description:      cw.Network.Description,
		Nodes:            cw.Network.Nodes,
		IPRange:          cw.Network.IPRange.String(),
		AddWGAccess:      cw.Network.AddWGAccess,
		SolutionType:     cw.Network.SolutionType,
		AccessWGConfig:   cw.Network.AccessWGConfig,
		PublicNodeID:     cw.Network.PublicNodeID,
		MyceliumKeys:     make(map[string]string),
		NodesIPRange:     make(map[string]string),
		NodeDeploymentID: make(map[string]uint64),
		WGPort:           make(map[string]int),
		Keys:             make(map[string]string),
	}

	// Convert ExternalIP
	if cw.Network.ExternalIP != nil {
		extIPStr := cw.Network.ExternalIP.String()
		networkWrapper.ExternalIP = &extIPStr
	}

	// Convert ExternalSK
	networkWrapper.ExternalSK = base64.StdEncoding.EncodeToString(cw.Network.ExternalSK[:])

	// Convert maps with uint32 keys to string keys
	for nodeID, myceliumKey := range cw.Network.MyceliumKeys {
		networkWrapper.MyceliumKeys[fmt.Sprintf("%d", nodeID)] = base64.StdEncoding.EncodeToString(myceliumKey)
	}

	for nodeID, ipRange := range cw.Network.NodesIPRange {
		networkWrapper.NodesIPRange[fmt.Sprintf("%d", nodeID)] = ipRange.String()
	}

	for nodeID, deploymentID := range cw.Network.NodeDeploymentID {
		networkWrapper.NodeDeploymentID[fmt.Sprintf("%d", nodeID)] = deploymentID
	}

	for nodeID, port := range cw.Network.WGPort {
		networkWrapper.WGPort[fmt.Sprintf("%d", nodeID)] = port
	}

	for nodeID, key := range cw.Network.Keys {
		networkWrapper.Keys[fmt.Sprintf("%d", nodeID)] = base64.StdEncoding.EncodeToString(key[:])
	}

	// Create the final wrapper structure
	wrapper := struct {
		*Alias
		Network interface{} `json:"network"`
	}{
		Alias:   (*Alias)(&cw.Cluster),
		Network: networkWrapper,
	}

	return json.Marshal(wrapper)
}

// UnmarshalJSON implements custom JSON unmarshaling for ClusterWrapper
func (cw *ClusterWrapper) UnmarshalJSON(data []byte) error {
	// First unmarshal into a temporary structure
	var temp struct {
		kubedeployer.Cluster
		Network struct {
			Name             string            `json:"name"`
			Description      string            `json:"description"`
			Nodes            []uint32          `json:"nodes"`
			IPRange          string            `json:"ip_range"`
			AddWGAccess      bool              `json:"add_wg_access"`
			MyceliumKeys     map[string]string `json:"mycelium_keys"`
			SolutionType     string            `json:"solution_type"`
			AccessWGConfig   string            `json:"access_wg_config"`
			ExternalIP       *string           `json:"external_ip"`
			ExternalSK       string            `json:"external_sk"`
			PublicNodeID     uint32            `json:"public_node_id"`
			NodesIPRange     map[string]string `json:"nodes_ip_range"`
			NodeDeploymentID map[string]uint64 `json:"node_deployment_id"`
			WGPort           map[string]int    `json:"wg_port"`
			Keys             map[string]string `json:"keys"`
		} `json:"network"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Copy the basic cluster fields
	cw.Cluster = temp.Cluster

	// Reconstruct the network with proper types
	cw.Network = workloads.ZNet{
		Name:             temp.Network.Name,
		Description:      temp.Network.Description,
		Nodes:            temp.Network.Nodes,
		AddWGAccess:      temp.Network.AddWGAccess,
		SolutionType:     temp.Network.SolutionType,
		AccessWGConfig:   temp.Network.AccessWGConfig,
		PublicNodeID:     temp.Network.PublicNodeID,
		MyceliumKeys:     make(map[uint32][]byte),
		NodesIPRange:     make(map[uint32]zos.IPNet),
		NodeDeploymentID: make(map[uint32]uint64),
		WGPort:           make(map[uint32]int),
		Keys:             make(map[uint32]wgtypes.Key),
	}

	// Parse IPRange
	if temp.Network.IPRange != "" {
		if ipNet, err := zos.ParseIPNet(temp.Network.IPRange); err == nil {
			cw.Network.IPRange = ipNet
		}
	}

	// Parse ExternalIP
	if temp.Network.ExternalIP != nil {
		if ipNet, err := zos.ParseIPNet(*temp.Network.ExternalIP); err == nil {
			cw.Network.ExternalIP = &ipNet
		}
	}

	// Parse ExternalSK
	if decoded, err := base64.StdEncoding.DecodeString(temp.Network.ExternalSK); err == nil && len(decoded) == 32 {
		var key [32]byte
		copy(key[:], decoded)
		cw.Network.ExternalSK = wgtypes.Key(key)
	}

	// Convert string-keyed maps back to uint32-keyed maps
	for nodeIDStr, myceliumKeyStr := range temp.Network.MyceliumKeys {
		if nodeID, err := strconv.ParseUint(nodeIDStr, 10, 32); err == nil {
			if decoded, err := base64.StdEncoding.DecodeString(myceliumKeyStr); err == nil {
				cw.Network.MyceliumKeys[uint32(nodeID)] = decoded
			}
		}
	}

	for nodeIDStr, ipRangeStr := range temp.Network.NodesIPRange {
		if nodeID, err := strconv.ParseUint(nodeIDStr, 10, 32); err == nil {
			if ipNet, err := zos.ParseIPNet(ipRangeStr); err == nil {
				cw.Network.NodesIPRange[uint32(nodeID)] = ipNet
			}
		}
	}

	for nodeIDStr, deploymentID := range temp.Network.NodeDeploymentID {
		if nodeID, err := strconv.ParseUint(nodeIDStr, 10, 32); err == nil {
			cw.Network.NodeDeploymentID[uint32(nodeID)] = deploymentID
		}
	}

	for nodeIDStr, port := range temp.Network.WGPort {
		if nodeID, err := strconv.ParseUint(nodeIDStr, 10, 32); err == nil {
			cw.Network.WGPort[uint32(nodeID)] = port
		}
	}

	for nodeIDStr, keyStr := range temp.Network.Keys {
		if nodeID, err := strconv.ParseUint(nodeIDStr, 10, 32); err == nil {
			if decoded, err := base64.StdEncoding.DecodeString(keyStr); err == nil && len(decoded) == 32 {
				var key [32]byte
				copy(key[:], decoded)
				cw.Network.Keys[uint32(nodeID)] = wgtypes.Key(key)
			}
		}
	}

	return nil
}

// GetCluster retrieves a cluster from workflow state with robust deserialization
func GetCluster(state ewf.State) (kubedeployer.Cluster, error) {
	value, ok := state["cluster"]
	if !ok {
		return kubedeployer.Cluster{}, fmt.Errorf("missing 'cluster' in state")
	}

	// Try direct type assertion first (for newly created clusters)
	if cluster, ok := value.(kubedeployer.Cluster); ok {
		return cluster, nil
	}

	// Check if it's stored as JSON with our custom wrapper
	if clusterStr, ok := value.(string); ok {
		if strings.HasPrefix(clusterStr, "wrap:") {
			jsonData := clusterStr[5:] // Remove "wrap:" prefix
			var wrapper ClusterWrapper
			if err := json.Unmarshal([]byte(jsonData), &wrapper); err != nil {
				return kubedeployer.Cluster{}, fmt.Errorf("failed to unmarshal cluster wrapper: %w", err)
			}
			return wrapper.Cluster, nil
		}
	}

	// Fallback: try to unmarshal as map and convert manually
	clusterBytes, err := json.Marshal(value)
	if err != nil {
		return kubedeployer.Cluster{}, fmt.Errorf("failed to marshal cluster value: %w", err)
	}

	var wrapper ClusterWrapper
	if err := json.Unmarshal(clusterBytes, &wrapper); err != nil {
		return kubedeployer.Cluster{}, fmt.Errorf("failed to unmarshal cluster: %w", err)
	}

	return wrapper.Cluster, nil
}

// StoreCluster safely stores the cluster in state using custom JSON marshaling
func StoreCluster(state ewf.State, cluster kubedeployer.Cluster) {
	// Wrap the cluster and marshal to JSON
	wrapper := ClusterWrapper{Cluster: cluster}
	if jsonData, err := json.Marshal(wrapper); err == nil {
		state["cluster"] = "wrap:" + string(jsonData)
	} else {
		// Fallback to direct storage if marshaling fails
		log.Warn().Err(err).Msg("Failed to marshal cluster wrapper, falling back to direct storage")
		state["cluster"] = cluster
	}
}
