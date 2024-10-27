package events

import "github.com/superkruger/nostr_app_data/app/utils"

type Tags []Tag
type Tag []string

type DBTags map[string]DBTag
type DBTag struct {
	Values []string   `bson:"values"`
	Other  [][]string `bson:"other"`
}

// Marshal Tag. Used for Serialization so string escaping should be as in RFC8259.
func (tag Tag) marshalTo(dst []byte) []byte {
	dst = append(dst, '[')
	for i, s := range tag {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst = utils.EscapeString(dst, s)
	}
	dst = append(dst, ']')
	return dst
}

// MarshalTo appends the JSON encoded byte of Tags as [][]string to dst.
// String escaping is as described in RFC8259.
func (tags Tags) marshalTo(dst []byte) []byte {
	dst = append(dst, '[')
	for i, tag := range tags {
		if i > 0 {
			dst = append(dst, ',')
		}
		dst = tag.marshalTo(dst)
	}
	dst = append(dst, ']')
	return dst
}

func (tags Tags) toDB() DBTags {
	result := make(DBTags)
	for _, tag := range tags {
		dbTag, _ := result[tag[0]]
		dbTag.Values = append(dbTag.Values, tag[1])
		dbTag.Other = append(dbTag.Other, tag[2:])
		result[tag[0]] = dbTag
	}
	return result
}

func (dbTags DBTags) toJson() Tags {
	result := make(Tags, 0, len(dbTags))
	for k, dbTag := range dbTags {
		for i, val := range dbTag.Values {
			tag := make(Tag, 0, len(dbTag.Other[i])+2)
			tag = append(tag, k, val)
			tag = append(tag, dbTag.Other[i]...)
			result = append(result, tag)
		}
	}
	return result
}
