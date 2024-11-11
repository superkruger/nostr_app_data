package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/goccy/go-json"

	"github.com/superkruger/nostr_app_data/app/domain"
	evts "github.com/superkruger/nostr_app_data/app/domain/events"
	req "github.com/superkruger/nostr_app_data/app/domain/requests"
	"github.com/superkruger/nostr_app_data/app/utils/aws/apigateway"
	"github.com/superkruger/nostr_app_data/app/utils/aws/notifiers"
	"github.com/superkruger/nostr_app_data/app/utils/aws/notifiers/messages"
	"github.com/superkruger/nostr_app_data/app/utils/env"
	"github.com/superkruger/nostr_app_data/app/utils/skmongo"
)

type handler struct {
	responder  apigateway.ProxyResponder
	evtService evts.Service
	reqService req.Service
	notifier   notifiers.Notifier
	shutdown   func()
}

func mustNewHandler() *handler {
	db, closeDb := skmongo.MustFromSecretWithClose(env.MustGetString("DB_SECRET"))
	sess := session.Must(session.NewSession())
	return &handler{
		evtService: evts.NewService(evts.WithRepo(evts.NewRepository(db))),
		reqService: req.NewService(req.WithRepo(req.NewRepository(db))),
		notifier:   notifiers.NewSNSNotifier(sns.New(sess), env.MustGetString("EVENT_FORWARD_TOPIC")),
		shutdown: func() {
			closeDb()
		},
	}
}

func (h *handler) handleRequest(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (apigateway.Response, error) {
	log.Printf("got event %+v", request.Body)
	var raw []json.RawMessage
	if err := json.Unmarshal([]byte(request.Body), &raw); err != nil {
		log.Printf("failed to unmarshal request body: %v", err)
		return h.responder.WithStatus(http.StatusBadRequest).WithJSONBody(eventResult(false, "", "error: could not unmarshal request body")), nil
	}
	if len(raw) != 2 {
		log.Printf("expected a length of 2")
		return h.responder.WithStatus(http.StatusBadRequest).WithJSONBody(eventResult(false, "", "error: wrong event format")), nil
	}
	var e domain.Event
	if err := json.Unmarshal(raw[1], &e); err != nil {
		log.Printf("failed to unmarshal event: %v", err)
		return h.responder.WithStatus(http.StatusBadRequest).WithJSONBody(eventResult(false, e.ID, "error: could not unmarshal event")), nil
	}
	if err := h.evtService.Add(ctx, e); err != nil {
		log.Printf("failed to add event: %v", err)
		return h.responder.WithStatus(http.StatusInternalServerError).WithJSONBody(eventResult(false, e.ID, "error: failed to store event")), nil
	}
	requests, err := h.reqService.Find(ctx, e)
	if err != nil {
		log.Printf("failed to find requests: %v", err)
		return h.responder.WithStatus(http.StatusInternalServerError).WithJSONBody(eventResult(true, e.ID, "error: failed to find matching requests")), nil
	}
	subscribers := make([]domain.Subscriber, 0, len(requests))
	for _, r := range requests {
		subscribers = append(subscribers, domain.Subscriber{
			ID:     r.ID,
			ConnID: r.ConnID,
		})
	}
	forwardEvent := domain.ForwardEvent{
		Subscribers: subscribers,
		Event:       request.Body,
	}
	log.Printf("forwarding event to %d subscribers", len(subscribers))
	if len(subscribers) > 0 {
		if err := h.notifier.Send(ctx, messages.NewForJSON(forwardEvent).WithFifoID(e.PubKey, e.ID)); err != nil {
			log.Printf("failed to send forward event: %v", err)
			return h.responder.WithStatus(http.StatusInternalServerError).WithJSONBody(eventResult(true, e.ID, "error: failed to forward event")), nil
		}
	}
	return h.responder.WithStatus(http.StatusOK).WithJSONBody(eventResult(true, e.ID, "")), nil
}

func eventResult(ok bool, id, reason string) string {
	return fmt.Sprintf("[\"OK\",\"%s\",%v,\"%s\"]", id, ok, reason)
}

func main() {
	h := mustNewHandler()
	lambda.StartWithOptions(h.handleRequest, lambda.WithEnableSIGTERM(h.shutdown))
}
