package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"github.com/davidw1457/httpfromtcp/internal/request"
	"github.com/davidw1457/httpfromtcp/internal/response"
)

type Server struct {
	closed   atomic.Bool
	listener net.Listener
	handler  Handler
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func Serve(port int, handler Handler) (*Server, error) {
	portString := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", portString)
	if err != nil {
		return nil, fmt.Errorf("serve: %w", err)
	}

	server := &Server{
		closed:   atomic.Bool{},
		listener: listener,
		handler:  handler,
	}
	go server.listen()

	return server, nil
}

func (s *Server) Close() error {
	if s.closed.Load() {
		return fmt.Errorf("server.close: server already closed")
	}

	s.closed.Store(true)

	err := s.listener.Close()
	if err != nil {
		return fmt.Errorf("server.Close: %w", err)
	}

	return nil
}

func (s *Server) listen() {
	for !s.closed.Load() {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("server.listen: %s\n", err)
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	request, err := request.RequestFromReader(conn)
	if err != nil {
		log.Printf("server.handle: %s\n", err)
		return
	}

	buf := bytes.NewBuffer([]byte{})
	handlerError := s.handler(buf, request)
	if handlerError != nil {
		s.writeError(conn, handlerError)
		return
	}

	header := response.GetDefaultHeaders(buf.Len())

	err = response.WriteStatusLine(conn, response.OK)
	if err != nil {
		log.Printf("server.handle: %s\n", err)
		return
	}

	err = response.WriteHeaders(conn, header)
	if err != nil {
		log.Printf("server.handle: %s\n", err)
	}

	_, err = conn.Write(buf.Bytes())
	if err != nil {
		log.Printf("server.handle: %s\n", err)
	}
}

func (s *Server) writeError(w io.Writer, handlerError *HandlerError) {
	message := []byte(handlerError.Message)
	header := response.GetDefaultHeaders(len(message))

	err := response.WriteStatusLine(w, handlerError.StatusCode)
	if err != nil {
		log.Printf("server.writeError: %s\n", err)
	}

	err = response.WriteHeaders(w, header)
	if err != nil {
		log.Printf("server.writeError: %s\n", err)
	}

	_, err = w.Write(message)
	if err != nil {
		log.Printf("server.writeError: %s\n", err)
	}
}
