package main

import (
	"context"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
	"github.com/aws/jsii-runtime-go"
	"github.com/superkruger/nostr_app_data/app/domain/connections"
	"github.com/superkruger/nostr_app_data/app/utils/env"
	"github.com/superkruger/nostr_app_data/app/utils/skmongo"

	"github.com/superkruger/nostr_app_data/app/utils/aws/apigateway"
)

type handler struct {
	responder           apigateway.ProxyResponder
	managementApiClient *apigatewaymanagementapi.ApiGatewayManagementApi
	//managementApiClient *apigatewaymanagementapi.Client
	connService connections.Service
	shutdown    func()
}

func mustNewHandler() *handler {
	log.Println("WS_API_ENDPOINT", env.MustGetString("WS_API_ENDPOINT"))
	log.Println("AWS_REGION", env.MustGetString("AWS_REGION"))
	db, closeDb := skmongo.MustFromSecretWithClose(env.MustGetString("DB_SECRET"))
	sess := session.Must(session.NewSession())
	return &handler{
		managementApiClient: apigatewaymanagementapi.New(sess, aws.NewConfig().WithRegion(env.MustGetString("AWS_REGION"))),
		//managementApiClient: apigatewaymanagementapi.New(apigatewaymanagementapi.Options{
		//	BaseEndpoint: jsii.String(env.MustGetString("WS_API_ENDPOINT")),
		//	Region:       env.MustGetString("AWS_REGION"),
		//}),
		connService: connections.NewService(connections.WithRepo(connections.NewRepository(db))),
		shutdown: func() {
			closeDb()
		},
	}
}

func (h *handler) handleRequest(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (apigateway.Response, error) {
	log.Printf("got event %+v", request)

	//api := apigatewaymanagementapi.New(apigatewaymanagementapi.Options{
	//	BaseEndpoint: jsii.String(fmt.Sprintf("https://%s/%s", request.RequestContext.DomainName, request.RequestContext.Stage)),
	//	Region:       env.MustGetString("AWS_REGION"),
	//})

	log.Printf("sending events to %s", h.managementApiClient.Endpoint)
	conns, err := h.connService.All(ctx)
	if err != nil {
		log.Printf("error getting connections: %v", err)
		return h.responder.WithStatus(http.StatusInternalServerError), nil
	}
	for _, conn := range conns {
		//if request.RequestContext.ConnectionID == conn.ID {
		//	continue
		//}
		log.Printf("sending event %s to %s", request.Body, conn.ID)
		res, err := h.managementApiClient.PostToConnection(&apigatewaymanagementapi.PostToConnectionInput{
			ConnectionId: jsii.String(conn.ID),
			Data:         []byte(request.Body),
		})
		//res, err := h.managementApiClient.PostToConnection(ctx, &apigatewaymanagementapi.PostToConnectionInput{
		//	ConnectionId: jsii.String(conn.ID),
		//	Data:         []byte(request.Body),
		//})
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
	lambda.StartWithOptions(h.handleRequest, lambda.WithEnableSIGTERM(h.shutdown))
}
