package kubedeployer

import (
	"fmt"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
)

type Client struct {
	GridClient deployer.TFPluginClient
}

func NewClient(mnemonic, gridNet string) (*Client, error) {
	tfplugin, err := deployer.NewTFPluginClient(
		mnemonic,
		deployer.WithNetwork(gridNet),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create TFPluginClient: %v", err)
	}

	return &Client{
		GridClient: tfplugin,
	}, nil
}

func (c *Client) Close() {
	c.GridClient.Close()
}
