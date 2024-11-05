package messages

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/go-test/deep"

	"github.com/superkruger/nostr_app_data/app/utils/compress"
)

var (
	topicArn = "topicArn"
	queueURL = "queueURL"
)

func TestSNSSimpleMessage(t *testing.T) {
	body := "simple"
	m := New(body)

	got, err := m.ToSNSInput(topicArn)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	expected := &sns.PublishInput{
		TopicArn: &topicArn,
		Message:  &body,
	}

	if diff := deep.Equal(expected, got); diff != nil {
		t.Logf("expected: %v", expected)
		t.Logf("got: %v", got)
		t.Error(diff)
	}

}

func TestSNSMessageWithAttributes(t *testing.T) {
	body := "body"
	m := New(body).WithAttributes(map[string]string{"attr": "value"})

	got, err := m.ToSNSInput(topicArn)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	var expectedAttr sns.MessageAttributeValue
	expectedAttr.SetDataType("String").SetStringValue("value")
	expected := &sns.PublishInput{
		TopicArn: &topicArn,
		Message:  &body,
		MessageAttributes: map[string]*sns.MessageAttributeValue{
			"attr": &expectedAttr,
		},
	}

	if diff := deep.Equal(expected, got); diff != nil {
		t.Logf("expected: %v", expected)
		t.Logf("got: %v", got)
		t.Error(diff)
	}
}

func TestSNSMessageWithAttribute(t *testing.T) {
	body := "body"
	m := New(body).WithAttribute("attr", "value")

	got, err := m.ToSNSInput(topicArn)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	var expectedAttr sns.MessageAttributeValue
	expectedAttr.SetDataType("String").SetStringValue("value")
	expected := &sns.PublishInput{
		TopicArn: &topicArn,
		Message:  &body,
		MessageAttributes: map[string]*sns.MessageAttributeValue{
			"attr": &expectedAttr,
		},
	}

	if diff := deep.Equal(expected, got); diff != nil {
		t.Logf("expected: %v", expected)
		t.Logf("got: %v", got)
		t.Error(diff)
	}
}

func TestSNSMessageWithAttributeOverrides(t *testing.T) {
	body := "body"
	m := New(body).
		WithAttribute("attr", "value").
		WithAttribute("attr", "newValue")

	got, err := m.ToSNSInput(topicArn)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	var expectedAttr sns.MessageAttributeValue
	expectedAttr.SetDataType("String").SetStringValue("newValue")
	expected := &sns.PublishInput{
		TopicArn: &topicArn,
		Message:  &body,
		MessageAttributes: map[string]*sns.MessageAttributeValue{
			"attr": &expectedAttr,
		},
	}

	if diff := deep.Equal(expected, got); diff != nil {
		t.Logf("expected: %v", expected)
		t.Logf("got: %v", got)
		t.Error(diff)
	}
}

func TestSNSMessageForFifo(t *testing.T) {
	body := "body"
	groupID := "groupID"
	deduplicationID := "dedupID"
	m := New(body).WithFifoID(groupID, deduplicationID)

	got, err := m.ToSNSInput(topicArn)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	expected := &sns.PublishInput{
		TopicArn:               &topicArn,
		Message:                &body,
		MessageGroupId:         &groupID,
		MessageDeduplicationId: &deduplicationID,
	}

	if diff := deep.Equal(expected, got); diff != nil {
		t.Logf("expected: %v", expected)
		t.Logf("got: %v", got)
		t.Error(diff)
	}
}

func TestSQSSimpleMessage(t *testing.T) {
	body := "simple"
	m := New(body)

	got, err := m.ToSQSInput(queueURL)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	expected := &sqs.SendMessageInput{
		QueueUrl:    &queueURL,
		MessageBody: &body,
	}

	if diff := deep.Equal(expected, got); diff != nil {
		t.Logf("expected: %v", expected)
		t.Logf("got: %v", got)
		t.Error(diff)
	}

}

func TestSQSMessageWithAttribute(t *testing.T) {
	body := "body"
	m := New(body).WithAttributes(map[string]string{"attr": "value"})

	got, err := m.ToSQSInput(queueURL)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	var expectedAttr sqs.MessageAttributeValue
	expectedAttr.SetDataType("String").SetStringValue("value")
	expected := &sqs.SendMessageInput{
		QueueUrl:    &queueURL,
		MessageBody: &body,
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"attr": &expectedAttr,
		},
	}

	if diff := deep.Equal(expected, got); diff != nil {
		t.Logf("expected: %v", expected)
		t.Logf("got: %v", got)
		t.Error(diff)
	}
}

func TestSQSMessageForFifo(t *testing.T) {
	body := "body"
	groupID := "groupID"
	deduplicationID := "dedupID"
	m := New(body).WithFifoID(groupID, deduplicationID)

	got, err := m.ToSQSInput(queueURL)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	expected := &sqs.SendMessageInput{
		QueueUrl:               &queueURL,
		MessageBody:            &body,
		MessageGroupId:         &groupID,
		MessageDeduplicationId: &deduplicationID,
	}

	if diff := deep.Equal(expected, got); diff != nil {
		t.Logf("expected: %v", expected)
		t.Logf("got: %v", got)
		t.Error(diff)
	}
}

func TestNewForJSON(t *testing.T) {
	obj := struct {
		Name string
	}{
		Name: "name",
	}

	got, err := NewForJSON(obj).ToSNSInput(topicArn)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	expectedBody := `{"Name":"name"}`
	expected := &sns.PublishInput{
		TopicArn: &topicArn,
		Message:  &expectedBody,
	}

	if diff := deep.Equal(expected, got); diff != nil {
		t.Logf("expected: %v", expected)
		t.Logf("got: %v", got)
		t.Error(diff)
	}
}

func TestNewCompressedForJSON(t *testing.T) {
	obj := struct {
		Name string
	}{
		Name: "name",
	}

	got, err := NewCompressedForJSON(obj).ToSNSInput(topicArn)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	asJSON := `{"Name":"name"}`
	expectedBody, err := compress.Encode(&asJSON)
	if err != nil {
		t.Fatalf("unexpected encode error %v", err)
	}
	expected := &sns.PublishInput{
		TopicArn: &topicArn,
		Message:  expectedBody,
	}

	if diff := deep.Equal(expected, got); diff != nil {
		t.Logf("expected: %v", expected)
		t.Logf("got: %v", got)
		t.Error(diff)
	}
}

func TestSNSErrInMessagePreventsCreatingInput(t *testing.T) {
	var u unmarshable
	_, err := NewForJSON(u).ToSNSInput(topicArn)
	if err == nil {
		t.Fatal("expected error but got nil")
	}
}

func TestSQSErrInMessagePreventsCreatingInput(t *testing.T) {
	var u unmarshable
	_, err := NewForJSON(u).ToSQSInput(queueURL)
	if err == nil {
		t.Fatal("expected error but got nil")
	}
}

type unmarshable struct {
}

func (u unmarshable) MarshalJSON() ([]byte, error) {
	return nil, errors.New("unmarshable")
}
