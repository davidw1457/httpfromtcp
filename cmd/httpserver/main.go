package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/davidw1457/httpfromtcp/internal/request"
	"github.com/davidw1457/httpfromtcp/internal/response"
	"github.com/davidw1457/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w io.Writer, req *request.Request) *server.HandlerError {
	target := req.RequestLine.RequestTarget
	if target == "/yourproblem" {
		return &server.HandlerError{
			StatusCode: response.BADREQUEST,
			Message:    "Your problem is not my problem\n",
		}
	} else if target == "/myproblem" {
		return &server.HandlerError{
			StatusCode: response.SERVERERROR,
			Message:    "Woopsie, my bad\n",
		}
	}

	_, err := w.Write([]byte("All good, frfr\n"))
	if err != nil {
		return &server.HandlerError{
			StatusCode: response.SERVERERROR,
			Message:    err.Error(),
		}
	}

	return nil
}
