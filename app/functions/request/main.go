package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/superkruger/nostr_app_data/app/utils/aws/notifiers"
	"github.com/superkruger/nostr_app_data/app/utils/aws/notifiers/messages"

	"github.com/superkruger/nostr_app_data/app/domain"
	evts "github.com/superkruger/nostr_app_data/app/domain/events"
	"github.com/superkruger/nostr_app_data/app/domain/requests"
	"github.com/superkruger/nostr_app_data/app/utils/aws/apigateway"
	"github.com/superkruger/nostr_app_data/app/utils/env"
	"github.com/superkruger/nostr_app_data/app/utils/skmongo"
)

type handler struct {
	responder      apigateway.ProxyResponder
	reqService     requests.Service
	evtService     evts.Service
	notifier       notifiers.Notifier
	notifierIssues notifiers.Notifier
	shutdown       func()
}

func mustNewHandler() *handler {
	db, closeDb := skmongo.MustFromSecretWithClose(env.MustGetString("DB_SECRET"))
	sess := session.Must(session.NewSession())
	return &handler{
		reqService:     requests.NewService(requests.WithRepo(requests.NewRepository(db))),
		evtService:     evts.NewService(evts.WithRepo(evts.NewRepository(db))),
		notifier:       notifiers.NewSNSNotifier(sns.New(sess), env.MustGetString("EVENT_FORWARD_TOPIC")),
		notifierIssues: notifiers.NewSNSNotifier(sns.New(sess), env.MustGetString("ISSUES_TOPIC")),
		shutdown: func() {
			closeDb()
		},
	}
}

func (h *handler) handleRequest(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (apigateway.Response, error) {
	log.Printf("got request %+v", request.Body)
	var r domain.Request
	if err := r.Unmarshal(request.Body, request.RequestContext.ConnectionID); err != nil {
		log.Printf("failed to unmarshal request body: %v", err)
		return h.responder.WithStatus(http.StatusBadRequest), nil
	}
	unIndexedTags := r.UnIndexedTags()
	if len(unIndexedTags) > 0 {
		if err := h.notifierIssues.Send(ctx, messages.NewForJSON(unIndexedTags)); err != nil {
			log.Printf("failed to send issues event: %v", err)
		}
	}
	if err := h.reqService.Add(ctx, r); err != nil {
		log.Printf("failed to add request: %v", err)
		return h.responder.WithStatus(http.StatusInternalServerError), nil
	}
	reqEvents, err := h.evtService.FindForRequest(ctx, r)
	if err != nil {
		log.Printf("failed to find request events: %v", err)
		return h.responder.WithStatus(http.StatusInternalServerError), nil
	}
	log.Printf("forwarding %d events", len(reqEvents))
	for _, reqEvent := range reqEvents {
		eventJson, err := json.Marshal(reqEvent)
		if err != nil {
			log.Printf("failed to marshal event: %v", err)
			return h.responder.WithStatus(http.StatusInternalServerError), nil
		}
		forwardEvent := domain.ForwardEvent{
			Subscribers: []domain.Subscriber{r.Subscriber},
			Event:       string(eventJson),
		}
		if err := h.notifier.Send(ctx, messages.NewForJSON(forwardEvent).WithFifoID(r.ID, reqEvent.ID)); err != nil {
			log.Printf("failed to send forward event: %v", err)
		}
	}
	if len(reqEvents) > 0 {
		log.Printf("forwarding EOSE event")
		forwardEvent := domain.ForwardEvent{
			Subscribers: []domain.Subscriber{r.Subscriber},
			Event:       domain.EventTypeEOSE,
		}
		if err := h.notifier.Send(ctx, messages.NewForJSON(forwardEvent).WithFifoID(r.ID, domain.EventTypeEOSE)); err != nil {
			log.Printf("failed to send forward event: %v", err)
		}
	}
	return h.responder.WithStatus(http.StatusOK), nil
}

func main() {
	h := mustNewHandler()
	lambda.StartWithOptions(h.handleRequest, lambda.WithEnableSIGTERM(h.shutdown))
}
