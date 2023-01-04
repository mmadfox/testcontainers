package infra

import (
	"context"

	tc "github.com/romnn/testcontainers"

	"github.com/go-redis/redis"
	"github.com/testcontainers/testcontainers-go"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	DefaultMongo     = "test-mongo"
	DefaultRedis     = "test-redis"
	DefaultKafka     = "test-kafka"
	DefaultZookeeper = "test-zoo"
	DefaultNetwork   = "test-network"
)

type ContainerNames struct {
	Mongo     string
	Redis     string
	Kafka     string
	Zookeeper string
	Network   string
}

type Sets struct {
	ContainerNames ContainerNames

	redis          *redis.Client
	mongo          *mongo.Database
	kafkaAddr      []string
	kafkaVersion   string
	network        testcontainers.Network
	networkName    string
	terminates     []func()
	containerNames []string
	err            error
}

func NewSets() *Sets {
	return &Sets{
		ContainerNames: ContainerNames{
			Mongo:     DefaultMongo,
			Redis:     DefaultRedis,
			Kafka:     DefaultKafka,
			Zookeeper: DefaultZookeeper,
			Network:   DefaultNetwork,
		},
	}
}

func (i *Sets) Err() error {
	return i.err
}

func (i *Sets) RedisClient() *redis.Client {
	return i.redis
}

func (i *Sets) MongoDB() *mongo.Database {
	return i.mongo
}

func (i *Sets) Clear() {
}

func (i *Sets) Close() {
	for x := 0; x < len(i.terminates); x++ {
		i.terminates[x]()
	}
	tc.DropContainers(i.containerNames)
	if i.network != nil {
		_ = i.network.Remove(context.Background())
	}
}

func (i *Sets) KafkaAddr() []string {
	return i.kafkaAddr
}

func (i *Sets) KafkaVersion() string {
	return i.kafkaVersion
}

func (i *Sets) SetupBridgeNetwork(ctx context.Context) {
	if i.err != nil {
		return
	}

	i.networkName = i.ContainerNames.Network
	i.network, i.err = BridgeNetwork(ctx, i.networkName)
}

func (i *Sets) RemoveNetwork(ctx context.Context) error {
	if i.network == nil {
		return nil
	}
	return i.network.Remove(ctx)
}

func (i *Sets) SetupRedis(ctx context.Context) {
	if i.err != nil {
		return
	}

	opts := []RedisOption{
		RedisContainerName(i.ContainerNames.Redis),
		RedisContainerPort(3890),
	}
	if len(i.networkName) > 0 {
		opts = append(opts, RedisContainerNetwork([]string{i.networkName}))
	}
	conn, terminate, err := Redis(ctx, opts...)
	if err != nil {
		i.err = err
		return
	}

	i.redis = conn
	i.register(terminate, i.ContainerNames.Redis)
}

func (i *Sets) SetupMongo(ctx context.Context) {
	if i.err != nil {
		return
	}

	opts := []MongoOption{
		MongoContainerName(i.ContainerNames.Mongo),
		MongoContainerPort(2189),
	}
	if len(i.networkName) > 0 {
		opts = append(opts, MongoContainerNetwork([]string{i.networkName}))
	}
	db, terminate, err := Mongo(ctx, opts...)
	if err != nil {
		i.err = err
		return
	}

	i.mongo = db
	i.register(terminate, i.ContainerNames.Mongo)
}

func (i *Sets) SetupKafka(ctx context.Context) {
	if i.err != nil {
		return
	}

	opts := []KafkaOption{
		KafkaContainerName(i.ContainerNames.Kafka),
		ZookeeperContainerName(i.ContainerNames.Zookeeper),
	}
	if len(i.networkName) > 0 {
		opts = append(opts, KafkaContainerNetwork([]string{i.networkName}))
	}
	broker, terminate, err := Kafka(ctx, opts...)
	if err != nil {
		i.err = err
		return
	}

	i.kafkaAddr = broker.Addr
	i.kafkaVersion = broker.Version
	i.register(terminate, i.ContainerNames.Kafka, i.ContainerNames.Zookeeper)
}

func (i *Sets) register(terminate func(), containerName ...string) {
	i.terminates = append(i.terminates, terminate)
	i.containerNames = append(i.containerNames, containerName...)
}
