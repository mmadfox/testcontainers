package infra

import (
	"context"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

func TestMongo(t *testing.T) {
	ctx := context.Background()
	mongoPort := 27019
	testVal := "test"
	containerName := "infra-01-mongo-container"

	DropContainerIfExists(containerName)

	db, terminate, err := Mongo(ctx,
		MongoContainerPort(mongoPort),
		MongoContainerName(containerName),
	)
	require.NoError(t, err)
	require.NotNil(t, terminate)
	require.NotNil(t, db)

	assertPortIsOpened(t, mongoPort)
	assertContainerExists(t, containerName)

	type record struct {
		Value string `bson:"value"`
	}

	testCollection := db.Collection("test")
	res, err := testCollection.InsertOne(ctx, record{Value: testVal})
	require.NoError(t, err)
	require.NotNil(t, res)
	doc := testCollection.FindOne(ctx, bson.M{"_id": res.InsertedID})
	require.NotNil(t, doc)
	var val record
	require.NoError(t, doc.Decode(&val))
	require.Equal(t, testVal, val.Value)

	terminate()

	assertPortIsClosed(t, mongoPort)
	assertContainerNotExists(t, containerName)
}

func TestMongoMultiContainers(t *testing.T) {
	testCases := []struct {
		port int
		name string
	}{
		{
			port: 2198,
			name: "infra-011-mongo-container",
		},
		{
			port: 2199,
			name: "infra-012-mongo-container",
		},
	}

	ctx := context.Background()
	terminates := make([]func(), 0)
	for _, tc := range testCases {
		db, terminate, err := Mongo(ctx, MongoContainerPort(tc.port), MongoContainerName(tc.name))
		require.NoError(t, err)
		require.NotNil(t, terminate)
		require.NotNil(t, db)

		assertPortIsOpened(t, tc.port)
		assertContainerExists(t, tc.name)

		terminates = append(terminates, terminate)
	}

	for _, terminate := range terminates {
		terminate()
	}

	for _, tc := range testCases {
		assertPortIsClosed(t, tc.port)
		assertContainerNotExists(t, tc.name)
	}
}
