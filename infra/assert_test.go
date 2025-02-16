package infra

import (
	"net"
	"os"
	"strconv"
	"syscall"
	"testing"

	"github.com/mmadfox/testcontainers"

	"github.com/stretchr/testify/require"
)

func assertPortIsOpened(t *testing.T, port int) {
	var check bool
	_, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		if netErr, ok := err.(*net.OpError); ok {
			syscallErr, ok := netErr.Err.(*os.SyscallError)
			if ok {
				check = true
				require.Equal(t, syscallErr.Err, syscall.EADDRINUSE)
			}
		}
	}
	require.Truef(t, check, "port %d is closed", port)
}

func assertPortIsClosed(t *testing.T, port int) {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	require.NoError(t, err)
	require.NotNil(t, ln)
	_ = ln.Close()
}

func assertContainerExists(t *testing.T, name string) {
	exists, err := testcontainers.ContainerExists(name)
	require.NoError(t, err)
	require.True(t, exists)
}

func assertContainerNotExists(t *testing.T, name string) {
	exists, _ := testcontainers.ContainerExists(name)
	require.False(t, exists)
}
