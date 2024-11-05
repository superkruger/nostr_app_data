package notifiers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/superkruger/nostr_app_data/app/utils/aws/notifiers/messages"
)

func TestNotifierHappyPath(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	mock := NewMockNotifier(ctrl)
	h := Handler{notifier: mock}

	msg := "{\"name\":\"test\"}"
	var event interface{}

	if err := json.Unmarshal([]byte(msg), &event); err != nil {
		t.Fatal(err)
	}

	expected := messages.New(msg)
	mock.EXPECT().Send(gomock.Eq(ctx), gomock.Eq(expected)).Return(nil)

	err := h.HandleRequest(ctx, event)
	if err != nil {
		t.Fatal(err)
	}
}
