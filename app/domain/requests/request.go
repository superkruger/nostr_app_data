package requests

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Request struct {
	ID      string   `bson:"id"`
	ConnID  string   `bson:"connId"`
	Filters []Filter `bson:"filters"`
}

type Filter struct {
	Ids     []string            `bson:"ids"`
	Authors []string            `bson:"authors"`
	Kinds   []int               `bson:"kinds"`
	Tags    map[string][]string `bson:"tags"`
	Since   int                 `bson:"since"`
	Until   int                 `bson:"until"`
	Limit   int                 `bson:"limit"`
}

func (f *Filter) UnmarshalJSON(data []byte) error {
	var jsonValue map[string]interface{}
	err := json.Unmarshal(data, &jsonValue)
	if err != nil {
		return err
	}
	for k, v := range jsonValue {
		switch k {
		case "ids":
			for _, v := range v.([]interface{}) {
				f.Ids = append(f.Ids, v.(string))
			}
		case "authors":
			for _, v := range v.([]interface{}) {
				f.Authors = append(f.Authors, v.(string))
			}
		case "kinds":
			for _, val := range v.([]interface{}) {
				f.Kinds = append(f.Kinds, val.(int))
			}
		case "since":
			f.Since = v.(int)
		case "until":
			f.Until = v.(int)
		case "limit":
			f.Limit = v.(int)
		default:
			if !strings.HasPrefix(k, "#") {
				return fmt.Errorf("unknown filter type: %s", k)
			}
			if f.Tags == nil {
				f.Tags = make(map[string][]string)
			}
			tagName := strings.TrimPrefix(k, "#")
			for _, val := range v.([]interface{}) {
				f.Tags[tagName] = append(f.Tags[tagName], val.(string))
			}
		}
	}
	return nil
}
