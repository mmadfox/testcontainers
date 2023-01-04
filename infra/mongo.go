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
	container *tcmongo.Options
	logger    bool
}

func Mongo(ctx context.Context, opts ...MongoOption) (db *mongo.Database, terminate func(), err error) {
	tcOpts := &mongoOptions{
		container: &tcmongo.Options{},
	}
	for _, fn := range opts {
		fn(tcOpts)
	}
	tcOpts.container.AutoRemove = true

	container, err := tcmongo.Start(ctx, *tcOpts.container)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if err != nil {
			container.Terminate(ctx)
		}
	}()

	var logger tc.LogCollector

	if tcOpts.logger {
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
