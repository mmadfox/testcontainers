package main

import (
	"context"
	"log"

	"github.com/romnn/testcontainers"

	"github.com/romnn/testcontainers/infra"
)

func main() {
	myInfra := infra.NewSets()
	defer myInfra.Close()

	// reset containers if needed
	testcontainers.DropNetwork(myInfra.ContainerNames.Network)
	testcontainers.DropContainerIfExists(myInfra.ContainerNames.Redis)
	testcontainers.DropContainerIfExists(myInfra.ContainerNames.Mongo)
	testcontainers.DropContainerIfExists(myInfra.ContainerNames.Kafka)
	testcontainers.DropContainerIfExists(myInfra.ContainerNames.Zookeeper)

	ctx := context.Background()
	myInfra.SetupBridgeNetwork(ctx)
	myInfra.SetupRedis(ctx)
	myInfra.SetupMongo(ctx)
	myInfra.SetupKafka(ctx)

	if myInfra.Err() != nil {
		log.Fatal(myInfra.Err())
	}

	// your testing logic ...

	// myInfra.RedisClient() => storage, repository, etc
	// myInfra.MongoDB()     => storage, repository, etc
	// myInfra.KafkaAddr()   => producing, consuming, etc
}
