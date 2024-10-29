package main

import (
	"context"
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
	var r requests.Request
	if err := r.Unmarshal(request.Body, request.RequestContext.ConnectionID); err != nil {
		log.Printf("failed to unmarshal request body: %v", err)
		return h.responder.WithStatus(http.StatusBadRequest), nil
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
