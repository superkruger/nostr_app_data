package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

var indexedTags = map[string]struct{}{"e": {}, "p": {}, "a": {}, "d": {}}

type Request struct {
	ID       string    `json:"id" bson:"id"`
	ConnID   string    `json:"connId" bson:"connId"`
	Filters  []Filter  `json:"-" bson:"filters"`
	ExpireAt time.Time `json:"-" bson:"expire_at,omitempty"`
}

type Filter struct {
	Ids     []string            `json:"ids" bson:"ids,omitempty"`
	Authors []string            `json:"authors" bson:"authors,omitempty"`
	Kinds   []int               `json:"kinds" bson:"kinds,omitempty"`
	Tags    map[string][]string `json:"-" bson:"tags,omitempty"`
	Since   int                 `json:"since" bson:"since,omitempty"`
	Until   int                 `json:"until" bson:"until,omitempty"`
	Limit   int                 `json:"limit" bson:"limit,omitempty"`
}

func (r *Request) Unmarshal(body, connectionID string) error {
	var raw []json.RawMessage
	if err := json.Unmarshal([]byte(body), &raw); err != nil {
		return fmt.Errorf("failed to unmarshal request body: %w", err)
	}
	if len(raw) < 3 {
		return fmt.Errorf("expected a length of at least 3")
	}
	if err := json.Unmarshal(raw[1], &r.ID); err != nil {
	}
	r.ConnID = connectionID
	r.Filters = make([]Filter, len(raw[2:]))
	for i, rawFilter := range raw[2:] {
		if err := json.Unmarshal(rawFilter, &r.Filters[i]); err != nil {
			return fmt.Errorf("failed to unmarshal filter %d: %w", i, err)
		}
		if err := r.Filters[i].unmarshalTags(rawFilter); err != nil {
			return fmt.Errorf("failed to unmarshal filter tags %d: %w", i, err)
		}
		if r.Filters[i].isZero() {
			return fmt.Errorf("filter at %d is zero", i)
		}
	}
	return nil
}

func (r *Request) UnIndexedTags() []string {
	unIndexed := make(map[string]struct{})
	for _, f := range r.Filters {
		f.unIndexedTags(unIndexed)
	}
	result := make([]string, 0, len(unIndexed))
	for k := range unIndexed {
		result = append(result, k)
	}
	return result
}

func (f *Filter) unmarshalTags(data []byte) error {
	var jsonValue map[string]interface{}
	err := json.Unmarshal(data, &jsonValue)
	if err != nil {
		return err
	}
	for k, v := range jsonValue {
		if !strings.HasPrefix(k, "#") {
			continue
		}
		if f.Tags == nil {
			f.Tags = make(map[string][]string)
		}
		tagName := strings.TrimPrefix(k, "#")
		for _, val := range v.([]interface{}) {
			f.Tags[tagName] = append(f.Tags[tagName], val.(string))
		}
	}
	return nil
}

func (f *Filter) isZero() bool {
	return len(f.Ids) == 0 && len(f.Authors) == 0 && len(f.Kinds) == 0 && len(f.Tags) == 0 && f.Limit == 0
}

func (f *Filter) unIndexedTags(unIndexed map[string]struct{}) {
	for tag := range f.Tags {
		if _, ok := indexedTags[tag]; !ok {
			unIndexed[tag] = struct{}{}
		}
	}
}
