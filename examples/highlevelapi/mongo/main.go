package main

import (
	"context"
	"log"

	"github.com/romnn/testcontainers"
	tcinfra "github.com/romnn/testcontainers/infra"
)

func main() {
	ctx := context.Background()
	mongoContainerName := "mongo-01-test"
	mongoContainerPort := 2189

	testcontainers.DropContainerIfExists(mongoContainerName)

	db, terminate, err := tcinfra.Mongo(ctx,
		tcinfra.MongoContainerName(mongoContainerName),
		tcinfra.MongoContainerPort(mongoContainerPort),
	)
	if err != nil {
		log.Fatal(err)
	}

	// your testing logic ...
	_ = db

	terminate()
}
