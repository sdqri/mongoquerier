package mongoquerier

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type MongoAdapter struct {
	*zap.Logger
	Client   *mongo.Client
	Database string
}

func NewMongoAdapter(ctx context.Context, logger *zap.Logger, uri string, database string) (*MongoAdapter, error) {
	// Setting package specific fields for log entry
	logger = logger.With(zap.String("package", "adapters.MongoAdapter"))

	clientOptions := options.Client().ApplyURI(uri)

	// Connect to the MongoDB server
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		logger.Error("unable to connect to mongo", zap.Error(err))
		return nil, err
	}

	// Ping the MongoDB server to verify that the connection is working
	err = client.Ping(ctx, nil)
	if err != nil {
		logger.Error("unable to ping mongo", zap.Error(err))
		return nil, err
	}

	logger.Debug("successfully connected to MongoDB!")

	return &MongoAdapter{
		Logger:   logger,
		Client:   client,
		Database: database,
	}, nil
}

func (madp *MongoAdapter) GetDatabase() *mongo.Database {
	return madp.Client.Database(madp.Database)
}

func (madp *MongoAdapter) GetCollection(collection string, opts ...*options.CollectionOptions) *mongo.Collection {
	return madp.GetDatabase().Collection(collection, opts...)
}

func (madp *MongoAdapter) Disconnect(ctx context.Context) error {
	return madp.Client.Disconnect(ctx)
}
