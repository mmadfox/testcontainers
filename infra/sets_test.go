package infra

import (
	"context"
	"testing"

	"github.com/mmadfox/testcontainers"

	"github.com/stretchr/testify/require"
)

func TestSets(t *testing.T) {
	sets := NewSets()
	ctx := context.Background()
	defer sets.Close()

	testcontainers.DropNetwork(sets.ContainerNames.Network)

	sets.SetupBridgeNetwork(ctx)
	sets.SetupMongo(ctx)
	sets.SetupRedis(ctx)
	sets.SetupKafka(ctx)

	require.NoError(t, sets.Err())
	require.NotNil(t, sets.MongoDB())
	require.NotNil(t, sets.RedisClient())
	require.NotEmpty(t, sets.KafkaAddr())
}
