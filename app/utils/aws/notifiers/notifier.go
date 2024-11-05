package notifiers

import (
	"context"

	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"

	"github.com/superkruger/nostr_app_data/app/utils/aws/notifiers/messages"
)

// Notifier knows how to notify the messages
type Notifier interface {
	Send(ctx context.Context, message messages.Message) error
}

// NewSQSNotifier returns a notifier based on SQS
func NewSQSNotifier(sqsSvc *sqs.SQS, queueURL string) Notifier {
	return sqsNotifier{
		sqsSvc:   sqsSvc,
		queueURL: queueURL,
	}
}

// NewSNSNotifier returns a notifier based on SNS
func NewSNSNotifier(snsSvc *sns.SNS, topicArn string) Notifier {
	return snsNotifier{
		snsSvc:   snsSvc,
		topicArn: topicArn,
	}
}

type sqsNotifier struct {
	sqsSvc   *sqs.SQS
	queueURL string
}

func (n sqsNotifier) Send(ctx context.Context, message messages.Message) error {
	input, err := message.ToSQSInput(n.queueURL)
	if err != nil {
		return err
	}
	_, err = n.sqsSvc.SendMessageWithContext(ctx, input)
	return err
}

type snsNotifier struct {
	snsSvc   *sns.SNS
	topicArn string
}

func (n snsNotifier) Send(ctx context.Context, message messages.Message) error {
	input, err := message.ToSNSInput(n.topicArn)
	if err != nil {
		return err
	}
	_, err = n.snsSvc.PublishWithContext(ctx, input)
	return err
}
