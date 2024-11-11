package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
	"github.com/aws/jsii-runtime-go"
	"github.com/goccy/go-json"
	"github.com/superkruger/nostr_app_data/app/domain"

	conns "github.com/superkruger/nostr_app_data/app/domain/connections"
	"github.com/superkruger/nostr_app_data/app/utils/aws/apigateway"
	"github.com/superkruger/nostr_app_data/app/utils/env"
	"github.com/superkruger/nostr_app_data/app/utils/skmongo"
)

type handler struct {
	responder           apigateway.ProxyResponder
	managementApiClient *apigatewaymanagementapi.ApiGatewayManagementApi
	connService         conns.Service
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
		shutdown: func() {
			closeDb()
		},
	}
}

func (h *handler) handleEvent(ctx context.Context, event events.SQSEvent) error {
	for _, record := range event.Records {
		//log.Printf("got record %+v", record)
		var recordMessage struct {
			Message string `json:"Message"`
		}
		if err := json.Unmarshal([]byte(record.Body), &recordMessage); err != nil {
			log.Printf("failed to unmarshal record body: %v", err)
			return err
		}
		var forwardEvent domain.ForwardEvent
		if err := json.Unmarshal([]byte(recordMessage.Message), &forwardEvent); err != nil {
			log.Printf("failed to unmarshal record message: %v", err)
			return err
		}
		log.Printf("sending event %v to %d subscribers", forwardEvent.Event, len(forwardEvent.Subscribers))
		for _, sub := range forwardEvent.Subscribers {
			_, err := h.managementApiClient.PostToConnection(&apigatewaymanagementapi.PostToConnectionInput{
				ConnectionId: jsii.String(sub.ConnID),
				Data:         []byte(eventBody(sub.ID, forwardEvent.Event)),
			})
			if err != nil {
				log.Printf("error posting to connection: %v", err)
				_ = h.connService.Remove(ctx, sub.ConnID)
			}
		}
	}
	return nil
}

func eventBody(subscriptionId, event string) string {
	if event == domain.EventTypeEOSE {
		return fmt.Sprintf("[\"%s\",\"%s\"]", domain.EventTypeEOSE, subscriptionId)
	}
	return fmt.Sprintf("[\"%s\",\"%s\",%s]", domain.EventTypeEvent, subscriptionId, event)
}

func main() {
	h := mustNewHandler()
	lambda.StartWithOptions(h.handleEvent, lambda.WithEnableSIGTERM(h.shutdown))
}
