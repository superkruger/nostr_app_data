package events

import (
	"context"
	"fmt"

	"github.com/aws/jsii-runtime-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/superkruger/nostr_app_data/app/domain"
	"github.com/superkruger/nostr_app_data/app/utils/skmongo"
)

const collectionName = "events"

type Repository interface {
	add(ctx context.Context, event domain.DBEvent) error
	remove(ctx context.Context, id string) error
	findForRequest(ctx context.Context, req domain.Request) ([]domain.DBEvent, error)
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

func (r *repository) add(ctx context.Context, event domain.DBEvent) error {
	_, err := r.c.UpdateOne(ctx, bson.M{"id": event.ID}, bson.M{"$set": event}, &options.UpdateOptions{Upsert: jsii.Bool(true)})
	return err
}

func (r *repository) remove(ctx context.Context, id string) error {
	_, err := r.c.DeleteOne(ctx, bson.M{"id": id})
	return err
}

func (r *repository) findForRequest(ctx context.Context, req domain.Request) ([]domain.DBEvent, error) {
	var results []domain.DBEvent
	eventMap := make(map[string]struct{})
	for _, reqFilter := range req.Filters {
		filterResults, err := r.findForRequestFilter(ctx, reqFilter)
		if err != nil {
			return nil, err
		}
		for _, event := range filterResults {
			if _, ok := eventMap[event.ID]; !ok {
				results = append(results, event)
			}
			eventMap[event.ID] = struct{}{}
		}
	}
	return results, nil
}

func (r *repository) findForRequestFilter(ctx context.Context, reqFilter domain.Filter) ([]domain.DBEvent, error) {
	var results []domain.DBEvent
	filter := requestFilter(reqFilter)
	limit := int64(reqFilter.Limit)
	cursor, err := r.c.Find(ctx, filter, &options.FindOptions{Limit: &limit})
	if err != nil {
		return nil, err
	}
	defer func() { _ = cursor.Close(ctx) }()
	err = cursor.All(ctx, &results)
	return results, err
}

func requestFilter(reqFilter domain.Filter) bson.M {
	filter := bson.M{}
	if len(reqFilter.Ids) > 0 {
		filter["id"] = bson.M{"$in": reqFilter.Ids}
	}
	if len(reqFilter.Authors) > 0 {
		filter["pubkey"] = bson.M{"$in": reqFilter.Authors}
	}
	if len(reqFilter.Kinds) > 0 {
		filter["kind"] = bson.M{"$in": reqFilter.Kinds}
	}
	if len(reqFilter.Tags) > 0 {
		for k, v := range reqFilter.Tags {
			filter[fmt.Sprintf("tags.%s.values", k)] = bson.M{"$in": v}
		}
	}
	if reqFilter.Since > 0 {
		filter["created_at"] = bson.M{"$gte": reqFilter.Since}
	}
	if reqFilter.Until > 0 {
		filter["created_at"] = bson.M{"$lte": reqFilter.Until}
	}
	return filter
}
