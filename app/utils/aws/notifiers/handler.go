package notifiers

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-xray-sdk-go/xray"

	"github.com/superkruger/nostr_app_data/app/utils/aws/notifiers/messages"
)

// Handler for handling a lambda notifier request, usually an error notifier.
type Handler struct {
	notifier Notifier
}

// MustNewHandler returns a new handler.
func MustNewHandler(importantIssueTopicArn string) Handler {
	snsClient := sns.New(session.Must(session.NewSession()))
	xray.AWS(snsClient.Client)

	return Handler{
		notifier: NewSNSNotifier(snsClient, importantIssueTopicArn),
	}
}

// HandleRequest handles a lambda request from the main.
// A typical error notifier will have the following form:
//
//	func main() {
//		h := MustNewHandler(os.Getenv("IMPORTANT_ISSUE_TOPIC_ARN"))
//		lambda.Start(h.handleRequest)
//	}
func (h *Handler) HandleRequest(ctx context.Context, event interface{}) error {
	return h.notifier.Send(ctx, messages.NewForJSON(event))
}
