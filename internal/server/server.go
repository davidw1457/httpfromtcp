package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/davidw1457/httpfromtcp/internal/response"
)

type Server struct {
	closed   atomic.Bool
	listener net.Listener
}

func Serve(port int) (*Server, error) {
	portString := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", portString)
	if err != nil {
		return nil, fmt.Errorf("serve: %w", err)
	}

	server := &Server{
		closed:   atomic.Bool{},
		listener: listener,
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

	header := response.GetDefaultHeaders(0)

	err := response.WriteStatusLine(conn, response.OK)
	if err != nil {
		log.Printf("server.handle: %s", err)
		return
	}

	err = response.WriteHeaders(conn, header)
	if err != nil {
		log.Printf("server.handle: %s", err)
	}
}
