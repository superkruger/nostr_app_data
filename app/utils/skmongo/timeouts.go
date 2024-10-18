package skmongo

import (
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"
)

// Timeouts configure the timeouts for connecting to Mongo
type Timeouts struct {
	// How long can it take to find the right server
	ServerSelection time.Duration
	// How long can it take to connect to the server
	Connect time.Duration
	// How long can it take to send or receive on a socket
	Socket time.Duration
	// How long connections can be open
	MaxIdleTimeMS time.Duration
}

// setValuesIn sets the relevant values in the query for constructing
// the connection string
func (t Timeouts) setValuesIn(options *options.ClientOptions) *options.ClientOptions {
	if t.ServerSelection != 0 {
		options.SetServerSelectionTimeout(t.ServerSelection)
	}
	if t.Connect != 0 {
		options.SetConnectTimeout(t.Connect)
	}
	if t.Socket != 0 {
		options.SetSocketTimeout(t.Socket)
	}
	if t.MaxIdleTimeMS != 0 {
		options.SetMaxConnIdleTime(t.MaxIdleTimeMS)
	}

	return options
}

// timeoutsForLambda are the typical timeouts for AWS Lambda functions
func timeoutsForLambda() Timeouts {
	return Timeouts{
		ServerSelection: 3 * time.Second,  // This is in the init of the function
		Connect:         1 * time.Second,  // This is typically 2 milliseconds
		Socket:          10 * time.Second, // This depends on the size of the request/response
		MaxIdleTimeMS:   15 * time.Minute, // This is the maximum time a connection can stay idle.
	}
}

// optionsForTest are the typical options for a testing environment.
func optionsForTest() *options.ClientOptions {
	return Timeouts{
		ServerSelection: 10 * time.Second, // This is in the init of the function
		Connect:         20 * time.Second, // This is typically 2 milliseconds
		Socket:          30 * time.Second, // This depends on the size of the request/response
		MaxIdleTimeMS:   15 * time.Minute, // This is the maximum time a connection can stay idle.
	}.setValuesIn(options.Client())
}
