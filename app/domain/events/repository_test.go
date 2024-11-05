package events

import (
	"context"
	"sort"
	"testing"

	"github.com/pascaldekloe/goe/verify"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/superkruger/nostr_app_data/app/domain"
	"github.com/superkruger/nostr_app_data/app/utils/skmongo"
)

const mongoSecret = "test/nostr/mongo/rw"

func TestFindForRequest(t *testing.T) {
	ctx := context.Background()
	testDB, testCollection, cleanupCallback := skmongo.DatabaseForTest(ctx, t, collectionName, mongoSecret)
	t.Cleanup(cleanupCallback)
	repo := &repository{
		c: testDB.Collection(testCollection),
	}

	tests := map[string]struct {
		given   []domain.DBEvent
		request domain.Request
		want    []domain.DBEvent
		wantErr bool
	}{
		"simple match": {
			given: []domain.DBEvent{
				{
					ID:        "id1",
					PubKey:    "auth1",
					Kind:      1,
					Tags:      map[string]domain.DBTag{"t1": {Values: []string{"t1val1"}}},
					CreatedAt: 20,
				},
			},
			request: domain.Request{
				Filters: []domain.Filter{
					{
						Ids:     []string{"id1"},
						Authors: []string{"auth1"},
						Kinds:   []int{1},
						Tags:    map[string][]string{"t1": {"t1val1", "t1val2"}},
						Since:   10,
						Until:   100,
					},
				},
			},
			want: []domain.DBEvent{
				{
					ID:        "id1",
					PubKey:    "auth1",
					Kind:      1,
					Tags:      map[string]domain.DBTag{"t1": {Values: []string{"t1val1"}}},
					CreatedAt: 20,
				},
			},
		},
	}
	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			skmongo.TransactionTest(ctx, t, testDB, func(sessCtx mongo.SessionContext) {
				for _, event := range testCase.given {
					if err := repo.add(sessCtx, event); err != nil {
						t.Fatal("problem inserting event", err)
					}
				}
				got, err := repo.findForRequest(sessCtx, testCase.request)
				verify.Values(t, name, err != nil, testCase.wantErr)
				sort.Slice(got, func(i, j int) bool {
					return got[i].ID < got[j].ID
				})
				verify.Values(t, name, got, testCase.want)
			})
		})
	}
}
