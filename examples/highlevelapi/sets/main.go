package main

import (
	"context"
	"log"

	"github.com/romnn/testcontainers/infra"

	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	myInfra := new(infrastructure)
	defer myInfra.close()

	ctx := context.Background()
	myInfra.setupRedis(ctx)
	myInfra.setupMongo(ctx)
	myInfra.setupKafka(ctx)

	if myInfra.err != nil {
		log.Fatal(myInfra.err)
	}

	// your testing logic ...
}

type infrastructure struct {
	redis          *redis.Client
	mongo          *mongo.Database
	kafkaAddr      []string
	terminates     []func()
	containerNames []string
	err            error
}

func (i *infrastructure) close() {
	for x := 0; x < len(i.terminates); x++ {
		i.terminates[x]()
	}
	infra.DropContainers(i.containerNames)
}

func (i *infrastructure) register(terminate func(), containerName ...string) {
	i.terminates = append(i.terminates, terminate)
	i.containerNames = append(i.containerNames, containerName...)
}

func (i *infrastructure) setupRedis(ctx context.Context) {
	if i.err != nil {
		return
	}

	containerName := "test-redis"
	conn, terminate, err := infra.Redis(ctx,
		infra.RedisContainerName(containerName),
		infra.RedisContainerPort(3890),
	)
	if err != nil {
		i.err = err
		return
	}

	i.redis = conn
	i.register(terminate, containerName)
}

func (i *infrastructure) setupMongo(ctx context.Context) {
	if i.err != nil {
		return
	}

	containerName := "test-mongo"
	db, terminate, err := infra.Mongo(ctx,
		infra.MongoContainerName(containerName),
		infra.MongoContainerPort(2189),
	)
	if err != nil {
		i.err = err
		return
	}

	i.mongo = db
	i.register(terminate, containerName)
}

func (i *infrastructure) setupKafka(ctx context.Context) {
	if i.err != nil {
		return
	}

	kafkaContainerName := "test-kafka"
	zookeeperContainerName := "test-zoo"
	broker, terminate, err := infra.Kafka(ctx,
		infra.KafkaContainerName(kafkaContainerName),
		infra.ZookeeperContainerName(zookeeperContainerName),
	)
	if err != nil {
		i.err = err
		return
	}

	i.kafkaAddr = broker.Addr
	i.register(terminate, kafkaContainerName, zookeeperContainerName)
}
