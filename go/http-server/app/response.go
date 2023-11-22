package main

import (
	"fmt"
	"net/http"
	"strings"
)

type Response struct {
	StatusCode int
	Path       string
	Headers    map[string]string
	Body       string
}

func (res *Response) SetHeader(k string, v string) {
	res.Headers[k] = v
}

func (res *Response) MarshalResponse() []byte {
	var response strings.Builder

	response.WriteString(fmt.Sprintf("%s %d %s"+lineEnding, httpVersion, res.StatusCode, http.StatusText(res.StatusCode)))
	for k, v := range res.Headers {
		response.WriteString(k + ": " + v + lineEnding)
	}
	response.WriteString(lineEnding)
	response.WriteString(res.Body)

	return []byte(response.String())
}

func NewResponse() Response {
	return Response{
		Headers: map[string]string{},
	}
}
