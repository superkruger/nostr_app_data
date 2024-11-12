package main

import (
	"context"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/superkruger/nostr_app_data/app/utils/aws/apigateway"
)

type handler struct {
	responder apigateway.ProxyResponder
	shutdown  func()
}

func mustNewHandler() *handler {
	return &handler{
		shutdown: func() {
		},
	}
}

func (h *handler) handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (apigateway.Response, error) {
	log.Printf("got request %+v", request)
	return h.responder.WithStatus(http.StatusOK), nil
}

func main() {
	h := mustNewHandler()
	lambda.StartWithOptions(h.handleRequest, lambda.WithEnableSIGTERM(h.shutdown))
}
