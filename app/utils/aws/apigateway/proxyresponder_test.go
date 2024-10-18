package apigateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/pascaldekloe/goe/verify"
)

func TestProxyResponderWithoutBody(t *testing.T) {
	t.Log("given a ProxyResponder with origin allowed")
	proxyResponder := NewProxyResponder("*")

	t.Log("when making a response with only a status")
	response := proxyResponder.WithStatus(http.StatusNoContent)

	t.Log("then there should be no body in the result")
	bytes, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("did not expect error %v", err)
	}
	expectBody(t, bytes, "")
}

func TestProxyResponderWithoutOriginAllowed(t *testing.T) {
	t.Log("given a ProxyResponder without origin allowed")
	proxyResponder := NewProxyResponder("")

	t.Log("when making a response with only a status")
	response := proxyResponder.WithStatus(http.StatusNoContent)

	t.Log("then there should be no Access-Control-Allow-Origin header in the result")
	bytes, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("did not expect error %v", err)
	}
	expectNoHeader(t, bytes, "Access-Control-Allow-Origin")
}

func TestProxyResponderWithBody(t *testing.T) {
	t.Log("given a ProxyResponder")
	proxyResponder := NewProxyResponder("*")

	t.Log("when making a response with a status and a body")
	response := proxyResponder.WithStatus(http.StatusNoContent).WithBody("123", "text/plain")

	t.Log("then the body and the content type should be part of the response")
	bytes, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("did not expect error %v", err)
	}
	expectBody(t, bytes, "123")
	expectHeader(t, bytes, "Content-Type", "text/plain")
}

func TestProxyResponderWithJSONBody(t *testing.T) {
	t.Log("given a ProxyResponder")
	proxyResponder := NewProxyResponder("*")

	t.Log("when making a response with a json body")
	body := struct{ Name string }{Name: "Sleeping Beauty"}
	response := proxyResponder.WithStatus(http.StatusNoContent).WithJSONBody(body)

	t.Log("then the body should be part of the response and the context type should be json")
	bytes, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("did not expect error %v", err)
	}
	expectBody(t, bytes, "{\"Name\":\"Sleeping Beauty\"}")
	expectHeader(t, bytes, "Content-Type", "application/json; charset=utf-8")
}

func TestProxyResponderWithJSONBodyGzip(t *testing.T) {
	proxyResponder := NewProxyResponder("*")
	longName := `long name more than 1024 characters long name more than 1024 characters long name more than 1024 characters long name more than 1024 characters long name more than 1024 characters long name more than 1024 characterslong name more than 1024 characters long name more than 1024 characters long name more than 1024 characters long name more than 1024 characters long name more than 1024 characters long name more than 1024 characters long name more than 1024 characterslong name more than 1024 characters long name more than 1024 characters long name more than 1024 characters long name more than 1024 characters long name more than 1024 characters long name more than 1024 characters long name more than 1024 characterslong name more than 1024 characters long name more than 1024 characters long name more than 1024 characters long name more than 1024 characters long name more than 1024 characters long name more than 1024 characters long name more than 1024 characterslong name more than 1024 characters long name more than 1024 characters long name more than 1024 characters long name more than 1024 characters long name more than 1024 characters long name more than 1024 characters long name more than 1024 characterslong name more than 1024 characters`
	body := struct {
		Name string `json:"name"`
	}{Name: longName}
	response := proxyResponder.WithStatus(http.StatusNoContent).WithJSONBody(body).WithGzip()
	wantDecodedBody := `H4sIAAAAAAAA/6pWykvMTVWyUsrJz0tXALEVcvOLUhVKMhLzFAwNjEwUkjMSixKTS1KLihWGoJpB5pxRr496fdTrA+d1pVpAAAAA//9YDWN48QQAAA==`
	verify.Values(t, "body", response.Body, wantDecodedBody)
}

func TestProxyResponderWithPlainTextBody(t *testing.T) {
	t.Log("given a ProxyResponder")
	proxyResponder := NewProxyResponder("*")

	t.Log("when making a response with a plain text body")
	body := "Sleeping Beauty"
	response := proxyResponder.WithStatus(http.StatusNoContent).WithPlainTextBody(body)

	t.Log("then the body should be part of the response and the context type should be plain text")
	bytes, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("did not expect error %v", err)
	}
	expectBody(t, bytes, "Sleeping Beauty")
	expectHeader(t, bytes, "Content-Type", "text/plain; charset=utf-8")
}

func TestProxyResponderWithErrorBody(t *testing.T) {
	t.Log("given a ProxyResponder")
	proxyResponder := NewProxyResponder("*")

	t.Log("when making a response with an error body")
	body := ErrSomethingWentWrong{Text: "oeps"}
	response := proxyResponder.WithStatus(http.StatusNoContent).WithErrorBody(body)

	t.Log("then the body should be part of the response and the context type should be plain text")
	bytes, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("did not expect error %v", err)
	}
	expectBody(t, bytes, body.Error())
	expectHeader(t, bytes, "Content-Type", "text/plain; charset=utf-8")
}

func TestProxyResponderWithJSONBodyTooBig(t *testing.T) {
	t.Log("given a ProxyResponder")
	proxyResponder := NewProxyResponder("*")

	t.Log("when making a response with a json body which is too big")
	lyricsBuilder := strings.Builder{}
	for i := 99999; i > 0; i-- {
		lyricsBuilder.WriteString(strconv.Itoa(i))
		lyricsBuilder.WriteString(" bottles of beer on the wall, ")
		lyricsBuilder.WriteString(strconv.Itoa(i))
		lyricsBuilder.WriteString(" bottles of beer. Take one down and pass it around, ")
		lyricsBuilder.WriteString(strconv.Itoa(i - 1))
		lyricsBuilder.WriteString(" bottles of beer on the wall.")
	}
	body := struct{ Lyrics string }{Lyrics: lyricsBuilder.String()}
	response := proxyResponder.WithStatus(http.StatusNoContent).WithJSONBody(body)

	t.Log("then an error is logged with the size of the body, no error is thrown, as the API Gateway will do that")

	t.Log(strconv.Itoa(response.StatusCode))
}

func TestProxyResponseShouldAlwaysPreventCaching(t *testing.T) {
	t.Log("given a ProxyResponder with origin allowed")
	proxyResponder := NewProxyResponder("*")

	t.Log("when making a response with only a status")
	response := proxyResponder.WithStatus(http.StatusNoContent)

	t.Log("then there should be a no-store directive")
	bytes, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("did not expect error %v", err)
	}
	expectHeader(t, bytes, "Cache-Control", "no-store")
}

/// Helper Functions ///

func expectHeader(t *testing.T, bytes []byte, header string, value string) {
	var result struct {
		Headers map[string]string `json:"headers"`
	}
	err := json.Unmarshal(bytes, &result)
	if err != nil {
		t.Fatalf("did not expect error %v", err)
	}
	if value != result.Headers[header] {
		t.Errorf("unexpected response %s", bytes)
	}
}

func expectBody(t *testing.T, bytes []byte, body string) {
	var result struct {
		Body string `json:"body"`
	}
	err := json.Unmarshal(bytes, &result)
	if err != nil {
		t.Fatalf("did not expect error %v", err)
	}
	if result.Body != body {
		t.Errorf("unexpected response %s", bytes)
	}
}

func expectNoHeader(t *testing.T, bytes []byte, header string) {
	var result struct {
		Headers map[string]string `json:"headers"`
	}
	err := json.Unmarshal(bytes, &result)
	if err != nil {
		t.Fatalf("did not expect error %v", err)
	}
	if _, found := result.Headers[header]; found {
		t.Errorf("unexpected response %s", bytes)
	}
}

// / Helper Types ///
type ErrSomethingWentWrong struct {
	Text string
}

func (e ErrSomethingWentWrong) Error() string {
	return fmt.Sprintf("There was an error: %v", e.Text)
}
