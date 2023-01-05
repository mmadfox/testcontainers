package infra

import (
	"context"
	"time"

	tc "github.com/mmadfox/testcontainers"
	tcmongo "github.com/mmadfox/testcontainers/mongo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoOption func(options *mongoOptions)

type mongoOptions struct {
	container  *tcmongo.Options
	logger     bool
	replicaSet bool
}

func Mongo(ctx context.Context, opts ...MongoOption) (db *mongo.Database, terminate func(), err error) {
	tcOpts := &mongoOptions{
		container: &tcmongo.Options{},
	}
	for _, fn := range opts {
		fn(tcOpts)
	}
	tcOpts.container.AutoRemove = true

	if tcOpts.replicaSet {
		return replicaSetMongo(ctx, tcOpts)
	} else {
		return standaloneMongo(ctx, tcOpts)
	}
}

func replicaSetMongo(ctx context.Context, opts *mongoOptions) (db *mongo.Database, terminate func(), err error) {
	container, err := tcmongo.StartReplicaSet(ctx, *opts.container)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if err != nil {
			container.Terminate(ctx)
			tc.DropContainers(container.ContainerNames)
		}
	}()

	mongoURI := container.MasterConnectionURI()
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, nil, err
	}
	if err = client.Connect(ctx); err != nil {
		return nil, nil, err
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, nil, err
	}
	database := client.Database("testdatabase")

	return database, func() {
		_ = client.Disconnect(ctx)
		container.Terminate(ctx)
		tc.DropContainers(container.ContainerNames)
	}, nil
}

func standaloneMongo(ctx context.Context, opts *mongoOptions) (db *mongo.Database, terminate func(), err error) {
	container, err := tcmongo.Start(ctx, *opts.container)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if err != nil {
			container.Terminate(ctx)
		}
	}()

	var logger tc.LogCollector

	if opts.logger {
		logger, err = tc.StartLogger(ctx, container.Container)
		if err != nil {
			return nil, nil, err
		} else {
			go logger.LogToStdout()
		}
	}

	mongoURI := container.ConnectionURI()
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, nil, err
	}
	if err = client.Connect(ctx); err != nil {
		return nil, nil, err
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, nil, err
	}
	database := client.Database("testdatabase")

	return database, func() {
		_ = client.Disconnect(ctx)
		if logger.LogChan != nil {
			logger.Stop()
		}
		container.Terminate(ctx)
	}, nil
}

func MongoEnableReplicaSet() MongoOption {
	return func(opts *mongoOptions) {
		opts.replicaSet = true
	}
}

func MongoEnableLogger() MongoOption {
	return func(opts *mongoOptions) {
		opts.logger = true
	}
}

func MongoContainerNetwork(networks []string) MongoOption {
	return func(opts *mongoOptions) {
		opts.container.Networks = networks
	}
}

func MongoContainerName(name string) MongoOption {
	return func(opts *mongoOptions) {
		opts.container.Name = name
	}
}

func MongoContainerPort(port int) MongoOption {
	return func(opts *mongoOptions) {
		opts.container.Port = port
	}
}

func MongoImageTag(tag string) MongoOption {
	return func(opts *mongoOptions) {
		opts.container.ImageTag = tag
	}
}

func MongoContainerBootstrapTimeout(timeout time.Duration) MongoOption {
	return func(opts *mongoOptions) {
		opts.container.StartupTimeout = timeout
	}
}

func MongoContainerEnv(envs map[string]string) MongoOption {
	return func(opts *mongoOptions) {
		opts.container.Env = envs
	}
}
