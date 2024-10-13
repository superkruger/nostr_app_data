package main

import (
	"context"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/superkruger/nostr_app_data/app/utils"
)

type handler struct {
	responder utils.ProxyResponder
}

func mustNewHandler() *handler {
	return &handler{}
}

func (h *handler) handleRequest(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (utils.Response, error) {
	log.Printf("disconnecting: %s", request.RequestContext.ConnectionID)
	return h.responder.WithStatus(http.StatusOK), nil
}

func main() {
	h := mustNewHandler()
	lambda.Start(h.handleRequest)
}
