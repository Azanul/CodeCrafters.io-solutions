package main

import (
	"fmt"
	"strings"
)

type Request struct {
	method  string
	path    string
	headers map[string]string
	body    string
}

func (req *Request) GetMethod() string {
	return req.method
}

func (req *Request) GetPath() string {
	return req.path
}

func (req *Request) GetHeader(key string) string {
	return req.headers[key]
}

func (req *Request) GetBody() string {
	return req.body
}

func UnmarshalRequest(b []byte) Request {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	bStr := string(b)
	reqLines := strings.Split(bStr, lineEnding)

	firstLine := strings.SplitN(reqLines[0], " ", 3)

	headers := map[string]string{}
	for _, line := range reqLines[1 : len(reqLines)-1] {
		if line == "" {
			break
		}
		x := strings.Split(line, ": ")
		headers[x[0]] = x[1]
	}

	return Request{
		method:  firstLine[0],
		path:    firstLine[1],
		headers: headers,
		body:    reqLines[len(reqLines)-1],
	}
}
