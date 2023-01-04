package redis

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/docker/go-connections/nat"
	tc "github.com/romnn/testcontainers"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Options ...
type Options struct {
	tc.ContainerOptions

	Port     int
	Password string
	ImageTag string
}

// Container ...
type Container struct {
	Container testcontainers.Container
	tc.ContainerConfig
	Host     string
	Port     int64
	Password string
}

// Terminate ...
func (c *Container) Terminate(ctx context.Context) {
	if c.Container != nil {
		_ = c.Container.Terminate(ctx)
	}
}

// ConnectionURI ...
func (c *Container) ConnectionURI() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Start ...
func Start(ctx context.Context, options Options) (Container, error) {
	var container Container
	if options.Port <= 0 {
		options.Port = 6379
	}
	port, err := nat.NewPort("", strconv.Itoa(options.Port))
	if err != nil {
		return container, fmt.Errorf("failed to build port: %v", err)
	}

	timeout := options.ContainerOptions.StartupTimeout
	if int64(timeout) < 1 {
		timeout = time.Minute // Default timeout
	}
	rawPort := strings.Trim(string(port), "/")

	tag := "latest"
	if options.ImageTag != "" {
		tag = options.ImageTag
	}
	exposedPorts := []string{
		fmt.Sprintf("%s:%s", rawPort, "6379"),
	}
	req := testcontainers.ContainerRequest{
		Image:        fmt.Sprintf("redis:%s", tag),
		ExposedPorts: exposedPorts,
		WaitingFor:   wait.ForListeningPort("6379").WithStartupTimeout(timeout),
	}

	if options.Password != "" {
		req.Cmd = []string{fmt.Sprintf("redis-server --requirepass %s", options.Password)}
		container.Password = options.Password
	}

	tc.MergeRequest(&req, &options.ContainerOptions.ContainerRequest)

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	container.Container = redisContainer

	if err != nil {
		return container, fmt.Errorf("failed to start container: %v", err)
	}

	host, err := redisContainer.Host(ctx)
	if err != nil {
		return container, fmt.Errorf("failed to get container host: %v", err)
	}
	container.Host = host

	realPort, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		return container, fmt.Errorf("failed to get exposed container port: %v", err)
	}
	container.Port = int64(realPort.Int())

	return container, nil
}
