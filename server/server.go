package server

import (
	"context"
	"log"
	"net"
	"os"
	"proj2_f5w9a_h6v9a_q7w9a_r8u8_w1c0b/serverpb"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

var ErrUnimplemented = errors.New("unimplemented")

// Server is the main server struct.
type Server struct {
	log        *log.Logger
	grpcServer *grpc.Server
}

// New returns a new server.
func New() (*Server, error) {
	s := &Server{
		log:        log.New(os.Stderr, "", log.Flags()|log.Lshortfile),
		grpcServer: grpc.NewServer(),
	}

	serverpb.RegisterNodeServer(s.grpcServer, s)

	return s, nil
}

func (s *Server) Hello(ctx context.Context, req *serverpb.HelloRequest) (*serverpb.HelloResponse, error) {
	return nil, ErrUnimplemented
}

// Listen causes the server to listen on the specified IP and port.
func (s *Server) Listen(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.log.Printf("Listening to %s", l.Addr().String())
	return s.grpcServer.Serve(l)
}
