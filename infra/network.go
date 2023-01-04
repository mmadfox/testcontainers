package infra

import (
	"context"
	"fmt"
	"time"

	tc "github.com/mmadfox/testcontainers"

	"github.com/testcontainers/testcontainers-go"
)

const networkTimeout = 10 * time.Second

func BridgeNetwork(_ context.Context, name string) (testcontainers.Network, error) {
	net, err := tc.CreateNetwork(testcontainers.NetworkRequest{
		Driver:         "bridge",
		Name:           name,
		Attachable:     true,
		CheckDuplicate: true,
	}, networkTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to create network: %v", err)
	}
	return net, nil
}
