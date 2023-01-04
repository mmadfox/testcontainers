package infra

import (
	"context"

	tc "github.com/romnn/testcontainers"

	tckafka "github.com/romnn/testcontainers/kafka"
)

type KafkaOption func(options *kafkaOptions)

type kafkaOptions struct {
	container *tckafka.Options
	logger    bool
}

type KafkaBroker struct {
	Addr    []string
	Version string
}

func Kafka(ctx context.Context, opts ...KafkaOption) (broker KafkaBroker, terminate func(), err error) {
	tcOpts := &kafkaOptions{
		container: &tckafka.Options{},
	}
	for _, fn := range opts {
		fn(tcOpts)
	}
	tcOpts.container.AutoRemove = true

	container, err := tckafka.Start(ctx, *tcOpts.container)
	if err != nil {
		return broker, nil, err
	}
	defer func() {
		if err != nil {
			container.Terminate(ctx)
		}
	}()

	var kafkaLogger, zookeeperLogger tc.LogCollector

	if tcOpts.logger {
		kafkaLogger, err = tc.StartLogger(ctx, container.Kafka.Container)
		if err != nil {
			return broker, nil, err
		} else {
			go kafkaLogger.LogToStdout()
		}

		zookeeperLogger, err = tc.StartLogger(ctx, container.Zookeeper.Container)
		if err != nil {
			return broker, nil, err
		} else {
			go zookeeperLogger.LogToStdout()
		}
	}

	broker.Addr = container.Kafka.Brokers
	broker.Version = container.Kafka.Version

	return broker, func() {
		container.Terminate(ctx)

		if kafkaLogger.LogChan != nil {
			kafkaLogger.Stop()
		}
		if zookeeperLogger.LogChan != nil {
			zookeeperLogger.Stop()
		}
	}, nil
}

func KafkaEnableLogger() KafkaOption {
	return func(opts *kafkaOptions) {
		opts.logger = true
	}
}

func KafkaContainerName(name string) KafkaOption {
	return func(opts *kafkaOptions) {
		opts.container.Name = name
	}
}

func ZookeeperContainerName(name string) KafkaOption {
	return func(opts *kafkaOptions) {
		opts.container.ZookeeperName = name
	}
}

func KafkaImageTag(tag string) KafkaOption {
	return func(opts *kafkaOptions) {
		opts.container.KafkaImageTag = tag
	}
}

func ZookeeperImageTag(tag string) KafkaOption {
	return func(opts *kafkaOptions) {
		opts.container.ZookeeperImageTag = tag
	}
}

func KafkaContainerNetwork(networks []string) KafkaOption {
	return func(opts *kafkaOptions) {
		opts.container.Networks = networks
	}
}
