package main

import (
	"context"
	"fmt"
	tcinfra "github.com/romnn/testcontainers/infra"
	"log"
)

func main() {
	ctx := context.Background()
	redisContainerName := "redis-01-test"
	redisContainerPort := 6718

	tcinfra.DropContainerIfExists(redisContainerName)

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
