package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"
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
			log.Printf("s.listen: %s\n", err)
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	response := []byte("HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain\r\n" +
//		"Content-Length: 13\r\n" +
		"\r\n" +
		"Hello World!")
	_, err := conn.Write(response)
	if err != nil {
		log.Printf("s.handle: %s", err)
	}

	err = conn.Close()
	if err != nil {
		log.Printf("s.handle: %s", err)
	}
}
