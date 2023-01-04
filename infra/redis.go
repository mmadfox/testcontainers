package infra

import (
	"context"
	"time"

	"github.com/go-redis/redis"
	"github.com/mmadfox/testcontainers"
	tcredis "github.com/mmadfox/testcontainers/redis"
)

type RedisOption func(*redisOptions)

type redisOptions struct {
	container *tcredis.Options
	server    *redis.Options
	logger    bool
}

func Redis(ctx context.Context, opts ...RedisOption) (cli *redis.Client, terminate func(), err error) {
	tcOpts := &redisOptions{
		container: &tcredis.Options{},
		server: &redis.Options{
			DB: 1,
		},
	}
	for _, fn := range opts {
		fn(tcOpts)
	}
	tcOpts.container.AutoRemove = true

	container, err := tcredis.Start(ctx, *tcOpts.container)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if err != nil {
			container.Terminate(ctx)
		}
	}()

	var logger testcontainers.LogCollector

	if tcOpts.logger {
		logger, err = testcontainers.StartLogger(ctx, container.Container)
		if err != nil {
			return nil, nil, err
		} else {
			go logger.LogToStdout()
		}
	}

	tcOpts.server.Addr = container.ConnectionURI()
	db := redis.NewClient(tcOpts.server)

	return db, func() {
		_ = db.Close()
		if logger.LogChan != nil {
			logger.Stop()
		}
		container.Terminate(ctx)
	}, nil
}

func RedisEnableLogger() RedisOption {
	return func(opts *redisOptions) {
		opts.logger = true
	}
}

func RedisServerOptions(serverOpts *redis.Options) RedisOption {
	return func(opts *redisOptions) {
		opts.server = serverOpts
	}
}

func RedisContainerNetwork(networks []string) RedisOption {
	return func(opts *redisOptions) {
		opts.container.Networks = networks
	}
}

func RedisContainerName(name string) RedisOption {
	return func(opts *redisOptions) {
		opts.container.Name = name
	}
}

func RedisContainerPort(port int) RedisOption {
	return func(opts *redisOptions) {
		opts.container.Port = port
	}
}

func RedisImageTag(tag string) RedisOption {
	return func(opts *redisOptions) {
		opts.container.ImageTag = tag
	}
}

func RedisContainerBootstrapTimeout(timeout time.Duration) RedisOption {
	return func(opts *redisOptions) {
		opts.container.StartupTimeout = timeout
	}
}

func RedisContainerEnv(envs map[string]string) RedisOption {
	return func(opts *redisOptions) {
		opts.container.Env = envs
	}
}
