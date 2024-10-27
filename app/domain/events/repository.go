package events

import (
	"context"

	"github.com/aws/aws-xray-sdk-go/xray"
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
	return xray.Capture(ctx, "DB - add event", func(ctx1 context.Context) error {
		_, err := r.c.UpdateOne(ctx1, bson.M{"id": event.ID}, bson.M{"$set": event}, &options.UpdateOptions{Upsert: jsii.Bool(true)})
		return err
	})
}

func (r *repository) remove(ctx context.Context, id string) error {
	return xray.Capture(ctx, "DB - remove event", func(ctx1 context.Context) error {
		_, err := r.c.DeleteOne(ctx1, bson.M{"id": id})
		return err
	})
}
