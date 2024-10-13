package main

import (
	"context"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/superkruger/nostr_app_data/utils"
)

type handler struct {
	responder utils.ProxyResponder
}

func mustNewHandler() *handler {
	return &handler{}
}

func (h *handler) handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (utils.Response, error) {
	log.Printf("got request: %+v", request)
	return h.responder.WithStatus(http.StatusOK), nil
}

func main() {
	h := mustNewHandler()
	lambda.Start(h.handleRequest)
}
