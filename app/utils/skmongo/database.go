package skmongo

import (
	"context"
	"strings"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type database struct {
	name    string
	client  *mongo.Client
	options *options.ClientOptions
}

func databaseFor(ctx context.Context, name string, clientOptions *options.ClientOptions) (*mongo.Database, error) {
	db := database{name: name, options: clientOptions}

	if err := db.verifyConnection(ctx); err != nil {
		log.WithField("db", db).WithError(err).Warning("problem occurs in databaseFor()")
		return nil, err
	}

	return db.database(), nil
}

func (db *database) database() *mongo.Database {
	return db.client.Database(db.name)
}

// verifyConnection verifies that we can connect to the database
func (db *database) verifyConnection(ctx context.Context) error {
	if err := db.initClient(ctx); err != nil {
		log.WithError(err).Warning("problem occurs in initClient()")
		return err
	}
	if err := db.client.Ping(ctx, readpref.PrimaryPreferred()); err != nil {
		log.WithError(err).Warning("problem occurs in client.Ping() in ping()")
		return err
	}
	return nil
}

// initClient initializes the client, based on the connection string from the config
func (db *database) initClient(ctx context.Context) error {
	client, err := mongo.Connect(ctx, db.options)
	if err != nil {
		return err
	}
	db.client = client
	return nil
}

// IsConnectionErr will validate if the given error is an error that occurred while connecting to MongoDB
func IsConnectionErr(err error) bool {
	return mongo.IsNetworkError(err) ||
		mongo.IsTimeout(err) ||
		strings.Contains(err.Error(), "error occured during connection") || // MongoDB driver misspells occurred
		strings.Contains(err.Error(), "error occurred during connection")
}
