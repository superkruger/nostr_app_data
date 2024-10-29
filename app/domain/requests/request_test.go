package requests

import (
	"testing"

	"github.com/pascaldekloe/goe/verify"
)

func TestUnmarshal(t *testing.T) {
	tests := map[string]struct {
		given        string
		connectionID string
		want         Request
		wantErr      bool
	}{
		"empty filter": {
			given:        "[\"REQ\", \"subId1\", {}]",
			connectionID: "connId1",
			want: Request{
				ID:      "subId1",
				ConnID:  "connId1",
				Filters: []Filter{{}},
			},
			wantErr: true,
		},
		"filter with tags": {
			given:        "[\"REQ\", \"subId1\", {\"#e\":[\"e1\", \"e2\"]}]",
			connectionID: "connId1",
			want: Request{
				ID:     "subId1",
				ConnID: "connId1",
				Filters: []Filter{{
					Tags: map[string][]string{"e": {"e1", "e2"}},
				}},
			},
			wantErr: false,
		},
		"normal filter": {
			given:        "[\"REQ\", \"subId1\", {\"#e\":[\"e1\", \"e2\"], \"ids\":[\"id1\", \"id2\"], \"authors\":[\"author1\", \"author2\"], \"kinds\":[1,2,3], \"since\":100, \"until\":200, \"limit\":300}]",
			connectionID: "connId1",
			want: Request{
				ID:     "subId1",
				ConnID: "connId1",
				Filters: []Filter{{
					Ids:     []string{"id1", "id2"},
					Authors: []string{"author1", "author2"},
					Kinds:   []int{1, 2, 3},
					Tags:    map[string][]string{"e": {"e1", "e2"}},
					Since:   100,
					Until:   200,
					Limit:   300,
				}},
			},
			wantErr: false,
		},
		"multiple filters": {
			given: "[\"REQ\", \"subId1\", " +
				"{\"#e\":[\"e1\", \"e2\"], \"ids\":[\"id1\", \"id2\"], \"authors\":[\"author1\", \"author2\"], \"kinds\":[1,2,3], \"since\":100, \"until\":200, \"limit\":300}," +
				"{\"#p\":[\"p1\", \"p2\"], \"#foo\":[\"foo1\", \"foo2\"], \"ids\":[\"id3\", \"id4\"], \"authors\":[\"author3\", \"author4\"], \"kinds\":[4,5,6], \"since\":400, \"until\":500, \"limit\":600}" +
				"]",
			connectionID: "connId1",
			want: Request{
				ID:     "subId1",
				ConnID: "connId1",
				Filters: []Filter{
					{
						Ids:     []string{"id1", "id2"},
						Authors: []string{"author1", "author2"},
						Kinds:   []int{1, 2, 3},
						Tags:    map[string][]string{"e": {"e1", "e2"}},
						Since:   100,
						Until:   200,
						Limit:   300,
					},
					{
						Ids:     []string{"id3", "id4"},
						Authors: []string{"author3", "author4"},
						Kinds:   []int{4, 5, 6},
						Tags:    map[string][]string{"p": {"p1", "p2"}, "foo": {"foo1", "foo2"}},
						Since:   400,
						Until:   500,
						Limit:   600,
					},
				},
			},
			wantErr: false,
		},
		"filter with unknown field": {
			given:        "[\"REQ\", \"subId1\", {\"dummy\":[\"e1\", \"e2\"], \"limit\":1}]",
			connectionID: "connId1",
			want: Request{
				ID:      "subId1",
				ConnID:  "connId1",
				Filters: []Filter{{Limit: 1}},
			},
			wantErr: false,
		},
		"request with wrong format": {
			given:        "[\"REQ\", {\"dummy\":[\"e1\", \"e2\"], \"limit\":1}]",
			connectionID: "connId1",
			wantErr:      true,
		},
	}
	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			var r Request
			err := r.Unmarshal(testCase.given, testCase.connectionID)
			verify.Values(t, name, err != nil, testCase.wantErr)
			verify.Values(t, name, r, testCase.want)
		})
	}
}
