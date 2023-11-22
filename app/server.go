package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	httpVersion = "HTTP/1.1"
	lineEnding  = "\r\n"
	directory   *string
)

func main() {
	directory = flag.String("directory", "", "absolute path of directory to serve")
	flag.Parse()

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	data := make([]byte, 1024)
	n, err := conn.Read(data)
	if err != nil {
		fmt.Println("Error reading data from connection: ", err.Error())
		return
	}

	fmt.Println(string(data))
	req := UnmarshalRequest(data[:n])
	fmt.Println(req)

	path := req.GetPath()
	res := NewResponse()
	res.SetHeader("Content-Type", "text/plain")
	if path == "/" {
		res.StatusCode = http.StatusOK
	} else if strings.HasPrefix(path, "/echo") {
		res.StatusCode = http.StatusOK
		res.Body = path[6:]
	} else if strings.HasPrefix(path, "/user-agent") {
		res.StatusCode = http.StatusOK
		res.Body = req.GetHeader("User-Agent")
	} else if strings.HasPrefix(path, "/files") && *directory != "" {
		targetFile := *directory + path[6:]
		if req.GetMethod() == "GET" {
			res.StatusCode = http.StatusOK
			res.SetHeader("Content-Type", "application/octet-stream")
			fileContent, err := os.ReadFile(targetFile)
			if err != nil {
				fmt.Println("Error reading data from file: ", err.Error())
				res.StatusCode = http.StatusNotFound
			}
			res.Body = string(fileContent)
		} else if req.GetMethod() == "POST" {
			res.StatusCode = http.StatusCreated
			f, err := os.Create(targetFile)
			if err != nil {
				fmt.Println("Error create file: ", err.Error())
				res.StatusCode = http.StatusNotModified
			}
			_, err = f.WriteString(req.GetBody())
			if err != nil {
				fmt.Println("Error writing data to file: ", err.Error())
				res.StatusCode = http.StatusNotModified
			}
		}
	} else {
		res.StatusCode = http.StatusNotFound
	}
	respond(conn, res)
}

func respond(conn net.Conn, res Response) {
	res.SetHeader("Content-Length", strconv.Itoa(len(res.Body)))
	_, err := conn.Write(res.MarshalResponse())
	if err != nil {
		fmt.Println("Error writing data to connection: ", err.Error())
		return
	}
}
