package skmongo

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Mongo provides access to the mongo database.
// After creating it with New(...), you can perform action on collections
// by calling Collection(name). That returns a regular mongo.Collection.
type Mongo struct {
	database *mongo.Database
}

// New creates a new mongo that connects to the mongo DB as described in the secret.
func New(secretName string) (Mongo, error) {
	return NewWithContext(context.Background(), secretName)
}

// NewWithOptions creates a new mongo that connects to the mongo DB as described in the secret.
func NewWithOptions(secretName string, clientOptions *options.ClientOptions) (Mongo, error) {
	return NewWithContextWithOptions(context.Background(), secretName, clientOptions)
}

// NewWithContext creates a new mongo that connects to the mongo DB as described
// in the secret. It uses the given context.
func NewWithContext(ctx context.Context, secretName string) (Mongo, error) {
	return NewWithContextWithOptions(ctx, secretName, nil)
}

// NewWithContextWithOptions connects to the database behind the secret name
// with the default configuration for Lambda functions or panics when
// an error occurred.
//
// Typically, used in the init of a Lambda function like this:
//
//	repo := NewRepository(dpgmongo.NewWithContextWithOptions(ctx, dbSecret, options.Client().SetMaxPoolSize(200))
func NewWithContextWithOptions(ctx context.Context, secretName string, clientOptions *options.ClientOptions) (Mongo, error) {
	cfg, err := ConfigFromSecret(secretName)
	if err != nil {
		return Mongo{}, fmt.Errorf("problem getting configuration from secret: %w", err)
	}

	db, err := databaseFor(ctx, cfg.DatabaseName(), cfg.WithClientOptions(clientOptions).toClientOptions())
	if err != nil {
		return Mongo{}, fmt.Errorf("problem connecting to database: %w", err)
	}
	return Mongo{db}, nil
}

// Client returns the client which can be used to create transactions.
func (m Mongo) Client() *mongo.Client {
	return m.database.Client()
}

// Collection returns the collection with the given name, to perform operations on.
func (m Mongo) Collection(name string) *mongo.Collection {
	return m.database.Collection(name)
}

// CreateCollection creates a new collection. This allows for the creation of a
// collection in a multi-document transaction while testing.
func (m Mongo) CreateCollection(ctx context.Context, name string, opts ...*options.CreateCollectionOptions) error {
	return m.database.CreateCollection(ctx, name, opts...)
}

// Must is a helper function to ensure the mongo is valid and there was no
// error when calling a New function. It will panic on error.
//
// This helper is intended to be used in variable initialization to load the
// mongo and configuration at startup. Such as:
//
//	var db = dpgmongo.Must(dpgmongo.New(secretName))
func Must(mongo Mongo, err error) Mongo {
	if err != nil {
		panic(err)
	}
	return mongo
}

// MustWithClose is a helper function to ensure the mongo is valid and there was no
// error when calling a New function. It will panic on error.
//
// This helper is intended to be used in variable initialization to load the
// mongo and configuration at startup. Such as:
//
//	db, closeDb := dpgmongo.Must(dpgmongo.New(secretName))
func MustWithClose(mongo Mongo, err error) (Mongo, func()) {
	if err != nil {
		panic(err)
	}
	return mongo, func() {
		logrus.Trace("closing mongo connection")
		err := mongo.database.Client().Disconnect(nil)
		if err != nil {
			logrus.WithField("db", mongo.database).Errorf("error closing mongo connection: %v", err)
		}
	}
}

// MustFromSecret connects to the database behind the secret name
// with the default configuration for Lambda functions or panics when
// an error occurred.
//
// Typically, used in the init of a Lambda function like this:
//
//	repo := NewRepository(dpgmongo.MustFromSecret(ctx, dbSecret))
func MustFromSecret(secretName string) Mongo {
	return MustFromSecretWithOptions(secretName, nil)
}

// MustFromSecretWithClose connects to the database behind the secret name
// with the default configuration for Lambda functions or panics when
// an error occurred.
//
// Typically, used in the init of a Lambda function like this:
//
//	db, closeDb := dpgmongo.MustFromSecretWithClose(ctx, dbSecret)
//	return &handler{
//		recurringOrderService: recurringorders.NewService(NewRepository(db)),
//		shutdown:              func() { closeDb() },
//	}
func MustFromSecretWithClose(secretName string) (Mongo, func()) {
	return MustFromSecretWithOptionsAndClose(secretName, nil)
}

// MustFromSecretReplica connects Lambda function to the database behind the secret name
// with the preference of replica for reads
// or panics when an error occurred.
//
// Typically, used in the init of a Lambda function like this:
//
//	repo := NewRepository(dpgmongo.MustFromSecretReplica(ctx, dbSecret))
func MustFromSecretReplica(secretName string) Mongo {
	return MustFromSecretWithOptions(secretName, options.Client().SetReadPreference(readpref.SecondaryPreferred()))
}

// MustFromSecretReplicaWithClose connects Lambda function to the database behind the secret name
// with the preference of replica for reads
// or panics when an error occurred.
//
// Typically, used in the init of a Lambda function like this:
//
//	db, closeDb := dpgmongo.MustFromSecretReplicaWithClose(ctx, dbSecret)
//	return &handler{
//		recurringOrderService: recurringorders.NewService(NewRepository(db)),
//		shutdown:              func() { closeDb() },
//	}
func MustFromSecretReplicaWithClose(secretName string) (Mongo, func()) {
	return MustFromSecretWithOptionsAndClose(secretName, options.Client().SetReadPreference(readpref.SecondaryPreferred()))
}

// MustFromSecretWithOptions connects to the database behind the secret name
// with the default configuration for Lambda functions or panics when
// an error occurred.
//
// Typically, used in the init of a Lambda function like this:
//
//	repo := NewRepository(dpgmongo.MustFromSecretWithOptions(ctx, dbSecret, options.Client().SetMaxPoolSize(200)))
func MustFromSecretWithOptions(secretName string, clientOptions *options.ClientOptions) Mongo {
	db, err := NewWithContextWithOptions(context.Background(), secretName, clientOptions)
	return Must(db, err)
}

// MustFromSecretWithOptionsAndClose connects to the database behind the secret name
// with the default configuration for Lambda functions or panics when
// an error occurred.
//
// Typically, used in the init of a Lambda function like this:
//
//	db, closeDb := dpgmongo.MustFromSecretWithOptionsAndClose(ctx, dbSecret, options.Client().SetMaxPoolSize(200))
//	return &handler{
//		recurringOrderService: recurringorders.NewService(NewRepository(db)),
//		shutdown:              func() { closeDb() },
//	}
func MustFromSecretWithOptionsAndClose(secretName string, clientOptions *options.ClientOptions) (Mongo, func()) {
	db, err := NewWithContextWithOptions(context.Background(), secretName, clientOptions)
	return MustWithClose(db, err)
}
