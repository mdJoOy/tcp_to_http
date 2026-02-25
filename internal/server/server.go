package server

import (
	"fmt"
	"io"
	"net"

	"github.com/mdjOoy/tcptohttp/internal/response"
)

type Server struct {
	closed bool
}

func runConnection(_s *Server, conn io.ReadWriteCloser) {
	defer conn.Close()
	err := response.WriteStatusLine(conn, response.StatusOK)
	if err != nil {
		return
	}
	headers := response.GetDefaultHeaders(0)
	err = response.WriteHeaders(conn, headers)
	if err != nil {
		return
	}

}

func runServer(s *Server, listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if s.closed {
			return
		}
		if err != nil {
			return
		}
		go runConnection(s, conn)
	}
}
func Serve(port uint16) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	server := &Server{closed: false}
	go runServer(server, listener)

	return server, nil
}
func (s *Server) Close() error {
	s.closed = true
	return nil
}
