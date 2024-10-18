/*
Package apigateway contains proxy responder to easily create proxy responses
*/
package apigateway

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	log "github.com/sirupsen/logrus"
)

const minCompressSize = 1024

// ProxyResponder is the responder that gives API Gateway Proxy responses
type ProxyResponder struct {
	originAllowed string
}

// NewProxyResponder creates a new proxy responder
func NewProxyResponder(originAllowed string) ProxyResponder {
	return ProxyResponder{
		originAllowed: originAllowed,
	}
}

// WithStatus creates a response with the give status code
// You can add to the response later on
func (r ProxyResponder) WithStatus(statusCode int) Response {
	response := Response{}
	response.StatusCode = statusCode
	response.Headers = map[string]string{"Cache-Control": "no-store"}
	if r.originAllowed != "" {
		response.Headers["Access-Control-Allow-Origin"] = r.originAllowed
	}
	return response
}

/***************
* The Response *
****************/

// Response is the response that will marshal to a correct api gateway proxy response
type Response struct {
	events.APIGatewayProxyResponse
}

// WithJSONBody adds the body to the response
// It will panic when the body can not be marshalled to json
func (r Response) WithJSONBody(body interface{}) Response {
	bytes, err := json.Marshal(body)
	if err != nil {
		panic(err) // preserve the fluent interface, this should never happen
	}
	return r.WithBody(string(bytes), "application/json; charset=utf-8")
}

// WithPlainTextBody adds the plain text body to the response
func (r Response) WithPlainTextBody(body string) Response {
	return r.WithBody(body, "text/plain; charset=utf-8")
}

// WithErrorBody adds the error body to the response
func (r Response) WithErrorBody(body error) Response {
	return r.WithBody(body.Error(), "text/plain; charset=utf-8")
}

// WithBody adds the body to the response, and sets the content-type
func (r Response) WithBody(body string, contentType string) Response {
	r.Body = body
	r.Headers["Content-Type"] = contentType
	return r
}

func (r Response) WithGzip() Response {
	if len([]byte(r.Body)) < minCompressSize {
		return r
	}
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	if _, err := w.Write([]byte(r.Body)); err != nil {
		log.WithError(err).Error("failed to compress gzipped body")
		r.StatusCode = http.StatusInternalServerError
		return r
	}
	_ = w.Close()
	r.Body = base64.StdEncoding.EncodeToString(b.Bytes())
	r.IsBase64Encoded = true
	r.Headers["Content-Encoding"] = "gzip"
	delete(r.Headers, "Content-Length")
	return r
}
