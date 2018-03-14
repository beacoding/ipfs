package server

import (
	"context"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"proj2_f5w9a_h6v9a_q7w9a_r8u8_w1c0b/serverpb"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	dialTimeout = 2 * time.Second
)

func (s *Server) Hello(ctx context.Context, req *serverpb.HelloRequest) (*serverpb.HelloResponse, error) {
	meta, err := s.NodeMeta()
	if err != nil {
		return nil, err
	}

	resp := serverpb.HelloResponse{
		Meta: &meta,
	}

	if err := s.AddNode(*req.Meta); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for id := range s.mu.peers {
		meta := s.mu.peerMeta[id]
		resp.ConnectedPeers = append(resp.ConnectedPeers, &meta)
	}
	for id, meta := range s.mu.peerMeta {
		if _, ok := s.mu.peers[id]; ok {
			continue
		}
		meta := meta
		resp.KnownPeers = append(resp.KnownPeers, &meta)
	}

	return &resp, nil
}

// addNodeMeta adds a node meta object to the server and returns whether or not
// that node has been seen before.
func (s *Server) addNodeMeta(meta serverpb.NodeMeta) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	old, ok := s.mu.peerMeta[meta.Id]
	if !ok || old.Updated < meta.Updated {
		s.mu.peerMeta[meta.Id] = meta
	}
	return !ok
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

func (s *Server) connectNode(ctx context.Context, meta serverpb.NodeMeta) (serverpb.NodeClient, error) {

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(meta.Cert))
	if !ok {
		return nil, errors.Errorf("failed to parse certificate for node %+v", meta)
	}

	creds := credentials.NewClientTLSFromCert(roots, "")
	var err error
	var conn *grpc.ClientConn
	for _, addr := range meta.Addrs {
		ctx, _ := context.WithTimeout(ctx, dialTimeout)
		conn, err = grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(creds), grpc.WithBlock())
		if err != nil {
			s.log.Printf("error dialing %+v: %+v", addr, err)
		}
	}
	if err != nil {
		return nil, errors.Wrapf(err, "dialing %s", meta.Id)
	}
	return serverpb.NewNodeClient(conn), nil
}

// getOutboundIP sets up a UDP connection (but doesn't send anything) and uses
// the local IP addressed assigned.
func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	return conn.LocalAddr().(*net.UDPAddr).IP
}

func (s *Server) NumConnections() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.mu.peers)
}

// AddNode adds a node to the server.
func (s *Server) AddNode(meta serverpb.NodeMeta) error {
	localMeta, err := s.NodeMeta()
	if err != nil {
		return err
	}

	if localMeta.Id == meta.Id {
		return nil
	}

	if err := validateNodeMeta(meta); err != nil {
		return err
	}

	s.log.Printf("AddNode %s", color.RedString(meta.Id))

	new := s.addNodeMeta(meta)
	if err := s.persistNodeMeta(meta); err != nil {
		return err
	}
	if !new {
		return nil
	}

	if s.NumConnections() >= int(s.config.MaxPeers) {
		return nil
	}

	ctx := context.TODO()
	conn, err := s.connectNode(ctx, meta)
	if err != nil {
		return err
	}
	resp, err := conn.Hello(ctx, &serverpb.HelloRequest{
		Meta: &localMeta,
	})
	if err != nil {
		return errors.Wrapf(err, "Hello")
	}
	if resp.Meta.Id != meta.Id {
		return errors.Errorf("expected node with ID %+v; got %+v", meta, resp.Meta)
	}

	s.mu.Lock()
	s.mu.peers[meta.Id] = conn
	s.mu.Unlock()

	if err := s.AddNodes(resp.ConnectedPeers, resp.KnownPeers); err != nil {
		return err
	}

	return nil
}

// AddNodes adds a list of connected and known peers. Connected means that one
// of our peers is connected to them and known just means we know they exist.
// The server should prefer to connect to known first since that maximizes
// the cross section bandwidth of the graph.
func (s *Server) AddNodes(connected []*serverpb.NodeMeta, known []*serverpb.NodeMeta) error {
	for _, meta := range known {
		if err := s.AddNode(*meta); err != nil {
			return err
		}
	}
	for _, meta := range connected {
		if err := s.AddNode(*meta); err != nil {
			return err
		}
	}
	return nil
}
