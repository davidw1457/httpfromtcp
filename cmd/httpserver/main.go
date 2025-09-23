package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	//    "time"

	"github.com/davidw1457/httpfromtcp/internal/headers"
	"github.com/davidw1457/httpfromtcp/internal/request"
	"github.com/davidw1457/httpfromtcp/internal/response"
	"github.com/davidw1457/httpfromtcp/internal/server"
)

const port = 42069
const bufferSize = 32

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

func handler(w *response.Writer, req *request.Request) {
	var statusCode response.StatusCode
	var body []byte
	headers := response.GetDefaultHeaders(0)
	target := req.RequestLine.RequestTarget
	var suffix string
	var proxyUrl string
	var err error
	if strings.HasPrefix(target, "/httpbin/") {
		suffix = strings.TrimPrefix(target, "/httpbin/")
		target = "/httpbin/"
		proxyUrl = "https://httpbin.org/"
	}

	switch target {
	case "/yourproblem":
		statusCode = response.BADREQUEST
		body = []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
		headers.Set("Content-Type", "text/html")
	case "/myproblem":
		statusCode = response.SERVERERROR
		body = []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
		headers.Set("Content-Type", "text/html")
	case "/httpbin/":
		headers.Delete("Content-Length")
		headers.Set("Transfer-Encoding", "chunked")
		handleProxy(w, headers, fmt.Sprintf("%s%s", proxyUrl, suffix))
		return
	case "/video":
		statusCode = response.OK
		headers.Set("Content-Type", "video/mp4")
		body, err = getVideo()
		if err != nil {
			log.Printf("handler: %s\n", err)
			return
		}
	default:
		statusCode = response.OK
		body = []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
		headers.Set("Content-Type", "text/html")
	}

	headers.Set("Content-Length", strconv.Itoa(len(body)))

	err = w.WriteStatusLine(statusCode)
	if err != nil {
		log.Printf("handler: %s\n", err)
		return
	}

	err = w.WriteHeaders(headers)
	if err != nil {
		log.Printf("handler: %s\n", err)
		return
	}

	_, err = w.WriteBody(body)
	if err != nil {
		log.Printf("handler: %s\n", err)
	}
}

func handleProxy(w *response.Writer, h headers.Headers, url string) {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("handleProxy: %s\n", err)
		return
	}

	//    for resp.StatusCode == 503 {
	//        log.Printf("error retrieving %s: %s", url, resp.Status)
	//        log.Println("retrying in 10 seconds...")
	//        time.Sleep(10 * time.Second)
	//        resp, err = http.Get(url)
	//	    if err != nil {
	//	    	log.Printf("handleProxy: %s\n", err)
	//	    	return
	//	    }
	//    }

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Printf("error retrieving %s: %s", url, resp.Status)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "" {
		h.Set("Content-Type", contentType)
	}

	h.Set("Trailer", "X-Content-SHA256, X-Content-Length")

	err = w.WriteStatusLine(response.OK)
	if err != nil {
		log.Printf("handleProxy: %s\n", err)
		return
	}

	err = w.WriteHeaders(h)
	if err != nil {
		log.Printf("handleProxy: %s\n", err)
		return
	}

	p := make([]byte, bufferSize)
	fullBody := []byte{}
	bodySize := 0

	n, err := resp.Body.Read(p)
	for ; n > 0 && (err == nil || errors.Is(err, io.EOF)); n, err = resp.Body.Read(p) {
		fullBody = append(fullBody, p[:n]...)
		bodySize += n

		_, err = w.WriteChunkedBody(p[:n])
		if err != nil {
			log.Printf("handleProxy: %s\n", err)
			return
		}
	}

	hash := sha256.Sum256(fullBody)
	h.Add("X-Content-SHA256", fmt.Sprintf("%x", hash))
	h.Add("X-Content-Length", strconv.Itoa(bodySize))

	if errors.Is(err, io.EOF) {
		_, err = w.WriteChunkedBodyDone()
		if err != nil {
			log.Printf("handleProxy: %s\n", err)
			return
		}

		err = w.WriteTrailers(h)
		if err != nil {
			log.Printf("handleProxy: %s\n", err)
		}
	} else {
		log.Printf("handleProxy: %s\n", err)
	}
}

func getVideo() ([]byte, error) {
	video, err := os.ReadFile("assets/vim.mp4")
	if err != nil {
		return nil, fmt.Errorf("getVideo: %w", err)
	}

	return video, nil
}
