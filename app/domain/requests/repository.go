package requests

import (
	"context"
	"fmt"

	"github.com/aws/jsii-runtime-go"
	"github.com/superkruger/nostr_app_data/app/domain/events"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/superkruger/nostr_app_data/app/utils/skmongo"
)

const collectionName = "requests"

type Repository interface {
	add(ctx context.Context, request Request) error
	remove(ctx context.Context, subID string) error
	findForEvent(ctx context.Context, event events.Event) ([]Request, error)
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

func (r *repository) add(ctx context.Context, request Request) error {
	_, err := r.c.UpdateOne(ctx, bson.M{"id": request.ID}, bson.M{"$set": request}, &options.UpdateOptions{Upsert: jsii.Bool(true)})
	return err
}

func (r *repository) remove(ctx context.Context, id string) error {
	_, err := r.c.DeleteOne(ctx, bson.M{"id": id})
	return err
}

func (r *repository) findForEvent(ctx context.Context, event events.Event) ([]Request, error) {
	var results []Request
	filter := bson.M{
		"filters": bson.M{
			"$elemMatch": bson.M{
				"$and": []bson.M{
					{"$or": []bson.M{
						{"ids": bson.M{"$exists": false}},
						{"ids": event.ID},
					}},
					{"$or": []bson.M{
						{"authors": bson.M{"$exists": false}},
						{"authors": event.PubKey},
					}},
					{"$or": []bson.M{
						{"kinds": bson.M{"$exists": false}},
						{"kinds": event.Kind},
					}},
					{"$or": tagsFilter(event)},
					{"$or": []bson.M{
						{"since": bson.M{"$exists": false}},
						{"since": bson.M{"$lte": event.CreatedAt}},
					}},
					{"$or": []bson.M{
						{"until": bson.M{"$exists": false}},
						{"until": bson.M{"$gte": event.CreatedAt}},
					}},
				},
			},
		},
	}
	cursor, err := r.c.Find(ctx, filter, &options.FindOptions{Projection: bson.M{"id": 1, "connId": 1}})
	if err != nil {
		return nil, err
	}
	err = cursor.All(ctx, &results)
	return results, err
}

func tagsFilter(event events.Event) []bson.M {
	filter := []bson.M{
		{"tags": bson.M{"$exists": false}},
	}
	for _, tag := range event.Tags {
		filter = append(filter, bson.M{"tags." + tag[0]: tag[1]})
	}
	fmt.Printf("- filter: %+v\n", filter)
	return filter
}
