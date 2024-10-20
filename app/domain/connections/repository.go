package connections

import (
	"context"
	"time"

	"github.com/aws/aws-xray-sdk-go/xray"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/superkruger/nostr_app_data/app/utils/skmongo"
)

const collectionName = "connections"

type connection struct {
	ID        string    `json:"id" bson:"id"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

type Repository interface {
	add(ctx context.Context, con connection) error
	remove(ctx context.Context, id string) error
	all(ctx context.Context) ([]connection, error)
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

func (r *repository) add(ctx context.Context, con connection) error {
	return xray.Capture(ctx, "DB - add connection", func(ctx1 context.Context) error {
		_, err := r.c.InsertOne(ctx1, con)
		return err
	})
}

func (r *repository) remove(ctx context.Context, id string) error {
	return xray.Capture(ctx, "DB - remove connection", func(ctx1 context.Context) error {
		_, err := r.c.DeleteOne(ctx1, bson.M{"id": id})
		return err
	})
}

func (r *repository) all(ctx context.Context) ([]connection, error) {
	var res []connection
	err := xray.Capture(ctx, "DB - all connections", func(ctx1 context.Context) error {
		cursor, err := r.c.Find(ctx1, bson.M{})
		if err != nil {
			return err
		}
		return cursor.All(ctx1, &res)
	})
	return res, err
}
