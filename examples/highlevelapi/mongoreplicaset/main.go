package main

import (
	"context"
	"log"

	tcinfra "github.com/mmadfox/testcontainers/infra"
)

func main() {
	ctx := context.Background()
	db, terminate, err := tcinfra.Mongo(ctx, tcinfra.MongoEnableReplicaSet())
	if err != nil {
		log.Fatal(err)
	}

	// your testing logic ...
	_ = db

	terminate()
}
