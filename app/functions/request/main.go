package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/superkruger/nostr_app_data/app/domain/requests"
	"github.com/superkruger/nostr_app_data/app/utils/env"
	"github.com/superkruger/nostr_app_data/app/utils/skmongo"

	"github.com/superkruger/nostr_app_data/app/utils/aws/apigateway"
)

type handler struct {
	responder  apigateway.ProxyResponder
	reqService requests.Service
	shutdown   func()
}

func mustNewHandler() *handler {
	db, closeDb := skmongo.MustFromSecretWithClose(env.MustGetString("DB_SECRET"))
	return &handler{
		reqService: requests.NewService(requests.WithRepo(requests.NewRepository(db))),
		shutdown: func() {
			closeDb()
		},
	}
}

func (h *handler) handleRequest(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (apigateway.Response, error) {
	log.Printf("got request %+v", request.Body)
	var raw []json.RawMessage
	if err := json.Unmarshal([]byte(request.Body), &raw); err != nil {
		log.Printf("failed to unmarshal request body: %v", err)
		return h.responder.WithStatus(http.StatusBadRequest), nil
	}
	if len(raw) < 3 {
		log.Printf("expected a length of at least 3")
		return h.responder.WithStatus(http.StatusBadRequest), nil
	}
	r := requests.Request{
		ID:      string(raw[1]),
		ConnID:  request.RequestContext.ConnectionID,
		Filters: make([]requests.Filter, len(raw[2:])),
	}
	for _, rawFilter := range raw[2:] {
		var f requests.Filter
		if err := json.Unmarshal(rawFilter, &f); err != nil {
			log.Printf("failed to unmarshal event: %v", err)
			return h.responder.WithStatus(http.StatusBadRequest), nil
		}
		r.Filters = append(r.Filters, f)
	}
	if err := h.reqService.Add(ctx, r); err != nil {
		log.Printf("failed to add request: %v", err)
		return h.responder.WithStatus(http.StatusInternalServerError), nil
	}
	return h.responder.WithStatus(http.StatusOK), nil
}

func main() {
	h := mustNewHandler()
	lambda.StartWithOptions(h.handleRequest, lambda.WithEnableSIGTERM(h.shutdown))
}
