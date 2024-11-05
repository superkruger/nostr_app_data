package requests

import (
	"context"
	"sort"
	"testing"

	"github.com/pascaldekloe/goe/verify"
	"github.com/superkruger/nostr_app_data/app/domain/events"
	"github.com/superkruger/nostr_app_data/app/utils/skmongo"
	"go.mongodb.org/mongo-driver/mongo"
)

const mongoSecret = "test/nostr/mongo/rw"

func TestFindForEvent(t *testing.T) {
	ctx := context.Background()
	testDB, testCollection, cleanupCallback := skmongo.DatabaseForTest(ctx, t, collectionName, mongoSecret)
	t.Cleanup(cleanupCallback)
	repo := &repository{
		c: testDB.Collection(testCollection),
	}

	tests := map[string]struct {
		given   []Request
		event   events.Event
		want    []Request
		wantErr bool
	}{
		"simple match": {
			given: []Request{
				{
					ID:     "r1",
					ConnID: "c1",
					Filters: []Filter{
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
			},
			event: events.Event{
				ID:        "id1",
				PubKey:    "auth1",
				Kind:      1,
				Tags:      events.Tags{{"t1", "t1val1"}}, //map[string]events.DBTag{"t1": {Values: []string{"t1val1"}}},
				CreatedAt: 20,
			},
			want: []Request{{ID: "r1", ConnID: "c1"}},
		},
		"no match over two filters": {
			given: []Request{
				{
					ID:     "r1",
					ConnID: "c1",
					Filters: []Filter{
						{
							Ids:     []string{"id1"},
							Authors: []string{"auth1"},
							Kinds:   []int{1},
							Tags:    map[string][]string{"t1": {"t1val1", "t1val2"}},
							Since:   10,
							Until:   100,
						},
						{
							Ids:     []string{"id2"},
							Authors: []string{"auth2"},
							Kinds:   []int{2},
							Tags:    map[string][]string{"t2": {"t2val1", "t2val2"}},
							Since:   20,
							Until:   200,
						},
					},
				},
			},
			event: events.Event{
				ID:        "id2",
				PubKey:    "auth1",
				Kind:      2,
				Tags:      events.Tags{{"t1", "t1val1"}}, //mmap[string]events.DBTag{"t1": {Values: []string{"t1val1"}}},
				CreatedAt: 150,
			},
			want: []Request{},
		},
		"match one request": {
			given: []Request{
				{
					ID:     "r1",
					ConnID: "c1",
					Filters: []Filter{
						{
							Ids:     []string{"id1"},
							Authors: []string{"auth1"},
							Kinds:   []int{1},
							Tags:    map[string][]string{"t1": {"t1val1", "t1val2"}},
							Since:   10,
							Until:   100,
						},
						{
							Ids:     []string{"id2"},
							Authors: []string{"auth2"},
							Kinds:   []int{2},
							Tags:    map[string][]string{"t2": {"t2val1", "t2val2"}},
							Since:   20,
							Until:   100,
						},
					},
				},
				{
					ID:     "r2",
					ConnID: "c2",
					Filters: []Filter{
						{
							Ids:     []string{"id3"},
							Authors: []string{"auth3"},
							Kinds:   []int{3},
							Tags:    map[string][]string{"t3": {"t3val1", "t3val2"}},
							Since:   30,
							Until:   100,
						},
						{
							Ids:     []string{"id4"},
							Authors: []string{"auth4"},
							Kinds:   []int{4},
							Tags:    map[string][]string{"t4": {"t4val1", "t4val2"}},
							Since:   40,
							Until:   100,
						},
					},
				},
			},
			event: events.Event{
				ID:        "id4",
				PubKey:    "auth4",
				Kind:      4,
				Tags:      events.Tags{{"t4", "t4val1"}}, //m map[string]events.DBTag{"t4": {Values: []string{"t4val1"}}},
				CreatedAt: 50,
			},
			want: []Request{{ID: "r2", ConnID: "c2"}},
		},
		"multiple tags": {
			given: []Request{
				{
					ID:     "r1",
					ConnID: "c1",
					Filters: []Filter{
						{
							Tags: map[string][]string{"t": {"tval1"}},
						},
						{
							Tags: map[string][]string{"t": {"tval2"}},
						},
					},
				},
				{
					ID:     "r2",
					ConnID: "c2",
					Filters: []Filter{
						{
							Tags: map[string][]string{"t": {"tval3"}},
						},
						{
							Tags: map[string][]string{"t": {"tval4"}},
						},
					},
				},
			},
			event: events.Event{
				ID:        "id4",
				PubKey:    "auth4",
				Kind:      4,
				Tags:      []events.Tag{{"t", "tval1"}, {"t", "tval4"}}, //map[string]events.DBTag{"t": {Values: []string{"tval1", "tval4"}}},
				CreatedAt: 50,
			},
			want: []Request{{ID: "r1", ConnID: "c1"}, {ID: "r2", ConnID: "c2"}},
		},
		"single tag": {
			given: []Request{
				{
					ID:     "r1",
					ConnID: "c1",
					Filters: []Filter{
						{
							Tags: map[string][]string{"t": {"tval1"}},
						},
						{
							Tags: map[string][]string{"t": {"tval2"}},
						},
					},
				},
				{
					ID:     "r2",
					ConnID: "c2",
					Filters: []Filter{
						{
							Tags: map[string][]string{"t": {"tval3"}},
						},
						{
							Tags: map[string][]string{"t": {"tval4"}},
						},
					},
				},
			},
			event: events.Event{
				ID:        "id4",
				PubKey:    "auth4",
				Kind:      4,
				Tags:      []events.Tag{{"t", "tval4"}}, //map[string]events.DBTag{"t": {Values: []string{"tval5"}}},
				CreatedAt: 50,
			},
			want: []Request{{ID: "r2", ConnID: "c2"}},
		},
		"open filters": {
			given: []Request{
				{
					ID:     "r1",
					ConnID: "c1",
					Filters: []Filter{
						{
							Since: 10,
						},
						{
							Since: 20,
						},
					},
				},
				{
					ID:     "r2",
					ConnID: "c2",
					Filters: []Filter{
						{
							Since: 30,
						},
						{
							Since: 40,
						},
					},
				},
			},
			event: events.Event{
				ID:        "id4",
				PubKey:    "auth4",
				Kind:      4,
				Tags:      events.Tags{{"t4", "t4val1"}}, //mmap[string]events.DBTag{"t4": {Values: []string{"t4val1"}}},
				CreatedAt: 10,
			},
			want: []Request{{ID: "r1", ConnID: "c1"}},
		},
	}
	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			skmongo.TransactionTest(ctx, t, testDB, func(sessCtx mongo.SessionContext) {
				for _, request := range testCase.given {
					if err := repo.add(sessCtx, request); err != nil {
						t.Fatal("problem inserting request", err)
					}
				}
				got, err := repo.findForEvent(sessCtx, testCase.event)
				verify.Values(t, name, err != nil, testCase.wantErr)
				sort.Slice(got, func(i, j int) bool {
					return got[i].ID < got[j].ID
				})
				verify.Values(t, name, got, testCase.want)
			})
		})
	}
}
