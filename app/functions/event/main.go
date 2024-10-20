package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"github.com/aws/jsii-runtime-go"
	"github.com/superkruger/nostr_app_data/app/domain/connections"

	"github.com/superkruger/nostr_app_data/app/utils/aws/apigateway"
)

type handler struct {
	responder           apigateway.ProxyResponder
	managementApiClient *apigatewaymanagementapi.Client
	connService         connections.Service
	shutdown            func()
}

func mustNewHandler() *handler {
	return &handler{
		managementApiClient: apigatewaymanagementapi.New(apigatewaymanagementapi.Options{
			BaseEndpoint: jsii.String(os.Getenv("WS_API_ENDPOINT")),
			Region:       os.Getenv("AWS_REGION"),
		}),
	}
}

func (h *handler) handleRequest(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (apigateway.Response, error) {
	log.Printf("got event %+v", request)
	log.Printf("sending events to %s", *h.managementApiClient.Options().BaseEndpoint)
	conns, err := h.connService.All(ctx)
	if err != nil {
		log.Printf("error getting connections: %v", err)
		return h.responder.WithStatus(http.StatusInternalServerError), nil
	}
	for _, conn := range conns {
		if request.RequestContext.ConnectionID == conn.ID {
			continue
		}
		res, err := h.managementApiClient.PostToConnection(ctx, &apigatewaymanagementapi.PostToConnectionInput{
			ConnectionId: &conn.ID,
			Data:         []byte(request.Body),
		})
		if err != nil {
			log.Printf("error posting to connection: %v", err)
			return h.responder.WithStatus(http.StatusInternalServerError), nil
		}
		log.Printf("post result %+v", res)
	}
	return h.responder.WithStatus(http.StatusOK), nil
}

func main() {
	h := mustNewHandler()
	lambda.Start(h.handleRequest)
}
