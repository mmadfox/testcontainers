package infra

import (
	"context"
	"testing"
	"time"

	"github.com/mmadfox/testcontainers"
	"github.com/stretchr/testify/require"
)

func TestRedis(t *testing.T) {
	ctx := context.Background()
	redisPort := 6382
	testVal := "test"
	containerName := "infra-01-redis-container"

	testcontainers.DropContainerIfExists(containerName)

	db, terminate, err := Redis(ctx,
		RedisContainerPort(redisPort),
		RedisContainerName(containerName),
	)
	require.NoError(t, err)
	require.NotNil(t, terminate)
	require.NotNil(t, db)

	assertPortIsOpened(t, redisPort)
	assertContainerExists(t, containerName)

	db.Set(testVal, testVal, time.Minute)
	cmd := db.Get(testVal)
	require.NotNil(t, cmd)
	require.EqualValues(t, testVal, cmd.Val())

	terminate()

	assertPortIsClosed(t, redisPort)
	assertContainerNotExists(t, containerName)
}
