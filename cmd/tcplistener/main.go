package main

import (
	"fmt"
	"log"
	"net"

	"github.com/davidw1457/httpfromtcp/internal/request"
)

func main() {
	l, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalln(err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println("Connection accepted")

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatalln(err)
		}

		printRequest(req)

		fmt.Println("Connection closed")
	}
}

func printRequest(req *request.Request) {
	fmt.Println("Request line:")
	fmt.Printf("- Method: %s\n", req.RequestLine.Method)
	fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
	fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)
	fmt.Println("Headers:")
	for k, v := range req.Headers {
		fmt.Printf("- %s: %s\n", k, v)
	}
	fmt.Println("Body:")
	fmt.Println(string(req.Body))
}
