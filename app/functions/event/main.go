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

	"github.com/superkruger/nostr_app_data/app/utils"
)

type handler struct {
	responder           utils.ProxyResponder
	managementApiClient *apigatewaymanagementapi.Client
}

func mustNewHandler() *handler {
	return &handler{
		managementApiClient: apigatewaymanagementapi.New(apigatewaymanagementapi.Options{
			BaseEndpoint: jsii.String(os.Getenv("WS_API_ENDPOINT")),
			Region:       os.Getenv("AWS_REGION"),
		}),
	}
}

func (h *handler) handleRequest(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (utils.Response, error) {
	log.Printf("got event %+v", request)
	log.Printf("sending events to %s", *h.managementApiClient.Options().BaseEndpoint)
	// for each connection with a request matching the event
	//_, err := h.managementApiClient.PostToConnection(ctx, &apigatewaymanagementapi.PostToConnectionInput{
	//	ConnectionId: nil, //
	//	Data:         nil,
	//})
	//if err != nil {
	//	log.Println(err)
	//	return h.responder.WithStatus(http.StatusInternalServerError), err
	//}
	return h.responder.WithStatus(http.StatusOK), nil
}

func main() {
	h := mustNewHandler()
	lambda.Start(h.handleRequest)
}
