package server

import (
	"context"
	"fmt"
	"proj2_f5w9a_h6v9a_q7w9a_r8u8_w1c0b/serverpb"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

const (
	dialTimeout = 2 * time.Second
)

func (s *Server) Hello(ctx context.Context, req *serverpb.HelloRequest) (*serverpb.HelloResponse, error) {
	return nil, ErrUnimplemented
}

// addNodeMeta adds a node meta object to the server and returns whether or not
func (s *Server) addNodeMeta(meta serverpb.NodeMeta) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.mu.peerMeta[meta.Id]; ok {
		return false
	}
	s.mu.peerMeta[meta.Id] = meta
	return true
}

func (s *Server) persistNodeMeta(meta serverpb.NodeMeta) error {
	body, err := meta.Marshal()
	if err != nil {
		return err
	}
	key := fmt.Sprintf("/NodeMeta/%s", meta.Id)
	if err := s.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), body)
	}); err != nil {
		return err
	}
	return nil
}

func validateNodeMeta(meta serverpb.NodeMeta) error {
	if meta.Id == "" {
		return errors.New("NodeMeta missing ID")
	}
	if len(meta.Addrs) == 0 {
		return errors.Errorf("%#v missing addresses", meta)
	}
	return nil
}

func (s *Server) connectNode(ctx context.Context, meta serverpb.NodeMeta) (serverpb.NodeClient, error) {
	var err error
	var conn *grpc.ClientConn
	for _, addr := range meta.Addrs {
		ctx, _ := context.WithTimeout(ctx, dialTimeout)
		conn, err = grpc.DialContext(ctx, addr)
	}
	if err != nil {
		return nil, err
	}
	return serverpb.NewNodeClient(conn), nil
}

func (s *Server) NodeMeta() serverpb.NodeMeta {
	s.mu.Lock()
	defer s.mu.Unlock()

	// TODO
	return serverpb.NodeMeta{}
}

// AddNode adds a node to the server.
func (s *Server) AddNode(meta serverpb.NodeMeta) error {
	if err := validateNodeMeta(meta); err != nil {
		return err
	}
	if !s.addNodeMeta(meta) {
		return nil
	}
	if err := s.persistNodeMeta(meta); err != nil {
		return err
	}

	ctx := context.TODO()
	conn, err := s.connectNode(ctx, meta)
	if err != nil {
		return err
	}
	localMeta := s.NodeMeta()
	resp, err := conn.Hello(ctx, &serverpb.HelloRequest{
		Meta: &localMeta,
	})
	if err != nil {
		return err
	}
	if resp.Meta.Id != meta.Id {
		return errors.Errorf("expected node with ID %+v; got %+v", meta, resp.Meta)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.mu.peers[meta.Id] = conn

	return nil
}
