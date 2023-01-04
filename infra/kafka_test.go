package infra

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKafka(t *testing.T) {
	ctx := context.Background()
	kafkaContainerName := "infra-01-kafka-container"
	zookeeperContainerName := "infra-01-zookeeper-container"

	DropContainers([]string{
		kafkaContainerName,
		zookeeperContainerName,
	})

	broker, terminate, err := Kafka(ctx,
		KafkaContainerName(kafkaContainerName),
		ZookeeperContainerName(zookeeperContainerName))

	require.NoError(t, err)
	require.NotNil(t, terminate)
	require.NotEmpty(t, broker.Addr)
	require.NotEmpty(t, broker.Version)

	kafkaContainerExists, err := ContainerExists(kafkaContainerName)
	require.NoError(t, err)
	require.True(t, kafkaContainerExists)

	zooContainerExists, err := ContainerExists(zookeeperContainerName)
	require.NoError(t, err)
	require.True(t, zooContainerExists)

	terminate()

	kafkaContainerExists, err = ContainerExists(kafkaContainerName)
	require.NoError(t, err)
	require.False(t, kafkaContainerExists)

	zooContainerExists, err = ContainerExists(zookeeperContainerName)
	require.NoError(t, err)
	require.False(t, zooContainerExists)
}
