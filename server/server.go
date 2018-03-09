package server

import (
	"log"
	"os"

	"github.com/pkg/errors"
)

var ErrUnimplemented = errors.New("unimplemented")

// Server is the main server struct.
type Server struct {
	log *log.Logger
}

// New returns a new server.
func New() (*Server, error) {
	s := &Server{
		log: log.New(os.Stderr, "", log.Flags()|log.Lshortfile),
	}

	return s, nil
}

// Listen causes the server to listen on the specified IP and port.
func (s *Server) Listen(addr string) error {
	return ErrUnimplemented
}
