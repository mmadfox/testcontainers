package infra

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBridgeNetwork(t *testing.T) {
	ctx := context.Background()
	net, err := BridgeNetwork(ctx, "test1")
	require.NoError(t, err)
	require.NotNil(t, net)
	_ = net.Remove(ctx)
}
