package main

import (
	"context"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/superkruger/nostr_app_data/app/domain/connections"
	"github.com/superkruger/nostr_app_data/app/utils/env"
	"github.com/superkruger/nostr_app_data/app/utils/skmongo"

	"github.com/superkruger/nostr_app_data/app/utils/aws/apigateway"
)

type handler struct {
	responder   apigateway.ProxyResponder
	connService connections.Service
	shutdown    func()
}

func mustNewHandler() *handler {
	db, closeDb := skmongo.MustFromSecretWithClose(env.MustGetString("DB_SECRET"))
	return &handler{
		connService: connections.NewService(connections.WithRepo(connections.NewRepository(db))),
		shutdown: func() {
			closeDb()
		},
	}
}

func (h *handler) handleRequest(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (apigateway.Response, error) {
	log.Printf("got request %+v", request)
	return h.responder.WithStatus(http.StatusOK), nil
}

func main() {
	h := mustNewHandler()
	lambda.StartWithOptions(h.handleRequest, lambda.WithEnableSIGTERM(h.shutdown))
}
