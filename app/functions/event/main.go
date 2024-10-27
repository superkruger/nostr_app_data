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
	"github.com/goccy/go-json"

	conns "github.com/superkruger/nostr_app_data/app/domain/connections"
	evts "github.com/superkruger/nostr_app_data/app/domain/events"
	"github.com/superkruger/nostr_app_data/app/utils/env"
	"github.com/superkruger/nostr_app_data/app/utils/skmongo"

	"github.com/superkruger/nostr_app_data/app/utils/aws/apigateway"
)

type handler struct {
	responder           apigateway.ProxyResponder
	managementApiClient *apigatewaymanagementapi.ApiGatewayManagementApi
	connService         conns.Service
	evtService          evts.Service
	shutdown            func()
}

func mustNewHandler() *handler {
	log.Println("WS_API_ENDPOINT", env.MustGetString("WS_API_ENDPOINT"))
	log.Println("AWS_REGION", env.MustGetString("AWS_REGION"))
	db, closeDb := skmongo.MustFromSecretWithClose(env.MustGetString("DB_SECRET"))
	sess := session.Must(session.NewSession())
	return &handler{
		managementApiClient: apigatewaymanagementapi.New(
			sess,
			aws.NewConfig().
				WithRegion(env.MustGetString("AWS_REGION")).
				WithEndpoint(env.MustGetString("WS_API_ENDPOINT"))),
		connService: conns.NewService(conns.WithRepo(conns.NewRepository(db))),
		evtService:  evts.NewService(evts.WithRepo(evts.NewRepository(db))),
		shutdown: func() {
			closeDb()
		},
	}
}

func (h *handler) handleRequest(ctx context.Context, request events.APIGatewayWebsocketProxyRequest) (apigateway.Response, error) {
	log.Printf("got event %+v", request.Body)

	var raw []json.RawMessage
	if err := json.Unmarshal([]byte(request.Body), &raw); err != nil {
		log.Printf("failed to unmarshal request body: %v", err)
		return h.responder.WithStatus(http.StatusBadRequest), nil
	}
	if len(raw) != 2 {
		log.Printf("expected a length of 2")
		return h.responder.WithStatus(http.StatusBadRequest), nil
	}
	var e evts.Event
	if err := json.Unmarshal(raw[1], &e); err != nil {
		log.Printf("failed to unmarshal event: %v", err)
		return h.responder.WithStatus(http.StatusBadRequest), nil
	}

	if err := h.evtService.Add(ctx, e); err != nil {
		log.Printf("failed to add event: %v", err)
		return h.responder.WithStatus(http.StatusInternalServerError), nil
	}

	//log.Printf("sending events to %s", h.managementApiClient.Endpoint)
	//conns, err := h.connService.All(ctx)
	//if err != nil {
	//	log.Printf("error getting connections: %v", err)
	//	return h.responder.WithStatus(http.StatusInternalServerError), nil
	//}
	//for _, conn := range conns {
	//	//if request.RequestContext.ConnectionID == conn.ID {
	//	//	continue
	//	//}
	//	log.Printf("sending event %s to %s", request.Body, conn.ID)
	//	_, err := h.managementApiClient.PostToConnection(&apigatewaymanagementapi.PostToConnectionInput{
	//		ConnectionId: jsii.String(conn.ID),
	//		Data:         []byte(request.Body),
	//	})
	//	if err != nil {
	//		log.Printf("error posting to connection: %v", err)
	//		_ = h.connService.Remove(ctx, conn.ID)
	//	}
	//}
	return h.responder.WithStatus(http.StatusOK), nil
}

func main() {
	h := mustNewHandler()
	lambda.StartWithOptions(h.handleRequest, lambda.WithEnableSIGTERM(h.shutdown))
}
