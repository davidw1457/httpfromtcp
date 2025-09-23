package server

import (
	"fmt"
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

type Handler func(w *response.Writer, req *request.Request)

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
	w := response.NewWriter(conn)
	request, err := request.RequestFromReader(conn)
	if err != nil {
		log.Printf("server.handle: %s\n", err)
		return
	}

	s.handler(w, request)
}
