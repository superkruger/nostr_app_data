package events

import "github.com/superkruger/nostr_app_data/app/utils"

type Tags []Tag
type Tag []string

type DBTags map[string]DBTag
type DBTag struct {
	Value string   `bson:"value"`
	Other []string `bson:"other"`
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
		result[tag[0]] = DBTag{
			Value: tag[1],
			Other: tag[2:],
		}
	}
	return result
}

func (dbTags DBTags) toJson() Tags {
	result := make(Tags, 0, len(dbTags))
	for k, dbTag := range dbTags {
		tag := make(Tag, 0, len(dbTag.Other)+2)
		tag = append(tag, k, dbTag.Value)
		tag = append(tag, dbTag.Other...)
		result = append(result, tag)
	}
	return result
}
