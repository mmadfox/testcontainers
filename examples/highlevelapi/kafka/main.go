package main

import (
	"context"
	"fmt"
	"log"

	tcinfra "github.com/romnn/testcontainers/infra"
)

func main() {
	ctx := context.Background()

	kafkaContainerName := "infra-01-kafka-container"
	zookeeperContainerName := "infra-01-zookeeper-container"

	tcinfra.DropContainers([]string{
		kafkaContainerName,
		zookeeperContainerName,
	})

	broker, terminate, err := tcinfra.Kafka(ctx,
		tcinfra.KafkaContainerName(kafkaContainerName),
		tcinfra.ZookeeperContainerName(zookeeperContainerName))
	if err != nil {
		log.Fatal(err)
	}
	defer terminate()

	// your testing logic ...
	// consuming, producing
	_ = broker.Addr

	kafkaExists, err := tcinfra.ContainerExists(kafkaContainerName)
	if err != nil {
		log.Fatal(err)
	}
	zooExists, err := tcinfra.ContainerExists(zookeeperContainerName)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("kafkaExists=%v, zookeeperExists=%v, kafkaAddr=%v \n",
		kafkaExists, zooExists, broker.Addr)
}
