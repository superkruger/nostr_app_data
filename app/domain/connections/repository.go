package connections

import (
	"context"
	"time"

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
	_, err := r.c.InsertOne(ctx, con)
	return err
}

func (r *repository) remove(ctx context.Context, id string) error {
	_, err := r.c.DeleteOne(ctx, bson.M{"id": id})
	return err
}

func (r *repository) all(ctx context.Context) ([]connection, error) {
	var res []connection
	cursor, err := r.c.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = cursor.Close(ctx) }()
	err = cursor.All(ctx, &res)
	return res, err
}
