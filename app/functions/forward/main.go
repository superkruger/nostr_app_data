package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
	"github.com/aws/jsii-runtime-go"
	"github.com/goccy/go-json"
	log "github.com/sirupsen/logrus"
	"github.com/superkruger/nostr_app_data/app/domain"
	"github.com/superkruger/nostr_app_data/app/domain/requests"
	"golang.org/x/sync/errgroup"

	conns "github.com/superkruger/nostr_app_data/app/domain/connections"
	"github.com/superkruger/nostr_app_data/app/utils/aws/apigateway"
	"github.com/superkruger/nostr_app_data/app/utils/env"
	"github.com/superkruger/nostr_app_data/app/utils/skmongo"
)

const workers = 5

type handler struct {
	responder           apigateway.ProxyResponder
	managementApiClient *apigatewaymanagementapi.ApiGatewayManagementApi
	connService         conns.Service
	reqService          requests.Service
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
		reqService:  requests.NewService(requests.WithRepo(requests.NewRepository(db))),
		shutdown: func() {
			closeDb()
		},
	}
}

func (h *handler) handleEvent(ctx context.Context, event events.SQSEvent) error {
	for _, record := range event.Records {
		//log.Infof("got record %+v", record)
		var recordMessage struct {
			Message string `json:"Message"`
		}
		if err := json.Unmarshal([]byte(record.Body), &recordMessage); err != nil {
			log.Infof("failed to unmarshal record body: %v", err)
			return err
		}
		var forwardEvent domain.ForwardEvent
		if err := json.Unmarshal([]byte(recordMessage.Message), &forwardEvent); err != nil {
			log.Infof("failed to unmarshal record message: %v", err)
			return err
		}
		log.Infof("sending event %v to %d subscribers", forwardEvent.Event, len(forwardEvent.Subscribers))

		g, gCtx := errgroup.WithContext(ctx)
		subChan := make(chan domain.Subscriber)
		go func() {
			for _, sub := range forwardEvent.Subscribers {
				subChan <- sub
			}
		}()
		for i := 0; i < calcWorkers(len(forwardEvent.Subscribers)); i++ {
			g.Go(func() error {
				for sub := range subChan {
					_, err := h.managementApiClient.PostToConnectionWithContext(gCtx, &apigatewaymanagementapi.PostToConnectionInput{
						ConnectionId: jsii.String(sub.ConnID),
						Data:         []byte(eventBody(sub.ID, forwardEvent.Event)),
					})
					if err != nil {
						log.Infof("error posting to connection: %v", err)
						_ = h.connService.Remove(gCtx, sub.ConnID)
						_ = h.reqService.Remove(gCtx, sub.ID)
					}
				}
				return nil
			})
		}
		if err := g.Wait(); err != nil {
			log.Errorf("error waiting for errgroup: %v", err)
			return err
		}
		log.Infof("forwarded event")
	}
	return nil
}

func calcWorkers(taskLen int) int {
	if taskLen <= workers {
		return taskLen
	}
	return workers
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
