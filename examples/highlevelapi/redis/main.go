package main

import (
	"context"
	"fmt"
	"log"

	"github.com/romnn/testcontainers"
	tcinfra "github.com/romnn/testcontainers/infra"
)

func main() {
	ctx := context.Background()
	redisContainerName := "redis-01-test"
	redisContainerPort := 6718

	testcontainers.DropContainerIfExists(redisContainerName)

	db, terminate, err := tcinfra.Redis(ctx,
		tcinfra.RedisContainerName(redisContainerName),
		tcinfra.RedisContainerPort(redisContainerPort),
	)
	if err != nil {
		log.Fatal(err)
	}

	// your testing logic ...
	db.Set("key", "value", 0)
	cmd := db.Get("key")

	fmt.Println("value", "==", cmd.Val())

	terminate()
}
