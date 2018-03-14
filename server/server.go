package server

import (
	"log"
	"net"
	"os"
	"path/filepath"
	"proj2_f5w9a_h6v9a_q7w9a_r8u8_w1c0b/serverpb"
	"sync"

	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

var ErrUnimplemented = errors.New("unimplemented")

// Server is the main server struct.
type Server struct {
	log        *log.Logger
	grpcServer *grpc.Server
	config     serverpb.NodeConfig
	db         *badger.DB
	l          net.Listener

	mu struct {
		sync.Mutex

		peerMeta map[string]serverpb.NodeMeta
		peers    map[string]serverpb.NodeClient
	}
}

// New returns a new server.
func New(c serverpb.NodeConfig) (*Server, error) {
	s := &Server{
		log:        log.New(os.Stderr, "", log.Flags()|log.Lshortfile),
		grpcServer: grpc.NewServer(),
		config:     c,
	}
	s.mu.peerMeta = map[string]serverpb.NodeMeta{}
	s.mu.peers = map[string]serverpb.NodeClient{}

	if len(c.Path) == 0 {
		return nil, errors.Errorf("config: path must not be empty")
	}
	if err := os.MkdirAll(c.Path, 0700); err != nil {
		return nil, err
	}

	badgerDir := filepath.Join(c.Path, "badger")
	opts := badger.DefaultOptions
	opts.Dir = badgerDir
	opts.ValueDir = badgerDir
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	s.db = db

	serverpb.RegisterNodeServer(s.grpcServer, s)
	serverpb.RegisterClientServer(s.grpcServer, s)

	return s, nil
}

func (s *Server) Close() error {
	if err := s.db.Close(); err != nil {
		return err
	}
	s.grpcServer.GracefulStop()
	if err := s.l.Close(); err != nil {
		return err
	}
	return nil
}

// Listen causes the server to listen on the specified IP and port.
func (s *Server) Listen(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.l = l
	s.log.Printf("Listening to %s", l.Addr().String())
	return s.grpcServer.Serve(l)
}
