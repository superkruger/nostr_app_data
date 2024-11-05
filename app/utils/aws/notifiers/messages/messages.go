package messages

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"

	"github.com/superkruger/nostr_app_data/app/utils/compress"
)

// Message represents a message we want to notify, using a notifier
type Message struct {
	body            string
	attributes      map[string]string
	groupID         string
	deduplicationID string
	err             error
}

// WithAttributes returns a new Message with the attributes
func (m Message) WithAttributes(attributes map[string]string) Message {
	m.attributes = attributes
	return m
}

// WithAttribute returns a new Message with an additional attribute
func (m Message) WithAttribute(key string, value string) Message {
	if m.attributes == nil {
		m.attributes = make(map[string]string)
	}
	m.attributes[key] = value
	return m
}

// WithFifoID returns a new Message with the fifo ids
func (m Message) WithFifoID(groupID, deduplicationID string) Message {
	m.groupID = groupID
	m.deduplicationID = deduplicationID
	return m
}

// ToSQSInput converts the message to SQS Input
func (m Message) ToSQSInput(queueURL string) (*sqs.SendMessageInput, error) {
	if m.err != nil {
		return nil, m.err
	}
	i := sqs.SendMessageInput{
		QueueUrl:    &queueURL,
		MessageBody: &(m.body),
	}
	if len(m.attributes) > 0 {
		i.MessageAttributes = attributesToSQS(m.attributes)
	}
	if m.groupID != "" || m.deduplicationID != "" {
		i.MessageGroupId = &m.groupID
		i.MessageDeduplicationId = &m.deduplicationID
	}
	return &i, nil
}

// ToSNSInput converts the message to SNS Input
func (m Message) ToSNSInput(topicArn string) (*sns.PublishInput, error) {
	if m.err != nil {
		return nil, m.err
	}
	i := sns.PublishInput{
		TopicArn: &topicArn,
		Message:  &(m.body),
	}
	if len(m.attributes) > 0 {
		i.MessageAttributes = attributesToSNS(m.attributes)
	}
	if m.groupID != "" || m.deduplicationID != "" {
		i.MessageGroupId = &m.groupID
		i.MessageDeduplicationId = &m.deduplicationID
	}
	return &i, nil
}

// New creates a new message based on the text you want to notify
//
// If you are sending JSON, use NewForJSON(obj) instead.
func New(message string) Message {
	return Message{
		body: message,
	}
}

// NewForJSON creates a new message based on the json of the object
func NewForJSON(obj interface{}) Message {
	message, err := json.Marshal(obj)
	if err != nil {
		return Message{
			err: err,
		}
	}
	return Message{
		body: string(message),
	}
}

func NewCompressedForJSON(obj interface{}) Message {
	message, err := json.Marshal(obj)
	if err != nil {
		return Message{err: err}
	}
	compressed, err := compress.EncodeBytes(message)
	if err != nil {
		return Message{err: err}
	}
	return Message{
		body: string(compressed),
	}

}

func attributesToSQS(attributes map[string]string) map[string]*sqs.MessageAttributeValue {
	attrs := make(map[string]*sqs.MessageAttributeValue, len(attributes))
	for key, value := range attributes {
		var attrValue sqs.MessageAttributeValue
		attrValue.SetDataType("String").SetStringValue(value)
		attrs[key] = &attrValue
	}
	return attrs
}

func attributesToSNS(attributes map[string]string) map[string]*sns.MessageAttributeValue {
	attrs := make(map[string]*sns.MessageAttributeValue, len(attributes))
	for key, value := range attributes {
		var attrValue sns.MessageAttributeValue
		attrValue.SetDataType("String").SetStringValue(value)
		attrs[key] = &attrValue
	}
	return attrs
}
