package kubedeployer

import (
	"context"
	"fmt"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
)

type Client struct {
	ctx          context.Context
	GridClient   deployer.TFPluginClient
	gridNet      string
	mnemonic     string
	masterPubKey string
	UserID       string
}

func NewClient(ctx context.Context, mnemonic, gridNet, masterPubKey, userID string) (*Client, error) {
	tfplugin, err := deployer.NewTFPluginClient(
		mnemonic,
		deployer.WithNetwork(gridNet),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create TFPluginClient: %v", err)
	}

	return &Client{
		ctx:          ctx,
		GridClient:   tfplugin,
		gridNet:      gridNet,
		mnemonic:     mnemonic,
		masterPubKey: masterPubKey,
		UserID:       userID,
	}, nil
}

func (c *Client) Close() {
	c.GridClient.Close()
}
