package mongo

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	tc "github.com/mmadfox/testcontainers"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type ReplicaSetContainer struct {
	MasterContainer     testcontainers.Container
	ReplicaSet1         testcontainers.Container
	ReplicaSet2         testcontainers.Container
	MasterContainerAddr Addr
	ReplicaSet1Addr     Addr
	ReplicaSet2Addr     Addr
	ContainerNames      []string
	NetworkName         string
	Network             testcontainers.Network
	User                string
	Password            string
}

type Addr struct {
	Host string
	Port uint
}

func (c *ReplicaSetContainer) MasterConnectionURI() string {
	return c.connectionURIFrom(c.MasterContainerAddr, true)
}

func (c *ReplicaSetContainer) ReplicaSet1ConnectionURI() string {
	return c.connectionURIFrom(c.ReplicaSet1Addr, false)
}

func (c *ReplicaSetContainer) ReplicaSet2ConnectionURI() string {
	return c.connectionURIFrom(c.ReplicaSet2Addr, false)
}

func (c *ReplicaSetContainer) connectionURIFrom(a Addr, master bool) string {
	var databaseAuth string
	if c.User != "" && c.Password != "" {
		databaseAuth = fmt.Sprintf("%s:%s@", c.User, c.Password)
	}
	databaseHost := fmt.Sprintf("%s:%d", a.Host, a.Port)
	return fmt.Sprintf("mongodb://%s%s/?connect=direct&retryWrites=true&w=majority&readPreference=primaryPreferred&replicaSet=rs0", databaseAuth, databaseHost)
}

func (c *ReplicaSetContainer) Terminate(ctx context.Context) {
	if c.MasterContainer != nil {
		_ = c.MasterContainer.Terminate(ctx)
	}
	if c.ReplicaSet1 != nil {
		_ = c.ReplicaSet1.Terminate(ctx)
	}
	if c.ReplicaSet2 != nil {
		_ = c.ReplicaSet2.Terminate(ctx)
	}
	if c.Network != nil {
		_ = c.Network.Remove(ctx)
	}
}

func StartReplicaSet(ctx context.Context, options Options) (cont *ReplicaSetContainer, err error) {
	if options.StartupTimeout <= 0 {
		options.StartupTimeout = DefaultStartupTimeout
	}

	tag := "latest"
	if options.ImageTag != "" {
		tag = options.ImageTag
	}

	var m1, rs2, rs3 testcontainers.Container
	var net testcontainers.Network
	var networkName string
	var m1Name, rs2Name, rs3Name string

	defer func() {
		if err == nil {
			return
		}
		if m1 != nil {
			_ = m1.Terminate(ctx)
		}
		if rs2 != nil {
			_ = rs2.Terminate(ctx)
		}
		if rs3 != nil {
			_ = rs3.Terminate(ctx)
		}
		if net != nil {
			_ = net.Remove(ctx)
		}
	}()

	if len(options.Name) == 0 {
		options.Name = "mongo-replicaset-" + tc.UniqueID()
	}

	m1Name = options.Name + "-m1"
	rs2Name = options.Name + "-rs2"
	rs3Name = options.Name + "-rs3"

	if len(options.Networks) < 1 {
		networkName = fmt.Sprintf("mongo-replicaset-%s", tc.UniqueID())
		net, err = tc.CreateNetwork(testcontainers.NetworkRequest{
			Driver:         "bridge",
			Name:           networkName,
			Attachable:     true,
			CheckDuplicate: true,
		}, 2)
		if err != nil {
			return nil, fmt.Errorf("failed to create network: %v", err)
		}
	} else {
		networkName = options.Networks[0]
	}

	req1 := testcontainers.ContainerRequest{
		Image: fmt.Sprintf("mongo:%s", tag),
		NetworkAliases: map[string][]string{
			networkName: {"master"},
		},
		Networks:     []string{networkName},
		Name:         m1Name,
		ExposedPorts: []string{"27017/"},
		Hostname:     "master",
		Cmd:          []string{"--replSet", "rs0", "--bind_ip", "localhost,master"},
		WaitingFor:   wait.ForListeningPort("27017").WithStartupTimeout(options.StartupTimeout),
	}
	m1, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req1,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %v", err)
	}

	req2 := testcontainers.ContainerRequest{
		Image: fmt.Sprintf("mongo:%s", tag),
		NetworkAliases: map[string][]string{
			networkName: {"rs2"},
		},
		Name:         rs2Name,
		Networks:     []string{networkName},
		ExposedPorts: []string{"27017/"},
		Hostname:     "rs2",
		Cmd:          []string{"--replSet", "rs0", "--bind_ip", "localhost,rs2"},
		WaitingFor:   wait.ForListeningPort("27017").WithStartupTimeout(options.StartupTimeout),
	}
	rs2, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req2,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %v", err)
	}

	req3 := testcontainers.ContainerRequest{
		Image:        fmt.Sprintf("mongo:%s", tag),
		Name:         rs3Name,
		ExposedPorts: []string{"27017/"},
		Networks:     []string{networkName},
		NetworkAliases: map[string][]string{
			networkName: {"rs3"},
		},
		Hostname:   "rs3",
		Cmd:        []string{"--replSet", "rs0", "--bind_ip", "localhost,rs3"},
		WaitingFor: wait.ForListeningPort("27017").WithStartupTimeout(options.StartupTimeout),
	}
	rs3, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req3,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %v", err)
	}

	if err = m1.Start(ctx); err != nil {
		return nil, err
	}
	if err = rs2.Start(ctx); err != nil {
		return nil, err
	}
	if err = rs3.Start(ctx); err != nil {
		return nil, err
	}

	cont = &ReplicaSetContainer{
		MasterContainer: m1,
		ReplicaSet1:     rs2,
		ReplicaSet2:     rs3,
		ContainerNames:  []string{m1Name, rs2Name, rs3Name},
		NetworkName:     networkName,
		Network:         net,
	}

	if cont.MasterContainerAddr, err = containerAddr(ctx, m1); err != nil {
		return nil, err
	}
	if cont.ReplicaSet1Addr, err = containerAddr(ctx, rs2); err != nil {
		return nil, err
	}
	if cont.ReplicaSet2Addr, err = containerAddr(ctx, rs3); err != nil {
		return nil, err
	}

	if err = runCreateReplicaSet(ctx, m1); err != nil {
		return nil, err
	}
	if err = runCheckIsMasterNode(ctx, m1); err != nil {
		return nil, err
	}

	for i := 0; i < 60; i++ {
		if ok := waitPrimaryNode(ctx, rs3); ok {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	return cont, nil
}

func runCreateReplicaSet(ctx context.Context, c testcontainers.Container) error {
	var wrapErr = func(err error) error {
		return fmt.Errorf("failed to create replica set. error: %w", err)
	}
	var cmd = []string{
		`mongosh`,
		`--eval`,
		`'printjson(rs.initiate(
{_id:"rs0","members":[
{_id:0,host:"master:27017"},
{_id:1,host:"rs2:27017"},
{_id:2,host:"rs3:27017"}
]}))'`,
		`--quiet`,
	}
	_, r, err := c.Exec(ctx, cmd)
	if err != nil {
		return wrapErr(err)
	}
	b, err := io.ReadAll(r)
	if err != nil {
		return wrapErr(err)
	}
	if strings.ContainsAny(string(b), "ok") {
		return nil
	}
	return fmt.Errorf("failed to create replica set")
}

func runCheckIsMasterNode(ctx context.Context, c testcontainers.Container) error {
	var wrapErr = func(err error) error {
		return fmt.Errorf("failed to check master node. error: %w", err)
	}
	cmd := []string{`mongosh`, `--eval`, `'printjson(rs.isMaster())'`}
	_, r, err := c.Exec(ctx, cmd)
	if err != nil {
		return wrapErr(err)
	}
	b, err := io.ReadAll(r)
	if err != nil {
		return wrapErr(err)
	}
	if strings.ContainsAny(string(b), "ismaster: true") {
		return nil
	}
	return fmt.Errorf("failed to check master node")
}

func waitPrimaryNode(ctx context.Context, c testcontainers.Container) bool {
	cmd := []string{`mongosh`, `--eval`, `rs.status()`, `--quiet`}
	_, r, _ := c.Exec(ctx, cmd)
	b, _ := io.ReadAll(r)
	if strings.Contains(string(b), "PRIMARY") {
		return true
	}
	return false
}

func containerAddr(ctx context.Context, c testcontainers.Container) (Addr, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return Addr{}, fmt.Errorf("failed to get container host: %v", err)
	}
	realPort, err := c.MappedPort(ctx, "27017")
	if err != nil {
		return Addr{}, fmt.Errorf("failed to get exposed container port: %v", err)
	}
	port := uint(realPort.Int())
	return Addr{
		Host: host,
		Port: port,
	}, nil
}
