package mongo

import (
	"context"
	"testing"

	tc "github.com/mmadfox/testcontainers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/stretchr/testify/require"
)

func TestMongoReplicaSet(t *testing.T) {
	tc.PruneNetwork()

	opts := Options{}
	opts.Name = "test"
	cont, err := StartReplicaSet(context.Background(), opts)
	require.NoError(t, err)
	require.NotNil(t, cont.MasterContainer)
	require.NotNil(t, cont.ReplicaSet1)
	require.NotNil(t, cont.ReplicaSet2)
	require.NotEmpty(t, cont.Network)
	require.NotEmpty(t, cont.ContainerNames)

	containers := []string{
		opts.Name + "-m1",
		opts.Name + "-rs2",
		opts.Name + "-rs3",
	}

	for _, contName := range containers {
		ok, err := tc.ContainerExists(contName)
		require.NoError(t, err)
		require.True(t, ok)
	}

	assertConnection(t, cont)
	assertTransaction(t, cont)

	cont.Terminate(context.Background())

	for _, contName := range containers {
		ok, err := tc.ContainerExists(contName)
		require.NoError(t, err)
		require.False(t, ok)
	}
}

func assertConnection(t *testing.T, c *ReplicaSetContainer) {
	m1, err := connect(c.MasterConnectionURI())
	require.NoError(t, err)
	require.NotNil(t, m1)

	rs2, err := connect(c.ReplicaSet1ConnectionURI())
	require.NoError(t, err)
	require.NotNil(t, rs2)

	rs3, err := connect(c.ReplicaSet2ConnectionURI())
	require.NoError(t, err)
	require.NotNil(t, rs3)

	res, err := m1.Client().Database("waitPrimaryNode").Collection("waitPrimaryNode").
		InsertOne(context.Background(), bson.M{"key": "val"})
	require.NoError(t, err)
	require.NotEmpty(t, res.InsertedID)
}

func assertTransaction(t *testing.T, c *ReplicaSetContainer) {
	db, err := connect(c.MasterConnectionURI())
	require.NoError(t, err)
	require.NotNil(t, db)
	ctx := context.Background()

	txnOpts := options.Transaction()

	session, err := db.Client().StartSession()
	require.NoError(t, err)
	defer session.EndSession(ctx)

	err = mongo.WithSession(ctx, session, func(sessionContext mongo.SessionContext) error {
		require.NoError(t, session.StartTransaction(txnOpts))
		res, err := db.Collection("waitPrimaryNode").InsertOne(sessionContext, bson.M{"key": "value"}, nil)
		require.NoError(t, err)
		require.NotEmpty(t, res.InsertedID)
		require.NoError(t, session.CommitTransaction(sessionContext))
		return nil
	})
	require.NoError(t, err)
}

func connect(uri string) (*mongo.Database, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	if err = client.Connect(ctx); err != nil {
		return nil, err
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}
	return client.Database("waitPrimaryNode"), nil
}
