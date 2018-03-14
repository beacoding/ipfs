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
	localMeta, err := s.NodeMeta()
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
	defer s.mu.Unlock()

	s.mu.peers[meta.Id] = conn

	return nil
}
