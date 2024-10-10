package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
)

func handleRequest(ctx context.Context) error {
	log.Println("test function")
	return nil
}

func main() {
	lambda.Start(handleRequest)
}
