package testsuite

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/go-redis/redis"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/require"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/mmadfox/testcontainers/infra"

	"github.com/stretchr/testify/suite"
)

func TestSomeSuite(t *testing.T) {
	suite.Run(t, new(someTestSuite))
}

type someTestSuite struct {
	suite.Suite

	infra *infra.Sets

	producer     sarama.AsyncProducer
	redisStorage *redisStorage
	mongoStorage *mongoStorage
}

func (s *someTestSuite) SetupSuite() {
	ctx := context.Background()

	s.infra = infra.NewSets()
	s.infra.SetupBridgeNetwork(ctx)
	s.infra.SetupKafka(ctx)
	s.infra.SetupMongoReplicaSet(ctx)
	s.infra.SetupRedis(ctx)

	require.NoError(s.T(), s.infra.Err())

	conf := sarama.NewConfig()
	conf.Version = sarama.V1_0_2_0
	conf.Producer.Return.Successes = true
	conf.Producer.Return.Errors = true
	producer, err := sarama.NewAsyncProducer(s.infra.KafkaAddr(), conf)
	require.NoError(s.T(), err)
	s.producer = producer

	s.redisStorage = &redisStorage{s.infra.RedisClient()}
	s.mongoStorage = &mongoStorage{s.infra.MongoDB().Collection("test")}

}

func (s *someTestSuite) TearDownSuite() {
	s.infra.Close()
}

func (s *someTestSuite) SetupTest() {

}

func (s *someTestSuite) TestWithMongo() {
	s.mongoStorage.Insert()
}

func (s *someTestSuite) TestWithRedis() {
	s.redisStorage.Set("k", "v")
}

func (s *someTestSuite) TestWithKafka() {
	s.producer.Input() <- &sarama.ProducerMessage{
		Topic: "test",
		Value: sarama.StringEncoder("test"),
	}
}

type mongoStorage struct {
	col *mongo.Collection
}

func (ms *mongoStorage) Insert() {
	_, _ = ms.col.InsertOne(context.Background(), bson.M{"key": "val"})
}

type redisStorage struct {
	cli *redis.Client
}

func (rs *redisStorage) Set(key, value string) {
	rs.cli.Set(key, value, 0)
}
