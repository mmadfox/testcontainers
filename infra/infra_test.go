package infra

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDockerExists(t *testing.T) {
	require.True(t, DockerExists())
}
