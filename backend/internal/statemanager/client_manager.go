package statemanager

import (
	"fmt"

	"kubecloud/kubedeployer"

	"github.com/rs/zerolog/log"
	"github.com/xmonader/ewf"
)

// GetKubeClient retrieves or creates a kubeclient with proper state management
func GetKubeClient(state ewf.State, config ClientConfig) (*kubedeployer.Client, error) {
	// Try to get existing kubeclient from state
	if value, ok := state["kubeclient"]; ok {
		if client, ok := value.(*kubedeployer.Client); ok && client != nil {
			log.Debug().Msg("Using existing kubeclient from state")
			return client, nil
		}
	}

	// If no existing client or it's invalid, create a fresh one
	if err := ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	kubeClient, err := kubedeployer.NewClient(config.Mnemonic, config.Network)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubeclient: %w", err)
	}

	// Restore GridClient state if it exists
	if err := RestoreGridClientState(state, kubeClient); err != nil {
		log.Warn().Err(err).Msg("Failed to restore GridClient state, continuing with fresh state")
	}

	// Store the new client in state for reuse
	state["kubeclient"] = kubeClient

	// Save the GridClient state for restart safety
	SaveGridClientState(state, kubeClient)

	log.Debug().Msg("Created and stored fresh kubeclient")
	return kubeClient, nil
}

// ClientConfig represents the configuration needed to create a kubeclient
type ClientConfig struct {
	SSHPublicKey string `json:"ssh_public_key"`
	Mnemonic     string `json:"mnemonic"`
	UserID       string `json:"user_id"`
	Network      string `json:"network"`
}

// ValidateConfig validates the client configuration
func ValidateConfig(config ClientConfig) error {
	if config.SSHPublicKey == "" {
		return fmt.Errorf("missing SSH public key in config")
	}
	if config.Mnemonic == "" {
		return fmt.Errorf("missing mnemonic in config")
	}
	if config.UserID == "" {
		return fmt.Errorf("missing user ID in config")
	}
	if config.Network == "" {
		return fmt.Errorf("missing network in config")
	}
	return nil
}

// EnsureClient ensures a kubeclient is available and ready for use
func EnsureClient(state ewf.State, config ClientConfig) error {
	// Get or create kubeclient (this will handle state restoration)
	_, err := GetKubeClient(state, config)
	if err != nil {
		return fmt.Errorf("failed to ensure kubeclient: %w", err)
	}

	log.Debug().Msg("Kubeclient ensured and ready for use")
	return nil
}

// SaveClientStateAfterOperation saves the GridClient state after a deployment operation
func SaveClientStateAfterOperation(state ewf.State, kubeClient *kubedeployer.Client) {
	SaveGridClientState(state, kubeClient)
}

// CloseClient properly closes a kubeclient and saves final state
func CloseClient(state ewf.State, kubeClient *kubedeployer.Client) {
	if kubeClient != nil {
		// Save final GridClient state before closing
		SaveGridClientState(state, kubeClient)
		kubeClient.Close()
		delete(state, "kubeclient")
		log.Debug().Msg("Closed kubeclient and cleaned up state")
	}
}
