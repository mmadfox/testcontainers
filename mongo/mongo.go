package mongo

import (
	"context"
	"fmt"
	"time"

	tc "github.com/mmadfox/testcontainers"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const DefaultStartupTimeout = 30 * time.Second

// Options ...
type Options struct {
	tc.ContainerOptions
	User           string
	Port           int
	Password       string
	ImageTag       string
	StartupTimeout time.Duration
}

// Container ...
type Container struct {
	Container testcontainers.Container
	tc.ContainerConfig
	Host     string
	Port     uint
	User     string
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
	var databaseAuth string
	if c.User != "" && c.Password != "" {
		databaseAuth = fmt.Sprintf("%s:%s@", c.User, c.Password)
	}
	databaseHost := fmt.Sprintf("%s:%d", c.Host, c.Port)
	return fmt.Sprintf("mongodb://%s%s/?connect=direct", databaseAuth, databaseHost)
}

// Start ...
func Start(ctx context.Context, options Options) (Container, error) {
	var container Container
	container.User = options.User
	container.Password = options.Password

	if options.Port <= 0 {
		options.Port = 27017
	}
	//
	//port, err := nat.NewPort("", strconv.Itoa(options.Port))
	//if err != nil {
	//	return container, fmt.Errorf("failed to build port: %v", err)
	//}

	env := make(map[string]string)
	if options.User != "" && options.Password != "" {
		env["MONGO_INITDB_ROOT_USERNAME"] = options.User
		env["MONGO_INITDB_ROOT_PASSWORD"] = options.Password
	}

	timeout := options.ContainerOptions.StartupTimeout
	if int64(timeout) < 1 {
		timeout = 5 * time.Minute // Default timeout
	}

	tag := "latest"
	if options.ImageTag != "" {
		tag = options.ImageTag
	}

	exposedPorts := []string{
		fmt.Sprintf("%d:%s", options.Port, "27017"),
	}

	req := testcontainers.ContainerRequest{
		Image:        fmt.Sprintf("mongo:%s", tag),
		Env:          env,
		ExposedPorts: exposedPorts,
		Cmd:          []string{},
		WaitingFor:   wait.ForListeningPort("27017").WithStartupTimeout(timeout),
	}

	tc.MergeRequest(&req, &options.ContainerOptions.ContainerRequest)

	mongoContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return container, fmt.Errorf("failed to start container: %v", err)
	}
	container.Container = mongoContainer

	host, err := mongoContainer.Host(ctx)
	if err != nil {
		return container, fmt.Errorf("failed to get container host: %v", err)
	}
	container.Host = host

	realPort, err := mongoContainer.MappedPort(ctx, "27017")
	if err != nil {
		return container, fmt.Errorf("failed to get exposed container port: %v", err)
	}
	container.Port = uint(realPort.Int())

	return container, nil
}
