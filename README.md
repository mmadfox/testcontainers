## testcontainers, pre-configured

[![GitHub](https://img.shields.io/github/license/romnn/testcontainers)](https://github.com/romnn/testcontainers)
[![GoDoc](https://godoc.org/github.com/romnn/testcontainers?status.svg)](https://godoc.org/github.com/romnn/testcontainers)
[![Test Coverage](https://codecov.io/gh/romnn/testcontainers/branch/master/graph/badge.svg)](https://codecov.io/gh/romnn/testcontainers)

A collection of pre-configured [testcontainers](https://github.com/testcontainers/testcontainers-go) for your golang integration tests.

Available containers (feel free to contribute):

- MongoDB (based on [mongo](https://hub.docker.com/_/mongo))
- Kafka (based on [confluentinc/cp-kafka](https://hub.docker.com/r/confluentinc/cp-kafka) and [bitnami/zookeeper](https://hub.docker.com/r/bitnami/zookeeper))
- RabbitMQ (based on [rabbitmq](https://hub.docker.com/_/rabbitmq))
- Redis (based on [redis](https://hub.docker.com/_/redis/))
- Minio (based on [minio/minio](https://hub.docker.com/r/minio/minio))

### Usage

#### HighLevel API

##### Sets
```go
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
```

##### Redis container
```go
package main

import (
	"context"
	"fmt"
	"github.com/romnn/testcontainers"
	tcinfra "github.com/romnn/testcontainers/infra"
	"log"
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
```

##### MongoDB container
```go
package main

import (
	"context"
	tcinfra "github.com/romnn/testcontainers/infra"
	"github.com/romnn/testcontainers"
	"log"
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
```

##### Kafka container
```go
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

	kafkaContainerName := "infra-01-kafka-container"
	zookeeperContainerName := "infra-01-zookeeper-container"

	testcontainers.DropContainers([]string{
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

	kafkaExists, err := testcontainers.ContainerExists(kafkaContainerName)
	if err != nil {
		log.Fatal(err)
	}
	zooExists, err := testcontainers.ContainerExists(zookeeperContainerName)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("kafkaExists=%v, zookeeperExists=%v, kafkaAddr=%v \n",
		kafkaExists, zooExists, broker.Addr)
}
```

#### LowLevel API
##### Redis

```go
// examples/redis/redis.go

package main

import (
	"context"
	"log"

	"github.com/go-redis/redis"
	tc "github.com/romnn/testcontainers"
	tcredis "github.com/romnn/testcontainers/redis"
)

func main() {
	container, err := tcredis.Start(context.Background(), tcredis.Options{
		ImageTag: "7.0.5", // you could use latest here
	})
	if err != nil {
		log.Fatalf("failed to start container: %v", err)
	}
	defer container.Terminate(context.Background())

	// start logger
	logger, err := tc.StartLogger(context.Background(), container.Container)
	if err != nil {
		log.Printf("failed to start logger: %v", err)
	} else {
		defer logger.Stop()
		go logger.LogToStdout()
	}

	// connect to redis
	db := redis.NewClient(&redis.Options{
		Addr:     container.ConnectionURI(),
		Password: container.Password,
		DB:       1,
	})

	// set some data
	db.HSet("my-hash-key", "key", "Hello World!")

	// get the data back
	value, err := db.HGet("my-hash-key", "key").Result()
	if err != nil {
		log.Fatalf("failed to get value: %v", err)
	}
	if value != "Hello World!" {
		log.Fatalf(`received %q instead of "Hello World!"`, value)
	}

	log.Printf("received %q from redis", value)
}

```

##### MongoDB

```go
// examples/mongo/mongo.go

package main

import (
	"context"
	"log"
	"time"

	tcmongo "github.com/romnn/testcontainers/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	container, err := tcmongo.Start(context.Background(), tcmongo.Options{
		ImageTag: "6.0.2", // you could use latest here
	})
	if err != nil {
		log.Fatalf("failed to start container: %v", err)
	}
	defer container.Terminate(context.Background())

	// connect to the container
	uri := container.ConnectionURI()
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client.Connect(ctx)

	// count documents in collection
	collection := client.Database("testdatabase").Collection("my-collection")
	opts := options.Count().SetMaxTime(2 * time.Second)
	count, err := collection.CountDocuments(
		context.TODO(),
		bson.D{},
		opts,
	)
	if err != nil {
		log.Fatalf("failed to count docs in collection %q: %v", collection.Name(), err)
	}
	log.Printf("collection %q contains %d documents", collection.Name(), count)
}

```

For more examples, see `examples/`.

### Development

Before you get started, make sure you have installed the following tools:

    $ python3 -m pip install pre-commit bump2version invoke
    $ go install golang.org/x/tools/cmd/goimports@latest
    $ go install golang.org/x/lint/golint@latest
    $ go install github.com/fzipp/gocyclo/cmd/gocyclo@latest

**Remember**: To be able to excecute the tools installed with `go install`,
make sure to include `$GOPATH/bin` in your `$PATH`.
If `echo $GOPATH` does not give you a path make sure to run
(`export GOPATH="$HOME/go"` to set it). In order for your changes to persist,
do not forget to add these to your shells `.bashrc`.

With the tools in place, it is strongly advised to install the git commit hooks to make sure checks are passing in CI:

```bash
invoke install-hooks
```

You can check if all checks pass at any time:

```bash
invoke pre-commit
```
