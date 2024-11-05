package events

import (
	"context"

	"github.com/aws/jsii-runtime-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/superkruger/nostr_app_data/app/utils/skmongo"
)

const collectionName = "events"

type Repository interface {
	add(ctx context.Context, event DBEvent) error
	remove(ctx context.Context, id string) error
}

type repository struct {
	c *mongo.Collection
}

func MustNewRepository(secret string) Repository {
	return &repository{
		c: skmongo.MustFromSecret(secret).Collection(collectionName),
	}
}

func NewRepository(db skmongo.Mongo) Repository {
	return &repository{
		c: db.Collection(collectionName),
	}
}

func (r *repository) add(ctx context.Context, event DBEvent) error {
	_, err := r.c.UpdateOne(ctx, bson.M{"id": event.ID}, bson.M{"$set": event}, &options.UpdateOptions{Upsert: jsii.Bool(true)})
	return err
}

func (r *repository) remove(ctx context.Context, id string) error {
	_, err := r.c.DeleteOne(ctx, bson.M{"id": id})
	return err
}
