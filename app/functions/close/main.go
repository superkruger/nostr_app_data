package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/superkruger/nostr_app_data/app/domain/requests"

	"github.com/superkruger/nostr_app_data/app/utils/aws/apigateway"
	"github.com/superkruger/nostr_app_data/app/utils/env"
	"github.com/superkruger/nostr_app_data/app/utils/skmongo"
)

type handler struct {
	responder apigateway.ProxyResponder
	service   requests.Service
	shutdown  func()
}

func mustNewHandler() *handler {
	db, closeDb := skmongo.MustFromSecretWithClose(env.MustGetString("DB_SECRET"))
	return &handler{
		service: requests.NewService(requests.WithRepo(requests.NewRepository(db))),
		shutdown: func() {
			closeDb()
		},
	}
}

func (h *handler) handleRequest(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (apigateway.Response, error) {
	log.Printf("got request %+v", request)
	var raw []json.RawMessage
	if err := json.Unmarshal([]byte(request.Body), &raw); err != nil {
		log.Printf("failed to unmarshal request body: %v", err)
		return h.responder.WithStatus(http.StatusBadRequest), nil
	}
	if len(raw) != 2 {
		log.Printf("expected a length of 2")
		return h.responder.WithStatus(http.StatusBadRequest), nil
	}
	var subID string
	if err := json.Unmarshal(raw[1], &subID); err != nil {
		log.Printf("failed to unmarshal subscription id: %v", err)
		return h.responder.WithStatus(http.StatusBadRequest), nil
	}
	if err := h.service.Remove(ctx, subID); err != nil {
		log.Printf("error removing request: %v", err)
		return h.responder.WithStatus(http.StatusInternalServerError), nil
	}
	return h.responder.WithStatus(http.StatusOK), nil
}

func main() {
	h := mustNewHandler()
	lambda.StartWithOptions(h.handleRequest, lambda.WithEnableSIGTERM(h.shutdown))
}
